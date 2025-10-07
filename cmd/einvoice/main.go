// Command einvoice validates electronic invoices against EN 16931 business rules.
package main

import (
	"fmt"
	"os"
)

const (
	exitOK         = 0 // Invoice is valid
	exitViolations = 1 // Invoice has validation violations
	exitError      = 2 // Error occurred (file not found, parse error, etc.)
)

func main() {
	os.Exit(run())
}

func run() int {
	// Check for subcommand
	if len(os.Args) < 2 {
		usage()
		return exitError
	}

	subcommand := os.Args[1]

	// Dispatch to subcommand
	switch subcommand {
	case "validate":
		return runValidate(os.Args[2:])
	default:
		fmt.Fprintf(os.Stderr, "Error: unknown command %q\n", subcommand)
		usage()
		return exitError
	}
}

func usage() {
	fmt.Fprintf(os.Stderr, `Usage: einvoice <command> [options]

Commands:
  validate    Validate an electronic invoice against EN 16931 business rules

Use "einvoice <command> --help" for more information about a command.
`)
}
