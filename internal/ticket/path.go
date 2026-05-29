package ticket

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

var ticketIDPattern = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9_-]*(?:\.[A-Za-z0-9][A-Za-z0-9_-]*)*$`)

var windowsReservedNames = map[string]struct{}{
	"CON":  {},
	"PRN":  {},
	"AUX":  {},
	"NUL":  {},
	"COM1": {},
	"COM2": {},
	"COM3": {},
	"COM4": {},
	"COM5": {},
	"COM6": {},
	"COM7": {},
	"COM8": {},
	"COM9": {},
	"LPT1": {},
	"LPT2": {},
	"LPT3": {},
	"LPT4": {},
	"LPT5": {},
	"LPT6": {},
	"LPT7": {},
	"LPT8": {},
	"LPT9": {},
}

func ResolveTicketPath(root Root, id string, mustExist bool) (string, error) {
	if err := ValidateID(id); err != nil {
		return "", err
	}

	ticketsDir, err := filepath.Abs(root.TicketsDir)
	if err != nil {
		return "", fmt.Errorf("resolve tickets directory: %w", err)
	}
	info, err := os.Lstat(ticketsDir)
	if err != nil {
		return "", fmt.Errorf("stat tickets directory: %w", err)
	}
	if info.Mode()&os.ModeSymlink != 0 {
		return "", fmt.Errorf("tickets directory is a symlink: %s", ticketsDir)
	}
	if !info.IsDir() {
		return "", fmt.Errorf("tickets directory is not a directory: %s", ticketsDir)
	}

	path := filepath.Join(ticketsDir, id+".md")
	resolved, err := filepath.Abs(path)
	if err != nil {
		return "", fmt.Errorf("resolve ticket path: %w", err)
	}

	rel, err := filepath.Rel(ticketsDir, resolved)
	if err != nil {
		return "", fmt.Errorf("check ticket path containment: %w", err)
	}
	if rel == ".." || strings.HasPrefix(rel, ".."+string(os.PathSeparator)) || filepath.IsAbs(rel) {
		return "", fmt.Errorf("ticket path escapes .tickets directory: %s", resolved)
	}

	info, err = os.Lstat(resolved)
	if err != nil {
		if os.IsNotExist(err) && !mustExist {
			return resolved, nil
		}
		return "", fmt.Errorf("stat ticket path: %w", err)
	}
	if info.Mode()&os.ModeSymlink != 0 {
		return "", fmt.Errorf("ticket file is a symlink: %s", resolved)
	}
	if !info.Mode().IsRegular() {
		return "", fmt.Errorf("ticket path is not a regular file: %s", resolved)
	}
	return resolved, nil
}
