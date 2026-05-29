package app

import (
	"bufio"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"runtime/debug"
	"sort"
	"strings"
	"time"

	"github.com/JimSycurity/go-ticket/internal/ticket"
)

var (
	Version   = "dev"
	Commit    = ""
	BuildDate = ""
)

const (
	maxBeadsImportBytes = 10 << 20
	maxBeadsLineBytes   = 1 << 20
)

const helpText = `gtk - Go ticket CLI

Usage:
  gtk <command> [args]

Commands:
  help                 Show this help text
  init                 Initialize .tickets in the current directory
  create               Create a ticket
  show                 Print a raw ticket Markdown file
  edit                 Open a ticket in a validated editor
  list, ls             List tickets; use --json for compact JSON summaries
  query                Emit ticket frontmatter as JSONL; filters are deferred
  start                Set a ticket to in_progress
  close                Set a ticket to closed
  reopen               Set a ticket to open
  status               Set a ticket status
  dep, undep           Add or remove a dependency
  dep tree, dep cycle  Show dependency trees or detect dependency cycles
  link, unlink         Add or remove ticket links
  ready                List tickets whose dependencies are resolved
  blocked              List tickets with unresolved dependencies
  closed               List recently closed tickets
  add-note             Append a timestamped note
  migrate-beads        Import .beads/issues.jsonl into .tickets
  super                Run a builtin command without plugin dispatch
  version              Show build and VCS metadata

Plugin execution is limited to reviewed PATH policy.
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

type QueryTicket struct {
	ID          string   `json:"id"`
	Status      string   `json:"status"`
	Deps        []string `json:"deps"`
	Links       []string `json:"links"`
	Created     string   `json:"created"`
	Type        string   `json:"type"`
	Priority    string   `json:"priority"`
	Assignee    string   `json:"assignee,omitempty"`
	ExternalRef string   `json:"external-ref,omitempty"`
	Parent      string   `json:"parent,omitempty"`
	Tags        []string `json:"tags,omitempty"`
}

var pluginCommandPattern = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9_-]*$`)

func Run(args []string, stdout io.Writer, stderr io.Writer) int {
	return RunWithIO(args, os.Stdin, stdout, stderr)
}

func RunWithIO(args []string, stdin io.Reader, stdout io.Writer, stderr io.Writer) int {
	return runCommand(args, stdin, stdout, stderr, true)
}

func runCommand(args []string, stdin io.Reader, stdout io.Writer, stderr io.Writer, allowPlugins bool) int {
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
	case "version", "--version":
		printVersion(stdout)
		return 0
	case "init":
		err = runInit(stdout)
	case "create":
		err = runCreate(args[1:], stdout, stderr)
	case "show":
		err = runShow(args[1:], stdout)
	case "edit":
		err = runEdit(args[1:], stdin, stdout, stderr)
	case "list", "ls":
		err = runList(args[1:], stdout, stderr)
	case "query":
		err = runQuery(args[1:], stdout, stderr)
	case "start":
		err = runStatus(args[1:], "in_progress", stdout)
	case "close":
		err = runStatus(args[1:], "closed", stdout)
	case "reopen":
		err = runStatus(args[1:], "open", stdout)
	case "status":
		err = runStatusArgs(args[1:], stdout)
	case "dep":
		err = runDep(args[1:], stdout, stderr)
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
	case "closed":
		err = runClosed(args[1:], stdout, stderr)
	case "add-note":
		err = runAddNote(args[1:], stdin, stdout)
	case "migrate-beads":
		err = runMigrateBeads(args[1:], stdout, stderr)
	case "super":
		if len(args) < 2 {
			fmt.Fprintln(stderr, "super requires a builtin command")
			return 1
		}
		return runCommand(args[1:], stdin, stdout, stderr, false)
	default:
		if allowPlugins {
			err = runPlugin(command, args[1:], stdin, stdout, stderr)
			break
		}
		fmt.Fprintf(stderr, "unsupported builtin command: %s\n", command)
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
	data, err := ticket.ReadRawFile(t.Path)
	if err != nil {
		return err
	}
	_, err = stdout.Write(data)
	return err
}

func runEdit(args []string, stdin io.Reader, stdout io.Writer, stderr io.Writer) error {
	if len(args) != 1 {
		return fmt.Errorf("edit requires exactly one ticket ID")
	}
	root, err := ticket.Discover(".", nil)
	if err != nil {
		return err
	}
	t, err := ticket.Resolve(root, args[0])
	if err != nil {
		return err
	}
	editor, err := resolveEditor()
	if err != nil {
		return err
	}
	cmd := exec.Command(editor, t.Path)
	cmd.Stdin = stdin
	cmd.Stdout = stdout
	cmd.Stderr = stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("editor failed: %w", err)
	}
	return nil
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

func runQuery(args []string, stdout io.Writer, stderr io.Writer) error {
	if len(args) > 0 {
		return fmt.Errorf("query filters are deferred for future feature parity; run query without a filter for JSONL output")
	}
	root, err := ticket.Discover(".", nil)
	if err != nil {
		return err
	}
	tickets, warnings := ticket.List(root)
	printWarnings(warnings, stderr)
	encoder := json.NewEncoder(stdout)
	for _, t := range tickets {
		if err := encoder.Encode(toQueryTicket(t)); err != nil {
			return err
		}
	}
	return nil
}

func runPlugin(command string, args []string, stdin io.Reader, stdout io.Writer, stderr io.Writer) error {
	if err := validatePluginCommand(command); err != nil {
		return err
	}
	root, err := ticket.Discover(".", nil)
	if err != nil {
		return err
	}
	path, ok, err := resolvePlugin(command, os.Getenv("PATH"), runtime.GOOS)
	if err != nil {
		return err
	}
	if !ok {
		return fmt.Errorf("unsupported command: %s", command)
	}
	cmd := exec.Command(path, args...)
	cmd.Stdin = stdin
	cmd.Stdout = stdout
	cmd.Stderr = stderr
	cmd.Env = pluginEnv(root)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("plugin %s failed: %w", command, err)
	}
	return nil
}

func validatePluginCommand(command string) error {
	if !pluginCommandPattern.MatchString(command) {
		return fmt.Errorf("invalid plugin command name: %s", command)
	}
	return nil
}

func resolvePlugin(command string, pathEnv string, goos string) (string, bool, error) {
	if err := validatePluginCommand(command); err != nil {
		return "", false, err
	}
	for _, dir := range filepath.SplitList(pathEnv) {
		if dir == "" || dir == "." || !filepath.IsAbs(dir) {
			continue
		}
		for _, base := range []string{"tk-" + command, "ticket-" + command} {
			for _, ext := range pluginExtensions(goos) {
				candidate := filepath.Join(dir, base+ext)
				ok, err := validPluginExecutable(candidate, goos)
				if err != nil {
					return "", false, err
				}
				if ok {
					return candidate, true, nil
				}
			}
		}
	}
	return "", false, nil
}

func pluginExtensions(goos string) []string {
	if goos == "windows" {
		return []string{".exe"}
	}
	return []string{""}
}

func validPluginExecutable(path string, goos string) (bool, error) {
	info, err := os.Lstat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}
	if info.Mode()&os.ModeSymlink != 0 || !info.Mode().IsRegular() {
		return false, nil
	}
	if goos != "windows" && info.Mode().Perm()&0o111 == 0 {
		return false, nil
	}
	return true, nil
}

func pluginEnv(root ticket.Root) []string {
	var env []string
	for _, item := range os.Environ() {
		key, _, ok := strings.Cut(item, "=")
		if !ok {
			continue
		}
		if keepPluginEnv(key) {
			env = append(env, item)
		}
	}
	env = append(env, "TICKETS_DIR="+root.TicketsDir)
	env = append(env, "TICKET_PROJECT_DIR="+root.ProjectDir)
	return env
}

func resolveEditor() (string, error) {
	for _, key := range []string{"GTK_EDITOR", "VISUAL", "EDITOR"} {
		value := strings.TrimSpace(os.Getenv(key))
		if value == "" {
			continue
		}
		editor, err := resolveEditorCommand(value, os.Getenv("PATH"), runtime.GOOS)
		if err != nil {
			return "", fmt.Errorf("invalid %s: %w", key, err)
		}
		return editor, nil
	}
	return "", fmt.Errorf("no editor configured; set GTK_EDITOR to an editor executable")
}

func resolveEditorCommand(value string, pathEnv string, goos string) (string, error) {
	if err := validateEditorCommand(value); err != nil {
		return "", err
	}
	if filepath.IsAbs(value) {
		ok, err := validPluginExecutable(value, goos)
		if err != nil {
			return "", err
		}
		if !ok {
			return "", fmt.Errorf("editor is not an executable regular file: %s", value)
		}
		return value, nil
	}
	path, ok, err := resolveCommandInPath(value, pathEnv, goos)
	if err != nil {
		return "", err
	}
	if !ok {
		return "", fmt.Errorf("editor command not found on safe PATH: %s", value)
	}
	return path, nil
}

func validateEditorCommand(value string) error {
	if value == "" {
		return fmt.Errorf("editor command cannot be empty")
	}
	if strings.ContainsAny(value, " \t\r\n") {
		return fmt.Errorf("editor command must not include inline arguments")
	}
	if strings.ContainsAny(value, "&|;<>()$`'\"") {
		return fmt.Errorf("editor command contains shell metacharacters")
	}
	if !filepath.IsAbs(value) && (strings.Contains(value, "/") || strings.Contains(value, "\\")) {
		return fmt.Errorf("editor command must be a command name or absolute path")
	}
	return nil
}

func keepPluginEnv(key string) bool {
	switch strings.ToUpper(key) {
	case "SYSTEMROOT", "WINDIR", "TEMP", "TMP":
		return true
	default:
		return false
	}
}

func resolveCommandInPath(command string, pathEnv string, goos string) (string, bool, error) {
	if command == "" || strings.Contains(command, ".") {
		return "", false, fmt.Errorf("unsupported command name: %s", command)
	}
	for _, dir := range filepath.SplitList(pathEnv) {
		if dir == "" || dir == "." || !filepath.IsAbs(dir) {
			continue
		}
		for _, ext := range pluginExtensions(goos) {
			candidate := filepath.Join(dir, command+ext)
			ok, err := validPluginExecutable(candidate, goos)
			if err != nil {
				return "", false, err
			}
			if ok {
				return candidate, true, nil
			}
		}
	}
	return "", false, nil
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

func runDep(args []string, stdout io.Writer, stderr io.Writer) error {
	if len(args) == 0 {
		return fmt.Errorf("dependency command requires arguments")
	}
	switch args[0] {
	case "tree":
		return runDepTree(args[1:], stdout, stderr)
	case "cycle":
		return runDepCycle(args[1:], stdout, stderr)
	default:
		return runDependency(args, true, stdout)
	}
}

func runDepTree(args []string, stdout io.Writer, stderr io.Writer) error {
	fs := flag.NewFlagSet("dep tree", flag.ContinueOnError)
	fs.SetOutput(stderr)
	full := fs.Bool("full", false, "show repeated dependencies")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() != 1 {
		return fmt.Errorf("dep tree requires exactly one ticket ID")
	}
	root, err := ticket.Discover(".", nil)
	if err != nil {
		return err
	}
	tickets, warnings := ticket.List(root)
	printWarnings(warnings, stderr)
	byID := ticketsByID(tickets)
	start, err := ticket.Resolve(root, fs.Arg(0))
	if err != nil {
		return err
	}
	seen := map[string]bool{}
	printDependencyTree(stdout, start, byID, "", true, *full, seen, map[string]bool{})
	return nil
}

func printDependencyTree(stdout io.Writer, t ticket.Ticket, byID map[string]ticket.Ticket, prefix string, isRoot bool, full bool, seen map[string]bool, path map[string]bool) {
	if !full && seen[t.ID] {
		return
	}
	if !full {
		seen[t.ID] = true
	}

	line := ticketLine(t)
	if isRoot {
		fmt.Fprintln(stdout, line)
	} else {
		connector := "└── "
		if !path["__last__"] {
			connector = "├── "
		}
		fmt.Fprintln(stdout, prefix+connector+line)
	}

	if path[t.ID] {
		return
	}
	nextPath := cloneBoolMap(path)
	nextPath[t.ID] = true
	delete(nextPath, "__last__")

	children := dependencyTickets(t, byID)
	for i, dep := range children {
		isLast := i == len(children)-1
		childPrefix := prefix
		if !isRoot {
			if path["__last__"] {
				childPrefix += "    "
			} else {
				childPrefix += "│   "
			}
		}
		childPath := cloneBoolMap(nextPath)
		childPath["__last__"] = isLast
		printDependencyTree(stdout, dep, byID, childPrefix, false, full, seen, childPath)
	}
}

func runDepCycle(args []string, stdout io.Writer, stderr io.Writer) error {
	if len(args) != 0 {
		return fmt.Errorf("dep cycle does not accept arguments")
	}
	root, err := ticket.Discover(".", nil)
	if err != nil {
		return err
	}
	tickets, warnings := ticket.List(root)
	printWarnings(warnings, stderr)
	cycles := findDependencyCycles(tickets)
	if len(cycles) == 0 {
		fmt.Fprintln(stdout, "No dependency cycles found")
		return nil
	}
	byID := ticketsByID(tickets)
	for i, cycle := range cycles {
		fmt.Fprintf(stdout, "Cycle %d: %s\n", i+1, strings.Join(cycle, " -> "))
		seen := map[string]bool{}
		for _, id := range cycle {
			if seen[id] {
				continue
			}
			seen[id] = true
			if t, ok := byID[id]; ok {
				fmt.Fprintf(stdout, "  %-8s [%s] %s\n", t.ID, t.Status, t.Title)
			}
		}
	}
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
	originalLeft := left
	if add {
		left.Links = ticket.AddUnique(left.Links, right.ID)
		right.Links = ticket.AddUnique(right.Links, left.ID)
	} else {
		left.Links = ticket.RemoveValue(left.Links, right.ID)
		right.Links = ticket.RemoveValue(right.Links, left.ID)
	}
	if err := ticket.PreflightWrite(root, left); err != nil {
		return err
	}
	if err := ticket.PreflightWrite(root, right); err != nil {
		return err
	}
	if err := ticket.Write(root, left); err != nil {
		return err
	}
	if err := ticket.Write(root, right); err != nil {
		if rollbackErr := ticket.Write(root, originalLeft); rollbackErr != nil {
			return fmt.Errorf("failed updating %s after updating %s: %w; rollback failed: %v", right.ID, left.ID, err, rollbackErr)
		}
		return fmt.Errorf("failed updating %s after updating %s: %w; rolled back %s", right.ID, left.ID, err, left.ID)
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

type closedTicket struct {
	ticket ticket.Ticket
	mtime  time.Time
}

func runClosed(args []string, stdout io.Writer, stderr io.Writer) error {
	fs := flag.NewFlagSet("closed", flag.ContinueOnError)
	fs.SetOutput(stderr)
	limit := fs.Int("limit", 20, "maximum closed tickets to show")
	assignee := fs.String("a", "", "filter assignee")
	ticketType := fs.String("T", "", "filter type")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() != 0 {
		return fmt.Errorf("closed does not accept positional arguments")
	}
	if *limit < 0 {
		return fmt.Errorf("closed limit must be non-negative")
	}
	root, err := ticket.Discover(".", nil)
	if err != nil {
		return err
	}
	tickets, warnings := ticket.List(root)
	printWarnings(warnings, stderr)
	var closed []closedTicket
	for _, t := range tickets {
		if t.Status != "closed" {
			continue
		}
		if *assignee != "" && t.Assignee != *assignee {
			continue
		}
		if *ticketType != "" && t.Type != *ticketType {
			continue
		}
		info, err := os.Stat(t.Path)
		if err != nil {
			printWarnings([]ticket.Warning{{Path: t.Path, Err: err}}, stderr)
			continue
		}
		closed = append(closed, closedTicket{ticket: t, mtime: info.ModTime()})
	}
	sort.SliceStable(closed, func(i, j int) bool {
		if !closed[i].mtime.Equal(closed[j].mtime) {
			return closed[i].mtime.After(closed[j].mtime)
		}
		return closed[i].ticket.ID < closed[j].ticket.ID
	})
	for i, item := range closed {
		if i >= *limit {
			break
		}
		fmt.Fprintf(stdout, "%-8s [%s] - %s\n", item.ticket.ID, item.ticket.Status, item.ticket.Title)
	}
	return nil
}

func runAddNote(args []string, stdin io.Reader, stdout io.Writer) error {
	if len(args) < 1 {
		return fmt.Errorf("add-note requires a ticket ID")
	}
	text := strings.TrimSpace(strings.Join(args[1:], " "))
	if text == "" {
		data, err := io.ReadAll(io.LimitReader(stdin, ticket.MaxNoteBytes+1))
		if err != nil {
			return err
		}
		if len(data) > ticket.MaxNoteBytes {
			return fmt.Errorf("note exceeds %d byte limit", ticket.MaxNoteBytes)
		}
		text = string(data)
	}
	if len([]byte(text)) > ticket.MaxNoteBytes {
		return fmt.Errorf("note exceeds %d byte limit", ticket.MaxNoteBytes)
	}
	return mutateOne(args[0], stdout, func(t ticket.Ticket) (ticket.Ticket, error) {
		return ticket.AppendNote(t, text, time.Now()), nil
	})
}

type beadsIssue struct {
	ID          string
	Title       string
	Description string
	Status      string
	Type        string
	Priority    string
	Assignee    string
	Created     string
	Parent      string
	Deps        []string
	Links       []string
	Tags        []string
}

type migrationReport struct {
	imported int
	skipped  int
	review   int
}

func runMigrateBeads(args []string, stdout io.Writer, stderr io.Writer) error {
	root, err := discoverForWrite()
	if err != nil {
		return err
	}
	fs := flag.NewFlagSet("migrate-beads", flag.ContinueOnError)
	fs.SetOutput(stderr)
	dryRun := fs.Bool("dry-run", false, "report import actions without writing tickets")
	source := fs.String("source", filepath.Join(".beads", "issues.jsonl"), "Beads JSONL export path under project root")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() != 0 {
		return fmt.Errorf("migrate-beads does not accept positional arguments")
	}
	sourcePath, err := resolveMigrationSource(root, *source)
	if err != nil {
		return err
	}
	report, err := migrateBeads(root, sourcePath, *dryRun, stdout)
	if err != nil {
		return err
	}
	mode := "imported"
	if *dryRun {
		mode = "dry-run"
	}
	fmt.Fprintf(stdout, "migrate-beads %s: imported=%d skipped=%d review=%d source=%s\n", mode, report.imported, report.skipped, report.review, sourcePath)
	return nil
}

func resolveMigrationSource(root ticket.Root, source string) (string, error) {
	if strings.TrimSpace(source) == "" {
		return "", fmt.Errorf("migration source cannot be empty")
	}
	var path string
	if filepath.IsAbs(source) {
		path = filepath.Clean(source)
	} else {
		path = filepath.Join(root.ProjectDir, filepath.Clean(source))
	}
	resolved, err := filepath.Abs(path)
	if err != nil {
		return "", fmt.Errorf("resolve migration source: %w", err)
	}
	project, err := filepath.Abs(root.ProjectDir)
	if err != nil {
		return "", fmt.Errorf("resolve project root: %w", err)
	}
	rel, err := filepath.Rel(project, resolved)
	if err != nil {
		return "", fmt.Errorf("check migration source containment: %w", err)
	}
	if rel == ".." || strings.HasPrefix(rel, ".."+string(os.PathSeparator)) || filepath.IsAbs(rel) {
		return "", fmt.Errorf("migration source must be inside project root: %s", resolved)
	}
	info, err := os.Lstat(resolved)
	if err != nil {
		return "", fmt.Errorf("stat migration source: %w", err)
	}
	if info.Mode()&os.ModeSymlink != 0 || !info.Mode().IsRegular() {
		return "", fmt.Errorf("migration source is not a regular file: %s", resolved)
	}
	if info.Size() > maxBeadsImportBytes {
		return "", fmt.Errorf("migration source exceeds %d byte limit: %s", maxBeadsImportBytes, resolved)
	}
	realSource, err := filepath.EvalSymlinks(resolved)
	if err != nil {
		return "", fmt.Errorf("canonicalize migration source: %w", err)
	}
	if filepath.Clean(realSource) != filepath.Clean(resolved) {
		return "", fmt.Errorf("migration source path contains symlink indirection: %s", resolved)
	}
	return resolved, nil
}

func migrateBeads(root ticket.Root, sourcePath string, dryRun bool, stdout io.Writer) (migrationReport, error) {
	file, err := os.Open(sourcePath)
	if err != nil {
		return migrationReport{}, err
	}
	defer file.Close()

	var report migrationReport
	scanner := bufio.NewScanner(file)
	scanner.Buffer(make([]byte, 64<<10), maxBeadsLineBytes)
	lineNumber := 0
	for scanner.Scan() {
		lineNumber++
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		issue, reviews, err := parseBeadsIssue(line)
		if err != nil {
			report.review++
			fmt.Fprintf(stdout, "review line %d: %v\n", lineNumber, err)
			continue
		}
		if reviews != "" {
			report.review++
			fmt.Fprintf(stdout, "review line %d: %s\n", lineNumber, reviews)
		}
		if _, err := ticket.ResolveTicketPath(root, issue.ID, true); err == nil {
			report.skipped++
			fmt.Fprintf(stdout, "skip %s: ticket already exists\n", issue.ID)
			continue
		} else if !errors.Is(err, os.ErrNotExist) {
			report.review++
			fmt.Fprintf(stdout, "review %s: %v\n", issue.ID, err)
			continue
		}
		t := ticket.Ticket{
			ID:          issue.ID,
			Status:      issue.Status,
			Deps:        issue.Deps,
			Links:       issue.Links,
			Created:     issue.Created,
			Type:        issue.Type,
			Priority:    issue.Priority,
			Assignee:    issue.Assignee,
			ExternalRef: "beads:" + issue.ID,
			Parent:      issue.Parent,
			Tags:        ticket.AddUnique(issue.Tags, "beads-import"),
			Title:       issue.Title,
			Body:        beadsBody(issue.Title, issue.Description),
		}
		if dryRun {
			report.imported++
			fmt.Fprintf(stdout, "would import %s: %s\n", issue.ID, issue.Title)
			continue
		}
		if err := ticket.Write(root, t); err != nil {
			report.review++
			fmt.Fprintf(stdout, "review %s: %v\n", issue.ID, err)
			continue
		}
		report.imported++
		fmt.Fprintf(stdout, "imported %s: %s\n", issue.ID, issue.Title)
	}
	if err := scanner.Err(); err != nil {
		return report, fmt.Errorf("read migration source: %w", err)
	}
	return report, nil
}

func parseBeadsIssue(line string) (beadsIssue, string, error) {
	var raw map[string]any
	if err := json.Unmarshal([]byte(line), &raw); err != nil {
		return beadsIssue{}, "", fmt.Errorf("invalid JSON: %w", err)
	}
	var reviews []string
	id, ok := stringField(raw, "id")
	if !ok || id == "" {
		return beadsIssue{}, "", fmt.Errorf("missing id")
	}
	if err := ticket.ValidateID(id); err != nil {
		return beadsIssue{}, "", err
	}
	title, ok := firstStringField(raw, "title", "summary", "name")
	if !ok || strings.TrimSpace(title) == "" {
		return beadsIssue{}, "", fmt.Errorf("missing title")
	}
	status, ok := stringField(raw, "status")
	if !ok || status == "" {
		status = "open"
	}
	status = normalizeBeadsStatus(status)
	if !ticket.IsValidStatus(status) {
		reviews = append(reviews, "unsupported status defaulted to open")
		status = "open"
	}
	ticketType, ok := firstStringField(raw, "type", "kind")
	if !ok || ticketType == "" {
		ticketType = "task"
	}
	if !ticket.IsValidType(ticketType) {
		reviews = append(reviews, "unsupported type defaulted to task")
		ticketType = "task"
	}
	priority := priorityField(raw)
	if !ticket.IsValidPriority(priority) {
		reviews = append(reviews, "unsupported priority defaulted to 2")
		priority = "2"
	}
	parent, _ := stringField(raw, "parent")
	if parent != "" {
		if err := ticket.ValidateID(parent); err != nil {
			return beadsIssue{}, "", fmt.Errorf("invalid parent: %w", err)
		}
	}
	deps, err := firstStringListField(raw, "deps", "dependencies", "blocked_by")
	if err != nil {
		return beadsIssue{}, "", err
	}
	for _, dep := range deps {
		if err := ticket.ValidateID(dep); err != nil {
			return beadsIssue{}, "", fmt.Errorf("invalid dependency %q: %w", dep, err)
		}
	}
	links, err := firstStringListField(raw, "links", "related")
	if err != nil {
		return beadsIssue{}, "", err
	}
	for _, link := range links {
		if err := ticket.ValidateID(link); err != nil {
			return beadsIssue{}, "", fmt.Errorf("invalid link %q: %w", link, err)
		}
	}
	tags, err := firstStringListField(raw, "tags", "labels")
	if err != nil {
		return beadsIssue{}, "", err
	}
	for _, tag := range tags {
		if strings.ContainsAny(tag, "\r\n,[]") {
			return beadsIssue{}, "", fmt.Errorf("invalid tag %q", tag)
		}
	}
	description, _ := firstStringField(raw, "description", "body", "content")
	assignee, _ := stringField(raw, "assignee")
	created, _ := firstStringField(raw, "created", "created_at", "createdAt")
	return beadsIssue{
		ID:          id,
		Title:       strings.TrimSpace(title),
		Description: strings.TrimSpace(description),
		Status:      status,
		Type:        ticketType,
		Priority:    priority,
		Assignee:    assignee,
		Created:     created,
		Parent:      parent,
		Deps:        deps,
		Links:       links,
		Tags:        tags,
	}, strings.Join(reviews, "; "), nil
}

func stringField(raw map[string]any, key string) (string, bool) {
	value, ok := raw[key]
	if !ok || value == nil {
		return "", false
	}
	text, ok := value.(string)
	if !ok {
		return "", false
	}
	return strings.TrimSpace(text), true
}

func firstStringField(raw map[string]any, keys ...string) (string, bool) {
	for _, key := range keys {
		if value, ok := stringField(raw, key); ok {
			return value, true
		}
	}
	return "", false
}

func firstStringListField(raw map[string]any, keys ...string) ([]string, error) {
	for _, key := range keys {
		value, ok := raw[key]
		if !ok || value == nil {
			continue
		}
		list, err := stringList(value, key)
		if err != nil {
			return nil, err
		}
		return list, nil
	}
	return []string{}, nil
}

func stringList(value any, key string) ([]string, error) {
	items, ok := value.([]any)
	if !ok {
		return nil, fmt.Errorf("%s must be a string array", key)
	}
	var out []string
	for _, item := range items {
		text, ok := item.(string)
		if !ok {
			return nil, fmt.Errorf("%s must contain only strings", key)
		}
		text = strings.TrimSpace(text)
		if text != "" {
			out = append(out, text)
		}
	}
	return out, nil
}

func priorityField(raw map[string]any) string {
	value, ok := raw["priority"]
	if !ok || value == nil {
		return "2"
	}
	switch typed := value.(type) {
	case string:
		return strings.TrimSpace(typed)
	case float64:
		if typed == float64(int(typed)) {
			return fmt.Sprintf("%d", int(typed))
		}
	}
	return "2"
}

func normalizeBeadsStatus(status string) string {
	switch strings.ToLower(strings.TrimSpace(status)) {
	case "in-progress", "in progress", "started", "active":
		return "in_progress"
	case "done", "resolved", "complete", "completed":
		return "closed"
	default:
		return status
	}
}

func beadsBody(title string, description string) string {
	if strings.TrimSpace(description) == "" {
		return "# " + title + "\n"
	}
	return "# " + title + "\n\n" + strings.TrimSpace(description) + "\n"
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

func printVersion(stdout io.Writer) {
	settings := buildSettings()
	commit := firstNonEmpty(Commit, settings["vcs.revision"], "unknown")
	vcsTime := firstNonEmpty(settings["vcs.time"], "unknown")
	buildDate := firstNonEmpty(BuildDate, "unknown")
	dirty := firstNonEmpty(settings["vcs.modified"], "unknown")
	binaryPath, binaryMTime := executableMetadata()

	fmt.Fprintf(stdout, "gtk version: %s\n", Version)
	fmt.Fprintf(stdout, "commit: %s\n", commit)
	fmt.Fprintf(stdout, "dirty: %s\n", dirty)
	fmt.Fprintf(stdout, "vcs_time: %s\n", vcsTime)
	fmt.Fprintf(stdout, "build_date: %s\n", buildDate)
	fmt.Fprintf(stdout, "binary: %s\n", binaryPath)
	fmt.Fprintf(stdout, "binary_mtime: %s\n", binaryMTime)
}

func buildSettings() map[string]string {
	settings := map[string]string{}
	info, ok := debug.ReadBuildInfo()
	if !ok {
		return settings
	}
	for _, setting := range info.Settings {
		settings[setting.Key] = setting.Value
	}
	return settings
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}

func executableMetadata() (string, string) {
	path, err := os.Executable()
	if err != nil {
		return "unknown", "unknown"
	}
	info, err := os.Stat(path)
	if err != nil {
		return path, "unknown"
	}
	return path, info.ModTime().UTC().Format(time.RFC3339)
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

func toQueryTicket(t ticket.Ticket) QueryTicket {
	return QueryTicket{
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
	}
}

func ticketsByID(tickets []ticket.Ticket) map[string]ticket.Ticket {
	byID := map[string]ticket.Ticket{}
	for _, t := range tickets {
		byID[t.ID] = t
	}
	return byID
}

func dependencyTickets(t ticket.Ticket, byID map[string]ticket.Ticket) []ticket.Ticket {
	var deps []ticket.Ticket
	for _, depID := range t.Deps {
		if dep, ok := byID[depID]; ok {
			deps = append(deps, dep)
		}
	}
	sort.SliceStable(deps, func(i, j int) bool {
		return deps[i].ID < deps[j].ID
	})
	return deps
}

func ticketLine(t ticket.Ticket) string {
	return fmt.Sprintf("%s [%s] %s", t.ID, t.Status, t.Title)
}

func cloneBoolMap(in map[string]bool) map[string]bool {
	out := make(map[string]bool, len(in))
	for key, value := range in {
		out[key] = value
	}
	return out
}

func findDependencyCycles(tickets []ticket.Ticket) [][]string {
	byID := ticketsByID(tickets)
	var ids []string
	for _, t := range tickets {
		if ticket.IsWorkStatus(t.Status) {
			ids = append(ids, t.ID)
		}
	}
	sort.Strings(ids)

	var cycles [][]string
	seenCycles := map[string]bool{}
	visited := map[string]bool{}
	var visit func(string, []string)
	visit = func(id string, stack []string) {
		if idx := indexOf(stack, id); idx >= 0 {
			cycle := append([]string{}, stack[idx:]...)
			cycle = append(cycle, id)
			key := canonicalCycleKey(cycle)
			if !seenCycles[key] {
				seenCycles[key] = true
				cycles = append(cycles, cycle)
			}
			return
		}
		if visited[id] {
			return
		}
		t, ok := byID[id]
		if !ok || !ticket.IsWorkStatus(t.Status) {
			return
		}
		visited[id] = true
		nextStack := append(stack, id)
		for _, dep := range dependencyTickets(t, byID) {
			if ticket.IsWorkStatus(dep.Status) {
				visit(dep.ID, nextStack)
			}
		}
	}
	for _, id := range ids {
		visit(id, nil)
	}
	return cycles
}

func indexOf(values []string, value string) int {
	for i, candidate := range values {
		if candidate == value {
			return i
		}
	}
	return -1
}

func canonicalCycleKey(cycle []string) string {
	if len(cycle) <= 1 {
		return strings.Join(cycle, "->")
	}
	nodes := append([]string{}, cycle[:len(cycle)-1]...)
	if len(nodes) == 0 {
		return ""
	}
	minIdx := 0
	for i := range nodes {
		if nodes[i] < nodes[minIdx] {
			minIdx = i
		}
	}
	rotated := append([]string{}, nodes[minIdx:]...)
	rotated = append(rotated, nodes[:minIdx]...)
	return strings.Join(rotated, "->")
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
