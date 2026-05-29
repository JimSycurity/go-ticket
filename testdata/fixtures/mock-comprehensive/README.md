# Mock Comprehensive Fixture

Hand-authored fixture for broad command and parser behavior.

Unlike `upstream-comprehensive`, this fixture intentionally includes malformed
ticket-shaped files so tests can verify that commands report warnings and keep
processing valid tickets.

Valid examples cover:

- parent/child relationships;
- dependency chains and cycles;
- symmetric links;
- open, in_progress, and closed statuses;
- common frontmatter fields and note sections.

Malformed examples cover:

- missing YAML frontmatter;
- missing required `id`;
- malformed inline lists.
