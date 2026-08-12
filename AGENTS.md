# ONTOVELA Open Agent Instructions

## Development Harness

This repository uses the [obra/superpowers](https://github.com/obra/superpowers) OpenCode plugin as its development harness. `opencode.json` loads the upstream methodology. ONTOVELA-specific Reality Integrity and PostgreSQL skills live in the core repository's `.superpowers/skills/`.

Use Superpowers workflows for non-trivial design, planning, test-driven implementation, and verification before completion.

## Engineering Rules

- The public wire contract in `api/openapi.yaml` is the source of truth; the `contract/` drift guard enforces SDK parity.
- SDKs, adapters, and examples must never promote `predicted` or `simulated` state to `observed`.
- Run `GOWORK=off go test ./...` in each Go module, `python -m unittest tests.test_client -v` in `sdk/python`, and `npm test` in `sdk/typescript` before reporting a completed change.
- Tenant scope, idempotency keys, source bindings, and evidence references are integrity invariants in examples and adapters.
- Keep the public surface Apache-2.0; never copy enterprise-only code into this repository.
