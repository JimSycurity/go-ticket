package app

import (
	"bytes"
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
