## 1. Compatibility Baseline

- [ ] 1.1 Create a command compatibility matrix covering upstream documented commands and plugin behavior.
- [ ] 1.2 Add fixture repositories generated from representative upstream-compatible `.tickets/` files.
- [ ] 1.3 Add golden command-output tests for parity-sensitive behavior.
- [ ] 1.4 Decide and document whether byte-for-byte output compatibility is required per command.

## 2. Parity Commands

- [ ] 2.1 Implement `dep tree` and `dep cycle` with tests for deduplication, full traversal, missing tickets, and cycles.
- [ ] 2.2 Implement `closed --limit=N` with mtime or documented ordering behavior.
- [ ] 2.3 Implement `query` JSON output and the selected filter strategy.
- [ ] 2.3a Define the plugin/editor execution security policy, including PATH lookup order, extension allowlist, `.ps1` handling, environment variables, editor argument handling, and shell interpolation rules.
- [ ] 2.4 Implement plugin discovery and `super` bypass behavior for Unix and Windows command extensions.
- [ ] 2.5 Implement `edit` with safe editor resolution and clear errors when no editor is configured.
- [ ] 2.6 Implement `migrate-beads` with import reporting and conflict handling.
- [ ] 2.7 Implement optional `.tickets/` settings for ID prefix and documented project behavior overrides.

## 3. Security and Migration Confidence

- [ ] 3.1 Review plugin execution, editor launching, PATH lookup, Beads input parsing, and ticket file writes for security risks.
- [ ] 3.2 Add tests that prevent ticket path traversal and accidental writes outside the active `.tickets/` directory.
- [ ] 3.3 Add migration tests for malformed, partial, and conflicting Beads data.
- [ ] 3.4 Document rollback paths for `.tickets/` compatibility testing and Beads migration.

## 4. Release Pipeline

- [ ] 4.1 Add CI checks for formatting, `go test`, and platform matrix coverage on Windows, macOS, and Linux.
- [ ] 4.2 Add release workflow triggered by version tags.
- [ ] 4.3 Produce archives for supported OS/architecture targets.
- [ ] 4.4 Publish checksums and release notes with every tagged release.

## 5. Documentation

- [ ] 5.1 Expand README with project purpose, compatibility promise, install instructions, and basic usage.
- [ ] 5.2 Add compatibility, migration, plugin, and release documentation.
- [ ] 5.3 Add shell completion generation or installation guidance for supported shells.
- [ ] 5.4 Add attribution to `wedow/ticket` and clarify that `go-ticket` is a compatible reimplementation.
- [ ] 5.5 Add contributor guidance for preserving compatibility and adding command tests.
