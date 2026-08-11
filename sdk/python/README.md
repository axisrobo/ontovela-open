# ONTOVELA Python SDK

Standard-library Python client for the ONTOVELA v0.1 API.

```python
from datetime import datetime, timezone
from ontovela import Client, StateAssertion

client = Client("http://localhost:8080", "acme")

client.create_twin("robot:WH-17", "robot")
client.create_source_binding(...)

client.append_assertion(
    StateAssertion(
        subject_id="robot:WH-17", property="health", value="ready",
        state_kind="observed",
        event_time=datetime(2026, 8, 11, 10, 0, 0, tzinfo=timezone.utc),
        source="sensor:health", evidence_ref="event:9f2",
    ),
    idempotency_key="event-9f2",
)
```

Every request sends `X-Tenant-ID`; operational writes require an explicit
idempotency key. Non-2xx responses raise `ontovela.APIError`.

Run verification:

```powershell
python -m unittest tests.test_client -v
```
