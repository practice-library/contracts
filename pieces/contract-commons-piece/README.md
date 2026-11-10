# contract-commons-piece

Two JSON Schemas (draft 2020-12) that describe the repository's package metadata.

| Schema | File | Version |
|---|---|---|
| piece | `schemas/piece.schema.json` | 0.1.0 |
| inventory | `schemas/inventory.schema.json` | 0.1.0 |

Each schema's URN is the `$id` in its file. A contract package's own URNs are those `$id`s,
so its `piece.json` lists nothing under `implements`; that field names the contracts a code
package implements.

## piece

Describes one `piece.json`. Every field is required and no other field is allowed.

- `kind`: contract, binding, library, tool, plug or runtime.
- `implements`: unique versioned contract URNs. `consumes`, `testConsumes`,
  `generatorConsumes`: unique package paths (`pieces/<name>`) or versioned URNs.
- `thirdParty`: unique vendored modules, each with `name`, an exact `version` (semver with an
  optional `v`), a lowercase hex `sha256` and the vendored `path` under `vendor/`.
- `owner`: `<organisation>/<team>`. `tier`: an integer, 0 or more.

## inventory

Describes the generated `inventory.json`: an object with a `pieces` list. Each item is the
package `path` (`pieces/<name>`) and its `piece`, a whole piece document. The inventory schema
refers to the piece schema only by its URN, so a consumer resolves both from a set of URNs.

## Cases

`cases/<schema>/<keyword>-pass.json` must be valid and `<keyword>-fail.json` must be invalid.
There is one such pair for each assertion or applicator keyword the schema uses; identifiers
and annotations (`$id`, `$schema`, `title`, any `x-*`) have no pair. A fail case for an
applicator (`properties`, `items`, `$ref`) fails through a sub-schema under that keyword.
Further cases are named `<keyword>-<detail>-pass.json` or `-fail.json`. A validator should
accept every `*-pass.json` and reject every `*-fail.json`.
