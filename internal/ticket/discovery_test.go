package ticket

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func noEnv(string) string { return "" }

func TestDiscoverFindsAncestorTicketsDir(t *testing.T) {
	root := t.TempDir()
	mustMkdir(t, filepath.Join(root, TicketsDirName))
	nested := filepath.Join(root, "a", "b")
	mustMkdir(t, nested)

	discovered, err := Discover(nested, noEnv)
	if err != nil {
		t.Fatalf("Discover returned error: %v", err)
	}
	if discovered.ProjectDir != root {
		t.Fatalf("ProjectDir = %q, want %q", discovered.ProjectDir, root)
	}
	if discovered.TicketsDir != filepath.Join(root, TicketsDirName) {
		t.Fatalf("TicketsDir = %q", discovered.TicketsDir)
	}
	if discovered.Source != RootSourceDiscovered {
		t.Fatalf("Source = %q, want %q", discovered.Source, RootSourceDiscovered)
	}
}

func TestDiscoverFromTicketsDir(t *testing.T) {
	root := t.TempDir()
	tickets := filepath.Join(root, TicketsDirName)
	mustMkdir(t, tickets)

	discovered, err := Discover(tickets, noEnv)
	if err != nil {
		t.Fatalf("Discover returned error: %v", err)
	}
	if discovered.ProjectDir != root {
		t.Fatalf("ProjectDir = %q, want %q", discovered.ProjectDir, root)
	}
	if discovered.TicketsDir != tickets {
		t.Fatalf("TicketsDir = %q, want %q", discovered.TicketsDir, tickets)
	}
	if discovered.Source != RootSourceDiscovered {
		t.Fatalf("Source = %q, want %q", discovered.Source, RootSourceDiscovered)
	}
}

func TestDiscoverUsesAbsoluteTicketsDirOverride(t *testing.T) {
	root := t.TempDir()
	tickets := filepath.Join(root, TicketsDirName)
	mustMkdir(t, tickets)
	other := t.TempDir()

	discovered, err := Discover(other, func(key string) string {
		if key == "TICKETS_DIR" {
			return tickets
		}
		return ""
	})
	if err != nil {
		t.Fatalf("Discover returned error: %v", err)
	}
	if discovered.TicketsDir != tickets {
		t.Fatalf("TicketsDir = %q, want %q", discovered.TicketsDir, tickets)
	}
	if discovered.Source != RootSourceEnv {
		t.Fatalf("Source = %q, want %q", discovered.Source, RootSourceEnv)
	}
}

func TestDiscoverRejectsRelativeTicketsDirOverride(t *testing.T) {
	_, err := Discover(t.TempDir(), func(key string) string {
		if key == "TICKETS_DIR" {
			return ".tickets"
		}
		return ""
	})
	if err == nil {
		t.Fatal("Discover succeeded for relative TICKETS_DIR")
	}
	if !strings.Contains(err.Error(), "absolute") {
		t.Fatalf("error = %q, want absolute path message", err.Error())
	}
}

func TestDiscoverInvalidTicketsDirDoesNotFallback(t *testing.T) {
	root := t.TempDir()
	mustMkdir(t, filepath.Join(root, TicketsDirName))

	_, err := Discover(root, func(key string) string {
		if key == "TICKETS_DIR" {
			return filepath.Join(root, "missing")
		}
		return ""
	})
	if err == nil {
		t.Fatal("Discover succeeded for missing TICKETS_DIR")
	}
	if !strings.Contains(err.Error(), "TICKETS_DIR") {
		t.Fatalf("error = %q, want TICKETS_DIR message", err.Error())
	}
}

func TestDiscoverRejectsSymlinkedAncestorTicketsDir(t *testing.T) {
	root := t.TempDir()
	target := filepath.Join(root, "target")
	mustMkdir(t, target)
	if err := os.Symlink(target, filepath.Join(root, TicketsDirName)); err != nil {
		t.Skipf("symlink unavailable: %v", err)
	}

	_, err := Discover(root, noEnv)
	if err == nil {
		t.Fatal("Discover succeeded for symlinked .tickets")
	}
	if !strings.Contains(err.Error(), "path-security review") {
		t.Fatalf("error = %q, want path-security review message", err.Error())
	}
}

func TestDiscoverRejectsNonDirectoryTicketsDirWithoutFallback(t *testing.T) {
	parent := t.TempDir()
	mustMkdir(t, filepath.Join(parent, TicketsDirName))
	child := filepath.Join(parent, "child")
	mustMkdir(t, child)
	if err := os.WriteFile(filepath.Join(child, TicketsDirName), []byte("not a directory"), 0o644); err != nil {
		t.Fatalf("write blocking .tickets file: %v", err)
	}

	_, err := Discover(child, noEnv)
	if err == nil {
		t.Fatal("Discover succeeded by falling back past a non-directory .tickets")
	}
	if !strings.Contains(err.Error(), "not a directory") {
		t.Fatalf("error = %q, want not a directory message", err.Error())
	}
}

func TestDiscoverRejectsSymlinkedTicketsDirOverride(t *testing.T) {
	root := t.TempDir()
	target := filepath.Join(root, "target")
	mustMkdir(t, target)
	link := filepath.Join(root, "tickets-link")
	if err := os.Symlink(target, link); err != nil {
		t.Skipf("symlink unavailable: %v", err)
	}

	_, err := Discover(root, func(key string) string {
		if key == "TICKETS_DIR" {
			return link
		}
		return ""
	})
	if err == nil {
		t.Fatal("Discover succeeded for symlinked TICKETS_DIR")
	}
	if !strings.Contains(err.Error(), "symlinked TICKETS_DIR") {
		t.Fatalf("error = %q, want symlinked TICKETS_DIR message", err.Error())
	}
}

func TestDiscoverRejectsTicketsDirOverrideWithSymlinkedParent(t *testing.T) {
	root := t.TempDir()
	targetParent := filepath.Join(root, "target-parent")
	mustMkdir(t, filepath.Join(targetParent, TicketsDirName))
	linkParent := filepath.Join(root, "link-parent")
	if err := os.Symlink(targetParent, linkParent); err != nil {
		t.Skipf("symlink unavailable: %v", err)
	}

	_, err := Discover(root, func(key string) string {
		if key == "TICKETS_DIR" {
			return filepath.Join(linkParent, TicketsDirName)
		}
		return ""
	})
	if err == nil {
		t.Fatal("Discover succeeded for TICKETS_DIR with symlinked parent")
	}
	if !strings.Contains(err.Error(), "symlink indirection") {
		t.Fatalf("error = %q, want symlink indirection message", err.Error())
	}
}

func TestDiscoverRejectsSymlinkedTicketsDirWhenCwdIsTicketsDir(t *testing.T) {
	root := t.TempDir()
	target := filepath.Join(root, "target")
	mustMkdir(t, target)
	link := filepath.Join(root, TicketsDirName)
	if err := os.Symlink(target, link); err != nil {
		t.Skipf("symlink unavailable: %v", err)
	}

	_, err := Discover(link, noEnv)
	if err == nil {
		t.Fatal("Discover succeeded when cwd path was symlinked .tickets")
	}
	if !strings.Contains(err.Error(), "symlinked .tickets") {
		t.Fatalf("error = %q, want symlinked .tickets message", err.Error())
	}
}

func mustMkdir(t *testing.T, path string) {
	t.Helper()
	if err := os.MkdirAll(path, 0o755); err != nil {
		t.Fatalf("mkdir %s: %v", path, err)
	}
}
