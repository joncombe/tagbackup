// Package server implements the local, read-only web UI served by
// `tagbackup serve`. It exposes a small JSON API over the existing config and
// object-store layers and serves an embedded single-page application.
package server

import (
	"context"
	"embed"
	"encoding/json"
	"io/fs"
	"log/slog"
	"net"
	"net/http"
	"path"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/joncombe/tagbackup/internal/config"
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
}

// Server serves the read-only web UI and its JSON API.
type Server struct {
	opts Options
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
	mux.HandleFunc("GET /api/buckets/{alias}/objects", s.handleObjects)
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

// handleObjects lists all tagbackup objects for the given bucket alias.
func (s *Server) handleObjects(w http.ResponseWriter, r *http.Request) {
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
	if _, err := cfg.GetBucket(alias); err != nil {
		writeJSONError(w, http.StatusNotFound, err.Error())
		return
	}
	st, err := store.NewObjectStoreFromAlias(r.Context(), cfg, alias, s.opts.Debug)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if s.opts.Debug != nil {
		st.SetDebugLog(s.opts.Debug)
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
