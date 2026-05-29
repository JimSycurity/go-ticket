# cross-platform-ticket-cli Specification

## Purpose
TBD - created by archiving change core-cross-platform-port. Update Purpose after archive.
## Requirements
### Requirement: Native cross-platform executable
The system SHALL provide a Go-built `gtk` executable that runs without Bash, coreutils, jq, grep, or ripgrep on Windows, macOS, and Linux for the MVP command set.

#### Scenario: Windows user runs tk
- **WHEN** a user runs the built `gtk.exe` from PowerShell in a repository containing `.tickets/`
- **THEN** the command SHALL discover and operate on the ticket directory without requiring WSL or Git Bash

#### Scenario: Unix user runs gtk
- **WHEN** a user runs the built `gtk` binary on macOS or Linux in a repository containing `.tickets/`
- **THEN** the command SHALL provide the same MVP behavior as the Windows build

#### Scenario: Side-by-side upstream testing
- **WHEN** upstream `tk` is already installed on PATH
- **THEN** the MVP binary SHALL NOT replace it and SHALL remain invokable as `gtk`

### Requirement: Minimal cross-platform CI
The system SHALL run MVP tests in GitHub Actions on Windows, macOS, and Linux.

#### Scenario: CI test matrix
- **WHEN** code is pushed or a pull request is opened
- **THEN** GitHub Actions SHALL run `go test` on Windows, macOS, and Linux

### Requirement: MVP attribution
The system SHALL include basic README attribution to `wedow/ticket` as the inspiration and compatibility target.

#### Scenario: User reads README
- **WHEN** a user reads the MVP README
- **THEN** they SHALL see that `go-ticket` is inspired by and aims for file-format compatibility with `wedow/ticket`

### Requirement: Ticket directory discovery
The system SHALL locate the active ticket directory by first honoring `TICKETS_DIR`, then walking parent directories from the current working directory until it finds `.tickets/`.

#### Scenario: TICKETS_DIR override
- **WHEN** `TICKETS_DIR` points to an existing ticket directory
- **THEN** the CLI SHALL verify the override is absolute, existing, non-symlinked, and a directory, canonicalize it, mark the selected root as environment-selected, and use the resolved path

#### Scenario: Invalid TICKETS_DIR override
- **WHEN** `TICKETS_DIR` is set but does not resolve to an existing directory
- **THEN** the CLI SHALL fail without falling back to ancestor discovery or creating files elsewhere

#### Scenario: Parent directory discovery
- **WHEN** the user runs `gtk` in a nested subdirectory of a repository containing `.tickets/`
- **THEN** the CLI SHALL find the ancestor `.tickets/` directory and use it

#### Scenario: Non-directory tickets path
- **WHEN** parent walking encounters a `.tickets` path that exists but is not a directory
- **THEN** the CLI SHALL fail without walking past it to an ancestor ticket directory

#### Scenario: Current directory is tickets directory
- **WHEN** the user runs `gtk` from inside `.tickets/`
- **THEN** the CLI SHALL treat the current directory as the active ticket directory and its parent as the project root

#### Scenario: Security-sensitive ticket root
- **WHEN** implementation encounters symlinked ticket directories, relative overrides, or external roots
- **THEN** behavior SHALL follow the security-reviewed path policy before write-capable commands are enabled, including rejecting symlinked discovered roots and symlinked `TICKETS_DIR` overrides in the MVP

#### Scenario: Missing ticket directory
- **WHEN** the user runs a read command and no ticket directory can be found
- **THEN** the CLI SHALL return a clear error explaining how to run `gtk init` or point to `.tickets/`

### Requirement: Explicit initialization
The system SHALL require an explicit `gtk init` command to create a new `.tickets/` directory.

#### Scenario: Initialize ticket directory
- **WHEN** the user runs `gtk init` in a directory without `.tickets/`
- **THEN** the CLI SHALL create `.tickets/` in the current directory and report the initialized path

#### Scenario: Create without initialization
- **WHEN** the user runs `gtk create "New work"` and no ticket directory can be found
- **THEN** the CLI SHALL fail without creating files and tell the user to run `gtk init`

#### Scenario: Initialize existing ticket directory
- **WHEN** the user runs `gtk init` in or below a repository that already has `.tickets/`
- **THEN** the CLI SHALL report the existing ticket directory without overwriting ticket files

### Requirement: Compatible ticket file format
The system SHALL read and write Markdown ticket files using YAML frontmatter compatible with `wedow/ticket`.

#### Scenario: Read upstream ticket
- **WHEN** the CLI reads a ticket file created by `wedow/ticket`
- **THEN** it SHALL correctly parse known fields including `id`, `status`, `deps`, `links`, `created`, `type`, `priority`, `assignee`, `external-ref`, `parent`, and `tags`

#### Scenario: Write MVP ticket
- **WHEN** the CLI creates or mutates a ticket
- **THEN** it SHALL write a Markdown file that upstream-compatible tools can read as a normal ticket

#### Scenario: Preserve custom metadata
- **WHEN** the CLI mutates a ticket containing unknown frontmatter fields
- **THEN** it SHALL preserve those fields and values while allowing the YAML block to be emitted in a stable normalized order

#### Scenario: Atomic ticket mutation
- **WHEN** the CLI mutates an existing ticket file
- **THEN** it SHALL write the replacement through a temporary file in the same `.tickets/` directory and rename it over the original

#### Scenario: Contained ticket write target
- **WHEN** the CLI creates or mutates a ticket file
- **THEN** it SHALL resolve the ticket ID through a single containment-checked path resolver that rejects traversal, path separators, drive-letter syntax, Windows reserved device names, symlinked ticket files, non-regular existing targets, and paths outside the active `.tickets/` directory

#### Scenario: Newline normalization
- **WHEN** the CLI creates or mutates a ticket file
- **THEN** it SHALL write LF newlines regardless of the operating system

#### Scenario: CRLF input
- **WHEN** the CLI reads a ticket file with CRLF newlines
- **THEN** it SHALL parse the ticket correctly

#### Scenario: Upstream-produced fixtures
- **WHEN** the compatibility test suite reads fixtures generated by upstream `tk`
- **THEN** `gtk` SHALL parse and operate on those fixtures without requiring fixture normalization

#### Scenario: Hand-authored edge fixtures
- **WHEN** the compatibility test suite reads hand-authored edge fixtures for malformed, CRLF, ambiguous, or Windows-specific cases
- **THEN** `gtk` SHALL handle each case according to the documented MVP behavior

### Requirement: Partial ID resolution
The system SHALL support partial ticket ID matching for commands that accept ticket IDs.

#### Scenario: Unique partial ID
- **WHEN** the user references a partial ID that matches exactly one ticket file
- **THEN** the CLI SHALL resolve it to that ticket

#### Scenario: Ambiguous partial ID
- **WHEN** the user references a partial ID that matches multiple ticket files
- **THEN** the CLI SHALL fail without modifying any ticket and report the ambiguity

### Requirement: Ticket ID generation
The system SHALL generate ticket IDs using upstream-compatible default behavior when no project settings are defined.

#### Scenario: Create ticket without settings
- **WHEN** the user runs `gtk create "New work"` in a repository with no `.tickets/` settings file
- **THEN** the CLI SHALL generate a non-colliding ticket ID using the documented upstream-compatible default behavior

### Requirement: MVP command boundary
The system SHALL NOT execute upstream-style plugins during the MVP.

#### Scenario: Unknown command
- **WHEN** the user runs an unknown `gtk` command
- **THEN** the CLI SHALL fail with a clear unsupported-command error and SHALL NOT search PATH for a plugin

#### Scenario: Help text
- **WHEN** the user runs `gtk --help` or `gtk help`
- **THEN** the CLI SHALL show MVP-specific help for supported commands only

### Requirement: Malformed ticket handling
The system SHALL tolerate malformed ticket files during broad read operations and protect malformed targets from mutation.

#### Scenario: List with malformed ticket
- **WHEN** the user runs a broad read command such as `gtk list`, `gtk ready`, or `gtk blocked` and one ticket file is malformed
- **THEN** the CLI SHALL continue processing valid tickets and report a warning for the malformed file

#### Scenario: Mutate malformed target
- **WHEN** the user runs a mutation command targeting a malformed ticket file
- **THEN** the CLI SHALL fail without modifying the file and report that the target could not be parsed safely

