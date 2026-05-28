package app

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/JimSycurity/go-ticket/internal/ticket"
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

type JSONTicket struct {
	ID          string   `json:"id"`
	Status      string   `json:"status"`
	Deps        []string `json:"deps"`
	Links       []string `json:"links"`
	Created     string   `json:"created"`
	Type        string   `json:"type"`
	Priority    string   `json:"priority"`
	Assignee    string   `json:"assignee,omitempty"`
	ExternalRef string   `json:"external_ref,omitempty"`
	Parent      string   `json:"parent,omitempty"`
	Tags        []string `json:"tags,omitempty"`
	Title       string   `json:"title"`
	RelPath     string   `json:"rel_path"`
	AbsPath     string   `json:"abs_path"`
}

func Run(args []string, stdout io.Writer, stderr io.Writer) int {
	return RunWithIO(args, os.Stdin, stdout, stderr)
}

func RunWithIO(args []string, stdin io.Reader, stdout io.Writer, stderr io.Writer) int {
	if len(args) == 0 {
		fmt.Fprint(stdout, helpText)
		return 0
	}

	command := strings.TrimSpace(args[0])
	var err error
	switch command {
	case "help", "-h", "--help":
		fmt.Fprint(stdout, helpText)
		return 0
	case "init":
		err = runInit(stdout)
	case "create":
		err = runCreate(args[1:], stdout, stderr)
	case "show":
		err = runShow(args[1:], stdout)
	case "list", "ls":
		err = runList(args[1:], stdout, stderr)
	case "start":
		err = runStatus(args[1:], "in_progress", stdout)
	case "close":
		err = runStatus(args[1:], "closed", stdout)
	case "reopen":
		err = runStatus(args[1:], "open", stdout)
	case "status":
		err = runStatusArgs(args[1:], stdout)
	case "dep":
		err = runDependency(args[1:], true, stdout)
	case "undep":
		err = runDependency(args[1:], false, stdout)
	case "link":
		err = runLink(args[1:], true, stdout)
	case "unlink":
		err = runLink(args[1:], false, stdout)
	case "ready":
		err = runReadyBlocked(args[1:], false, stdout, stderr)
	case "blocked":
		err = runReadyBlocked(args[1:], true, stdout, stderr)
	case "add-note":
		err = runAddNote(args[1:], stdin, stdout)
	default:
		fmt.Fprintf(stderr, "unsupported command in gtk MVP: %s\n", command)
		return 2
	}
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	return 0
}

func runInit(stdout io.Writer) error {
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}
	root, err := ticket.Discover(cwd, nil)
	if err == nil {
		fmt.Fprintf(stdout, "Existing ticket directory: %s\n", root.TicketsDir)
		return nil
	}
	if !strings.Contains(err.Error(), "no .tickets directory found") {
		return err
	}
	path := filepath.Join(cwd, ticket.TicketsDirName)
	if err := os.Mkdir(path, 0o755); err != nil {
		return err
	}
	fmt.Fprintf(stdout, "Initialized ticket directory: %s\n", path)
	return nil
}

func runCreate(args []string, stdout io.Writer, stderr io.Writer) error {
	root, err := discoverForWrite()
	if err != nil {
		return err
	}
	fs := flag.NewFlagSet("create", flag.ContinueOnError)
	fs.SetOutput(stderr)
	description := fs.String("description", "", "description text")
	design := fs.String("design", "", "design text")
	acceptance := fs.String("acceptance", "", "acceptance text")
	ticketType := fs.String("type", "task", "ticket type")
	priority := fs.String("priority", "2", "priority 0-4")
	assignee := fs.String("assignee", "", "assignee")
	externalRef := fs.String("external-ref", "", "external reference")
	parent := fs.String("parent", "", "parent ticket ID")
	tags := fs.String("tags", "", "comma-separated tags")
	if err := fs.Parse(args); err != nil {
		return err
	}
	title := strings.TrimSpace(strings.Join(fs.Args(), " "))
	if title == "" {
		return fmt.Errorf("create requires a title")
	}
	if !ticket.IsValidType(*ticketType) {
		return fmt.Errorf("unsupported ticket type: %s", *ticketType)
	}
	if !ticket.IsValidPriority(*priority) {
		return fmt.Errorf("unsupported priority: %s", *priority)
	}
	t, err := ticket.NewTicket(root, title)
	if err != nil {
		return err
	}
	t.Type = *ticketType
	t.Priority = *priority
	t.Assignee = *assignee
	t.ExternalRef = *externalRef
	t.Parent = *parent
	t.Tags = splitCSV(*tags)
	t.Body = buildBody(title, *description, *design, *acceptance)
	if err := ticket.Write(root, t); err != nil {
		return err
	}
	fmt.Fprintln(stdout, t.ID)
	return nil
}

func runShow(args []string, stdout io.Writer) error {
	if len(args) != 1 {
		return fmt.Errorf("show requires exactly one ticket ID")
	}
	root, err := ticket.Discover(".", nil)
	if err != nil {
		return err
	}
	t, err := ticket.Resolve(root, args[0])
	if err != nil {
		return err
	}
	data, err := os.ReadFile(t.Path)
	if err != nil {
		return err
	}
	_, err = stdout.Write(data)
	return err
}

func runList(args []string, stdout io.Writer, stderr io.Writer) error {
	fs := flag.NewFlagSet("list", flag.ContinueOnError)
	fs.SetOutput(stderr)
	jsonOutput := fs.Bool("json", false, "emit JSON")
	status := fs.String("status", "", "filter status")
	assignee := fs.String("assignee", "", "filter assignee")
	ticketType := fs.String("type", "", "filter type")
	if err := fs.Parse(args); err != nil {
		return err
	}
	root, err := ticket.Discover(".", nil)
	if err != nil {
		return err
	}
	tickets, warnings := ticket.List(root)
	printWarnings(warnings, stderr)
	tickets = filterTickets(tickets, *status, *assignee, *ticketType)
	if *jsonOutput {
		out := make([]JSONTicket, 0, len(tickets))
		for _, t := range tickets {
			out = append(out, toJSONTicket(t))
		}
		encoder := json.NewEncoder(stdout)
		encoder.SetIndent("", "  ")
		return encoder.Encode(out)
	}
	for _, t := range tickets {
		fmt.Fprintf(stdout, "%s [P%s][%s] - %s\n", t.ID, t.Priority, t.Status, t.Title)
	}
	return nil
}

func runStatus(args []string, status string, stdout io.Writer) error {
	if len(args) != 1 {
		return fmt.Errorf("status update requires exactly one ticket ID")
	}
	return mutateOne(args[0], stdout, func(t ticket.Ticket) (ticket.Ticket, error) {
		t.Status = status
		return t, nil
	})
}

func runStatusArgs(args []string, stdout io.Writer) error {
	if len(args) != 2 {
		return fmt.Errorf("status requires a ticket ID and status")
	}
	status := args[1]
	if !ticket.IsValidStatus(status) {
		return fmt.Errorf("unsupported status: %s", status)
	}
	return runStatus([]string{args[0]}, status, stdout)
}

func runDependency(args []string, add bool, stdout io.Writer) error {
	if len(args) != 2 {
		return fmt.Errorf("dependency command requires ticket ID and dependency ID")
	}
	root, err := discoverForWrite()
	if err != nil {
		return err
	}
	t, err := ticket.Resolve(root, args[0])
	if err != nil {
		return err
	}
	dep, err := ticket.Resolve(root, args[1])
	if err != nil {
		return err
	}
	if add {
		t.Deps = ticket.AddUnique(t.Deps, dep.ID)
	} else {
		t.Deps = ticket.RemoveValue(t.Deps, dep.ID)
	}
	if err := ticket.Write(root, t); err != nil {
		return err
	}
	fmt.Fprintf(stdout, "Updated %s\n", t.ID)
	return nil
}

func runLink(args []string, add bool, stdout io.Writer) error {
	if len(args) != 2 {
		return fmt.Errorf("link command requires two ticket IDs")
	}
	root, err := discoverForWrite()
	if err != nil {
		return err
	}
	left, err := ticket.Resolve(root, args[0])
	if err != nil {
		return err
	}
	right, err := ticket.Resolve(root, args[1])
	if err != nil {
		return err
	}
	if add {
		left.Links = ticket.AddUnique(left.Links, right.ID)
		right.Links = ticket.AddUnique(right.Links, left.ID)
	} else {
		left.Links = ticket.RemoveValue(left.Links, right.ID)
		right.Links = ticket.RemoveValue(right.Links, left.ID)
	}
	if err := ticket.Write(root, left); err != nil {
		return err
	}
	if err := ticket.Write(root, right); err != nil {
		return err
	}
	fmt.Fprintf(stdout, "Updated %s and %s\n", left.ID, right.ID)
	return nil
}

func runReadyBlocked(args []string, blocked bool, stdout io.Writer, stderr io.Writer) error {
	if len(args) != 0 {
		return fmt.Errorf("ready/blocked do not accept arguments")
	}
	root, err := ticket.Discover(".", nil)
	if err != nil {
		return err
	}
	tickets, warnings := ticket.List(root)
	printWarnings(warnings, stderr)
	byID := map[string]ticket.Ticket{}
	malformed := map[string]bool{}
	for _, warning := range warnings {
		malformed[strings.TrimSuffix(filepath.Base(warning.Path), ".md")] = true
	}
	for _, t := range tickets {
		byID[t.ID] = t
	}
	for _, t := range tickets {
		if !ticket.IsWorkStatus(t.Status) {
			continue
		}
		blockers := blockersFor(t, byID, malformed)
		if blocked && len(blockers) > 0 {
			fmt.Fprintf(stdout, "%s [P%s][%s] - %s <- [%s]\n", t.ID, t.Priority, t.Status, t.Title, strings.Join(blockers, ", "))
		}
		if !blocked && len(blockers) == 0 {
			fmt.Fprintf(stdout, "%s [P%s][%s] - %s\n", t.ID, t.Priority, t.Status, t.Title)
		}
	}
	return nil
}

func runAddNote(args []string, stdin io.Reader, stdout io.Writer) error {
	if len(args) < 1 {
		return fmt.Errorf("add-note requires a ticket ID")
	}
	text := strings.TrimSpace(strings.Join(args[1:], " "))
	if text == "" {
		data, err := io.ReadAll(stdin)
		if err != nil {
			return err
		}
		text = string(data)
	}
	return mutateOne(args[0], stdout, func(t ticket.Ticket) (ticket.Ticket, error) {
		return ticket.AppendNote(t, text, time.Now()), nil
	})
}

func mutateOne(ref string, stdout io.Writer, mutate func(ticket.Ticket) (ticket.Ticket, error)) error {
	root, err := discoverForWrite()
	if err != nil {
		return err
	}
	t, err := ticket.Resolve(root, ref)
	if err != nil {
		return err
	}
	next, err := mutate(t)
	if err != nil {
		return err
	}
	if err := ticket.Write(root, next); err != nil {
		return err
	}
	fmt.Fprintf(stdout, "Updated %s\n", next.ID)
	return nil
}

func discoverForWrite() (ticket.Root, error) {
	root, err := ticket.Discover(".", nil)
	if err != nil {
		if strings.Contains(err.Error(), "no .tickets directory found") {
			return ticket.Root{}, fmt.Errorf("%w; run gtk init first", err)
		}
		return ticket.Root{}, err
	}
	return root, nil
}

func printWarnings(warnings []ticket.Warning, stderr io.Writer) {
	for _, warning := range warnings {
		fmt.Fprintf(stderr, "warning: %s\n", warning.Error())
	}
}

func filterTickets(tickets []ticket.Ticket, status string, assignee string, ticketType string) []ticket.Ticket {
	var out []ticket.Ticket
	for _, t := range tickets {
		if status != "" && t.Status != status {
			continue
		}
		if assignee != "" && t.Assignee != assignee {
			continue
		}
		if ticketType != "" && t.Type != ticketType {
			continue
		}
		out = append(out, t)
	}
	return out
}

func toJSONTicket(t ticket.Ticket) JSONTicket {
	return JSONTicket{
		ID:          t.ID,
		Status:      t.Status,
		Deps:        t.Deps,
		Links:       t.Links,
		Created:     t.Created,
		Type:        t.Type,
		Priority:    t.Priority,
		Assignee:    t.Assignee,
		ExternalRef: t.ExternalRef,
		Parent:      t.Parent,
		Tags:        t.Tags,
		Title:       t.Title,
		RelPath:     t.RelPath,
		AbsPath:     t.Path,
	}
}

func blockersFor(t ticket.Ticket, byID map[string]ticket.Ticket, malformed map[string]bool) []string {
	var blockers []string
	for _, depID := range t.Deps {
		dep, ok := byID[depID]
		if !ok || malformed[depID] {
			blockers = append(blockers, depID)
			continue
		}
		if dep.Status != "closed" {
			blockers = append(blockers, dep.ID)
		}
	}
	sort.Strings(blockers)
	return blockers
}

func buildBody(title string, description string, design string, acceptance string) string {
	var b strings.Builder
	b.WriteString("# ")
	b.WriteString(title)
	b.WriteString("\n")
	if strings.TrimSpace(description) != "" {
		b.WriteString("\n")
		b.WriteString(strings.TrimSpace(description))
		b.WriteString("\n")
	}
	if strings.TrimSpace(design) != "" {
		b.WriteString("\n## Design\n\n")
		b.WriteString(strings.TrimSpace(design))
		b.WriteString("\n")
	}
	if strings.TrimSpace(acceptance) != "" {
		b.WriteString("\n## Acceptance Criteria\n\n")
		b.WriteString(strings.TrimSpace(acceptance))
		b.WriteString("\n")
	}
	return b.String()
}

func splitCSV(value string) []string {
	if strings.TrimSpace(value) == "" {
		return nil
	}
	var out []string
	for _, part := range strings.Split(value, ",") {
		part = strings.TrimSpace(part)
		if part != "" {
			out = append(out, part)
		}
	}
	return out
}
