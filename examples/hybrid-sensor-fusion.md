# Hybrid Sensor Fusion (UWB + OPC UA + MQTT)

Fuse multiple telemetry protocols into one trusted robot state.

1. **UWB localization** via MQTT writes `robot:WH-17.location` (observed,
   `mqtt:uwb`).
2. **PLC state** via OPC UA writes `robot:WH-17.status` (observed, `opcua:line-1`,
   quality-gated good).
3. **Executive intent** via `effect` writes `robot:WH-17.task` (reported,
   `kinetovela:robot-1`).

Resolution:

- `location` from the highest-authority bound source.
- If two UWB gateways disagree at equal authority, the conflict is explicit
  (`GET /v1/conflicts`) instead of a silent winner.
- A simulated branch (`peiravela`) writes the same properties; `sim-to-real`
  and Reality Views exclude it.

```bash
curl -H 'X-Tenant-ID: acme' \
  http://localhost:8080/v1/twins/robot:WH-17/sim-to-real/location
```

The fused robot remains decision-ready only while every required property is
`fresh` and unconflicted.
