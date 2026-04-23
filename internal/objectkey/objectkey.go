// Package objectkey parses and builds tagbackup object key names.
package objectkey

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
)

// Parsed is a key after the optional bucket prefix has been removed.
type Parsed struct {
	Timestamp   int64  // epoch milliseconds
	RawTags     string // comma-separated, as in the key
	Tags        []string
	DisplayName string // original filename segment
}

const (
	MaxOriginalNameBytes = 255
	MaxKeyBytes          = 1024
)

// IsTagChar reports whether b is a valid tag character ([a-zA-Z0-9]).
func IsTagChar(b byte) bool {
	return b >= 'a' && b <= 'z' || b >= 'A' && b <= 'Z' || b >= '0' && b <= '9'
}

// ValidateTag returns an error if t is not a non-empty run of tag characters.
func ValidateTag(t string) error {
	if t == "" {
		return fmt.Errorf("empty tag")
	}
	for i := 0; i < len(t); i++ {
		if !IsTagChar(t[i]) {
			return fmt.Errorf("invalid tag %q (allowed: [a-zA-Z0-9])", t)
		}
	}
	return nil
}

// Parse parses "timestamp-tags-filename" (no prefix). Filename may contain hyphens.
func Parse(keySuffix string) (*Parsed, error) {
	// keySuffix is the part of the key after the configured prefix, without leading/trailing slashes.
	parts := strings.SplitN(keySuffix, "-", 3)
	if len(parts) < 3 {
		return nil, fmt.Errorf("not a tagbackup object key")
	}
	tsStr, tagSeg, name := parts[0], parts[1], parts[2]
	if len(tsStr) != 13 {
		return nil, fmt.Errorf("not a tagbackup object key")
	}
	ts, err := strconv.ParseInt(tsStr, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("not a tagbackup object key")
	}
	if tagSeg == "" {
		return nil, fmt.Errorf("not a tagbackup object key")
	}
	for _, t := range strings.Split(tagSeg, ",") {
		if ValidateTag(t) != nil {
			return nil, fmt.Errorf("not a tagbackup object key")
		}
	}
	tags := strings.Split(tagSeg, ",")
	return &Parsed{
		Timestamp:   ts,
		RawTags:     tagSeg,
		Tags:        tags,
		DisplayName: name,
	}, nil
}

// BuildKey assembles a full S3 key (with optional prefix) from components.
// Tag strings are sorted and comma-joined, per spec.
func BuildKey(prefix, originalName string, tagList []string, tsMs int64) (string, error) {
	if len([]byte(originalName)) > MaxOriginalNameBytes {
		return "", fmt.Errorf("original filename exceeds %d bytes", MaxOriginalNameBytes)
	}
	if len(tagList) == 0 {
		return "", fmt.Errorf("at least one tag is required")
	}
	tags := append([]string(nil), tagList...)
	for _, t := range tags {
		if err := ValidateTag(t); err != nil {
			return "", err
		}
	}
	sort.Strings(tags)
	tagSeg := strings.Join(tags, ",")
	body := fmt.Sprintf("%d-%s-%s", tsMs, tagSeg, originalName)
	if prefix == "" {
		if len([]byte(body)) > MaxKeyBytes {
			return "", fmt.Errorf("assembled S3 key exceeds %d bytes", MaxKeyBytes)
		}
		return body, nil
	}
	p := prefix
	if !strings.HasSuffix(p, "/") {
		p += "/"
	}
	full := p + body
	if len([]byte(full)) > MaxKeyBytes {
		return "", fmt.Errorf("assembled S3 key exceeds %d bytes", MaxKeyBytes)
	}
	return full, nil
}

// TagSet returns p.Tags as a set for tagexpr.
func (p *Parsed) TagSet() map[string]struct{} {
	m := make(map[string]struct{}, len(p.Tags))
	for _, t := range p.Tags {
		m[t] = struct{}{}
	}
	return m
}
