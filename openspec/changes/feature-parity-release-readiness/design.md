## Context

The MVP focuses on a native Go port of the highest-value ticket workflows. The follow-up parity phase turns that usable core into a credible replacement for upstream `wedow/ticket`, including commands that rely on shell behavior today, release engineering, documentation, and migration confidence.

## Goals / Non-Goals

**Goals:**
- Reach solid command and file-format compatibility for the documented upstream command surface.
- Preserve a single-binary native installation story across Windows, macOS, and Linux.
- Provide compatibility fixtures and tests that make behavioral drift visible.
- Ship versioned release artifacts with checksums and clear installation instructions.
- Document migration and rollback for existing `.tickets/` repositories.

**Non-Goals:**
- Perfect byte-for-byte output compatibility for every edge case unless tests prove users depend on it.
- Compatibility with arbitrary shell-only plugins on native Windows without adaptation.
- Adding server, daemon, database, or hosted issue-tracker behavior.
- Replacing project-specific VS Code extension work.

## Decisions

- Treat upstream README usage and behavior tests as the parity target.
  - Rationale: The project should be compatible with observable behavior and file format, not line-for-line Bash internals.
  - Alternative considered: Fork and translate the Bash script mechanically; rejected because that would preserve POSIX assumptions.

- Implement plugin lookup as a cross-platform command resolution layer.
  - Rationale: Unix users expect `tk-<cmd>` and `ticket-<cmd>` from PATH; Windows users need `.exe`, `.cmd`, `.bat`, and optionally `.ps1` handling.
  - Alternative considered: Drop plugins; rejected for parity, but plugin execution should receive explicit security review.

- Support `query` through either embedded jq-compatible evaluation or a documented external `jq` fallback.
  - Rationale: `query` is useful for agents and scripts, but native Windows should not require users to install Unix tooling for basic JSON output.
  - Alternative considered: JSON-only query with no filter language; acceptable only if documented as a compatibility limitation.

- Add an optional lightweight settings file under `.tickets/` for configurable behavior.
  - Rationale: Prefixes and project-level behavior should be customizable without changing ticket files or requiring environment variables, while absent settings preserve upstream-compatible defaults.
  - Alternative considered: Add settings during MVP; rejected because default compatibility matters more than configuration before the core is stable.

- Use release automation rather than manual artifact creation.
  - Rationale: Windows support is the point of the project; every release should prove Windows builds and tests still work.

## Risks / Trade-offs

- Plugin execution can run arbitrary PATH commands -> require explicit docs, tests, and security review before enabling.
- `query` compatibility may pull in a nontrivial dependency -> isolate evaluator logic behind a small interface.
- Beads migration input may be messy or untrusted -> parse defensively and never overwrite existing tickets without explicit behavior.
- Release workflows can fail from platform-specific packaging assumptions -> build simple archives first, then add installers later if needed.
- Compatibility can become vague -> maintain a visible command compatibility matrix in docs.

## Migration Plan

Users should be able to install `go-ticket`, run read-only parity checks against existing `.tickets/`, and then use write commands on a branch. Migration from Beads remains a separate command that imports `.beads/issues.jsonl` into `.tickets/`; rollback is preserving the original Beads data and removing generated ticket files if needed.

## Open Questions

- Should `query` embed a jq-compatible evaluator or require external `jq` only for filtered queries?
- Should release artifacts include shell completions, package manager manifests, or just archives/checksums at first?
- How much plugin compatibility should be enabled by default on Windows, especially PowerShell scripts?
- Should parity tests invoke the upstream Bash `ticket` script on POSIX as a differential oracle?
