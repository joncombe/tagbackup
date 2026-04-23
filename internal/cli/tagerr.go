package cli

import "fmt"

// tagErr returns "tagbackup: <command>: <err>" for user-facing error lines (spec: Handling errors).
func tagErr(cmd string, err error) string {
	return fmt.Sprintf("tagbackup: %s: %v", cmd, err)
}

// tagf returns "tagbackup: <command>: " + formatted message.
func tagf(cmd, format string, args ...any) string {
	return "tagbackup: " + cmd + ": " + fmt.Sprintf(format, args...)
}
