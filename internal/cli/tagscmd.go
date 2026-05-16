package cli

import (
	"fmt"
	"sort"
	"time"

	"github.com/joncombe/tagbackup/internal/config"
	"github.com/joncombe/tagbackup/internal/store"
	"github.com/spf13/cobra"
)

type tagSummary struct {
	Count  int
	Oldest int64
	Newest int64
}

func (g *Runtime) cmdTags() *cobra.Command {
	var bucket string
	cmd := &cobra.Command{
		Use:   "tags",
		Short: "List all tags in the bucket with file counts and date ranges",
		RunE: func(cmd *cobra.Command, args []string) error {
			return g.runTags(bucket)
		},
	}
	cmd.Flags().StringVar(&bucket, "bucket", "", "bucket alias (required)")
	_ = cmd.MarkFlagRequired("bucket")
	return cmd
}

func (g *Runtime) runTags(bucket string) error {
	const name = "tags"
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
	objs, err := st.ListObjectsAll(g.Ctx, func(_ map[string]struct{}) bool { return true })
	if err != nil {
		return exitS3(name, err)
	}
	if len(objs) == 0 {
		return exitNoMatches(name, "no files found")
	}

	summary := map[string]*tagSummary{}
	for _, o := range objs {
		for _, tag := range o.Parsed.Tags {
			s, ok := summary[tag]
			if !ok {
				s = &tagSummary{Oldest: o.Parsed.Timestamp, Newest: o.Parsed.Timestamp}
				summary[tag] = s
			}
			s.Count++
			if o.Parsed.Timestamp < s.Oldest {
				s.Oldest = o.Parsed.Timestamp
			}
			if o.Parsed.Timestamp > s.Newest {
				s.Newest = o.Parsed.Timestamp
			}
		}
	}

	names := make([]string, 0, len(summary))
	for tag := range summary {
		names = append(names, tag)
	}
	sort.Strings(names)

	maxLen := 3 // minimum width to fit "TAG" header
	for _, tag := range names {
		if len(tag) > maxLen {
			maxLen = len(tag)
		}
	}

	fmt.Printf("%-*s  %5s  %-20s  %-20s\n", maxLen, "TAG", "FILES", "OLDEST", "NEWEST")
	for _, tag := range names {
		s := summary[tag]
		oldest := time.UnixMilli(s.Oldest).UTC().Format("2006-01-02 15:04:05Z")
		newest := time.UnixMilli(s.Newest).UTC().Format("2006-01-02 15:04:05Z")
		fmt.Printf("%-*s  %5d  %s  %s\n", maxLen, tag, s.Count, oldest, newest)
	}
	return nil
}
