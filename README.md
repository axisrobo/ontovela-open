# ONTOVELA

English | [简体中文](README.zh-CN.md)

**Digital Enterprise Twin & Operational World Model Platform**

ONTOVELA fuses physical, digital, organizational, process, agent, and robot
state into an evidence-bearing temporal graph for planning, simulation,
intervention, and closed-loop autonomy.

## What ONTOVELA Solves

Autonomous and enterprise systems currently act on scattered, stale, or
unverifiable "latest values". A planner cannot tell an observation from an
inference, a prediction from a simulation, or an older fact from a newer one.
ONTOVELA makes enterprise reality **computable** — queryable, reconstructable,
and decision-ready — while preserving the distinction between what was
observed, what was inferred, what is predicted, and what exists only in
simulation.

It replaces fragmented IoT dashboards, CMDBs, knowledge graphs, and private
agent state with one authoritative operational-world plane that planners,
executors, robots, and human governors can all trust.

## Key Features

- **Evidence-first state**: every claim carries a source, evidence reference,
  event time, system time, state kind, and confidence.
- **Six state kinds, strictly separated**: `observed`, `reported`, `derived`,
  `inferred`, `predicted`, and `simulated` never conflate; `simulated` state
  can never resolve into real operational state.
- **Bitemporal history**: `as_of` (event time) and `as_known` (system time)
  reconstruct what happened and what the system knew at any point.
- **Explainable resolution**: authority-ranked resolution with explicit
  `unknown` / `conflicted` outcomes — never a silent overwrite.
- **Purpose-bound Reality Views**: planners request only the state they need,
  with per-property freshness gates and source-liveness (heartbeat) checks.
- **Signed snapshots**: immutable, verifiable reality slices for audit, plans,
  and PEIRAVELA simulation branches, with diff and policy-change detection.
- **Impact, causal lineage, and analytics**: traverse dependencies and only
  `causes` relations for impact and causal reasoning.
- **Sim-to-real comparison**: compare real resolved state with a simulated
  branch without contamination.
- **Durable subscriptions**: change feeds, consumer offsets, subscription
  definitions, filters, and audit export.
- **Source accountability**: tenant-scoped, source-bound, idempotent writes
  with optional authenticated source principals.
- **Go backend, PostgreSQL persistence**: append-only bitemporal ledger with
  embedded migrations.

## This Repository: ONTOVELA Open

Apache-2.0 licensed public developer surface for
[ONTOVELA](https://github.com/axisrobo/ONTOVELA) — the public adoption path with
stable API contracts, SDKs, examples, reference adapters, and local developer
binaries.

Included:

- Versioned public schemas and OpenAPI/event contracts
- Go, Python, and TypeScript SDKs
- Local developer binary and Docker quickstart
- Examples for assertions, temporal query, snapshot, subscriptions, and Reality Views
- HTTP/webhook, Kafka/NATS, SQL/REST, MQTT, ROS 2, OPC UA, and Harmovela reference adapters
- Contract compatibility and Reality Integrity fixtures

Excluded (they live in the core or enterprise repositories):

- The ONTOVELA temporal state kernel and reconciliation implementation
- Multi-tenant control plane, cross-region federation, HA, and commercial connectors
- Enterprise identity integration, compliance packs, and premium support tooling

See `docs/repository-boundary.md`.

## Adapters

| Adapter | Ingestion mode | Maps |
| --- | --- | --- |
| `harmovela` | pull (evidence) | `harmovela:event/<id>` evidence validation |
| `httpwebhook` | push (HTTP) | webhook bodies to assertions |
| `stream` | push (Kafka/NATS) | broker messages to assertions |
| `sqlrest` | pull (SQL/REST) | changed rows to assertions |
| `edge` | local buffer | offline spool with JSONL persistence |
| `effect` | push (executor) | EffectRecords to assertions |
| `mqtt` | push (MQTT) | topic messages to assertions |
| `ros2` | push (ROS 2) | robot state to assertions |
| `opcua` | push (OPC UA) | quality-gated node reads |
| `prediction` | pull (model) | predictions as `predicted` assertions |
| `amqp` / `amqp1` | push (AMQP) | broker messages to assertions |
| `coap` | push (CoAP) | constrained-node messages to assertions |
| `grpc` | push (gRPC) | stream messages to assertions |
| `websocket` / `sse` / `longpoll` | push (HTTP) | realtime and long-poll messages to assertions |
| `modbus` | pull/push (Modbus TCP) | register values to assertions |
| `can` | push (CAN bus) | frame data to assertions |
| `bacnet` | push (BACnet) | building-automation objects to assertions |
| `lorawan` | push (LoRaWAN) | uplinks to assertions |
| `ble` | push (BLE) | beacon data to assertions |
| `stomp` | push (STOMP) | broker messages to assertions |
| `zeromq` | push (ZeroMQ) | messages to assertions |
| `redis` | push (Redis Streams) | stream entries to assertions |
| `csv` | pull (file) | CSV records to assertions |
| `graphql` | push (GraphQL) | subscription payloads to assertions |
| `ethernetip` / `profinet` | push (industrial Ethernet) | PLC values to assertions |
| `mqttsn` | push (MQTT-SN) | sensor-network messages to assertions |
| `snmp` | pull (SNMP) | OID values to assertions |
| `dds` | push (DDS) | data-centric topic samples to assertions |

All protocol adapters share the `adapters/base` payload contract and preserve
tenant scope, idempotency, source bindings, evidence references, and state-kind
integrity.

## SDK Coverage Matrix

| Endpoint family | Go | Python | TypeScript |
| --- | --- | --- | --- |
| Twins + lifecycle | yes | yes | yes |
| Source bindings + revoke | yes | yes | yes |
| Assertions + relations + reads | yes | yes | yes |
| Resolved state + Reality View + sim-to-real | yes | yes | yes |
| Snapshots (create/get/list/verify/diff) | yes | yes | yes |
| Change feed + filters + audit export | yes | yes | yes |
| Subscriptions (offsets + definitions) | yes | yes | yes |
| Conflicts + impact + causal | yes | yes | yes |
| Twin types + heartbeats | yes | yes | yes |

Parity is enforced by `contract/` drift guards.

## Quickstart and Contract

- OpenAPI: `api/openapi.yaml`
- Quickstart: `docs/quickstart.md`
- Compatibility policy: `docs/compatibility.md`
- Reference scenarios: `examples/warehouse-robot.md`, `examples/incident-response.md`, `examples/supply-chain-counterfactual.md`
- Releases: tagged GitHub releases include the packaged developer binary for Windows.
