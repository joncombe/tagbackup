package cli

import (
	"fmt"
	"os"
)

func (g *Runtime) printVersion() {
	_, _ = fmt.Fprintln(os.Stdout, "tagbackup", version)
}
