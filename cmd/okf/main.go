// Command okf is the Open Knowledge Format toolkit CLI.
package main

import (
	"io"
	"os"
)

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}

// run executes a single CLI invocation and returns the process exit code:
// 0 on success, 1 when a validation-type command reports failure, 2 for
// usage errors (unknown command or bad flags).
func run(args []string, stdout, stderr io.Writer) int {
	panic("not implemented")
}
