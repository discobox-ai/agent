// Command discobox-agent is the manager agent's companion binary.
package main

import (
	"fmt"
	"os"

	"github.com/discobox-ai/agent/internal/version"
)

func main() {
	if len(os.Args) == 2 && (os.Args[1] == "--version" || os.Args[1] == "version") {
		fmt.Println(version.String())
		return
	}
	fmt.Fprintln(os.Stderr, "usage: discobox-agent --version")
	os.Exit(2)
}
