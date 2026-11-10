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
- `implements`, `consumes`, `testConsumes`, `generatorConsumes`: lists of unique package paths
  (`pieces/<name>`) or versioned URNs.
- `thirdParty`: vendored modules, each with `name`, an exact `version`, a lowercase hex `sha256`
  and the vendored `path`.
- `owner`: `<organisation>/<team>`. `tier`: an integer, 0 or more.

## inventory

Describes the generated `inventory.json`: an object with a `pieces` list. Each item is the
package `path` (`pieces/<name>`) and its `piece`, a whole piece document. The inventory schema
refers to the piece schema only by its URN, so a consumer resolves both from a set of URNs.

## Cases

`cases/<schema>/<keyword>-pass.json` is valid and `<keyword>-fail.json` is invalid for that
keyword only. There is one pair for each JSON Schema keyword the schema uses, under
`cases/piece/` and `cases/inventory/`. A validator should accept every pass file and reject
every fail file.
