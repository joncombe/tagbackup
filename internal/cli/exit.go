package cli

import "github.com/joncombe/tagbackup/internal/exitc"

// Exit is a Go error with a process exit code.
type Exit struct {
	Code  int
	Human string
}

func (e *Exit) Error() string { return e.Human }

// New returns an Exit error. Code must be non-zero.
func New(code int, human string) *Exit { return &Exit{Code: code, Human: human} }

func exitConfig(cmd string, err error) error {
	return &Exit{Code: exitc.Config, Human: tagErr(cmd, err)}
}

func exitConfigMsg(cmd, format string, a ...any) error {
	return &Exit{Code: exitc.Config, Human: tagf(cmd, format, a...)}
}

func exitUsage(cmd, format string, a ...any) error {
	return &Exit{Code: exitc.Usage, Human: tagf(cmd, format, a...)}
}

func exitUsageErr(cmd string, err error) error {
	return &Exit{Code: exitc.Usage, Human: tagErr(cmd, err)}
}

func exitS3(cmd string, err error) error {
	return &Exit{Code: exitc.S3, Human: tagErr(cmd, err)}
}

func exitS3Msg(cmd, format string, a ...any) error {
	return &Exit{Code: exitc.S3, Human: tagf(cmd, format, a...)}
}

func exitErr(cmd string, err error) error {
	return &Exit{Code: exitc.Err, Human: tagErr(cmd, err)}
}

func exitErrMsg(cmd, format string, a ...any) error {
	return &Exit{Code: exitc.Err, Human: tagf(cmd, format, a...)}
}

func exitNoMatches(cmd, format string, a ...any) error {
	return &Exit{Code: exitc.NoMatches, Human: tagf(cmd, format, a...)}
}
