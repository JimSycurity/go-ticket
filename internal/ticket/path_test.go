package ticket

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestResolveTicketPathAcceptsValidID(t *testing.T) {
	root := Root{TicketsDir: t.TempDir()}

	path, err := ResolveTicketPath(root, "gt-27pw", false)
	if err != nil {
		t.Fatalf("ResolveTicketPath returned error: %v", err)
	}
	want := filepath.Join(root.TicketsDir, "gt-27pw.md")
	if path != want {
		t.Fatalf("path = %q, want %q", path, want)
	}
}

func TestResolveTicketPathAcceptsDottedUpstreamID(t *testing.T) {
	root := Root{TicketsDir: t.TempDir()}

	path, err := ResolveTicketPath(root, "GlobalTech-k78c.1", false)
	if err != nil {
		t.Fatalf("ResolveTicketPath returned error: %v", err)
	}
	want := filepath.Join(root.TicketsDir, "GlobalTech-k78c.1.md")
	if path != want {
		t.Fatalf("path = %q, want %q", path, want)
	}
}

func TestResolveTicketPathRejectsUnsafeIDs(t *testing.T) {
	root := Root{TicketsDir: t.TempDir()}
	unsafeIDs := []string{
		"",
		".",
		"..",
		".gt-27pw",
		"gt-27pw.",
		"gt..27pw",
		"../gt-27pw",
		"nested/gt-27pw",
		`nested\gt-27pw`,
		"C:gt-27pw",
		"CON",
		"CON.1",
		"lpt1",
	}

	for _, id := range unsafeIDs {
		t.Run(id, func(t *testing.T) {
			_, err := ResolveTicketPath(root, id, false)
			if err == nil {
				t.Fatal("ResolveTicketPath succeeded for unsafe ID")
			}
		})
	}
}

func TestResolveTicketPathRejectsExistingNonRegularPath(t *testing.T) {
	root := Root{TicketsDir: t.TempDir()}
	if err := os.Mkdir(filepath.Join(root.TicketsDir, "gt-dir.md"), 0o755); err != nil {
		t.Fatalf("mkdir ticket-shaped directory: %v", err)
	}

	_, err := ResolveTicketPath(root, "gt-dir", false)
	if err == nil {
		t.Fatal("ResolveTicketPath succeeded for non-regular ticket path")
	}
	if !strings.Contains(err.Error(), "not a regular file") {
		t.Fatalf("error = %q, want not a regular file message", err.Error())
	}
}

func TestResolveTicketPathRejectsExistingSymlink(t *testing.T) {
	root := Root{TicketsDir: t.TempDir()}
	target := filepath.Join(root.TicketsDir, "target.md")
	if err := os.WriteFile(target, []byte("---\nid: target\n---\n"), 0o644); err != nil {
		t.Fatalf("write symlink target: %v", err)
	}
	link := filepath.Join(root.TicketsDir, "gt-link.md")
	if err := os.Symlink(target, link); err != nil {
		t.Skipf("symlink unavailable: %v", err)
	}

	_, err := ResolveTicketPath(root, "gt-link", true)
	if err == nil {
		t.Fatal("ResolveTicketPath succeeded for symlinked ticket path")
	}
	if !strings.Contains(err.Error(), "symlink") {
		t.Fatalf("error = %q, want symlink message", err.Error())
	}
}

func TestNestedSymlinkDirectoryIsNotTraversed(t *testing.T) {
	root := Root{TicketsDir: t.TempDir()}
	external := t.TempDir()
	if err := os.WriteFile(filepath.Join(external, "gt-escape.md"), []byte("---\nid: gt-escape\n---\n# Escape\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(external, filepath.Join(root.TicketsDir, "linked")); err != nil {
		t.Skipf("symlink unavailable: %v", err)
	}
	if _, err := Resolve(root, "gt-escape"); !errors.Is(err, ErrMissingID) {
		t.Fatalf("Resolve error = %v, want missing ID", err)
	}
	if tickets, warnings := List(root); len(tickets) != 0 || len(warnings) != 0 {
		t.Fatalf("tickets = %#v, warnings = %#v", tickets, warnings)
	}
}
