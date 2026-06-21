// Package server implements the local web UI served by `tagbackup serve`. It
// exposes a small JSON API over the existing config and object-store layers
// and serves an embedded single-page application.
package server

import (
	"context"
	"embed"
	"encoding/json"
	"fmt"
	"io/fs"
	"log/slog"
	"net"
	"net/http"
	"path"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/joncombe/tagbackup/internal/config"
	"github.com/joncombe/tagbackup/internal/objectkey"
	"github.com/joncombe/tagbackup/internal/store"
)

//go:generate sh -c "npm --prefix ../../web install && npm --prefix ../../web run build"

//go:embed dist
var distFS embed.FS

// Options configures a Server.
type Options struct {
	// ConfigPath is the resolved tagbackup config path ("" means default).
	ConfigPath string
	// Debug, when non-nil, enables DEBUG logging in the store layer.
	Debug *slog.Logger
	// Version is the running binary version ("" means "dev").
	Version string
}

// Server serves the web UI and its JSON API.
type Server struct {
	opts Options
}

type deleteObjectRequest struct {
	Key string `json:"key"`
}

// New returns a Server with the given options.
func New(opts Options) *Server {
	return &Server{opts: opts}
}

// objectDTO is the JSON shape returned for each listed object.
type objectDTO struct {
	Key       string   `json:"key"`
	Filename  string   `json:"filename"`
	Tags      []string `json:"tags"`
	Size      int64    `json:"size"`
	Timestamp int64    `json:"timestamp"` // epoch milliseconds
}

// Listen binds a TCP listener on host:port. It is kept separate from Serve so
// the caller can report the resolved address (and surface bind errors such as
// "address already in use") before the server starts accepting requests.
func (s *Server) Listen(host string, port int) (net.Listener, error) {
	addr := net.JoinHostPort(host, strconv.Itoa(port))
	return net.Listen("tcp", addr)
}

// Serve handles requests on ln until ctx is cancelled, then shuts down
// gracefully. It returns nil on a clean shutdown.
func (s *Server) Serve(ctx context.Context, ln net.Listener) error {
	httpSrv := &http.Server{
		Handler:           s.routes(),
		ReadHeaderTimeout: 10 * time.Second,
	}
	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = httpSrv.Shutdown(shutdownCtx)
	}()
	if err := httpSrv.Serve(ln); err != nil && err != http.ErrServerClosed {
		return err
	}
	return nil
}

func (s *Server) routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/buckets", s.handleBuckets)
	mux.HandleFunc("GET /api/buckets/{alias}", s.handleBucketConfig)
	mux.HandleFunc("GET /api/buckets/{alias}/objects", s.handleObjects)
	mux.HandleFunc("DELETE /api/buckets/{alias}/objects", s.handleDeleteObject)
	mux.HandleFunc("POST /api/buckets/{alias}/objects", s.handleUploadObject)
	mux.HandleFunc("GET /api/version", s.handleVersion)
	mux.Handle("/", s.staticHandler())
	return mux
}

// handleBuckets returns the configured bucket aliases sorted alphabetically.
// A missing config yields an empty list (so the SPA can show "no buckets").
func (s *Server) handleBuckets(w http.ResponseWriter, r *http.Request) {
	cfg, _, err := config.LoadOrEmpty(s.opts.ConfigPath)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "failed to load config")
		return
	}
	aliases := make([]string, 0, len(cfg.Buckets))
	for a := range cfg.Buckets {
		aliases = append(aliases, a)
	}
	sort.Strings(aliases)
	writeJSON(w, http.StatusOK, aliases)
}

// handleVersion returns the running binary version.
func (s *Server) handleVersion(w http.ResponseWriter, r *http.Request) {
	v := s.opts.Version
	if v == "" {
		v = "dev"
	}
	writeJSON(w, http.StatusOK, map[string]string{"version": v})
}

// handleBucketConfig returns sanitized configuration for the given bucket alias.
func (s *Server) handleBucketConfig(w http.ResponseWriter, r *http.Request) {
	alias := r.PathValue("alias")
	if err := config.ValidateAlias(alias); err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid bucket alias")
		return
	}
	cfg, _, err := config.LoadOrEmpty(s.opts.ConfigPath)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "failed to load config")
		return
	}
	bkt, err := cfg.GetBucket(alias)
	if err != nil {
		writeJSONError(w, http.StatusNotFound, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, bucketConfigFrom(alias, bkt))
}

// handleObjects lists all tagbackup objects for the given bucket alias.
func (s *Server) handleObjects(w http.ResponseWriter, r *http.Request) {
	_, st, err := s.objectStoreForAlias(w, r)
	if err != nil {
		return
	}
	objs, err := st.ListObjectsAll(r.Context(), func(map[string]struct{}) bool { return true })
	if err != nil {
		writeJSONError(w, http.StatusBadGateway, err.Error())
		return
	}
	out := make([]objectDTO, 0, len(objs))
	for _, o := range objs {
		out = append(out, objectDTO{
			Key:       o.Key,
			Filename:  o.Parsed.DisplayName,
			Tags:      o.Parsed.Tags,
			Size:      o.Size,
			Timestamp: o.Parsed.Timestamp,
		})
	}
	writeJSON(w, http.StatusOK, out)
}

func (s *Server) objectStoreForAlias(w http.ResponseWriter, r *http.Request) (string, store.ObjectStore, error) {
	alias := r.PathValue("alias")
	if err := config.ValidateAlias(alias); err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid bucket alias")
		return "", nil, err
	}
	cfg, _, err := config.LoadOrEmpty(s.opts.ConfigPath)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "failed to load config")
		return "", nil, err
	}
	if _, err := cfg.GetBucket(alias); err != nil {
		writeJSONError(w, http.StatusNotFound, err.Error())
		return "", nil, err
	}
	st, err := store.NewObjectStoreFromAlias(r.Context(), cfg, alias, s.opts.Debug)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, err.Error())
		return "", nil, err
	}
	if s.opts.Debug != nil {
		st.SetDebugLog(s.opts.Debug)
	}
	return alias, st, nil
}

func (s *Server) handleDeleteObject(w http.ResponseWriter, r *http.Request) {
	var req deleteObjectRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.Key == "" {
		writeJSONError(w, http.StatusBadRequest, "key is required")
		return
	}
	_, st, err := s.objectStoreForAlias(w, r)
	if err != nil {
		return
	}
	if err := st.DeleteObject(r.Context(), req.Key); err != nil {
		writeJSONError(w, http.StatusBadGateway, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) handleUploadObject(w http.ResponseWriter, r *http.Request) {
	alias, st, err := s.objectStoreForAlias(w, r)
	if err != nil {
		return
	}
	if err := r.ParseMultipartForm(32 << 20); err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid multipart form")
		return
	}
	file, header, err := r.FormFile("file")
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "file is required")
		return
	}
	defer file.Close()

	tagsJSON := r.FormValue("tags")
	var tags []string
	if err := json.Unmarshal([]byte(tagsJSON), &tags); err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid tags")
		return
	}
	if len(tags) == 0 {
		writeJSONError(w, http.StatusBadRequest, "at least one tag is required")
		return
	}
	for _, t := range tags {
		if err := objectkey.ValidateTag(t); err != nil {
			writeJSONError(w, http.StatusBadRequest, err.Error())
			return
		}
	}

	filename := filepath.Base(header.Filename)
	if filename == "" || filename == "." {
		writeJSONError(w, http.StatusBadRequest, "invalid filename")
		return
	}

	key, err := st.BuildPushKey(filename, tags)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, err.Error())
		return
	}
	if err := st.Upload(r.Context(), key, file, header.Size); err != nil {
		writeJSONError(w, http.StatusBadGateway, err.Error())
		return
	}

	cfg, _, err := config.LoadOrEmpty(s.opts.ConfigPath)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "failed to load config")
		return
	}
	bkt, err := cfg.GetBucket(alias)
	if err != nil {
		writeJSONError(w, http.StatusNotFound, err.Error())
		return
	}
	dto, err := objectDTOFromKey(key, bkt.Prefix, header.Size)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, dto)
}

func objectDTOFromKey(key, prefix string, size int64) (objectDTO, error) {
	np := prefix
	if np != "" && !strings.HasSuffix(np, "/") {
		np += "/"
	}
	rel := key
	if np != "" {
		if !strings.HasPrefix(key, np) {
			return objectDTO{}, fmt.Errorf("key %q does not match bucket prefix", key)
		}
		rel = strings.TrimPrefix(key, np)
	}
	parsed, err := objectkey.Parse(rel)
	if err != nil {
		return objectDTO{}, err
	}
	return objectDTO{
		Key:       key,
		Filename:  parsed.DisplayName,
		Tags:      parsed.Tags,
		Size:      size,
		Timestamp: parsed.Timestamp,
	}, nil
}

// staticHandler serves the embedded SPA, falling back to index.html for any
// path that is not a real asset (so client-side routing works).
func (s *Server) staticHandler() http.Handler {
	sub, err := fs.Sub(distFS, "dist")
	if err != nil {
		// Should never happen: the dist directory is embedded at build time.
		return http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			writeJSONError(w, http.StatusInternalServerError, "frontend assets unavailable")
		})
	}
	fileServer := http.FileServer(http.FS(sub))
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		clean := strings.TrimPrefix(path.Clean(r.URL.Path), "/")
		if clean == "" {
			clean = "index.html"
		}
		if f, openErr := sub.Open(clean); openErr == nil {
			_ = f.Close()
			fileServer.ServeHTTP(w, r)
			return
		}
		data, readErr := fs.ReadFile(sub, "index.html")
		if readErr != nil {
			http.Error(w, "frontend not built", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write(data)
	})
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeJSONError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}
