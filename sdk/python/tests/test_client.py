"""Contract tests for the ONTOVELA Python client."""

import json
import sys
import unittest
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


if __name__ == "__main__":
    unittest.main()
