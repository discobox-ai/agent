// Command discobox-agent seeds a development manager and runs what is not
// the agent itself (ADR 26-10-15-157).
package main

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/discobox-ai/agent/internal/instance"
	"github.com/discobox-ai/agent/internal/templates"
	"github.com/discobox-ai/agent/internal/version"
)

const usage = `usage:
  discobox-agent init <dir>   seed a new manager in <dir>
  discobox-agent --version    print the version
`

func main() {
	os.Exit(run(context.Background(), os.Args[1:], os.Stdout, os.Stderr))
}

func run(ctx context.Context, args []string, stdout, stderr io.Writer) int {
	switch {
	case len(args) == 1 && (args[0] == "--version" || args[0] == "version"):
		_, _ = fmt.Fprintln(stdout, version.String())
		return 0
	case len(args) == 2 && args[0] == "init":
		dir := args[1]
		if err := instance.Init(ctx, dir, templates.Instance(), version.String()); err != nil {
			_, _ = fmt.Fprintf(stderr, "discobox-agent init: %v\n", err)
			return 1
		}
		_, _ = fmt.Fprintf(stdout, "Seeded a manager in %s. Start Claude Code there; it sets itself up first.\n", dir)
		return 0
	}
	_, _ = fmt.Fprint(stderr, usage)
	return 2
}
