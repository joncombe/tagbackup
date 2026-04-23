package store

import "github.com/joncombe/tagbackup/internal/objectkey"

// Object is a listable tagbackup object.
type Object struct {
	Key    string
	Parsed *objectkey.Parsed
	Size   int64
}
