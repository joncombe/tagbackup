package store

import (
	"log/slog"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/retry"
)

// loggingRetryer wraps the default SDK retryer and emits a DEBUG log line each
// time a request error is about to trigger a retry attempt. The SDK itself
// handles the retry mechanics; we only observe via IsErrorRetryable.
type loggingRetryer struct {
	aws.Retryer
	log *slog.Logger
}

func (r *loggingRetryer) IsErrorRetryable(err error) bool {
	retryable := r.Retryer.IsErrorRetryable(err)
	if retryable && r.log != nil {
		r.log.Debug("sdk request is retryable", "err", err.Error())
	}
	return retryable
}

// newLoggingRetryer returns a retryer that delegates to the SDK default and
// emits a DEBUG log every time a retry decision is made.
func newLoggingRetryer(log *slog.Logger) aws.Retryer {
	return &loggingRetryer{Retryer: retry.NewStandard(), log: log}
}
