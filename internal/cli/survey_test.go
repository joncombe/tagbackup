package cli

import (
	"errors"
	"testing"

	"github.com/AlecAivazis/survey/v2/terminal"
	"github.com/joncombe/tagbackup/internal/exitc"
)

func TestMapSurveyErr_Interrupt(t *testing.T) {
	err := mapSurveyErr(terminal.InterruptErr)
	var exit *Exit
	if !errors.As(err, &exit) {
		t.Fatalf("expected *Exit, got %T", err)
	}
	if exit.Code != exitc.SIGINT {
		t.Fatalf("code = %d, want %d", exit.Code, exitc.SIGINT)
	}
	if exit.Human != "" {
		t.Fatalf("Human = %q, want empty", exit.Human)
	}
}

func TestMapSurveyErr_Other(t *testing.T) {
	orig := errors.New("validation failed")
	if got := mapSurveyErr(orig); got != orig {
		t.Fatalf("got %v, want %v", got, orig)
	}
}

func TestMapSurveyErr_Nil(t *testing.T) {
	if got := mapSurveyErr(nil); got != nil {
		t.Fatalf("got %v, want nil", got)
	}
}
