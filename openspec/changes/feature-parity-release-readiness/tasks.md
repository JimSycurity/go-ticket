## 1. Compatibility Baseline

- [x] 1.1 Create a command compatibility matrix covering upstream documented commands and plugin behavior.
- [x] 1.2 Add fixture repositories generated from representative upstream-compatible `.tickets/` files plus hand-authored mock fixtures with malformed examples.
- [x] 1.3 Add golden and fixture-backed command-output tests for parity-sensitive behavior and malformed-ticket warnings.
- [x] 1.4 Decide and document whether byte-for-byte output compatibility is required per command.

## 2. Parity Commands

- [x] 2.1 Implement `dep tree` and `dep cycle` with tests for deduplication, full traversal, missing tickets, and cycles.
- [x] 2.2 Implement `closed --limit=N` with mtime or documented ordering behavior.
- [x] 2.3 Implement `query` JSON output and the selected filter strategy. Security reference: `security.md#medium-filtered-query-strategy`.
- [x] 2.3a Define the plugin/editor execution security policy, including PATH lookup order, extension allowlist, `.ps1` handling, environment variables, editor argument handling, and shell interpolation rules. Security references: `security.md#high-plugin-and-super-execution`, `security.md#high-editor-integration`.
- [x] 2.4 Implement plugin discovery and `super` bypass behavior for Unix and Windows command extensions. Security reference: `security.md#high-plugin-and-super-execution`.
- [x] 2.5 Implement `edit` with safe editor resolution and clear errors when no editor is configured. Security reference: `security.md#high-editor-integration`.
- [x] 2.6 Implement `migrate-beads` with import reporting and conflict handling. Security references: `security.md#medium-beads-migration`, `security.md#medium-write-surface-invariants`.
- [x] 2.7 Implement optional `.tickets/` settings for ID prefix and documented project behavior overrides. Security references: `security.md#medium-optional-tickets-settings`, `security.md#medium-write-surface-invariants`.

## 3. Security and Migration Confidence

- [x] 3.1 Review plugin execution, editor launching, PATH lookup, Beads input parsing, and ticket file writes for security risks. Security references: `security.md#recommended-gate-order` and all feature-specific sections in `security.md`.
- [x] 3.2 Add tests that prevent ticket path traversal and accidental writes outside the active `.tickets/` directory. Security reference: `security.md#medium-write-surface-invariants`.
- [x] 3.3 Add migration tests for malformed, partial, and conflicting Beads data. Security reference: `security.md#medium-beads-migration`.
- [x] 3.4 Document rollback paths for `.tickets/` compatibility testing and Beads migration. Security reference: `security.md#medium-beads-migration`.

## 4. Release Pipeline

- [x] 4.1 Add CI checks for formatting, `go test`, and platform matrix coverage on Windows, macOS, and Linux. Security reference: `security.md#low-release-pipeline-supply-chain`.
- [x] 4.2 Add release workflow triggered by version tags. Security reference: `security.md#low-release-pipeline-supply-chain`.
- [x] 4.3 Produce archives for supported OS/architecture targets. Security reference: `security.md#low-release-pipeline-supply-chain`.
- [x] 4.4 Publish checksums and release notes with every tagged release. Security reference: `security.md#low-release-pipeline-supply-chain`.

## 5. Documentation

- [x] 5.1 Expand README with project purpose, compatibility promise, install instructions, and basic usage.
- [x] 5.2 Add compatibility, migration, plugin, security, and release documentation. Security references: `security.md#high-plugin-and-super-execution`, `security.md#high-editor-integration`, `security.md#medium-beads-migration`, `security.md#low-json-path-disclosure`, `security.md#low-release-pipeline-supply-chain`.
- [x] 5.3 Add shell completion generation or installation guidance for supported shells.
- [x] 5.4 Add attribution to `wedow/ticket` and clarify that `go-ticket` is a compatible reimplementation.
- [x] 5.5 Add contributor guidance for preserving compatibility and adding command tests. Security references: `security.md#medium-write-surface-invariants`, `security.md#medium-optional-tickets-settings`.
