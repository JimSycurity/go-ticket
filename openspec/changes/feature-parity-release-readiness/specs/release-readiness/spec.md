## ADDED Requirements

### Requirement: Cross-platform test matrix
The system SHALL run automated tests on Windows, macOS, and Linux for supported Go versions.

#### Scenario: Pull request validation
- **WHEN** a pull request is opened or updated
- **THEN** CI SHALL run formatting, lint or vet checks, and tests on all supported platforms

#### Scenario: Platform regression
- **WHEN** a platform-specific test fails
- **THEN** CI SHALL fail the change before release artifacts are produced

### Requirement: Versioned release artifacts
The system SHALL produce versioned release artifacts for supported operating systems and architectures.

#### Scenario: Tag release
- **WHEN** a version tag is pushed
- **THEN** the release workflow SHALL build platform binaries and attach archives to a GitHub release

#### Scenario: Checksums
- **WHEN** release artifacts are produced
- **THEN** the workflow SHALL publish checksums for every artifact

### Requirement: Installation documentation
The system SHALL document installation and update paths for Windows, macOS, and Linux.

#### Scenario: Windows install
- **WHEN** a Windows user reads the installation docs
- **THEN** they SHALL find steps for installing `gtk.exe` and adding it to PATH

#### Scenario: Unix install
- **WHEN** a macOS or Linux user reads the installation docs
- **THEN** they SHALL find steps for installing the `gtk` binary and verifying it

#### Scenario: Optional upstream replacement
- **WHEN** a user wants `go-ticket` to replace upstream `tk`
- **THEN** the docs SHALL explain that `gtk` is the side-by-side binary and any `tk` wrapper, copy, or symlink is an intentional local opt-in

### Requirement: Shell completions
The system SHALL provide shell completion generation or installation guidance for supported shells after core parity is stable.

#### Scenario: Completion documentation
- **WHEN** a user reads release documentation
- **THEN** they SHALL find the supported completion shells and installation steps

### Requirement: Compatibility and migration documentation
The system SHALL document compatibility expectations, migration steps, rollback, and known limitations.

#### Scenario: Existing ticket repo
- **WHEN** a user already has `.tickets/` created by `wedow/ticket`
- **THEN** the docs SHALL explain how to test read-only commands before using write commands

#### Scenario: Beads migration
- **WHEN** a user wants to migrate from Beads
- **THEN** the docs SHALL explain prerequisites, command behavior, review steps, and rollback

### Requirement: Attribution and license clarity
The system SHALL clearly attribute inspiration and compatibility targets to `wedow/ticket` while identifying `go-ticket` as a compatible reimplementation.

#### Scenario: User reads README
- **WHEN** a user reads the README
- **THEN** they SHALL see that `go-ticket` is a Go reimplementation compatible with and inspired by `wedow/ticket`

#### Scenario: User reads license docs
- **WHEN** a user reviews licensing and attribution files
- **THEN** they SHALL see the project license and upstream attribution without implying upstream endorsement
