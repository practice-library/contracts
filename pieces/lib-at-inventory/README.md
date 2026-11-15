# lib-at-inventory

Checks over the repository's pieces. Go package `atinventory`.

| File | What it gives |
|---|---|
| `src/inventory.go` | `Build` (the inventory from every `piece.json`), `Check`, `Canonical` |
| `src/affected.go` | `Affected(inventory, changedPaths)`: the pieces a change touches |

Like `lib-at-core`, it never copies a schema into Go: values are plain maps, validated through
the `Registry`, and the piece and inventory schemas are named by file path in `policy/commons.json`.

- `Check` compares `Build` with `inventory.json` as JSON values and byte for byte against
  `Canonical` (sorted keys, indent 2, trailing newline). A missing file is `commons.input_missing`.
- `Affected` runs no git. A path inside `pieces/<p>/` affects `<p>`; any other path affects every
  piece. Then every piece that consumes an affected one, through `consumes`, `testConsumes` or
  `generatorConsumes`, is added. Only path-form entries count; a schema-id entry is ignored. A
  cycle, or an entry naming a piece not in the inventory, is an error.

It imports no module of its own; the third-party modules come in through `lib-at-core`.
