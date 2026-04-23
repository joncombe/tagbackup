// Package exitc defines process exit status codes for tagbackup.
package exitc

const (
	OK        = 0
	Err       = 1
	Usage     = 2
	Config    = 3
	S3        = 4
	NoMatches = 5
	SIGINT    = 130
	SIGTERM   = 143
)
