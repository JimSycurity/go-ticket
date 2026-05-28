package app

import (
	"fmt"
	"io"
	"strings"
)

const helpText = `gtk - Go ticket MVP CLI

Usage:
  gtk <command> [args]

Commands:
  help                 Show this help text
  init                 Initialize .tickets in the current directory
  create               Create a ticket
  show                 Print a raw ticket Markdown file
  list, ls             List tickets; use --json for compact JSON summaries
  start                Set a ticket to in_progress
  close                Set a ticket to closed
  reopen               Set a ticket to open
  status               Set a ticket status
  dep, undep           Add or remove a dependency
  link, unlink         Add or remove ticket links
  ready                List tickets whose dependencies are resolved
  blocked              List tickets with unresolved dependencies
  add-note             Append a timestamped note

MVP intentionally does not execute plugins.
`

func Run(args []string, stdout io.Writer, stderr io.Writer) int {
	if len(args) == 0 {
		fmt.Fprint(stdout, helpText)
		return 0
	}

	command := strings.TrimSpace(args[0])
	switch command {
	case "help", "-h", "--help":
		fmt.Fprint(stdout, helpText)
		return 0
	default:
		fmt.Fprintf(stderr, "unsupported command in gtk MVP: %s\n", command)
		return 2
	}
}
