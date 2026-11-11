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

- `at`: `AT-01` to `AT-18`. `command`: the command that ran, not empty. `commit`: 40 hex digits.
- `status`: PASS, FAIL or NOT_RUN. `code` is null exactly when `status` is PASS.
- `counts`: named non-negative integers. `findings`: each a `message`, optionally `path`, `line`.

## codes

The outcome of the whole run, `{exit, code}`. Exit 0: every AT run passed. Exit 1: any AT is
FAIL or NOT_RUN, or the run could not complete (for example a missing input or schema).
`code` is null exactly when `exit` is 0, and otherwise one of the four `commons.*` codes
defined in `$defs`. The report schema refers to them by URN, not by file path.

## Cases

`cases/<schema>/<keyword>-pass.json` must be valid and `<keyword>-fail.json` must be invalid.
There is one such pair for each assertion or applicator keyword the schema uses; identifiers
and annotations (`$id`, `$schema`, `title`) have no pair. A fail case for an applicator
(`properties`, `items`, `anyOf`, `$ref`, `then`, `else`) fails through a sub-schema under it.
`if` has no failure of its own: its fail case fails through the branch it selects.
Further cases are named `<keyword>-<detail>-pass.json` or `-fail.json`. A validator should
accept every `*-pass.json` and reject every `*-fail.json`.
