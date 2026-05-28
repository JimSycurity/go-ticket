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
