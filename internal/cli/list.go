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

type listJSON struct {
	Key       string   `json:"key"`
	Tags      []string `json:"tags"`
	Size      int64    `json:"size"`
	Timestamp int64    `json:"timestamp"` // epoch ms from filename
}

func (g *Runtime) cmdList() *cobra.Command {
	var bucket, tagExpr string
	var asJSON bool
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List files in the bucket matching a tag expression",
		RunE: func(cmd *cobra.Command, args []string) error {
			return g.runList(bucket, tagExpr, asJSON)
		},
	}
	cmd.Flags().StringVar(&bucket, "bucket", "", "bucket alias (required)")
	cmd.Flags().StringVar(&tagExpr, "tag", "", "tag expression (required)")
	cmd.Flags().BoolVar(&asJSON, "json", false, "one JSON object per line")
	_ = cmd.MarkFlagRequired("bucket")
	_ = cmd.MarkFlagRequired("tag")
	return cmd
}

func (g *Runtime) runList(bucket, tagExpr string, asJSON bool) error {
	const name = "list"
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
		return exitNoMatches(name, "no matching files")
	}
	sort.Slice(objs, func(i, j int) bool { return objs[i].Parsed.Timestamp > objs[j].Parsed.Timestamp })

	if asJSON {
		enc := json.NewEncoder(os.Stdout)
		for _, o := range objs {
			line := listJSON{
				Key:       o.Key,
				Tags:      o.Parsed.Tags,
				Size:      o.Size,
				Timestamp: o.Parsed.Timestamp,
			}
			if e := enc.Encode(&line); e != nil {
				return exitErr(name, e)
			}
		}
		return nil
	}
	for _, o := range objs {
		t := time.UnixMilli(o.Parsed.Timestamp).UTC()
		_, _ = fmt.Fprintf(os.Stdout, "%s  %10s  %s  %s\n", t.Format(time.RFC3339), humanBytes(o.Size), o.Parsed.DisplayName, o.Parsed.RawTags)
	}
	return nil
}
