package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"time"

	"github.com/joncombe/tagbackup/internal/config"
	"github.com/joncombe/tagbackup/internal/store"
	"github.com/spf13/cobra"
)

type filesJSON struct {
	Key       string   `json:"key"`
	Tags      []string `json:"tags"`
	Size      int64    `json:"size"`
	Timestamp int64    `json:"timestamp"`
}

func (g *Runtime) cmdFiles() *cobra.Command {
	var bucket, tagExpr string
	var asJSON bool
	cmd := &cobra.Command{
		Use:   "files",
		Short: "List files in the bucket matching a tag expression",
		RunE: func(cmd *cobra.Command, args []string) error {
			return g.runFiles(bucket, tagExpr, asJSON)
		},
	}
	cmd.Flags().StringVar(&bucket, "bucket", "", "bucket alias (required)")
	cmd.Flags().StringVar(&tagExpr, "tag", "", "tag expression (required)")
	cmd.Flags().BoolVar(&asJSON, "json", false, "output one JSON object per line")
	_ = cmd.MarkFlagRequired("bucket")
	_ = cmd.MarkFlagRequired("tag")
	return cmd
}

func (g *Runtime) runFiles(bucket, tagExpr string, asJSON bool) error {
	const name = "files"
	ev, err := store.ParseTagExpr(tagExpr)
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
	objs, err := st.ListObjectsAll(g.Ctx, ev)
	if err != nil {
		return exitS3(name, err)
	}
	if len(objs) == 0 {
		return exitNoMatches(name, "no files found")
	}
	sort.Slice(objs, func(i, j int) bool { return objs[i].Parsed.Timestamp > objs[j].Parsed.Timestamp })

	if asJSON {
		enc := json.NewEncoder(os.Stdout)
		for _, o := range objs {
			line := filesJSON{
				Key:       o.Key,
				Tags:      o.Parsed.Tags,
				Size:      o.Size,
				Timestamp: o.Parsed.Timestamp,
			}
			if err := enc.Encode(line); err != nil {
				return exitErr(name, err)
			}
		}
		return nil
	}

	for _, o := range objs {
		ts := time.UnixMilli(o.Parsed.Timestamp).UTC().Format("2006-01-02 15:04:05Z")
		fmt.Printf("%s  %10s  %s  [%s]\n", ts, humanBytes(o.Size), o.Parsed.DisplayName, o.Parsed.RawTags)
	}
	return nil
}
