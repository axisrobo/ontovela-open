# ONTOVELA TypeScript SDK

Dependency-free TypeScript client for the ONTOVELA v0.1 API, built on `fetch`.

```ts
import { OntovelaClient } from "@axisrobo/ontovela";

const client = new OntovelaClient({ baseUrl: "http://localhost:8080", tenantId: "acme" });

const twin = await client.createTwin({ id: "robot:WH-17", type_ref: "robot" });
const state = await client.resolveState("robot:WH-17", "health", { as_of: "2026-08-11T10:00:00Z" });
```

Every request sends `X-Tenant-ID`; operational writes require an explicit idempotency key. Non-2xx responses throw `APIError`.

Run verification:

```powershell
npm install
npm test
```
