## ADDED Requirements

### Requirement: Documented command parity
The system SHALL implement the documented `wedow/ticket` command surface or explicitly mark unsupported behavior in a compatibility matrix.

#### Scenario: User checks parity status
- **WHEN** a user reads the compatibility documentation
- **THEN** each upstream command SHALL be marked as supported, partially supported, or unsupported with notes

#### Scenario: Supported command behaves compatibly
- **WHEN** a command is marked supported
- **THEN** tests SHALL cover its expected behavior using `.tickets/` fixtures

### Requirement: Dependency tree and cycle commands
The system SHALL support dependency tree and dependency cycle analysis.

#### Scenario: Show dependency tree
- **WHEN** the user runs `tk dep tree <id>`
- **THEN** the CLI SHALL print a readable dependency tree rooted at the resolved ticket

#### Scenario: Show full dependency tree
- **WHEN** the user runs `tk dep tree --full <id>`
- **THEN** the CLI SHALL print repeated dependencies without deduplicating them

#### Scenario: Detect dependency cycles
- **WHEN** the user runs `tk dep cycle`
- **THEN** the CLI SHALL report cycles among open or in-progress tickets

### Requirement: Query output
The system SHALL provide JSON output for tickets and support the chosen query filtering strategy.

#### Scenario: Query all tickets
- **WHEN** the user runs `tk query`
- **THEN** the CLI SHALL output ticket data as JSON

#### Scenario: Query with filter
- **WHEN** the user runs `tk query <filter>`
- **THEN** the CLI SHALL apply the documented filter behavior or return a clear unsupported-filter error

### Requirement: Plugin compatibility
The system SHALL support compatible plugin discovery and execution for custom commands.

#### Scenario: Unix plugin command
- **WHEN** a Unix user has `tk-hello` or `ticket-hello` on PATH
- **THEN** `tk hello` SHALL invoke the plugin with ticket environment variables

#### Scenario: Windows plugin command
- **WHEN** a Windows user has a compatible `tk-hello.exe`, `.cmd`, `.bat`, or documented script extension on PATH
- **THEN** `tk hello` SHALL invoke the plugin with ticket environment variables

#### Scenario: Bypass plugins
- **WHEN** the user runs `tk super <cmd>`
- **THEN** the CLI SHALL execute the built-in command even if a plugin with that name exists

### Requirement: Editor integration
The system SHALL support opening ticket files in the configured editor when `edit` parity is enabled.

#### Scenario: Edit ticket
- **WHEN** the user runs `tk edit <id>` with a configured editor
- **THEN** the CLI SHALL launch the editor for the resolved ticket file

#### Scenario: Missing editor
- **WHEN** the user runs `tk edit <id>` without an editor configuration
- **THEN** the CLI SHALL return a clear error without modifying the ticket

### Requirement: Beads migration
The system SHALL import Beads issue exports into compatible `.tickets/` Markdown files.

#### Scenario: Migrate beads
- **WHEN** the user runs `tk migrate-beads` in a repository containing `.beads/issues.jsonl`
- **THEN** the CLI SHALL create corresponding ticket files and preserve relevant metadata and relationships

#### Scenario: Migration review
- **WHEN** migration creates files
- **THEN** the CLI SHALL report what was imported and what requires manual review

### Requirement: Optional project settings
The system SHALL support an optional lightweight settings file under `.tickets/` for configurable project behavior.

#### Scenario: Settings define prefix
- **WHEN** a `.tickets/` settings file defines a ticket ID prefix
- **THEN** newly created tickets SHALL use that prefix while still avoiding collisions

#### Scenario: Settings absent
- **WHEN** no `.tickets/` settings file exists
- **THEN** the CLI SHALL use upstream-compatible default behavior
