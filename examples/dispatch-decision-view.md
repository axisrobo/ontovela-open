# Dispatch Decision Reality View

An ORCHADYN-style planner resolves a decision-ready Reality View before
dispatching a robot.

```bash
curl -X POST http://localhost:8080/v1/reality-views \
  -H 'X-Tenant-ID: acme' -H 'Content-Type: application/json' \
  -d '{
    "twin_id": "robot:WH-17",
    "purpose": "dispatch replenishment",
    "authorization_ref": "grant:8a2",
    "required_state": [
      {"property": "location", "max_age_seconds": 5},
      {"property": "health", "max_age_seconds": 60}
    ]
  }'
```

A `ready` status means every required property resolved and is within the
purpose window. `stale`, `unknown`, or `conflicted` stops the planner from
acting on unreliable reality. A dead source (no heartbeat) forces `stale`
even if the event time is recent.
