package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"time"

	"github.com/AlecAivazis/survey/v2"
	"github.com/joncombe/tagbackup/internal/config"
	"github.com/joncombe/tagbackup/internal/duration"
	"github.com/joncombe/tagbackup/internal/store"
	"github.com/spf13/cobra"
)

func (g *Runtime) cmdDelete() *cobra.Command {
	var bucket, tagExpr string
	var force, dry, asJSON bool
	var newer, older string
	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete files matching a tag expression",
		RunE: func(cmd *cobra.Command, args []string) error {
			return g.runDelete(bucket, tagExpr, force, dry, asJSON, newer, older)
		},
	}
	cmd.Flags().StringVar(&bucket, "bucket", "", "bucket alias (required)")
	cmd.Flags().StringVar(&tagExpr, "tag", "", "tag expression (required)")
	cmd.Flags().BoolVar(&force, "force", false, "do not prompt; delete all matches")
	cmd.Flags().BoolVar(&dry, "dry-run", false, "only show what would be deleted")
	cmd.Flags().BoolVar(&asJSON, "json", false, "one JSON object per line")
	cmd.Flags().StringVar(&newer, "newer-than", "", "only match files newer than duration, e.g. 2d")
	cmd.Flags().StringVar(&older, "older-than", "", "only match files older than duration, e.g. 30d")
	_ = cmd.MarkFlagRequired("bucket")
	_ = cmd.MarkFlagRequired("tag")
	return cmd
}

// filterAge returns objects whose embedded timestamp satisfies the chosen
// relative-duration boundary. The boundary is strictly inequality: an object
// whose timestamp equals the cutoff is not included, matching the spec.
// When newer is true the function keeps objects strictly newer than now-d;
// otherwise it keeps objects strictly older.
func filterAge(objs []store.Object, now time.Time, d time.Duration, newer bool) []store.Object {
	cutoff := now.Add(-d).UnixMilli()
	var out []store.Object
	for _, o := range objs {
		if newer {
			if o.Parsed.Timestamp > cutoff {
				out = append(out, o)
			}
		} else {
			if o.Parsed.Timestamp < cutoff {
				out = append(out, o)
			}
		}
	}
	return out
}

type delJSON struct {
	Key       string   `json:"key"`
	Tags      []string `json:"tags"`
	Size      int64    `json:"size"`
	Timestamp int64    `json:"timestamp"`
}

func (g *Runtime) runDelete(bucket, tagExpr string, force, dry, asJSON bool, newer, older string) error {
	const name = "delete"
	if newer != "" && older != "" {
		return exitUsage(name, "only one of --newer-than and --older-than is allowed")
	}
	if !force && g.NonInter && !dry {
		return exitUsage(name, "--force is required when --non-interactive is set (or use --dry-run)")
	}

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
	if newer != "" {
		d, e := duration.ParseRelativeUTC(newer)
		if e != nil {
			return exitUsageErr(name, e)
		}
		objs = filterAge(objs, time.Now().UTC(), d, true)
	}
	if older != "" {
		d, e := duration.ParseRelativeUTC(older)
		if e != nil {
			return exitUsageErr(name, e)
		}
		objs = filterAge(objs, time.Now().UTC(), d, false)
	}
	if len(objs) == 0 {
		if dry {
			return nil
		}
		return exitNoMatches(name, "no matching files")
	}
	sort.Slice(objs, func(i, j int) bool { return objs[i].Parsed.Timestamp > objs[j].Parsed.Timestamp })

	todo := objs
	if !dry && !force {
		if !StdinIsTTY() {
			return exitUsage(name, "confirmation requires a TTY or use --force")
		}
		opts := make([]string, len(objs))
		for i, o := range objs {
			ts := time.UnixMilli(o.Parsed.Timestamp).UTC().Format("2006-01-02 15:04:05Z")
			opts[i] = fmt.Sprintf("%s  %10s  %s  [%s]", ts, humanBytes(o.Size), o.Parsed.DisplayName, o.Parsed.RawTags)
		}
		var picked []int
		if err := askOneErr(name, &survey.MultiSelect{
			Message:  "Select files to delete",
			Options:  opts,
			PageSize: 20,
		}, &picked); err != nil {
			return err
		}
		if len(picked) == 0 {
			return nil
		}
		todo = make([]store.Object, 0, len(picked))
		for _, i := range picked {
			todo = append(todo, objs[i])
		}
	}

	enc := json.NewEncoder(os.Stdout)
	for _, o := range todo {
		if dry {
			if asJSON {
				_ = enc.Encode(delJSON{Key: o.Key, Tags: o.Parsed.Tags, Size: o.Size, Timestamp: o.Parsed.Timestamp})
			} else {
				_, _ = fmt.Fprintln(os.Stdout, "would delete", o.Key)
			}
			continue
		}
		if e := st.DeleteObject(g.Ctx, o.Key); e != nil {
			return exitS3(name, e)
		}
		if asJSON {
			_ = enc.Encode(delJSON{Key: o.Key, Tags: o.Parsed.Tags, Size: o.Size, Timestamp: o.Parsed.Timestamp})
		} else if !g.Quiet {
			_, _ = fmt.Fprintln(os.Stderr, "deleted", o.Key)
		}
	}
	return nil
}
