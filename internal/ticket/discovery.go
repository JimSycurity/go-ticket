package ticket

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

const TicketsDirName = ".tickets"

type Root struct {
	ProjectDir string
	TicketsDir string
	Source     RootSource
}

type RootSource string

const (
	RootSourceDiscovered RootSource = "discovered"
	RootSourceEnv        RootSource = "env"
)

type EnvLookup func(string) string

func Discover(startDir string, lookup EnvLookup) (Root, error) {
	if lookup == nil {
		lookup = os.Getenv
	}
	if override := lookup("TICKETS_DIR"); override != "" {
		return discoverOverride(override)
	}

	start, err := filepath.Abs(startDir)
	if err != nil {
		return Root{}, fmt.Errorf("resolve start directory: %w", err)
	}
	start, err = filepath.EvalSymlinks(start)
	if err != nil {
		return Root{}, fmt.Errorf("canonicalize start directory: %w", err)
	}

	info, err := os.Stat(start)
	if err != nil {
		return Root{}, fmt.Errorf("stat start directory: %w", err)
	}
	if !info.IsDir() {
		return Root{}, fmt.Errorf("start path is not a directory: %s", start)
	}

	if filepath.Base(start) == TicketsDirName {
		return Root{ProjectDir: filepath.Dir(start), TicketsDir: start, Source: RootSourceDiscovered}, nil
	}

	for dir := start; ; dir = filepath.Dir(dir) {
		candidate := filepath.Join(dir, TicketsDirName)
		root, err := rootFromCandidate(dir, candidate)
		if err == nil {
			return root, nil
		}
		if !errors.Is(err, os.ErrNotExist) {
			return Root{}, err
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
	}

	return Root{}, fmt.Errorf("no .tickets directory found from %s; run gtk init first or set TICKETS_DIR", start)
}

func discoverOverride(path string) (Root, error) {
	if !filepath.IsAbs(path) {
		return Root{}, fmt.Errorf("TICKETS_DIR must be an absolute path before path-security review is complete: %s", path)
	}
	info, err := os.Lstat(path)
	if err != nil {
		return Root{}, fmt.Errorf("stat TICKETS_DIR: %w", err)
	}
	if info.Mode()&os.ModeSymlink != 0 {
		return Root{}, fmt.Errorf("symlinked TICKETS_DIR requires path-security review before use: %s", path)
	}
	if !info.IsDir() {
		return Root{}, fmt.Errorf("TICKETS_DIR is not a directory: %s", path)
	}
	resolved, err := filepath.EvalSymlinks(path)
	if err != nil {
		return Root{}, fmt.Errorf("canonicalize TICKETS_DIR: %w", err)
	}
	info, err = os.Stat(resolved)
	if err != nil {
		return Root{}, fmt.Errorf("stat TICKETS_DIR: %w", err)
	}
	if !info.IsDir() {
		return Root{}, fmt.Errorf("TICKETS_DIR is not a directory: %s", resolved)
	}
	return Root{ProjectDir: filepath.Dir(resolved), TicketsDir: resolved, Source: RootSourceEnv}, nil
}

func rootFromCandidate(projectDir string, ticketsDir string) (Root, error) {
	info, err := os.Lstat(ticketsDir)
	if err != nil {
		return Root{}, err
	}
	if info.Mode()&os.ModeSymlink != 0 {
		return Root{}, fmt.Errorf("symlinked .tickets requires path-security review before use: %s", ticketsDir)
	}
	if !info.IsDir() {
		return Root{}, fmt.Errorf(".tickets exists but is not a directory: %s", ticketsDir)
	}
	resolved, err := filepath.EvalSymlinks(ticketsDir)
	if err != nil {
		return Root{}, fmt.Errorf("canonicalize .tickets: %w", err)
	}
	project, err := filepath.EvalSymlinks(projectDir)
	if err != nil {
		return Root{}, fmt.Errorf("canonicalize project root: %w", err)
	}
	return Root{ProjectDir: project, TicketsDir: resolved, Source: RootSourceDiscovered}, nil
}
