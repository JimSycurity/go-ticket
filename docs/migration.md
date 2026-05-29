# Migration and Rollback

## Existing `.tickets` Repositories

Before using write commands on an existing `wedow/ticket` repository:

1. Work on a branch or disposable copy.
2. Run read-only checks first:
   - `gtk list`
   - `gtk ready`
   - `gtk blocked`
   - `gtk query`
3. Review warnings for malformed, unknown-status, symlinked, or non-regular
   ticket files before running write commands.

Rollback for ordinary `.tickets` compatibility testing is to discard the branch
or restore the copied `.tickets` directory.

## Beads Migration

`gtk migrate-beads` imports `.beads/issues.jsonl` from inside the project root.
It treats Beads data as untrusted local input:

- the source must be a regular, non-symlink file under the project root;
- the source and each JSONL line are size-bounded;
- malformed, partial, or unsafe records are reported for review;
- existing ticket IDs are skipped instead of overwritten;
- generated tickets are written through the normal ticket writer.

Use `--dry-run` before writing:

```sh
gtk migrate-beads --dry-run
```

Then import:

```sh
gtk migrate-beads
```

Rollback after Beads migration:

1. Review the command report for `imported <id>` lines.
2. Remove the generated `.tickets/<id>.md` files listed in the report, or
   discard the working branch/copy.
3. Keep the original `.beads/issues.jsonl`; migration does not modify Beads
   data.

If the import report contains `review` lines, inspect those Beads records
manually before rerunning migration or creating replacement tickets by hand.
