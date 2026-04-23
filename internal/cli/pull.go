package cli

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/AlecAivazis/survey/v2"
	"github.com/joncombe/tagbackup/internal/config"
	"github.com/joncombe/tagbackup/internal/store"
	"github.com/schollz/progressbar/v3"
	"github.com/spf13/cobra"
)

func (g *Runtime) cmdPull() *cobra.Command {
	var bucket, tagExpr, output string
	var latest bool
	cmd := &cobra.Command{
		Use:   "pull",
		Short: "Download a file from the bucket matching tags",
		RunE: func(cmd *cobra.Command, args []string) error {
			return g.runPull(bucket, tagExpr, latest, output)
		},
	}
	cmd.Flags().StringVar(&bucket, "bucket", "", "bucket alias (required)")
	cmd.Flags().StringVar(&tagExpr, "tag", "", "tag expression (required)")
	cmd.Flags().BoolVar(&latest, "latest", false, "download the newest matching file")
	cmd.Flags().StringVar(&output, "output", "", "output path (- for stdout)")
	_ = cmd.MarkFlagRequired("bucket")
	_ = cmd.MarkFlagRequired("tag")
	return cmd
}

func (g *Runtime) runPull(bucket, tagExpr string, latest bool, output string) error {
	const name = "pull"
	ev, err := store.ParseTagExpr(tagExpr)
	if err != nil {
		return exitUsageErr(name, err)
	}
	if !latest && g.NonInter {
		return exitUsage(name, "--latest is required when --non-interactive is set")
	}
	cfg, _, err := config.Load(g.ConfigPath)
	if err != nil {
		return exitConfig(name, err)
	}
	var st store.ObjectStore
	st, err = store.NewObjectStoreFromAlias(g.Ctx, cfg, bucket, g.debugLogger())
	if err != nil {
		return exitConfig(name, err)
	}
	if g.Verbose {
		st.SetDebugLog(g.Log)
	}
	objs, err := st.ListObjectsAll(g.Ctx, ev)
	if err != nil {
		return exitS3(name, err)
	}
	if len(objs) == 0 {
		return exitNoMatches(name, "no file to download")
	}
	sort.Slice(objs, func(i, j int) bool { return objs[i].Parsed.Timestamp > objs[j].Parsed.Timestamp })

	var chosen *store.Object
	if latest {
		chosen = &objs[0]
	} else {
		if !StdinIsTTY() {
			return exitUsage(name, "interactive selection requires a TTY on stdin")
		}
		opts := make([]string, len(objs))
		for i, o := range objs {
			ts := time.UnixMilli(o.Parsed.Timestamp).UTC().Format("2006-01-02 15:04:05Z")
			opts[i] = fmt.Sprintf("%s  %10s  %s  [%s]", ts, humanBytes(o.Size), o.Parsed.DisplayName, o.Parsed.RawTags)
		}
		var pick int
		if e := survey.AskOne(&survey.Select{
			Message:  "Choose a file",
			Options:  opts,
			PageSize: 20,
		}, &pick); e != nil {
			return exitErr(name, e)
		}
		chosen = &objs[pick]
	}

	rc, n, err := st.GetObjectReader(g.Ctx, chosen.Key)
	if err != nil {
		return exitS3(name, err)
	}
	defer rc.Close()

	dest := output
	if dest == "" {
		dest = chosen.Parsed.DisplayName
	} else if dest != "-" {
		if st, err := os.Stat(dest); err == nil && st.IsDir() {
			dest = filepath.Join(dest, chosen.Parsed.DisplayName)
		} else if strings.HasSuffix(dest, string(os.PathSeparator)) {
			if err := os.MkdirAll(dest, 0o755); err != nil {
				return exitErr(name, err)
			}
			dest = filepath.Join(dest, chosen.Parsed.DisplayName)
		}
	}
	if dest == "-" {
		w := os.Stdout
		var src io.Reader = rc
		if g.showProgressBar() && n > 0 {
			pb := progressbar.DefaultBytes(n, "downloading")
			src = io.TeeReader(rc, pb)
		}
		_, err = io.Copy(w, src)
		if err != nil {
			return exitErr(name, err)
		}
		return nil
	}

	dir := filepath.Dir(dest)
	if dir == "" {
		dir = "."
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return exitErr(name, err)
	}
	f, err := os.CreateTemp(dir, filepath.Base(dest)+".")
	if err != nil {
		return exitErr(name, err)
	}
	tmp := f.Name()
	var src io.Reader = rc
	if g.showProgressBar() && n > 0 {
		pb := progressbar.DefaultBytes(n, "downloading")
		src = io.TeeReader(rc, pb)
	}
	_, copyErr := io.Copy(f, src)
	closeErr := f.Close()
	if copyErr != nil {
		_ = os.Remove(tmp)
		return exitErr(name, copyErr)
	}
	if closeErr != nil {
		_ = os.Remove(tmp)
		return exitErr(name, closeErr)
	}
	if err := os.Rename(tmp, dest); err != nil {
		_ = os.Remove(tmp)
		return exitErr(name, err)
	}
	if g.Log != nil && !g.Quiet {
		g.Log.Info("downloaded", "path", filepath.Clean(dest))
	}
	return nil
}
