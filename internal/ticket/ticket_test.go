package ticket

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestParseReadsUpstreamFixture(t *testing.T) {
	repo := filepath.Join("..", "..", "testdata", "fixtures", "upstream-basic")
	root, err := Discover(repo, noEnv)
	if err != nil {
		t.Fatalf("Discover returned error: %v", err)
	}

	ticket, err := ParseFile(root, filepath.Join(root.TicketsDir, "ub-0hf0.md"))
	if err != nil {
		t.Fatalf("ParseFile returned error: %v", err)
	}
	if ticket.ID != "ub-0hf0" || ticket.Parent != "ub-4vbh" || ticket.Title != "Upstream child ticket" {
		t.Fatalf("parsed ticket = %#v", ticket)
	}
	if len(ticket.Deps) != 1 || ticket.Deps[0] != "ub-05xk" {
		t.Fatalf("Deps = %#v", ticket.Deps)
	}
}

func TestListSkipsMalformedWithWarning(t *testing.T) {
	repo := filepath.Join("..", "..", "testdata", "fixtures", "edge-cases")
	root, err := Discover(repo, noEnv)
	if err != nil {
		t.Fatalf("Discover returned error: %v", err)
	}

	tickets, warnings := List(root)
	if len(tickets) != 2 {
		t.Fatalf("len(tickets) = %d, want 2", len(tickets))
	}
	if len(warnings) != 1 {
		t.Fatalf("len(warnings) = %d, want 1", len(warnings))
	}
	if !strings.Contains(warnings[0].Error(), "malformed inline list") {
		t.Fatalf("warning = %q", warnings[0].Error())
	}
}

func TestParseAcceptsCRLF(t *testing.T) {
	root := Root{ProjectDir: t.TempDir()}
	root.TicketsDir = filepath.Join(root.ProjectDir, TicketsDirName)
	mustMkdir(t, root.TicketsDir)
	content := "---\r\nid: gt-crlf\r\nstatus: open\r\ndeps: []\r\nlinks: []\r\ncreated: 2026-05-28T00:00:00Z\r\ntype: task\r\npriority: 2\r\n---\r\n# CRLF\r\n"

	ticket, err := Parse(root, filepath.Join(root.TicketsDir, "gt-crlf.md"), content)
	if err != nil {
		t.Fatalf("Parse returned error: %v", err)
	}
	if ticket.Title != "CRLF" {
		t.Fatalf("Title = %q, want CRLF", ticket.Title)
	}
}

func TestResolveIsCaseInsensitiveAndAmbiguous(t *testing.T) {
	root := Root{ProjectDir: t.TempDir()}
	root.TicketsDir = filepath.Join(root.ProjectDir, TicketsDirName)
	mustMkdir(t, root.TicketsDir)
	for _, id := range []string{"gt-Abcd", "gt-abef"} {
		ticket := Ticket{
			ID:       id,
			Status:   "open",
			Deps:     []string{},
			Links:    []string{},
			Created:  "2026-05-28T00:00:00Z",
			Type:     "task",
			Priority: "2",
			Title:    id,
			Body:     "# " + id + "\n",
		}
		if err := Write(root, ticket); err != nil {
			t.Fatalf("Write(%s) returned error: %v", id, err)
		}
	}

	_, err := Resolve(root, "GT-AB")
	if !errors.Is(err, ErrAmbiguousID) {
		t.Fatalf("Resolve error = %v, want ErrAmbiguousID", err)
	}
}

func TestWritePreservesUnknownFieldsAndNormalizesLF(t *testing.T) {
	root := Root{ProjectDir: t.TempDir()}
	root.TicketsDir = filepath.Join(root.ProjectDir, TicketsDirName)
	mustMkdir(t, root.TicketsDir)
	ticket := Ticket{
		ID:       "gt-test",
		Status:   "open",
		Deps:     []string{},
		Links:    []string{},
		Created:  "2026-05-28T00:00:00Z",
		Type:     "task",
		Priority: "2",
		Unknown:  []Field{{Key: "custom-field", Value: "retained"}},
		Title:    "Write test",
		Body:     "# Write test\r\n\r\nBody\r\n",
	}

	if err := Write(root, ticket); err != nil {
		t.Fatalf("Write returned error: %v", err)
	}
	data, err := os.ReadFile(filepath.Join(root.TicketsDir, "gt-test.md"))
	if err != nil {
		t.Fatalf("read written ticket: %v", err)
	}
	content := string(data)
	if strings.Contains(content, "\r\n") {
		t.Fatalf("content contains CRLF: %q", content)
	}
	if !strings.Contains(content, "custom-field: retained") {
		t.Fatalf("content did not preserve custom field:\n%s", content)
	}
}

func TestWriteRejectsFrontmatterNewlineInjection(t *testing.T) {
	root := Root{ProjectDir: t.TempDir()}
	root.TicketsDir = filepath.Join(root.ProjectDir, TicketsDirName)
	mustMkdir(t, root.TicketsDir)
	ticket := Ticket{
		ID:       "gt-inject",
		Status:   "open",
		Deps:     []string{},
		Links:    []string{},
		Created:  "2026-05-28T00:00:00Z",
		Type:     "task",
		Priority: "2",
		Assignee: "codex\nstatus: closed",
		Title:    "Injection",
		Body:     "# Injection\n",
	}

	err := Write(root, ticket)
	if err == nil {
		t.Fatal("Write succeeded with newline injection in frontmatter")
	}
	if !strings.Contains(err.Error(), "contains a newline") {
		t.Fatalf("error = %q, want newline message", err.Error())
	}
	if _, statErr := os.Stat(filepath.Join(root.TicketsDir, "gt-inject.md")); !os.IsNotExist(statErr) {
		t.Fatalf("ticket file exists after rejected write: %v", statErr)
	}
}

func TestResolveRejectsTargetedSymlinkBeforeRead(t *testing.T) {
	root := Root{ProjectDir: t.TempDir()}
	root.TicketsDir = filepath.Join(root.ProjectDir, TicketsDirName)
	mustMkdir(t, root.TicketsDir)
	target := filepath.Join(root.TicketsDir, "target.md")
	if err := os.WriteFile(target, []byte("---\nid: target\nstatus: open\ndeps: []\nlinks: []\n---\n# Target\n"), 0o644); err != nil {
		t.Fatalf("write target: %v", err)
	}
	link := filepath.Join(root.TicketsDir, "gt-link.md")
	if err := os.Symlink(target, link); err != nil {
		t.Skipf("symlink unavailable: %v", err)
	}

	_, err := Resolve(root, "gt-link")
	if err == nil {
		t.Fatal("Resolve succeeded for symlinked targeted ticket")
	}
	if !strings.Contains(err.Error(), "symlink") {
		t.Fatalf("error = %q, want symlink message", err.Error())
	}
}

func TestReadRawFileRejectsOversizedTicket(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "gt-large.md")
	content := strings.Repeat("x", MaxTicketFileBytes+1)
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write large ticket: %v", err)
	}

	_, err := ReadRawFile(path)
	if err == nil {
		t.Fatal("ReadRawFile succeeded for oversized ticket")
	}
	if !strings.Contains(err.Error(), "exceeds") {
		t.Fatalf("error = %q, want exceeds message", err.Error())
	}
}
