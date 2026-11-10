# contracts

Contracts, bindings, small libraries and conformance tools that independent programs share. Each piece is a small
package, and pieces depend on each other only through contracts.

This first commit is the repository's scaffolding. Packages arrive one pull request at a time, under `pieces/`.

## Layout

- `pieces/<package>/`: one package, with `piece.json` (what it is, what it implements and consumes), a short
  `README.md`, and its schemas, cases or source.
- `.github/workflows/gate.yml`: the checks every pull request runs.

## Rules

See [AGENTS.md](AGENTS.md). Licensed under the Apache License 2.0 ([LICENSE](LICENSE)).
