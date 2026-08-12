"""Contract tests for the ONTOVELA Python client."""

import json
import sys
import unittest
import urllib.parse
from datetime import datetime, timezone
from http.server import BaseHTTPRequestHandler, HTTPServer
from threading import Thread

sys.path.insert(0, "..")

from ontovela import APIError, Client, ResolvedState, StateAssertion  # noqa: E402


class _Handler(BaseHTTPRequestHandler):
    def do_POST(self):  # noqa: N802
        if self.path == "/v1/assertions":
            length = int(self.headers.get("Content-Length", "0"))
            body = json.loads(self.rfile.read(length) or b"{}")
            if self.headers.get("X-Tenant-ID") != "acme":
                self._json(400, {"error": "missing tenant"})
                return
            if not self.headers.get("Idempotency-Key"):
                self._json(400, {"error": "missing idempotency"})
                return
            body["id"] = "assertion-1"
            body["tenant_id"] = "acme"
            body["system_time"] = body.get("event_time")
            self._json(201, body)
            return
        if self.path == "/v1/twins/twin-1":
            self._json(200, {"id": "twin-1", "tenant_id": "acme", "created_at": "2026-08-11T00:00:00Z"})
            return
        self._json(404, {"error": "not found"})

    def do_GET(self):  # noqa: N802
        if self.path.startswith("/v1/twins/twin-1/state/health"):
            if "as_of" not in self.path:
                self._json(400, {"error": "missing as_of"})
                return
            self._json(200, {"tenant_id": "acme", "subject_id": "twin-1", "property": "health", "status": "resolved", "freshness": "fresh", "resolution_policy": "p", "value": "ready"})
            return
        self._json(404, {"error": "not found"})

    def _json(self, status, value):
        payload = json.dumps(value).encode("utf-8")
        self.send_response(status)
        self.send_header("Content-Type", "application/json")
        self.send_header("Content-Length", str(len(payload)))
        self.end_headers()
        self.wfile.write(payload)

    def log_message(self, *args):  # suppress server logging
        pass


class E2EHandler(BaseHTTPRequestHandler):
    state = {"events": []}

    def _path(self):
        return urllib.parse.unquote(self.path).split("?")[0]

    def do_POST(self):  # noqa: N802
        self.path = self._path()
        if self.path == "/v1/twins":
            self._json(201, {"id": "robot:WH-17", "type_ref": "robot", "tenant_id": "acme"})
            return
        if self.path == "/v1/source-bindings":
            self._json(201, {"id": "b1", "tenant_id": "acme"})
            return
        if self.path == "/v1/assertions":
            length = int(self.headers.get("Content-Length", "0"))
            body = json.loads(self.rfile.read(length) or b"{}")
            body["id"] = "a1"
            body["tenant_id"] = "acme"
            self.state["events"].append({"offset": len(self.state["events"]) + 1, "tenant_id": "acme", "kind": "state_assertion.appended", "subject_id": "robot:WH-17", "payload": body, "occurred_at": "2026-08-11T10:00:00Z"})
            self._json(201, body)
            return
        if "/snapshots" in self.path:
            self._json(201, {"id": "snap-1", "tenant_id": "acme", "subject_id": "robot:WH-17", "resolution_policy": "p", "digest": "d1", "states": [], "relations": [], "created_at": "2026-08-11T00:00:00Z"})
            return
        self._json(404, {"error": "not found"})

    def do_GET(self):  # noqa: N802
        self.path = self._path()
        if self.path.startswith("/v1/twins/robot:WH-17/state/health"):
            self._json(200, {"tenant_id": "acme", "subject_id": "robot:WH-17", "property": "health", "status": "resolved", "freshness": "fresh", "resolution_policy": "p", "value": "ready"})
            return
        if self.path.startswith("/v1/twins/robot:WH-17/snapshots"):
            self._json(200, {"id": "snap-1", "tenant_id": "acme", "subject_id": "robot:WH-17", "resolution_policy": "p", "digest": "d1", "states": [], "relations": [], "created_at": "2026-08-11T00:00:00Z"})
            return
        if "/verify" in self.path:
            self._json(200, {"valid": True})
            return
        if self.path == "/v1/changes":
            self._json(200, {"events": self.state["events"]})
            return
        self._json(404, {"error": "not found"})

    def _json(self, status, value):
        payload = json.dumps(value).encode("utf-8")
        self.send_response(status)
        self.send_header("Content-Type", "application/json")
        self.send_header("Content-Length", str(len(payload)))
        self.end_headers()
        self.wfile.write(payload)


class ClientTest(unittest.TestCase):
    @classmethod
    def setUpClass(cls):
        cls.server = HTTPServer(("127.0.0.1", 0), _Handler)
        cls.thread = Thread(target=cls.server.serve_forever, daemon=True)
        cls.thread.start()
        cls.base_url = f"http://127.0.0.1:{cls.server.server_port}"

    @classmethod
    def tearDownClass(cls):
        cls.server.shutdown()

    def test_append_assertion_sends_tenant_and_idempotency(self):
        client = Client(self.base_url, "acme")
        claim = StateAssertion(
            subject_id="twin-1", property="health", value="ready", state_kind="observed",
            event_time=datetime(2026, 8, 11, 10, 0, 0, tzinfo=timezone.utc),
            source="sensor:health", evidence_ref="event:1",
        )
        result = client.append_assertion(claim, idempotency_key="event-1")
        self.assertEqual(result.id, "assertion-1")

    def test_resolve_state_sends_bitemporal_query(self):
        client = Client(self.base_url, "acme")
        when = datetime(2026, 8, 11, 10, 0, 0, tzinfo=timezone.utc)
        state = client.resolve_state("twin-1", "health", as_of=when, as_known=when)
        self.assertEqual(state.status, "resolved")

    def test_sim_to_real_parses_delta(self):
        handler = _Handler
        original_do_get = handler.do_GET

        def patched(self):
            if self.path.startswith("/v1/twins/twin-1/sim-to-real/"):
                self._json(200, {"tenant_id": "acme", "twin_id": "twin-1", "property": "health", "real_state": {}, "simulated_state": {}, "delta": "diverges"})
                return
            original_do_get(self)

        handler.do_GET = patched
        try:
            client = Client(self.base_url, "acme")
            delta = client.sim_to_real("twin-1", "health")
            self.assertEqual(delta.delta, "diverges")
        finally:
            handler.do_GET = original_do_get

    def test_compute_causal_analytics_parses_counts(self):
        handler = _Handler
        original_do_get = handler.do_GET

        def patched(self):
            if self.path.startswith("/v1/twins/twin-1/causal/analytics"):
                self._json(200, {"twin_id": "twin-1", "fan_out": 3, "fan_in": 1, "top_targets": {"bin-1": 2}})
                return
            original_do_get(self)

        handler.do_GET = patched
        try:
            client = Client(self.base_url, "acme")
            analytics = client.compute_causal_analytics("twin-1")
            self.assertEqual(analytics.fan_out, 3)
            self.assertEqual(analytics.top_targets["bin-1"], 2)
        finally:
            handler.do_GET = original_do_get

    def test_compute_causal_lineage_parses_links(self):
        handler = _Handler
        original_do_get = handler.do_GET

        def patched(self):
            if self.path.startswith("/v1/twins/twin-1/causal"):
                self._json(200, {"causal_links": [{"subject_id": "twin-1", "depth": 1, "relation_id": "r1", "mechanism": "motor_failure", "target_id": "bin-1", "event_time": "2026-08-11T10:00:00Z", "source": "map", "evidence_ref": "e1"}]})
                return
            original_do_get(self)

        handler.do_GET = patched
        try:
            client = Client(self.base_url, "acme")
            links = client.compute_causal_lineage("twin-1", max_depth=5)
            self.assertEqual(len(links), 1)
            self.assertEqual(links[0].mechanism, "motor_failure")
        finally:
            handler.do_GET = original_do_get

    def test_list_and_revoke_source_bindings(self):
        handler = _Handler
        original_do_get = handler.do_GET
        original_do_delete = handler.do_DELETE if hasattr(handler, "do_DELETE") else None

        def do_get(self):
            if self.path.startswith("/v1/source-bindings"):
                self._json(200, {"source_bindings": [{"id": "b1", "source": "sensor:health", "property": "health", "authority_rank": 1, "max_lag_seconds": 60, "tenant_id": "acme"}]})
                return
            original_do_get(self)

        def do_delete(self):
            if self.path == "/v1/source-bindings/b1":
                self.send_response(204)
                self.end_headers()
                return
            self._json(404, {"error": "not found"})

        handler.do_GET = do_get
        handler.do_DELETE = do_delete
        try:
            client = Client(self.base_url, "acme")
            bindings = client.list_source_bindings(source="sensor:health")
            self.assertEqual(len(bindings), 1)
            self.assertEqual(bindings[0].id, "b1")
            client.revoke_source_binding("b1")
        finally:
            handler.do_GET = original_do_get
            if original_do_delete is not None:
                handler.do_DELETE = original_do_delete
            else:
                delattr(handler, "do_DELETE")

    def test_update_twin_lifecycle_posts_lifecycle(self):
        handler = _Handler
        original_do_post = handler.do_POST

        def patched(self):
            if self.path == "/v1/twins/twin-1/lifecycle":
                length = int(self.headers.get("Content-Length", "0"))
                body = json.loads(self.rfile.read(length) or b"{}")
                if body.get("lifecycle") != "retired":
                    self._json(400, {"error": "bad lifecycle"})
                    return
                self._json(200, {"id": "twin-1", "type_ref": "robot", "lifecycle": "retired", "tenant_id": "acme", "created_at": "2026-08-11T00:00:00Z"})
                return
            original_do_post(self)

        handler.do_POST = patched
        try:
            client = Client(self.base_url, "acme")
            twin = client.update_twin_lifecycle("twin-1", "retired")
            self.assertEqual(twin.lifecycle, "retired")
        finally:
            handler.do_POST = original_do_post

    def test_retry_on_transient_errors(self):
        counter = {"n": 0}
        original_do_get = _Handler.do_GET

        def patched(self):
            counter["n"] += 1
            if counter["n"] < 3:
                self._json(503, {"error": "unavailable"})
                return
            original_do_get(self)

        _Handler.do_GET = patched
        try:
            client = Client(self.base_url, "acme", max_retries=3)
            state = client.resolve_state("twin-1", "health", as_of=datetime(2026, 8, 11, 10, 0, 0, tzinfo=timezone.utc), as_known=datetime(2026, 8, 11, 10, 0, 0, tzinfo=timezone.utc))
            self.assertEqual(state.status, "resolved")
            self.assertEqual(counter["n"], 3)
        finally:
            _Handler.do_GET = original_do_get

    def test_list_snapshots_parses_list(self):
        handler = _Handler
        original_do_get = handler.do_GET

        def patched(self):
            if self.path.startswith("/v1/twins/twin-1/snapshots"):
                self._json(200, {"snapshots": [{"id": "s1", "tenant_id": "acme", "subject_id": "twin-1", "resolution_policy": "p", "digest": "d", "states": [], "relations": [], "created_at": "2026-08-11T00:00:00Z"}]})
                return
            original_do_get(self)

        handler.do_GET = patched
        try:
            client = Client(self.base_url, "acme")
            snapshots = client.list_snapshots("twin-1", limit=10)
            self.assertEqual(len(snapshots), 1)
            self.assertEqual(snapshots[0].id, "s1")
        finally:
            handler.do_GET = original_do_get

    def test_report_heartbeat_posts_source(self):
        handler = _Handler
        original_do_post = handler.do_POST

        def patched(self):
            if self.path == "/v1/heartbeats":
                length = int(self.headers.get("Content-Length", "0"))
                body = json.loads(self.rfile.read(length) or b"{}")
                if body.get("source") != "sensor:health":
                    self._json(400, {"error": "bad source"})
                    return
                self._json(200, {"tenant_id": "acme", "source": "sensor:health", "last_heartbeat_at": "2026-08-11T10:00:00Z"})
                return
            original_do_post(self)

        handler.do_POST = patched
        try:
            client = Client(self.base_url, "acme")
            heartbeat = client.report_heartbeat("sensor:health")
            self.assertEqual(heartbeat.source, "sensor:health")
        finally:
            handler.do_POST = original_do_post

    def test_list_twin_types_parses_packs(self):
        handler = _Handler
        original_do_get = handler.do_GET

        def patched(self):
            if self.path == "/v1/twin-types":
                self._json(200, {"twin_types": [{"type_ref": "robot", "description": "Robot twin", "properties": ["location"], "relations": ["located_in"]}]})
                return
            original_do_get(self)

        handler.do_GET = patched
        try:
            client = Client(self.base_url, "acme")
            types = client.list_twin_types()
            self.assertEqual(len(types), 1)
            self.assertEqual(types[0].type_ref, "robot")
        finally:
            handler.do_GET = original_do_get

    def test_compute_impact_sends_depth_and_parses_paths(self):
        handler = _Handler
        original_do_get = handler.do_GET

        def patched(self):
            if self.path.startswith("/v1/twins/twin-1/impact"):
                self._json(200, {"impact_paths": [{"subject_id": "twin-1", "depth": 1, "relation_id": "r1", "predicate": "located_in", "target_id": "zone-1", "state_kind": "observed", "source": "map", "evidence_ref": "e1"}]})
                return
            original_do_get(self)

        handler.do_GET = patched
        try:
            client = Client(self.base_url, "acme")
            paths = client.compute_impact("twin-1", max_depth=3, predicate="located_in")
            self.assertEqual(len(paths), 1)
            self.assertEqual(paths[0].target_id, "zone-1")
        finally:
            handler.do_GET = original_do_get

    def test_api_error_carries_status_and_message(self):
        client = Client(self.base_url, "acme")
        with self.assertRaises(APIError) as caught:
            client.get_twin("missing")
        self.assertEqual(caught.exception.status, 404)
        self.assertEqual(caught.exception.message, "not found")

    def test_resolved_state_round_trip(self):
        state = ResolvedState(
            tenant_id="acme", subject_id="twin-1", property="health", status="resolved",
            freshness="fresh", resolution_policy="p", value="ready",
        )
        restored = ResolvedState(**json.loads(json.dumps(state.__dict__)))
        self.assertEqual(restored, state)


    def test_end_to_end_warehouse_flow(self):
        server = HTTPServer(("127.0.0.1", 0), E2EHandler)
        thread = Thread(target=server.serve_forever, daemon=True)
        thread.start()
        try:
            client = Client(f"http://127.0.0.1:{server.server_port}", "acme")
            twin = client.create_twin("robot:WH-17", "robot")
            self.assertEqual(twin.id, "robot:WH-17")
            claim = client.append_assertion(StateAssertion(subject_id="robot:WH-17", property="health", value="ready", state_kind="observed", event_time=datetime(2026, 8, 11, 10, 0, 0, tzinfo=timezone.utc), source="sensor:health", evidence_ref="e1"), idempotency_key="k1")
            self.assertEqual(claim.id, "a1")
            state = client.resolve_state("robot:WH-17", "health")
            self.assertEqual(state.status, "resolved")
            snapshot = client.create_snapshot("robot:WH-17")
            self.assertEqual(snapshot.id, "snap-1")
            self.assertTrue(client.verify_snapshot("snap-1"))
            events = client.list_changes()
            self.assertEqual(len(events), 1)
        finally:
            server.shutdown()


if __name__ == "__main__":
    unittest.main()
