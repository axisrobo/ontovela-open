# Contributing to ONTOVELA Open

## Scope

ONTOVELA-open hosts the Apache-2.0 public surface: API contracts, SDKs,
adapters, examples, and developer documentation. Core runtime changes belong in
the AGPL ONTOVELA repository; enterprise capabilities belong in ONTOVELA-ee.

## Requirements

- Every new or changed endpoint must update `api/openapi.yaml` and the Go,
  Python, and TypeScript SDKs, and keep the `contract/` drift guard green.
- Every adapter must preserve tenant scope, idempotency, source bindings,
  evidence references, and state-kind integrity.
- Run the full verification matrix:
  - `GOWORK=off go test ./...` per Go module
  - `python -m unittest tests.test_client -v`
  - `npm test`
  - `GOWORK=off go test ./...` in `contract/`

## Process

1. Add a focused test first.
2. Implement the minimal change.
3. Verify all checks pass.
4. Open a pull request describing the contract and integrity impact.
