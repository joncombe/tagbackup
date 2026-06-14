package cli

import (
	"io"
	"os"
	"path/filepath"

	"github.com/joncombe/tagbackup/internal/config"
	"github.com/joncombe/tagbackup/internal/store"
	"github.com/schollz/progressbar/v3"
	"github.com/spf13/cobra"
)

func (g *Runtime) cmdPush() *cobra.Command {
	var bucket, tagStr, filename string
	cmd := &cobra.Command{
		Use:   "push <path>",
		Short: "Upload a file to the bucket with tags",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return g.runPush(args[0], bucket, tagStr, filename)
		},
	}
	cmd.Flags().StringVar(&bucket, "bucket", "", "bucket alias (required)")
	cmd.Flags().StringVar(&tagStr, "tag", "", "comma-separated tags (required)")
	cmd.Flags().StringVar(&filename, "filename", "", "original filename when path is '-' (required for stdin)")
	_ = cmd.MarkFlagRequired("bucket")
	_ = cmd.MarkFlagRequired("tag")
	return cmd
}

func (g *Runtime) runPush(path, bucket, tagStr, filename string) error {
	const name = "push"
	tags, err := parseTagCSV(tagStr)
	if err != nil {
		return exitUsageErr(name, err)
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

	var r io.Reader
	var orig string
	var size int64 = -1

	if path == "-" {
		if filename == "" {
			return exitUsage(name, "--filename is required when reading from stdin")
		}
		orig = filename
		r = os.Stdin
	} else {
		f, e := os.Open(path)
		if e != nil {
			if os.IsNotExist(e) {
				return exitErrMsg(name, "file not found: %s", path)
			}
			return exitErr(name, e)
		}
		defer f.Close()
		stt, e := f.Stat()
		if e != nil {
			return exitErr(name, e)
		}
		size = stt.Size()
		orig = filepath.Base(path)
		r = f
	}

	key, err := st.BuildPushKey(orig, tags)
	if err != nil {
		return exitUsageErr(name, err)
	}

	body := io.Reader(r)
	if g.showProgressBar() {
		if size >= 0 {
			pb := progressbar.DefaultBytes(size, "uploading")
			body = io.TeeReader(r, pb)
		} else {
			pb := progressbar.NewOptions(-1,
				progressbar.OptionSetDescription("uploading (stdin)"),
				progressbar.OptionSetWriter(os.Stderr),
				progressbar.OptionShowBytes(true),
			)
			body = io.TeeReader(r, pb)
		}
	}

	if e := st.Upload(g.Ctx, key, body, size); e != nil {
		return exitS3(name, e)
	}
	return nil
}
