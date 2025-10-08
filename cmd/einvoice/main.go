// Command einvoice validates electronic invoices against EN 16931 business rules.
package main

import (
	"fmt"
	"os"
)

const (
	exitOK         = 0 // Success
	exitError      = 1 // Error occurred (file not found, parse error, etc.)
	exitViolations = 2 // Invoice has validation violations (validate command only)
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
	case "info":
		return runInfo(os.Args[2:])
	default:
		fmt.Fprintf(os.Stderr, "Error: unknown command %q\n", subcommand)
		usage()
		return exitError
	}
}

func usage() {
	fmt.Fprintf(os.Stderr, `Usage: einvoice <command> [options]

Commands:
  info        Display detailed information about an electronic invoice
  validate    Validate an electronic invoice against EN 16931 business rules

Use "einvoice <command> --help" for more information about a command.
`)
}
