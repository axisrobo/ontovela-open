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
