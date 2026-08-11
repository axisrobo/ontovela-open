# ONTOVELA Open

English | [简体中文](README.zh-CN.md)

Apache-2.0 licensed public developer surface for [ONTOVELA](https://github.com/axisrobo/ONTOVELA), the Digital Enterprise Twin and Operational World Model Platform.

This repository is the public adoption path: stable API contracts, Go/Python/TypeScript SDKs, examples, reference adapters, local developer binaries, documentation, and interoperability test fixtures.

## Scope

Included:

- Versioned public schemas and OpenAPI/event contracts
- Go, Python, and TypeScript SDKs
- Local developer binary and Docker quickstart
- Examples for assertions, temporal query, snapshot, subscriptions, and Reality Views
- HTTP/webhook, Kafka/NATS, SQL/REST, and Harmovela reference adapters
- Contract compatibility and Reality Integrity fixtures

Excluded:

- The ONTOVELA temporal state kernel and reconciliation implementation
- Multi-tenant control plane, cross-region federation, HA, and commercial connectors
- Enterprise identity integration, compliance packs, and premium support tooling

See `docs/repository-boundary.md`.

## v0.1 Contract

- OpenAPI: `api/openapi.yaml`
- Example: `examples/warehouse-robot.md`
- Compatibility policy: `docs/compatibility.md`
- Go SDK: `sdk/go/`
- Python SDK: `sdk/python/`
- TypeScript SDK: `sdk/typescript/`
- Harmovela evidence adapter: `adapters/harmovela/`
- Prediction adapter: `adapters/prediction/`
- HTTP/webhook adapter: `adapters/httpwebhook/`
- Stream ingest adapter (Kafka/NATS): `adapters/stream/`
- SQL/REST polling adapter: `adapters/sqlrest/`
- Edge spool adapter: `adapters/edge/`
- Executor effect adapter: `adapters/effect/`
- MQTT adapter: `adapters/mqtt/`
- ROS 2 adapter: `adapters/ros2/`
- OPC UA adapter: `adapters/opcua/`
- Quickstart: `docs/quickstart.md`
- Reference scenarios: `examples/warehouse-robot.md`, `examples/incident-response.md`, `examples/supply-chain-counterfactual.md`
