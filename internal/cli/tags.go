package cli

import (
	"fmt"
	"strings"

	"github.com/joncombe/tagbackup/internal/objectkey"
)

func parseTagCSV(s string) ([]string, error) {
	if s == "" {
		return nil, fmt.Errorf("at least one tag is required")
	}
	if strings.ContainsAny(s, " \t\r\n") {
		return nil, fmt.Errorf("tags must not contain spaces; use commas between tags")
	}
	parts := strings.Split(s, ",")
	for _, p := range parts {
		if err := objectkey.ValidateTag(p); err != nil {
			return nil, err
		}
	}
	return parts, nil
}
