package app

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestHelpOnlyShowsMVPCommands(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := Run([]string{"--help"}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("Run returned %d, want 0", code)
	}
	output := stdout.String()
	for _, want := range []string{"gtk - Go ticket MVP CLI", "create", "list, ls", "add-note"} {
		if !strings.Contains(output, want) {
			t.Fatalf("help missing %q in:\n%s", want, output)
		}
	}
	for _, unexpected := range []string{"query", "edit", "migrate-beads", "super"} {
		if strings.Contains(output, unexpected) {
			t.Fatalf("help unexpectedly includes %q in:\n%s", unexpected, output)
		}
	}
	if stderr.Len() != 0 {
		t.Fatalf("stderr = %q, want empty", stderr.String())
	}
}

func TestUnknownCommandDoesNotDispatchPlugin(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := Run([]string{"made-up-plugin"}, &stdout, &stderr)
	if code == 0 {
		t.Fatal("Run returned success for unsupported command")
	}
	if stdout.Len() != 0 {
		t.Fatalf("stdout = %q, want empty", stdout.String())
	}
	if !strings.Contains(stderr.String(), "unsupported command") {
		t.Fatalf("stderr = %q, want unsupported command error", stderr.String())
	}
}

func TestCreateListShowAndLifecycleWorkflow(t *testing.T) {
	chdir(t, t.TempDir())

	stdout, stderr, code := run(nil, "init")
	if code != 0 {
		t.Fatalf("init returned %d stderr=%q", code, stderr)
	}
	if !strings.Contains(stdout, "Initialized ticket directory") {
		t.Fatalf("init stdout = %q", stdout)
	}

	stdout, stderr, code = run(nil, "create", "--type", "feature", "--priority", "1", "--tags", "mvp,cli", "Add Windows support")
	if code != 0 {
		t.Fatalf("create returned %d stderr=%q", code, stderr)
	}
	id := strings.TrimSpace(stdout)
	if id == "" {
		t.Fatal("create did not print an ID")
	}

	stdout, stderr, code = run(nil, "list", "--type", "feature")
	if code != 0 {
		t.Fatalf("list returned %d stderr=%q", code, stderr)
	}
	if !strings.Contains(stdout, id) || !strings.Contains(stdout, "Add Windows support") {
		t.Fatalf("list stdout = %q", stdout)
	}

	stdout, stderr, code = run(nil, "show", id[:4])
	if code != 0 {
		t.Fatalf("show returned %d stderr=%q", code, stderr)
	}
	if !strings.Contains(stdout, "# Add Windows support") {
		t.Fatalf("show stdout = %q", stdout)
	}

	_, stderr, code = run(nil, "start", id)
	if code != 0 {
		t.Fatalf("start returned %d stderr=%q", code, stderr)
	}
	stdout, stderr, code = run(nil, "list", "--status", "in_progress")
	if code != 0 {
		t.Fatalf("list status returned %d stderr=%q", code, stderr)
	}
	if !strings.Contains(stdout, "[in_progress]") {
		t.Fatalf("list status stdout = %q", stdout)
	}
}

func TestRelationshipsReadyBlockedAndNotes(t *testing.T) {
	chdir(t, t.TempDir())
	mustRun(t, "init")
	parent := strings.TrimSpace(mustRun(t, "create", "Parent"))
	child := strings.TrimSpace(mustRun(t, "create", "Child"))

	if _, stderr, code := run(nil, "dep", child, parent); code != 0 {
		t.Fatalf("dep returned %d stderr=%q", code, stderr)
	}
	if _, stderr, code := run(nil, "link", child, parent); code != 0 {
		t.Fatalf("link returned %d stderr=%q", code, stderr)
	}
	stdout, stderr, code := run(nil, "blocked")
	if code != 0 {
		t.Fatalf("blocked returned %d stderr=%q", code, stderr)
	}
	if !strings.Contains(stdout, child) || !strings.Contains(stdout, parent) {
		t.Fatalf("blocked stdout = %q", stdout)
	}

	mustRun(t, "close", parent)
	stdout, stderr, code = run(nil, "ready")
	if code != 0 {
		t.Fatalf("ready returned %d stderr=%q", code, stderr)
	}
	if !strings.Contains(stdout, child) {
		t.Fatalf("ready stdout = %q", stdout)
	}

	if _, stderr, code := run(strings.NewReader("note from stdin"), "add-note", child); code != 0 {
		t.Fatalf("add-note returned %d stderr=%q", code, stderr)
	}
	data, err := os.ReadFile(filepath.Join(".tickets", child+".md"))
	if err != nil {
		t.Fatalf("read child ticket: %v", err)
	}
	if !strings.Contains(string(data), "note from stdin") {
		t.Fatalf("note not written:\n%s", string(data))
	}
}

func TestListWarnsAndContinuesForMalformedTicket(t *testing.T) {
	dir := t.TempDir()
	chdir(t, dir)
	mustRun(t, "init")
	mustRun(t, "create", "Good")
	if err := os.WriteFile(filepath.Join(dir, ".tickets", "bad.md"), []byte("---\nid: bad\ndeps: [\n---\n# Bad\n"), 0o644); err != nil {
		t.Fatalf("write malformed ticket: %v", err)
	}

	stdout, stderr, code := run(nil, "list")
	if code != 0 {
		t.Fatalf("list returned %d stderr=%q", code, stderr)
	}
	if !strings.Contains(stdout, "Good") {
		t.Fatalf("list stdout = %q", stdout)
	}
	if !strings.Contains(stderr, "warning:") {
		t.Fatalf("stderr = %q, want warning", stderr)
	}
}

func run(stdin *strings.Reader, args ...string) (string, string, int) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	var input *strings.Reader
	if stdin == nil {
		input = strings.NewReader("")
	} else {
		input = stdin
	}
	code := RunWithIO(args, input, &stdout, &stderr)
	return stdout.String(), stderr.String(), code
}

func mustRun(t *testing.T, args ...string) string {
	t.Helper()
	stdout, stderr, code := run(nil, args...)
	if code != 0 {
		t.Fatalf("%v returned %d stderr=%q", args, code, stderr)
	}
	return stdout
}

func chdir(t *testing.T, dir string) {
	t.Helper()
	old, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("chdir %s: %v", dir, err)
	}
	t.Cleanup(func() {
		if err := os.Chdir(old); err != nil {
			t.Fatalf("restore cwd: %v", err)
		}
	})
}
