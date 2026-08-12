import { createServer, type Server } from "node:http";
import { strict as assert } from "node:assert";
import { test, before, after } from "node:test";
import { APIError, OntovelaClient } from "../client";

let server: Server;
let baseUrl: string;

function sendJson(response: import("node:http").ServerResponse, status: number, value: unknown): void {
  const payload = JSON.stringify(value);
  response.writeHead(status, { "Content-Type": "application/json", "Content-Length": Buffer.byteLength(payload) });
  response.end(payload);
}

before(async () => {
  server = createServer((request, response) => {
    if (request.method === "POST" && request.url === "/v1/assertions") {
      let body = "";
      request.on("data", (chunk) => (body += chunk));
      request.on("end", () => {
        if (request.headers["x-tenant-id"] !== "acme" || !request.headers["idempotency-key"]) {
          sendJson(response, 400, { error: "missing tenant or idempotency" });
          return;
        }
        const parsed = JSON.parse(body) as Record<string, unknown>;
        parsed.id = "assertion-1";
        parsed.tenant_id = "acme";
        parsed.system_time = parsed.event_time;
        sendJson(response, 201, parsed);
      });
      return;
    }
    if (request.method === "GET" && request.url?.startsWith("/v1/twins/twin-1/state/health")) {
      if (!request.url.includes("as_of=")) {
        sendJson(response, 400, { error: "missing as_of" });
        return;
      }
      sendJson(response, 200, { tenant_id: "acme", subject_id: "twin-1", property: "health", status: "resolved", freshness: "fresh", resolution_policy: "p", value: "ready" });
      return;
    }
    if (request.method === "GET" && request.url?.startsWith("/v1/twins/twin-1/impact")) {
      sendJson(response, 200, { impact_paths: [{ subject_id: "twin-1", depth: 1, relation_id: "r1", predicate: "located_in", target_id: "zone-1", state_kind: "observed", source: "map", evidence_ref: "e1" }] });
      return;
    }
    if (request.method === "GET" && request.url?.startsWith("/v1/twins/twin-1/sim-to-real/")) {
      sendJson(response, 200, { tenant_id: "acme", twin_id: "twin-1", property: "health", real_state: {}, simulated_state: {}, delta: "diverges" });
      return;
    }
    if (request.method === "GET" && request.url?.startsWith("/v1/twins/twin-1/causal/analytics")) {
      sendJson(response, 200, { twin_id: "twin-1", fan_out: 3, fan_in: 1, top_targets: { "bin-1": 2 } });
      return;
    }
    if (request.method === "GET" && request.url?.startsWith("/v1/twins/twin-1/causal")) {
      sendJson(response, 200, { causal_links: [{ subject_id: "twin-1", depth: 1, relation_id: "r1", mechanism: "motor_failure", target_id: "bin-1", event_time: "2026-08-11T10:00:00Z", source: "map", evidence_ref: "e1" }] });
      return;
    }
    if (request.method === "GET" && request.url?.startsWith("/v1/twins/twin-1/snapshots")) {
      sendJson(response, 200, { snapshots: [{ id: "s1", tenant_id: "acme", subject_id: "twin-1", resolution_policy: "p", digest: "d", states: [], relations: [], created_at: "2026-08-11T00:00:00Z" }] });
      return;
    }
    if (request.method === "POST" && request.url?.endsWith("/lifecycle")) {
      let body = "";
      request.on("data", (chunk) => (body += chunk));
      request.on("end", () => {
        const parsed = JSON.parse(body) as { lifecycle: string };
        sendJson(response, 200, { id: "twin-1", type_ref: "robot", lifecycle: parsed.lifecycle, tenant_id: "acme" });
      });
      return;
    }
    if (request.method === "POST" && request.url === "/v1/heartbeats") {
      let body = "";
      request.on("data", (chunk) => (body += chunk));
      request.on("end", () => {
        const parsed = JSON.parse(body) as { source: string };
        sendJson(response, 200, { tenant_id: "acme", source: parsed.source, last_heartbeat_at: "2026-08-11T10:00:00Z" });
      });
      return;
    }
    if (request.method === "GET" && request.url?.startsWith("/v1/source-bindings")) {
      sendJson(response, 200, { source_bindings: [{ id: "b1", source: "sensor:health", property: "health", authority_rank: 1, max_lag_seconds: 60, tenant_id: "acme" }] });
      return;
    }
    if (request.method === "DELETE" && request.url === "/v1/source-bindings/b1") {
      response.writeHead(204);
      response.end();
      return;
    }
    if (request.method === "GET" && request.url === "/v1/twin-types") {
      sendJson(response, 200, { twin_types: [{ type_ref: "robot", description: "Robot twin", properties: ["location"], relations: ["located_in"] }] });
      return;
    }
    if (request.method === "GET" && request.url === "/v1/twins/missing") {
      sendJson(response, 404, { error: "not found" });
      return;
    }
    sendJson(response, 404, { error: "not found" });
  });
  await new Promise<void>((resolve) => server.listen(0, "127.0.0.1", resolve));
  const address = server.address() as { port: number };
  baseUrl = `http://127.0.0.1:${address.port}`;
});

after(() => server.close());

test("appendAssertion sends tenant and idempotency headers", async () => {
  const client = new OntovelaClient({ baseUrl, tenantId: "acme" });
  const claim = await client.appendAssertion(
    { subject_id: "twin-1", property: "health", value: "ready", state_kind: "observed", event_time: "2026-08-11T10:00:00Z", source: "sensor:health", evidence_ref: "event:1" },
    "event-1",
  );
  assert.equal(claim.id, "assertion-1");
});

test("resolveState sends bitemporal query", async () => {
  const client = new OntovelaClient({ baseUrl, tenantId: "acme" });
  const state = await client.resolveState("twin-1", "health", { as_of: "2026-08-11T10:00:00Z", as_known: "2026-08-11T10:00:00Z" });
  assert.equal(state.status, "resolved");
});

test("APIError carries status and message", async () => {
  const client = new OntovelaClient({ baseUrl, tenantId: "acme" });
  await assert.rejects(
    () => client.getTwin("missing"),
    (error: unknown) => error instanceof APIError && error.status === 404 && error.serverMessage === "not found",
  );
});

test("end-to-end warehouse flow", async () => {
  const events: Array<Record<string, unknown>> = [];
  const fake = createServer((request, response) => {
    const path = decodeURIComponent((request.url ?? "").split("?")[0]);
    const send = (status: number, value?: unknown) => {
      const payload = JSON.stringify(value ?? {});
      response.writeHead(status, { "Content-Type": "application/json", "Content-Length": Buffer.byteLength(payload) });
      response.end(payload);
    };
    if (request.method === "POST" && path === "/v1/twins") { send(201, { id: "robot:WH-17", type_ref: "robot", tenant_id: "acme" }); return; }
    if (request.method === "POST" && path === "/v1/source-bindings") { send(201, { id: "b1", tenant_id: "acme" }); return; }
    if (request.method === "POST" && path === "/v1/assertions") {
      let body = "";
      request.on("data", (chunk) => (body += chunk));
      request.on("end", () => {
        const parsed = JSON.parse(body) as Record<string, unknown>;
        events.push({ offset: events.length + 1, tenant_id: "acme", kind: "state_assertion.appended", subject_id: parsed.subject_id, payload: parsed, occurred_at: "2026-08-11T10:00:00Z" });
        send(201, { ...parsed, id: "a1", tenant_id: "acme", system_time: parsed.event_time });
      });
      return;
    }
    if (request.method === "GET" && path.startsWith("/v1/twins/robot:WH-17/state/health")) { send(200, { tenant_id: "acme", subject_id: "robot:WH-17", property: "health", status: "resolved", freshness: "fresh", resolution_policy: "p", value: "ready" }); return; }
    if (request.method === "POST" && path.endsWith("/snapshots")) { send(201, { id: "snap-1", tenant_id: "acme", subject_id: "robot:WH-17", resolution_policy: "p", digest: "d1", states: [], relations: [], created_at: "2026-08-11T00:00:00Z" }); return; }
    if (request.method === "GET" && path.includes("/verify")) { send(200, { valid: true }); return; }
    if (request.method === "GET" && path === "/v1/changes") { send(200, { events }); return; }
    send(404, { error: "not found" });
  });
  await new Promise<void>((resolve) => fake.listen(0, "127.0.0.1", resolve));
  const address = fake.address() as { port: number };
  try {
    const client = new OntovelaClient({ baseUrl: `http://127.0.0.1:${address.port}`, tenantId: "acme" });
    await client.createTwin({ id: "robot:WH-17", type_ref: "robot" });
    const claim = await client.appendAssertion({ subject_id: "robot:WH-17", property: "health", value: "ready", state_kind: "observed", event_time: "2026-08-11T10:00:00Z", source: "sensor:health", evidence_ref: "e1" }, "k1");
    assert.equal(claim.id, "a1");
    const state = await client.resolveState("robot:WH-17", "health");
    assert.equal(state.status, "resolved");
    const snapshot = await client.createSnapshot("robot:WH-17");
    assert.equal(snapshot.id, "snap-1");
    assert.equal(await client.verifySnapshot("snap-1"), true);
    const changeEvents = await client.listChanges();
    assert.equal(changeEvents.length, 1);
  } finally {
    fake.close();
  }
});

test("list and revoke source bindings", async () => {
  const client = new OntovelaClient({ baseUrl, tenantId: "acme" });
  const bindings = await client.listSourceBindings("sensor:health");
  assert.equal(bindings.length, 1);
  assert.equal(bindings[0].id, "b1");
  await client.revokeSourceBinding("b1");
});

test("updateTwinLifecycle posts lifecycle", async () => {
  const client = new OntovelaClient({ baseUrl, tenantId: "acme" });
  const twin = await client.updateTwinLifecycle("twin-1", "retired");
  assert.equal(twin.lifecycle, "retired");
});

test("simToReal parses delta", async () => {
  const client = new OntovelaClient({ baseUrl, tenantId: "acme" });
  const delta = await client.simToReal("twin-1", "health");
  assert.equal(delta.delta, "diverges");
});

test("computeCausalAnalytics parses counts", async () => {
  const client = new OntovelaClient({ baseUrl, tenantId: "acme" });
  const analytics = await client.computeCausalAnalytics("twin-1");
  assert.equal(analytics.fan_out, 3);
  assert.equal(analytics.top_targets["bin-1"], 2);
});

test("computeCausalLineage parses links", async () => {
  const client = new OntovelaClient({ baseUrl, tenantId: "acme" });
  const links = await client.computeCausalLineage("twin-1", 5);
  assert.equal(links.length, 1);
  assert.equal(links[0].mechanism, "motor_failure");
});

test("listSnapshots parses list", async () => {
  const client = new OntovelaClient({ baseUrl, tenantId: "acme" });
  const snapshots = await client.listSnapshots("twin-1", 10);
  assert.equal(snapshots.length, 1);
  assert.equal(snapshots[0].id, "s1");
});

test("reportHeartbeat posts source", async () => {
  const client = new OntovelaClient({ baseUrl, tenantId: "acme" });
  const heartbeat = await client.reportHeartbeat("sensor:health");
  assert.equal(heartbeat.source, "sensor:health");
});

test("retries transient errors", async () => {
  let attempts = 0;
  const client = new OntovelaClient({
    baseUrl,
    tenantId: "acme",
    maxRetries: 3,
    fetch: async () => {
      attempts++;
      if (attempts < 3) {
        return new Response(JSON.stringify({ error: "unavailable" }), { status: 503 });
      }
      return new Response(JSON.stringify({ status: "resolved" }), { status: 200 });
    },
  });
  const state = await client.resolveState("twin-1", "health", { as_of: "2026-08-11T10:00:00Z" });
  assert.equal(state.status, "resolved");
  assert.equal(attempts, 3);
});

test("listTwinTypes parses packs", async () => {
  const client = new OntovelaClient({ baseUrl, tenantId: "acme" });
  const types = await client.listTwinTypes();
  assert.equal(types.length, 1);
  assert.equal(types[0].type_ref, "robot");
});

test("computeImpact parses impact paths", async () => {
  const client = new OntovelaClient({ baseUrl, tenantId: "acme" });
  const paths = await client.computeImpact("twin-1", { maxDepth: 3, predicate: "located_in" });
  assert.equal(paths.length, 1);
  assert.equal(paths[0].target_id, "zone-1");
});

test("tenantId is required", () => {
  assert.throws(() => new OntovelaClient({ baseUrl, tenantId: " " }), /tenantId is required/);
});
