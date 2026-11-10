# Guide for contributors and automated agents

Humans and agents follow the same rules here.

## Pieces

- A package lives in `pieces/<package>/` and declares itself in `piece.json`.
- Kinds: contract (schemas and cases, no code), binding (generated), library, tool, plug or runtime.
- A package depends only on contracts, bindings, libraries, the language standard library and pinned, vendored
  third-party modules. A tool never imports another tool, and a plug or runtime never imports another plug or runtime.
- Size: at most 1,500 hand-written production lines per code package and 400 per file; at most 40 schema files per
  contract package; a README of at most 40 lines.
- Generated files carry a "Generated from … Do not edit." header and are never edited by hand.

## Public hygiene

This repository is public.

- Commit messages and pull request text describe the change plainly, with no tool or assistant attribution trailers.
- No personal names, no names of internal systems or private repositories, anywhere.
- Do not state or imply which vendors anyone uses.
- No secrets, tokens or credentials. The leak scan and the denylist guard run on every change.
- Do not merge or approve your own pull request.

The organisation-wide rules are in the
[`.github` repository's AGENTS.md](https://github.com/practice-library/.github/blob/fa14c82916585c4e3357985a0433e60123f3b91d/AGENTS.md).
