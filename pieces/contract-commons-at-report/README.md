# contract-commons-at-report

Two JSON Schemas (draft 2020-12) that describe the outcome of acceptance-test (AT) runs.

| Schema | File | Version |
|---|---|---|
| at-report | `schemas/at-report.schema.json` | 0.1.0 |
| codes | `schemas/codes.schema.json` | 0.1.0 |

Each schema's URN is the `$id` in its file. A contract package's own URNs are those `$id`s,
so its `piece.json` lists nothing under `implements`.

## at-report

One AT result; every field is required and no other field is allowed.

- `at`: `AT-01` to `AT-18`. `command`: the command that ran, not empty. `commit`: 40 lowercase
  hex digits. `counts`: named non-negative integers.
- `status`: PASS, FAIL or NOT_RUN. `code` is null for PASS and one of the four `commons.*`
  codes for NOT_RUN. A FAIL with a verdict has a null `code` and says why in `findings`; a
  FAIL with no verdict possible (for example a missing input) carries the code.
- `findings`: each a `message`, optionally a `path`, and a `line` only together with a `path`.

## codes

`$defs` holds the four `commons.*` codes and the exit codes. Exit 0: every AT run passed.
Exit 1: any AT is FAIL or NOT_RUN, or the run could not complete (for example a missing
schema). The root object `{exit, code}` records how a run ended: `code` is null for exit 0,
and null or a code for exit 1. Nothing in W0 prints that object; it states the relation.

## Cases

`cases/<schema>/<keyword>-pass.json` must be valid and `<keyword>-fail.json` must be invalid.
There is one such pair for each assertion or applicator keyword the schema uses; identifiers,
containers and annotations (`$id`, `$schema`, `$defs`, `title`, any `x-*`) have no pair. A fail
case for an applicator (`properties`, `items`, `anyOf`, `$ref`, `then`, `else`) fails through a
sub-schema under it.
`if` has no failure of its own: its fail case fails through the branch it selects.
Further cases are named `<keyword>-<detail>-pass.json` or `-fail.json`. A validator should
accept every `*-pass.json` and reject every `*-fail.json`.
