# Contributing

## Project Roles

- `grav-charge`: traditional/reference engine.
- `para-nbody`: experimental engine and API work.
- `docs/`: shared planning and decision records.

## Development Workflow

Work in small, testable changes.

Before starting a change:

1. Check current milestone.
2. Inspect only relevant files.
3. Prefer minimal edits.
4. Avoid unrelated refactors.

Before finishing a change:

1. Run targeted tests.
2. Run `go test ./...` when practical.
3. Update docs if behavior or architecture changed.

## Multi-AI Collaboration Rules

This project may be edited with help from multiple AI tools.

Rules:

1. The repository is the source of truth.
2. Tests are the referee.
3. Prefer small commits.
4. Do not make broad rewrites without an explicit milestone.
5. Document design decisions in `docs/decisions.md`.
6. If another AI generated code, review it before building on it.
7. Avoid simultaneous conflicting edits.

## Go Style

- Prefer simple, idiomatic Go.
- Keep packages small and focused.
- Favor deterministic tests.
- Avoid hidden global state in simulation logic.
- Pass configuration explicitly.
- Use `context.Context` for API/server boundaries where appropriate.
- Return errors rather than panicking in library code.

## Testing

Minimum expectation:
