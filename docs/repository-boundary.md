# ONTOVELA Open Repository Boundary

## Mission

Make ONTOVELA easy to evaluate, integrate, and extend without making the core's operational-state semantics proprietary or ambiguous.

## Compatibility contract

The open repository is the source of public wire-contract releases. Every SDK and adapter must preserve these invariants:

- A state assertion carries source, event time, system time, state kind, and evidence reference.
- `observed`, `reported`, `derived`, `inferred`, `predicted`, and `simulated` remain distinct values.
- Snapshot digests and public schema semantics are compatible with core and EE.
- No SDK convenience API silently upgrades a predicted or simulated value to observed state.

## Publishing policy

- API and schema breaking changes require a new major version.
- Examples must run against the released local developer binary.
- Adapters belong here when they are general-purpose and can be maintained publicly.
- This repository contains no enterprise license checks, tenant policy implementations, or private connector code.
