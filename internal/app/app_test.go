package app

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"
)

func TestHelpOnlyShowsMVPCommands(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := Run([]string{"--help"}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("Run returned %d, want 0", code)
	}
	output := stdout.String()
	for _, want := range []string{"gtk - Go ticket CLI", "create", "edit", "list, ls", "query", "add-note", "migrate-beads", "super", "version"} {
		if !strings.Contains(output, want) {
			t.Fatalf("help missing %q in:\n%s", want, output)
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

func TestUnknownCommandWithoutPluginFails(t *testing.T) {
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

func TestPluginExecutesUnknownCommandWithMinimalTicketEnv(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("test uses a POSIX shebang plugin")
	}
	dir := t.TempDir()
	chdir(t, dir)
	mustRun(t, "init")

	pluginDir := filepath.Join(dir, "plugins")
	if err := os.Mkdir(pluginDir, 0o755); err != nil {
		t.Fatalf("mkdir plugin dir: %v", err)
	}
	writeExecutable(t, filepath.Join(pluginDir, "tk-hello"), "#!/bin/sh\nprintf 'plugin:%s:%s\\n' \"$1\" \"$TICKETS_DIR\"\nprintf 'project:%s\\n' \"$TICKET_PROJECT_DIR\" >&2\n")
	t.Setenv("PATH", pluginDir)

	stdout, stderr, code := run(nil, "hello", "world")
	if code != 0 {
		t.Fatalf("plugin returned %d stderr=%q", code, stderr)
	}
	wantTicketsDir := canonicalTestPath(t, filepath.Join(dir, ".tickets"))
	wantProjectDir := canonicalTestPath(t, dir)
	if !strings.Contains(stdout, "plugin:world:"+wantTicketsDir) {
		t.Fatalf("stdout = %q, want plugin output with TICKETS_DIR", stdout)
	}
	if !strings.Contains(stderr, "project:"+wantProjectDir) {
		t.Fatalf("stderr = %q, want project env", stderr)
	}
}

func TestPluginLookupIgnoresRelativePathEntries(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("test uses POSIX shebang plugins")
	}
	dir := t.TempDir()
	chdir(t, dir)
	mustRun(t, "init")
	writeExecutable(t, filepath.Join(dir, "tk-hello"), "#!/bin/sh\nprintf 'relative\\n'\n")

	pluginDir := filepath.Join(dir, "absolute-plugins")
	if err := os.Mkdir(pluginDir, 0o755); err != nil {
		t.Fatalf("mkdir plugin dir: %v", err)
	}
	writeExecutable(t, filepath.Join(pluginDir, "tk-hello"), "#!/bin/sh\nprintf 'absolute\\n'\n")
	t.Setenv("PATH", "."+string(os.PathListSeparator)+pluginDir)

	stdout, stderr, code := run(nil, "hello")
	if code != 0 {
		t.Fatalf("plugin returned %d stderr=%q", code, stderr)
	}
	if stdout != "absolute\n" {
		t.Fatalf("stdout = %q, want absolute plugin", stdout)
	}
}

func TestSuperBypassesPluginsAndRunsBuiltinsOnly(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("test uses a POSIX shebang plugin")
	}
	dir := t.TempDir()
	chdir(t, dir)
	mustRun(t, "init")
	id := strings.TrimSpace(mustRun(t, "create", "Builtin ticket"))

	pluginDir := filepath.Join(dir, "plugins")
	if err := os.Mkdir(pluginDir, 0o755); err != nil {
		t.Fatalf("mkdir plugin dir: %v", err)
	}
	writeExecutable(t, filepath.Join(pluginDir, "tk-list"), "#!/bin/sh\nprintf 'plugin-list\\n'\n")
	writeExecutable(t, filepath.Join(pluginDir, "tk-hello"), "#!/bin/sh\nprintf 'plugin-hello\\n'\n")
	t.Setenv("PATH", pluginDir)

	stdout, stderr, code := run(nil, "super", "list")
	if code != 0 {
		t.Fatalf("super list returned %d stderr=%q", code, stderr)
	}
	if !strings.Contains(stdout, id) || strings.Contains(stdout, "plugin-list") {
		t.Fatalf("super list stdout = %q", stdout)
	}

	stdout, stderr, code = run(nil, "super", "hello")
	if code == 0 {
		t.Fatalf("super hello succeeded stdout=%q", stdout)
	}
	if !strings.Contains(stderr, "unsupported builtin command") {
		t.Fatalf("super hello stderr = %q, want unsupported builtin", stderr)
	}
}

func TestPluginResolutionPolicy(t *testing.T) {
	if got := pluginExtensions("windows"); len(got) != 1 || got[0] != ".exe" {
		t.Fatalf("windows plugin extensions = %v, want .exe only", got)
	}
	if err := validatePluginCommand("../hello"); err == nil {
		t.Fatal("validatePluginCommand accepted path traversal")
	}
	if err := validatePluginCommand("hello.world"); err == nil {
		t.Fatal("validatePluginCommand accepted dotted command")
	}
	if err := validatePluginCommand("hello;world"); err == nil {
		t.Fatal("validatePluginCommand accepted shell metacharacters")
	}
}

func TestEditUsesValidatedEditorAndPassesTicketPathAsSingleArg(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("test uses a POSIX shebang editor")
	}
	dir := filepath.Join(t.TempDir(), "repo with spaces")
	if err := os.Mkdir(dir, 0o755); err != nil {
		t.Fatalf("mkdir spaced repo: %v", err)
	}
	chdir(t, dir)
	mustRun(t, "init")
	id := strings.TrimSpace(mustRun(t, "create", "Editable ticket"))

	editorPath := filepath.Join(t.TempDir(), "editor")
	markerPath := filepath.Join(t.TempDir(), "opened.txt")
	writeExecutable(t, editorPath, "#!/bin/sh\nprintf '%s\\n%s\\n' \"$#\" \"$1\" > "+shellQuote(markerPath)+"\n")
	t.Setenv("GTK_EDITOR", editorPath)

	stdout, stderr, code := run(nil, "edit", id)
	if code != 0 {
		t.Fatalf("edit returned %d stdout=%q stderr=%q", code, stdout, stderr)
	}
	data, err := os.ReadFile(markerPath)
	if err != nil {
		t.Fatalf("read editor marker: %v", err)
	}
	wantPath := canonicalTestPath(t, filepath.Join(dir, ".tickets", id+".md"))
	if string(data) != "1\n"+wantPath+"\n" {
		t.Fatalf("editor marker = %q, want one arg path %q", string(data), wantPath)
	}
}

func TestEditRejectsUnsafeOrMissingEditor(t *testing.T) {
	chdir(t, t.TempDir())
	mustRun(t, "init")
	id := strings.TrimSpace(mustRun(t, "create", "Editable ticket"))

	t.Setenv("GTK_EDITOR", "code -w")
	_, stderr, code := run(nil, "edit", id)
	if code == 0 {
		t.Fatal("edit succeeded with inline editor arguments")
	}
	if !strings.Contains(stderr, "inline arguments") {
		t.Fatalf("stderr = %q, want inline arguments error", stderr)
	}

	t.Setenv("GTK_EDITOR", "")
	t.Setenv("VISUAL", "")
	t.Setenv("EDITOR", "")
	_, stderr, code = run(nil, "edit", id)
	if code == 0 {
		t.Fatal("edit succeeded without editor")
	}
	if !strings.Contains(stderr, "no editor configured") {
		t.Fatalf("stderr = %q, want no editor configured", stderr)
	}
}

func TestMigrateBeadsImportsReportsConflictsAndSupportsDryRun(t *testing.T) {
	dir := t.TempDir()
	chdir(t, dir)
	mustRun(t, "init")
	beadsDir := filepath.Join(dir, ".beads")
	if err := os.Mkdir(beadsDir, 0o755); err != nil {
		t.Fatalf("mkdir beads dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, ".tickets", "bd-existing.md"), []byte("---\nid: bd-existing\nstatus: open\ndeps: []\nlinks: []\ncreated: 2026-05-29T00:00:00Z\ntype: task\npriority: 2\n---\n# Existing\n"), 0o644); err != nil {
		t.Fatalf("write existing ticket: %v", err)
	}
	source := strings.Join([]string{
		`{"id":"bd-one","title":"Import one","description":"Body text","status":"in-progress","type":"bug","priority":1,"assignee":"Jim","deps":["bd-existing"],"tags":["beads"]}`,
		`{"id":"bd-existing","title":"Conflict"}`,
		`{"id":"../bad","title":"Bad ID"}`,
		`not json`,
		`{"id":"bd-partial"}`,
		"",
	}, "\n")
	if err := os.WriteFile(filepath.Join(beadsDir, "issues.jsonl"), []byte(source), 0o644); err != nil {
		t.Fatalf("write beads source: %v", err)
	}

	stdout, stderr, code := run(nil, "migrate-beads", "--dry-run")
	if code != 0 {
		t.Fatalf("migrate-beads dry-run returned %d stderr=%q", code, stderr)
	}
	if !strings.Contains(stdout, "would import bd-one") || !strings.Contains(stdout, "skip bd-existing") || !strings.Contains(stdout, "review line 3") || !strings.Contains(stdout, "review line 4") || !strings.Contains(stdout, "review line 5") {
		t.Fatalf("dry-run stdout missing report details:\n%s", stdout)
	}
	if _, err := os.Stat(filepath.Join(dir, ".tickets", "bd-one.md")); !os.IsNotExist(err) {
		t.Fatalf("dry-run created bd-one.md, stat err=%v", err)
	}

	stdout, stderr, code = run(nil, "migrate-beads")
	if code != 0 {
		t.Fatalf("migrate-beads returned %d stderr=%q", code, stderr)
	}
	if !strings.Contains(stdout, "imported bd-one") || !strings.Contains(stdout, "imported=1 skipped=1 review=3") {
		t.Fatalf("stdout missing import summary:\n%s", stdout)
	}
	data, err := os.ReadFile(filepath.Join(dir, ".tickets", "bd-one.md"))
	if err != nil {
		t.Fatalf("read imported ticket: %v", err)
	}
	content := string(data)
	for _, want := range []string{"id: bd-one", "status: in_progress", "type: bug", "priority: 1", "assignee: Jim", "external-ref: beads:bd-one", "deps: [bd-existing]", "tags: [beads, beads-import]", "# Import one", "Body text"} {
		if !strings.Contains(content, want) {
			t.Fatalf("imported ticket missing %q:\n%s", want, content)
		}
	}
}

func TestMigrateBeadsRejectsUnsafeSource(t *testing.T) {
	dir := t.TempDir()
	chdir(t, dir)
	mustRun(t, "init")
	outside := filepath.Join(t.TempDir(), "issues.jsonl")
	if err := os.WriteFile(outside, []byte(`{"id":"bd-one","title":"Outside"}`+"\n"), 0o644); err != nil {
		t.Fatalf("write outside source: %v", err)
	}

	_, stderr, code := run(nil, "migrate-beads", "--source", outside)
	if code == 0 {
		t.Fatal("migrate-beads accepted source outside project")
	}
	if !strings.Contains(stderr, "inside project root") {
		t.Fatalf("stderr = %q, want containment error", stderr)
	}
}

func TestMigrateBeadsRejectsSourceSymlinkIndirection(t *testing.T) {
	dir := t.TempDir()
	chdir(t, dir)
	mustRun(t, "init")
	targetDir := filepath.Join(dir, "target")
	if err := os.Mkdir(targetDir, 0o755); err != nil {
		t.Fatalf("mkdir target: %v", err)
	}
	if err := os.WriteFile(filepath.Join(targetDir, "issues.jsonl"), []byte(`{"id":"bd-one","title":"One"}`+"\n"), 0o644); err != nil {
		t.Fatalf("write target source: %v", err)
	}
	if err := os.Symlink(targetDir, filepath.Join(dir, ".beads")); err != nil {
		t.Skipf("symlink unavailable: %v", err)
	}

	_, stderr, code := run(nil, "migrate-beads")
	if code == 0 {
		t.Fatal("migrate-beads accepted source with symlinked parent")
	}
	if !strings.Contains(stderr, "symlink indirection") {
		t.Fatalf("stderr = %q, want symlink indirection error", stderr)
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

func TestLinkPreflightsBothTicketsBeforeWriting(t *testing.T) {
	chdir(t, t.TempDir())
	mustRun(t, "init")
	left := strings.TrimSpace(mustRun(t, "create", "Left"))
	right := strings.TrimSpace(mustRun(t, "create", "Right"))
	rightPath := filepath.Join(".tickets", right+".md")
	data, err := os.ReadFile(rightPath)
	if err != nil {
		t.Fatalf("read right ticket: %v", err)
	}
	content := strings.Replace(string(data), "priority: 2\n", "priority: 2\nunsafe key: value\n", 1)
	if err := os.WriteFile(rightPath, []byte(content), 0o644); err != nil {
		t.Fatalf("write unsafe right ticket: %v", err)
	}

	_, stderr, code := run(nil, "link", left, right)
	if code == 0 {
		t.Fatal("link succeeded despite unsafe right ticket frontmatter")
	}
	if !strings.Contains(stderr, "invalid frontmatter key") {
		t.Fatalf("stderr = %q, want invalid frontmatter key", stderr)
	}
	leftData, err := os.ReadFile(filepath.Join(".tickets", left+".md"))
	if err != nil {
		t.Fatalf("read left ticket: %v", err)
	}
	if strings.Contains(string(leftData), right) {
		t.Fatalf("left ticket was partially updated after failed link:\n%s", string(leftData))
	}
}

func TestDependencyTreeDedupFullMissingAndCycles(t *testing.T) {
	chdir(t, t.TempDir())
	mustRun(t, "init")
	root := strings.TrimSpace(mustRun(t, "create", "Root"))
	left := strings.TrimSpace(mustRun(t, "create", "Left"))
	right := strings.TrimSpace(mustRun(t, "create", "Right"))
	shared := strings.TrimSpace(mustRun(t, "create", "Shared"))
	missing := strings.TrimSpace(mustRun(t, "create", "Missing dep holder"))

	mustRun(t, "dep", root, left)
	mustRun(t, "dep", root, right)
	mustRun(t, "dep", left, shared)
	mustRun(t, "dep", right, shared)
	addDependencyByHand(t, missing, "gt-missing")

	stdout, stderr, code := run(nil, "dep", "tree", root)
	if code != 0 {
		t.Fatalf("dep tree returned %d stderr=%q", code, stderr)
	}
	if countSubstring(stdout, shared) != 1 {
		t.Fatalf("dep tree stdout = %q, want shared dependency once", stdout)
	}

	stdout, stderr, code = run(nil, "dep", "tree", "--full", root)
	if code != 0 {
		t.Fatalf("dep tree --full returned %d stderr=%q", code, stderr)
	}
	if countSubstring(stdout, shared) != 2 {
		t.Fatalf("dep tree --full stdout = %q, want shared dependency twice", stdout)
	}

	stdout, stderr, code = run(nil, "dep", "tree", missing)
	if code != 0 {
		t.Fatalf("dep tree missing returned %d stderr=%q", code, stderr)
	}
	if strings.Contains(stdout, "gt-missing") {
		t.Fatalf("dep tree stdout = %q, missing dependency should be omitted", stdout)
	}

	a := strings.TrimSpace(mustRun(t, "create", "Cycle A"))
	b := strings.TrimSpace(mustRun(t, "create", "Cycle B"))
	c := strings.TrimSpace(mustRun(t, "create", "Cycle C"))
	mustRun(t, "dep", a, b)
	mustRun(t, "dep", b, c)
	mustRun(t, "dep", c, a)

	stdout, stderr, code = run(nil, "dep", "cycle")
	if code != 0 {
		t.Fatalf("dep cycle returned %d stderr=%q", code, stderr)
	}
	for _, want := range []string{"Cycle 1:", a, b, c} {
		if !strings.Contains(stdout, want) {
			t.Fatalf("dep cycle stdout = %q, want %q", stdout, want)
		}
	}

	mustRun(t, "close", c)
	stdout, stderr, code = run(nil, "dep", "cycle")
	if code != 0 {
		t.Fatalf("dep cycle after close returned %d stderr=%q", code, stderr)
	}
	if !strings.Contains(stdout, "No dependency cycles found") {
		t.Fatalf("dep cycle stdout = %q, want no cycles", stdout)
	}
}

func TestClosedListsRecentlyClosedByMTimeWithLimitAndFilters(t *testing.T) {
	chdir(t, t.TempDir())
	mustRun(t, "init")
	first := strings.TrimSpace(mustRun(t, "create", "--assignee", "Jim", "--type", "task", "First closed"))
	second := strings.TrimSpace(mustRun(t, "create", "--assignee", "Jim", "--type", "bug", "Second closed"))
	third := strings.TrimSpace(mustRun(t, "create", "--assignee", "Other", "--type", "task", "Third closed"))
	open := strings.TrimSpace(mustRun(t, "create", "Still open"))

	mustRun(t, "close", first)
	mustRun(t, "close", second)
	mustRun(t, "close", third)
	setTicketMTime(t, first, time.Unix(100, 0))
	setTicketMTime(t, second, time.Unix(300, 0))
	setTicketMTime(t, third, time.Unix(200, 0))

	stdout, stderr, code := run(nil, "closed", "--limit=2")
	if code != 0 {
		t.Fatalf("closed returned %d stderr=%q", code, stderr)
	}
	want := fmt.Sprintf("%-8s [closed] - Second closed\n%-8s [closed] - Third closed\n", second, third)
	if stdout != want {
		t.Fatalf("closed stdout mismatch\nwant:\n%sgot:\n%s", want, stdout)
	}
	if strings.Contains(stdout, open) {
		t.Fatalf("closed stdout includes open ticket: %q", stdout)
	}

	stdout, stderr, code = run(nil, "closed", "-a", "Jim", "-T", "task")
	if code != 0 {
		t.Fatalf("closed filters returned %d stderr=%q", code, stderr)
	}
	want = fmt.Sprintf("%-8s [closed] - First closed\n", first)
	if stdout != want {
		t.Fatalf("closed filter stdout mismatch\nwant:\n%sgot:\n%s", want, stdout)
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

func TestGoldenOutputForUpstreamBasicQuery(t *testing.T) {
	assertGoldenCommand(t, "upstream-basic", "query.jsonl", nil, "query")
}

func TestGoldenOutputForUpstreamRelationshipsReadyBlocked(t *testing.T) {
	assertGoldenCommand(t, "upstream-relationships", "ready.txt", nil, "ready")
	assertGoldenCommand(t, "upstream-relationships", "blocked.txt", nil, "blocked")
}

func TestUpstreamComprehensiveFixtureExercisesParityCommands(t *testing.T) {
	root := repoRoot(t)
	fixtureDir := filepath.Join(root, "testdata", "fixtures", "upstream-comprehensive")
	chdir(t, fixtureDir)

	stdout, stderr, code := run(nil, "list")
	if code != 0 {
		t.Fatalf("list returned %d stderr=%q", code, stderr)
	}
	for _, want := range []string{
		"tmp-u1ds [P0][open] - Upstream comprehensive epic",
		"tmp-2hq1 [P1][open] - Upstream comprehensive feature",
		"tmp-9n81 [P3][in_progress] - Upstream comprehensive task",
		"tmp-kkz9 [P4][closed] - Upstream comprehensive chore",
	} {
		if !strings.Contains(stdout, want) {
			t.Fatalf("list output missing %q:\n%s", want, stdout)
		}
	}

	stdout, stderr, code = run(nil, "ready")
	if code != 0 {
		t.Fatalf("ready returned %d stderr=%q", code, stderr)
	}
	for _, want := range []string{"tmp-u1ds", "tmp-qwtq", "tmp-9n81"} {
		if !strings.Contains(stdout, want) {
			t.Fatalf("ready output missing %q:\n%s", want, stdout)
		}
	}

	stdout, stderr, code = run(nil, "blocked")
	if code != 0 {
		t.Fatalf("blocked returned %d stderr=%q", code, stderr)
	}
	for _, want := range []string{
		"tmp-2hq1 [P1][open] - Upstream comprehensive feature <- [tmp-9n81]",
		"tmp-d22y [P2][open] - Upstream comprehensive blocked <- [tmp-2hq1]",
		"tmp-8tbz [P2][open] - Upstream comprehensive cycle a <- [tmp-5uhs]",
	} {
		if !strings.Contains(stdout, want) {
			t.Fatalf("blocked output missing %q:\n%s", want, stdout)
		}
	}

	stdout, stderr, code = run(nil, "dep", "tree", "--full", "tmp-d22y")
	if code != 0 {
		t.Fatalf("dep tree returned %d stderr=%q", code, stderr)
	}
	for _, want := range []string{"tmp-d22y [open] Upstream comprehensive blocked", "tmp-2hq1 [open] Upstream comprehensive feature", "tmp-9n81 [in_progress] Upstream comprehensive task"} {
		if !strings.Contains(stdout, want) {
			t.Fatalf("dep tree output missing %q:\n%s", want, stdout)
		}
	}

	stdout, stderr, code = run(nil, "dep", "cycle")
	if code != 0 {
		t.Fatalf("dep cycle returned %d stderr=%q", code, stderr)
	}
	if !strings.Contains(stdout, "tmp-5uhs -> tmp-8tbz -> tmp-5uhs") {
		t.Fatalf("dep cycle output missing expected cycle:\n%s", stdout)
	}

	stdout, stderr, code = run(nil, "closed", "--limit=5")
	if code != 0 {
		t.Fatalf("closed returned %d stderr=%q", code, stderr)
	}
	if !strings.Contains(stdout, "tmp-kkz9 [closed] - Upstream comprehensive chore") {
		t.Fatalf("closed output missing closed chore:\n%s", stdout)
	}

	stdout, stderr, code = run(nil, "query")
	if code != 0 {
		t.Fatalf("query returned %d stderr=%q", code, stderr)
	}
	lines := strings.Split(strings.TrimSpace(stdout), "\n")
	if len(lines) != 8 {
		t.Fatalf("query emitted %d records, want 8:\n%s", len(lines), stdout)
	}
	var foundEpic bool
	for _, line := range lines {
		var record map[string]any
		if err := json.Unmarshal([]byte(line), &record); err != nil {
			t.Fatalf("query line is not JSON: %q: %v", line, err)
		}
		if record["id"] == "tmp-u1ds" {
			foundEpic = true
			if record["type"] != "epic" || record["priority"] != "0" || record["external-ref"] != "gh-100" {
				t.Fatalf("epic query record mismatch: %#v", record)
			}
		}
	}
	if !foundEpic {
		t.Fatalf("query output did not include epic record:\n%s", stdout)
	}
}

func TestMockComprehensiveFixtureIncludesMalformedExamples(t *testing.T) {
	root := repoRoot(t)
	fixtureDir := filepath.Join(root, "testdata", "fixtures", "mock-comprehensive")
	chdir(t, fixtureDir)

	stdout, stderr, code := run(nil, "list")
	if code != 0 {
		t.Fatalf("list returned %d stderr=%q", code, stderr)
	}
	for _, want := range []string{
		"mk-epic [P0][open] - Mock comprehensive epic",
		"mk-feature [P1][open] - Mock comprehensive feature",
		"mk-bug [P2][open] - Mock comprehensive bug",
		"mk-cycle-a [P2][open] - Mock comprehensive cycle a",
		"mk-cycle-b [P2][open] - Mock comprehensive cycle b",
		"mk-task [P3][in_progress] - Mock comprehensive task",
		"mk-closed [P4][closed] - Mock comprehensive closed chore",
	} {
		if !strings.Contains(stdout, want) {
			t.Fatalf("list output missing %q:\n%s", want, stdout)
		}
	}
	assertMockComprehensiveWarnings(t, stderr)

	stdout, stderr, code = run(nil, "ready")
	if code != 0 {
		t.Fatalf("ready returned %d stderr=%q", code, stderr)
	}
	for _, want := range []string{
		"mk-epic [P0][open] - Mock comprehensive epic",
		"mk-bug [P2][open] - Mock comprehensive bug",
		"mk-task [P3][in_progress] - Mock comprehensive task",
	} {
		if !strings.Contains(stdout, want) {
			t.Fatalf("ready output missing %q:\n%s", want, stdout)
		}
	}
	assertMockComprehensiveWarnings(t, stderr)

	stdout, stderr, code = run(nil, "blocked")
	if code != 0 {
		t.Fatalf("blocked returned %d stderr=%q", code, stderr)
	}
	for _, want := range []string{
		"mk-feature [P1][open] - Mock comprehensive feature <- [mk-task]",
		"mk-cycle-a [P2][open] - Mock comprehensive cycle a <- [mk-cycle-b]",
		"mk-cycle-b [P2][open] - Mock comprehensive cycle b <- [mk-cycle-a]",
	} {
		if !strings.Contains(stdout, want) {
			t.Fatalf("blocked output missing %q:\n%s", want, stdout)
		}
	}
	assertMockComprehensiveWarnings(t, stderr)

	stdout, stderr, code = run(nil, "dep", "cycle")
	if code != 0 {
		t.Fatalf("dep cycle returned %d stderr=%q", code, stderr)
	}
	if !strings.Contains(stdout, "Cycle 1: mk-cycle-a -> mk-cycle-b -> mk-cycle-a") {
		t.Fatalf("dep cycle output missing expected cycle:\n%s", stdout)
	}
	assertMockComprehensiveWarnings(t, stderr)

	stdout, stderr, code = run(nil, "query")
	if code != 0 {
		t.Fatalf("query returned %d stderr=%q", code, stderr)
	}
	assertMockComprehensiveWarnings(t, stderr)
	lines := strings.Split(strings.TrimSpace(stdout), "\n")
	if len(lines) != 7 {
		t.Fatalf("query emitted %d records, want 7:\n%s", len(lines), stdout)
	}
	seen := map[string]bool{}
	for _, line := range lines {
		var record map[string]any
		if err := json.Unmarshal([]byte(line), &record); err != nil {
			t.Fatalf("query line is not JSON: %q: %v", line, err)
		}
		id, _ := record["id"].(string)
		seen[id] = true
		if strings.HasPrefix(id, "mk-malformed") || id == "" {
			t.Fatalf("query emitted malformed or missing ID record: %#v", record)
		}
	}
	for _, want := range []string{"mk-epic", "mk-feature", "mk-bug", "mk-cycle-a", "mk-cycle-b", "mk-task", "mk-closed"} {
		if !seen[want] {
			t.Fatalf("query output missing %s:\n%s", want, stdout)
		}
	}
}

func TestQueryOutputsJSONLAndRejectsFiltersAsDeferred(t *testing.T) {
	root := repoRoot(t)
	fixtureDir := filepath.Join(root, "testdata", "fixtures", "upstream-basic")
	chdir(t, fixtureDir)

	stdout, stderr, code := run(nil, "query")
	if code != 0 {
		t.Fatalf("query returned %d stderr=%q", code, stderr)
	}
	if stderr != "" {
		t.Fatalf("query stderr = %q, want empty", stderr)
	}
	lines := strings.Split(strings.TrimSpace(stdout), "\n")
	if len(lines) != 3 {
		t.Fatalf("query emitted %d JSONL records, want 3:\n%s", len(lines), stdout)
	}
	for _, line := range lines {
		var record map[string]any
		if err := json.Unmarshal([]byte(line), &record); err != nil {
			t.Fatalf("query line is not JSON: %q: %v", line, err)
		}
		if _, ok := record["abs_path"]; ok {
			t.Fatalf("query record leaked list-only abs_path field: %v", record)
		}
		if _, ok := record["rel_path"]; ok {
			t.Fatalf("query record leaked list-only rel_path field: %v", record)
		}
		if _, ok := record["path"]; ok {
			t.Fatalf("query record leaked path field: %v", record)
		}
	}

	_, stderr, code = run(nil, "query", `.status == "open"`)
	if code == 0 {
		t.Fatal("query filter unexpectedly succeeded")
	}
	if !strings.Contains(stderr, "filters are deferred") || !strings.Contains(stderr, "future feature parity") {
		t.Fatalf("query filter stderr = %q, want deferred feature parity error", stderr)
	}
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
	want := strings.ReplaceAll(string(wantBytes), "\r\n", "\n")
	if normalize != nil {
		got = normalize(absFixture, got)
		want = normalize(absFixture, want)
	}
	if got != want {
		t.Fatalf("%v output mismatch\nwant:\n%s\ngot:\n%s", args, want, got)
	}
}

func assertMockComprehensiveWarnings(t *testing.T, stderr string) {
	t.Helper()
	for _, want := range []string{
		"warning:",
		"mk-malformed-list.md",
		`malformed inline list "["`,
		"mk-missing-frontmatter.md",
		"missing YAML frontmatter",
		"mk-missing-id.md",
		"missing required id",
	} {
		if !strings.Contains(stderr, want) {
			t.Fatalf("stderr missing %q:\n%s", want, stderr)
		}
	}
}

func normalizeFixturePaths(fixtureDir string, value string) string {
	normalized := strings.ReplaceAll(value, `\\`, "/")
	normalized = strings.ReplaceAll(normalized, "\\", "/")
	fixture := strings.ReplaceAll(fixtureDir, "\\", "/")
	return strings.ReplaceAll(normalized, fixture, "<FIXTURE>")
}

func addDependencyByHand(t *testing.T, id string, dep string) {
	t.Helper()
	path := filepath.Join(".tickets", id+".md")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read ticket %s: %v", id, err)
	}
	content := strings.Replace(string(data), "deps: []", "deps: ["+dep+"]", 1)
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write ticket %s: %v", id, err)
	}
}

func setTicketMTime(t *testing.T, id string, mtime time.Time) {
	t.Helper()
	path := filepath.Join(".tickets", id+".md")
	if err := os.Chtimes(path, mtime, mtime); err != nil {
		t.Fatalf("set mtime for %s: %v", id, err)
	}
}

func countSubstring(value string, sub string) int {
	count := 0
	for {
		idx := strings.Index(value, sub)
		if idx < 0 {
			return count
		}
		count++
		value = value[idx+len(sub):]
	}
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

func canonicalTestPath(t *testing.T, path string) string {
	t.Helper()
	resolved, err := filepath.EvalSymlinks(path)
	if err != nil {
		t.Fatalf("canonicalize %s: %v", path, err)
	}
	return resolved
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

func writeExecutable(t *testing.T, path string, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o755); err != nil {
		t.Fatalf("write executable %s: %v", path, err)
	}
}

func shellQuote(value string) string {
	return "'" + strings.ReplaceAll(value, "'", "'\\''") + "'"
}
