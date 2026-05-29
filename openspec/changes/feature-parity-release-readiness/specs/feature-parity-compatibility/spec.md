## ADDED Requirements

### Requirement: Documented command parity
The system SHALL implement the documented `wedow/ticket` command surface or explicitly mark unsupported behavior in a compatibility matrix.

#### Scenario: User checks parity status
- **WHEN** a user reads the compatibility documentation
- **THEN** each upstream command SHALL be marked as supported, partially supported, or unsupported with notes

#### Scenario: Supported command behaves compatibly
- **WHEN** a command is marked supported
- **THEN** tests SHALL cover its expected behavior using `.tickets/` fixtures

#### Scenario: Malformed ticket fixtures remain non-fatal
- **WHEN** a `.tickets/` fixture contains malformed ticket-shaped files
- **THEN** read-only commands SHALL warn about malformed files while continuing to process valid tickets

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
The system SHALL provide JSONL output for tickets and clearly document the selected query filtering strategy.

#### Scenario: Query all tickets
- **WHEN** the user runs `tk query`
- **THEN** the CLI SHALL output one JSON object per ticket line using ticket frontmatter fields

#### Scenario: Query with filter
- **WHEN** the user runs `tk query <filter>`
- **THEN** the CLI SHALL apply the documented filter behavior or return a clear unsupported-filter error that identifies filtered query as future feature-parity work

### Requirement: Plugin compatibility
The system SHALL support compatible plugin discovery and execution for custom commands only after a documented plugin/editor execution security policy exists.

#### Scenario: Plugin execution policy
- **WHEN** plugin compatibility is implemented
- **THEN** the implementation SHALL document and test PATH lookup order, extension allowlist, deferred PowerShell script handling, environment variables passed to plugins, builtin precedence, command-name validation, and no implicit shell interpolation

#### Scenario: Unix plugin command
- **WHEN** a Unix user has `tk-hello` or `ticket-hello` on PATH
- **THEN** `tk hello` SHALL invoke the plugin with ticket environment variables

#### Scenario: Windows plugin command
- **WHEN** a Windows user has a compatible `tk-hello.exe` on PATH
- **THEN** `tk hello` SHALL invoke the plugin with ticket environment variables

#### Scenario: Deferred Windows script plugins
- **WHEN** a Windows user has only `.cmd`, `.bat`, or `.ps1` plugin candidates
- **THEN** the CLI SHALL not execute them until a documented wrapper policy and tests are added

#### Scenario: Bypass plugins
- **WHEN** the user runs `tk super <cmd>`
- **THEN** the CLI SHALL execute the built-in command even if a plugin with that name exists

### Requirement: Editor integration
The system SHALL support opening ticket files in the configured editor when `edit` parity is enabled.

#### Scenario: Edit ticket
- **WHEN** the user runs `tk edit <id>` with a validated editor command
- **THEN** the CLI SHALL launch the editor for the resolved ticket file

#### Scenario: Unsafe editor command
- **WHEN** the configured editor contains inline arguments, shell metacharacters, or repository-controlled editor configuration
- **THEN** the CLI SHALL return a clear error without launching a process

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
