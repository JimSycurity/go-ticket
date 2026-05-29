# core-ticket-workflows Specification

## Purpose
TBD - created by archiving change core-cross-platform-port. Update Purpose after archive.
## Requirements
### Requirement: Ticket creation
The system SHALL create Markdown ticket files with compatible frontmatter and a Markdown title.

#### Scenario: Create requires ticket directory
- **WHEN** the user runs `gtk create "Add Windows support"` and no `.tickets/` directory is active
- **THEN** the CLI SHALL fail without creating files and tell the user to run `gtk init`

#### Scenario: Create ticket with defaults
- **WHEN** the user runs `gtk create "Add Windows support"`
- **THEN** the CLI SHALL create a new open task ticket in `.tickets/` and print its generated ID

#### Scenario: Create ticket with metadata
- **WHEN** the user supplies type, priority, assignee, external reference, parent, tags, description, design, or acceptance text
- **THEN** the CLI SHALL write those values into the created ticket using compatible fields and Markdown sections

#### Scenario: Reject invalid type write
- **WHEN** the user supplies a ticket type other than `bug`, `feature`, `task`, `epic`, or `chore`
- **THEN** the CLI SHALL fail without creating or modifying a ticket

#### Scenario: Read unknown type
- **WHEN** the CLI reads an existing ticket with an unknown type
- **THEN** broad read commands SHALL surface the type without treating the ticket file as malformed

#### Scenario: Reject invalid priority write
- **WHEN** the user supplies a priority outside integer range `0` through `4`
- **THEN** the CLI SHALL fail without creating or modifying a ticket

#### Scenario: Read unusual priority
- **WHEN** the CLI reads an existing ticket with an unusual priority value
- **THEN** broad read commands SHALL surface the priority without treating the ticket file as malformed

### Requirement: Ticket lifecycle updates
The system SHALL update ticket status through the MVP lifecycle commands.

#### Scenario: Start ticket
- **WHEN** the user runs `gtk start <id>`
- **THEN** the ticket status SHALL become `in_progress`

#### Scenario: Close ticket
- **WHEN** the user runs `gtk close <id>`
- **THEN** the ticket status SHALL become `closed`

#### Scenario: Reopen ticket
- **WHEN** the user runs `gtk reopen <id>`
- **THEN** the ticket status SHALL become `open`

#### Scenario: Set valid status
- **WHEN** the user runs `gtk status <id> <status>` with `open`, `in_progress`, or `closed`
- **THEN** the ticket status SHALL become the requested status

#### Scenario: Reject invalid status write
- **WHEN** the user runs `gtk status <id> <status>` with an unsupported MVP status
- **THEN** the CLI SHALL fail without modifying the ticket

#### Scenario: Read unknown status
- **WHEN** the CLI reads an existing ticket with an unknown status
- **THEN** broad read commands SHALL surface the status without treating the ticket file as malformed

### Requirement: Ticket relationships
The system SHALL support dependency, link, and parent relationship fields compatible with upstream ticket files.

#### Scenario: Add dependency
- **WHEN** the user runs `gtk dep <id> <dep-id>`
- **THEN** `<dep-id>` SHALL be added to the dependent ticket's `deps` list once

#### Scenario: Remove dependency
- **WHEN** the user runs `gtk undep <id> <dep-id>`
- **THEN** `<dep-id>` SHALL be removed from the dependent ticket's `deps` list

#### Scenario: Link tickets
- **WHEN** the user runs `gtk link <id> <target-id>`
- **THEN** each ticket SHALL reference the other in `links`

#### Scenario: Unlink tickets
- **WHEN** the user runs `gtk unlink <id> <target-id>`
- **THEN** each ticket SHALL remove the other from `links`

#### Scenario: Parent metadata
- **WHEN** the CLI reads, creates, shows, or lists a ticket with a `parent` field
- **THEN** it SHALL preserve and display the parent metadata without requiring parent/child traversal commands in MVP

### Requirement: Ticket listing and display
The system SHALL provide core read commands for humans and agents.

#### Scenario: Show ticket
- **WHEN** the user runs `gtk show <id>`
- **THEN** the CLI SHALL print the raw ticket Markdown file content exactly as stored

#### Scenario: List tickets
- **WHEN** the user runs `gtk list` or `gtk ls`
- **THEN** the CLI SHALL print all tickets, including closed tickets, with ID, priority, status, and title in a stable order as an intentional go-ticket MVP convenience command

#### Scenario: Filter list
- **WHEN** the user runs list filters for status, assignee, or type
- **THEN** the CLI SHALL only include tickets matching the requested filters

#### Scenario: JSON list output
- **WHEN** the user runs `gtk list --json`
- **THEN** the CLI SHALL print compact machine-readable JSON ticket summaries with raw ticket fields, project-relative file paths, absolute file paths, and no raw Markdown bodies or derived graph fields

#### Scenario: JSON absolute path disclosure
- **WHEN** the user runs `gtk list --json`
- **THEN** the CLI documentation SHALL identify absolute paths as local-machine metadata that may expose usernames or repository locations when copied into logs or reports

### Requirement: Ready and blocked views
The system SHALL classify open and in-progress tickets based on unresolved dependencies.

#### Scenario: Ready ticket
- **WHEN** an open or in-progress ticket has no dependencies or all dependencies are closed
- **THEN** `gtk ready` SHALL include the ticket

#### Scenario: Blocked ticket
- **WHEN** an open or in-progress ticket depends on at least one open or in-progress ticket
- **THEN** `gtk blocked` SHALL include the ticket and identify blocking dependencies

#### Scenario: Missing dependency target
- **WHEN** an open or in-progress ticket depends on a missing or malformed ticket ID
- **THEN** `gtk blocked` SHALL include the ticket and identify the unresolved dependency reference

#### Scenario: Ready excludes missing dependency target
- **WHEN** an open or in-progress ticket depends on a missing or malformed ticket ID
- **THEN** `gtk ready` SHALL NOT include the ticket

#### Scenario: Unknown dependency status
- **WHEN** a ticket depends on a ticket whose status is neither `open`, `in_progress`, nor `closed`
- **THEN** `gtk blocked` SHALL treat that dependency as unresolved unless it is exactly `closed`

### Requirement: Notes
The system SHALL append timestamped notes to ticket Markdown without disturbing frontmatter.

#### Scenario: Add note argument
- **WHEN** the user runs `gtk add-note <id> "Investigated parser"`
- **THEN** the note SHALL be appended under a notes section with an upstream-compatible timestamp when confirmed by fixtures, otherwise a documented UTC ISO-8601 timestamp

#### Scenario: Add note from stdin
- **WHEN** the user pipes note text into `gtk add-note <id>`
- **THEN** the piped text SHALL be appended as the note body

