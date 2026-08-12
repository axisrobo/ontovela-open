# Design-Partner Onboarding

## Who

Teams evaluating ONTOVELA for autonomous operations, digital twins, or
enterprise reality infrastructure. Focus areas: warehouse robotics, enterprise
incident response, and supply-chain counterfactuals.

## Onboarding track

1. **Sandbox** — run the local developer binary (`v0.53.0+` releases include a
   Windows binary) and `docs/quickstart.md`.
2. **Contract review** — validate `api/openapi.yaml` against your source
   classes; the drift guard and `contract/` fixtures keep SDKs aligned.
3. **Protocol pilot** — pick 1-2 sources and wire them via the adapter matrix;
   `docs/adapter-conformance.md` defines the integrity contract.
4. **Scenario** — adopt a reference template (`examples/incident-response.md`,
   `examples/supply-chain-counterfactual.md`) and reproduce the observe →
   plan → execute → replan loop.
5. **Feedback** — report API gaps, freshness/authority policies, and SLO
   requirements; they drive the GA backlog.

## Success criteria

- Reality Views consumed by a real planner/executor.
- Signed snapshots reconstructed and verified in an audit.
- Simulated branches never contaminate real resolution.
