package cli

import (
	"errors"

	"github.com/AlecAivazis/survey/v2"
	"github.com/AlecAivazis/survey/v2/terminal"
	"github.com/joncombe/tagbackup/internal/exitc"
)

// askOne runs a single survey prompt. Ctrl-C returns a silent *Exit with code 130.
func askOne(q survey.Prompt, response interface{}, opts ...survey.AskOpt) error {
	if err := survey.AskOne(q, response, opts...); err != nil {
		return mapSurveyErr(err)
	}
	return nil
}

// askOneErr is like askOne but wraps non-interrupt survey errors with exitErr.
func askOneErr(cmd string, q survey.Prompt, response interface{}, opts ...survey.AskOpt) error {
	if err := askOne(q, response, opts...); err != nil {
		if _, ok := err.(*Exit); ok {
			return err
		}
		return exitErr(cmd, err)
	}
	return nil
}

func mapSurveyErr(err error) error {
	if err != nil && errors.Is(err, terminal.InterruptErr) {
		return &Exit{Code: exitc.SIGINT}
	}
	return err
}
