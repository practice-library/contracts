# lib-at-core

Shared plumbing for the acceptance-test (AT) libraries. Go package `atcore`.

| File | What it gives |
|---|---|
| `src/args.go` | `Parse`: `--root` (default `.`), `--base` (optional revision), `--commit` |
| `src/schemas.go` | `LoadPolicy`, the schema `Registry` (`Load`, `Schema`) and `CodedError` |
| `src/validate.go` | `Registry.Validate(path, value)` against the schema at a file path |
| `src/emit.go` | `Emit` (validate, then one JSON line), `Code`, `Exit` |

Schemas are never copied into Go. `Load` reads `pieces/contract-commons-*/schemas/*.schema.json`
under `--root` at runtime (only the commons contract packages), takes the `$id` from each file
and registers it, so a schema is named by its file path. Reports are plain maps, validated
against the report schema named in `policy/commons.json` when they are emitted; no Go type
mirrors a schema.

- `--commit`, when given, must be 40 lowercase hex digits; an unknown flag is an error.
- `Code(r, name)` returns `name` only if it is in the code table's `$defs.code` enum, read at
  runtime from the schema named by `codeTable` in the policy.
- `Exit` is 0 when every report's status is PASS, and 1 otherwise, or for an empty list.
- A missing schema (also a referenced one) or policy file is exit 1 with code
  `commons.input_missing`.

The package consumes the contract packages `contract-commons-at-report` and
`contract-commons-piece` only at runtime, from the checkout. Its third-party modules are the
JSON Schema validator and its text dependency. They are not vendored: the build downloads them
and checks each against `go.sum`. `thirdParty[].path` is the module's Go vendor-layout path.
