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
	for _, want := range []string{"gtk - Go ticket MVP CLI", "create", "list, ls", "add-note", "version"} {
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

func TestVersionShowsBuildMetadata(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := Run([]string{"version"}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("Run returned %d, want 0", code)
	}
	output := stdout.String()
	for _, want := range []string{"gtk version:", "commit:", "dirty:", "vcs_time:", "build_date:", "binary:", "binary_mtime:"} {
		if !strings.Contains(output, want) {
			t.Fatalf("version missing %q in:\n%s", want, output)
		}
	}
	if stderr.Len() != 0 {
		t.Fatalf("stderr = %q, want empty", stderr.String())
	}
}

func TestDashDashVersionShowsBuildMetadata(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := Run([]string{"--version"}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("Run returned %d, want 0", code)
	}
	if !strings.Contains(stdout.String(), "gtk version:") {
		t.Fatalf("stdout = %q, want version output", stdout.String())
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

func TestGoldenOutputForUpstreamBasicList(t *testing.T) {
	assertGoldenCommand(t, "upstream-basic", "list.txt", nil, "list")
}

func TestGoldenOutputForUpstreamBasicListJSON(t *testing.T) {
	assertGoldenCommand(t, "upstream-basic", "list.json", normalizeFixturePaths, "list", "--json")
}

func TestGoldenOutputForUpstreamRelationshipsReadyBlocked(t *testing.T) {
	assertGoldenCommand(t, "upstream-relationships", "ready.txt", nil, "ready")
	assertGoldenCommand(t, "upstream-relationships", "blocked.txt", nil, "blocked")
}

func TestAddNoteRejectsOversizedStdin(t *testing.T) {
	chdir(t, t.TempDir())
	mustRun(t, "init")
	id := strings.TrimSpace(mustRun(t, "create", "Big note"))
	bigNote := strings.NewReader(strings.Repeat("x", (64<<10)+1))

	_, stderr, code := run(bigNote, "add-note", id)
	if code == 0 {
		t.Fatal("add-note succeeded for oversized stdin")
	}
	if !strings.Contains(stderr, "note exceeds") {
		t.Fatalf("stderr = %q, want note exceeds", stderr)
	}
}

func assertGoldenCommand(t *testing.T, fixtureName string, goldenName string, normalize func(string, string) string, args ...string) {
	t.Helper()
	root := repoRoot(t)
	fixtureDir := filepath.Join(root, "testdata", "fixtures", fixtureName)
	goldenPath := filepath.Join(root, "testdata", "golden", fixtureName, goldenName)
	absFixture, err := filepath.Abs(fixtureDir)
	if err != nil {
		t.Fatalf("resolve fixture path: %v", err)
	}
	chdir(t, fixtureDir)

	stdout, stderr, code := run(nil, args...)
	if code != 0 {
		t.Fatalf("%v returned %d stderr=%q", args, code, stderr)
	}
	if stderr != "" {
		t.Fatalf("%v stderr = %q, want empty", args, stderr)
	}
	wantBytes, err := os.ReadFile(goldenPath)
	if err != nil {
		t.Fatalf("read golden %s: %v", goldenPath, err)
	}
	got := stdout
	want := string(wantBytes)
	if normalize != nil {
		got = normalize(absFixture, got)
		want = normalize(absFixture, want)
	}
	if got != want {
		t.Fatalf("%v output mismatch\nwant:\n%s\ngot:\n%s", args, want, got)
	}
}

func normalizeFixturePaths(fixtureDir string, value string) string {
	normalized := strings.ReplaceAll(value, "\\", "/")
	fixture := strings.ReplaceAll(fixtureDir, "\\", "/")
	return strings.ReplaceAll(normalized, fixture, "<FIXTURE>")
}

func repoRoot(t *testing.T) string {
	t.Helper()
	dir, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatalf("could not find repo root from %s", dir)
		}
		dir = parent
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
