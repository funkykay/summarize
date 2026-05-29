# Technical documentation: summarize

`summarize` is a Go CLI that walks a directory tree and emits selected files as deterministic, machine-readable plain text.

## Chapters

- [Overview](overview.md) describes purpose, scope, and non-goals.
- [CLI usage](cli.md) documents commands, options, output, and exit behavior.
- [Configuration](configuration.md) documents `summarize.json`, profiles, merge behavior, and patterns.
- [Architecture](architecture.md) describes the Go packages and their responsibilities.
- [Data flow](data-flow.md) follows execution from CLI invocation to output.
- [Update mechanism](update.md) documents the GitHub Release based standalone-binary updater.
- [Development](development.md) documents build, test, and project conventions.
- [Risks and technical debt](risks.md) lists known limitations and open decisions.
- [Glossary](glossary.md) defines project-specific terms.

## Recommended reading paths

For usage: [CLI usage](cli.md) and [Configuration](configuration.md).

For implementation work: [Architecture](architecture.md), [Data flow](data-flow.md), and [Development](development.md).

For release or update changes: [Update mechanism](update.md) and [Risks and technical debt](risks.md).


