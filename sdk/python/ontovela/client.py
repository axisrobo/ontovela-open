"""Standard-library ONTOVELA v0.1 client."""

from __future__ import annotations

import json
import urllib.error
import urllib.parse
import urllib.request
from datetime import datetime
from typing import Any, Optional

from .models import (
    CausalAnalytics,
    CausalLink,
    ChangeEvent,
    ConflictRecord,
    ImpactPath,
    SourceHeartbeat,
    TwinType,
    RealityView,
    RealityViewItem,
    RealityViewRequest,
    RelationAssertion,
    RequiredState,
    ResolvedState,
    Snapshot,
    SnapshotDiff,
    SourceBinding,
    StateAssertion,
    SubscriptionOffset,
    Twin,
    to_dict,
    to_iso,
)


class APIError(Exception):
    """Raised for non-2xx ONTOVELA responses."""

    def __init__(self, status: int, message: str):
        super().__init__(f"ONTOVELA API {status}: {message}")
        self.status = status
        self.message = message


def _prune(value: Any) -> Any:
    if isinstance(value, dict):
        return {k: _prune(v) for k, v in value.items() if v is not None}
    if isinstance(value, list):
        return [_prune(v) for v in value]
    return value


class Client:
    """Tenant-scoped ONTOVELA client. Safe to reuse across requests."""

    def __init__(self, base_url: str, tenant_id: str, timeout: int = 30):
        if not tenant_id or not tenant_id.strip():
            raise ValueError("tenant_id is required")
        self.base_url = base_url.rstrip("/")
        self.tenant_id = tenant_id
        self.timeout = timeout

    def list_twin_types(self) -> list:
        body = self._request("GET", "/v1/twin-types")
        return [TwinType(**item) for item in body.get("twin_types", [])]

    def list_source_bindings(self, source: Optional[str] = None) -> list:
        body = self._request("GET", "/v1/source-bindings", query=_query(source=source))
        return [SourceBinding(**item) for item in body.get("source_bindings", [])]

    def revoke_source_binding(self, binding_id: str) -> None:
        self._request("DELETE", f"/v1/source-bindings/{urllib.parse.quote(binding_id, safe='')}")

    def report_heartbeat(self, source: str) -> SourceHeartbeat:
        return self._post("/v1/heartbeats", {"source": source}, SourceHeartbeat)

    def create_twin(self, twin_id: str, type_ref: str, lifecycle: Optional[str] = None) -> Twin:
        return self._post("/v1/twins", _prune({"id": twin_id, "type_ref": type_ref, "lifecycle": lifecycle}), Twin)

    def update_twin_lifecycle(self, twin_id: str, lifecycle: str) -> Twin:
        return self._post(f"/v1/twins/{urllib.parse.quote(twin_id, safe='')}/lifecycle", {"lifecycle": lifecycle}, Twin)

    def get_twin(self, twin_id: str) -> Twin:
        return self._get(f"/v1/twins/{urllib.parse.quote(twin_id, safe='')}", Twin)

    def create_source_binding(self, payload: SourceBinding) -> SourceBinding:
        return self._post("/v1/source-bindings", to_dict(payload), SourceBinding)

    def append_assertion(self, payload: StateAssertion, idempotency_key: str) -> StateAssertion:
        return self._post("/v1/assertions", self._assertion_dict(payload), StateAssertion, idempotency_key=idempotency_key)

    def append_relation(self, payload: RelationAssertion, idempotency_key: str) -> RelationAssertion:
        return self._post("/v1/relations", self._relation_dict(payload), RelationAssertion, idempotency_key=idempotency_key)

    def list_assertions(self, twin_id: str, property: Optional[str] = None, as_of: Optional[datetime] = None, as_known: Optional[datetime] = None) -> list:
        query = _query(property=property, as_of=to_iso(as_of), as_known=to_iso(as_known))
        body = self._request("GET", f"/v1/twins/{urllib.parse.quote(twin_id, safe='')}/assertions", query=query)
        return [StateAssertion(**item) for item in body.get("assertions", [])]

    def list_relations(self, twin_id: str, predicate: Optional[str] = None, as_of: Optional[datetime] = None, as_known: Optional[datetime] = None) -> list:
        query = _query(predicate=predicate, as_of=to_iso(as_of), as_known=to_iso(as_known))
        body = self._request("GET", f"/v1/twins/{urllib.parse.quote(twin_id, safe='')}/relations", query=query)
        return [RelationAssertion(**item) for item in body.get("relations", [])]

    def resolve_state(self, twin_id: str, property: str, as_of: Optional[datetime] = None, as_known: Optional[datetime] = None) -> ResolvedState:
        query = _query(as_of=to_iso(as_of), as_known=to_iso(as_known))
        return self._request(
            "GET",
            f"/v1/twins/{urllib.parse.quote(twin_id, safe='')}/state/{urllib.parse.quote(property, safe='')}",
            query=query,
            model=ResolvedState,
        )

    def create_reality_view(self, payload: RealityViewRequest, as_of: Optional[datetime] = None, as_known: Optional[datetime] = None) -> RealityView:
        body = to_dict(payload)
        body["required_state"] = [to_dict(item) for item in payload.required_state]
        query = _query(as_of=to_iso(as_of), as_known=to_iso(as_known))
        raw = self._request("POST", "/v1/reality-views", query=query, body=body)
        return self._reality_view(raw)

    def list_snapshots(self, twin_id: str, limit: int = 100) -> list:
        body = self._request("GET", f"/v1/twins/{urllib.parse.quote(twin_id, safe='')}/snapshots", query=_query(limit=limit))
        return [Snapshot(**item) for item in body.get("snapshots", [])]

    def create_snapshot(self, twin_id: str, as_of: Optional[datetime] = None, as_known: Optional[datetime] = None) -> Snapshot:
        query = _query(as_of=to_iso(as_of), as_known=to_iso(as_known))
        return self._request("POST", f"/v1/twins/{urllib.parse.quote(twin_id, safe='')}/snapshots", query=query, model=Snapshot)

    def get_snapshot(self, snapshot_id: str) -> Snapshot:
        return self._request("GET", f"/v1/snapshots/{urllib.parse.quote(snapshot_id, safe='')}", model=Snapshot)

    def verify_snapshot(self, snapshot_id: str) -> bool:
        raw = self._request("GET", f"/v1/snapshots/{urllib.parse.quote(snapshot_id, safe='')}/verify")
        return bool(raw.get("valid"))

    def diff_snapshots(self, from_snapshot_id: str, to_snapshot_id: str) -> SnapshotDiff:
        return self._request(
            "GET",
            f"/v1/snapshots/{urllib.parse.quote(from_snapshot_id, safe='')}/diff/{urllib.parse.quote(to_snapshot_id, safe='')}",
            model=SnapshotDiff,
        )

    def list_changes(self, after: int = 0, limit: int = 100, kind: Optional[str] = None, subject_id: Optional[str] = None, property: Optional[str] = None) -> list:
        body = self._request("GET", "/v1/changes", query=_query(after=after, limit=limit, kind=kind, subject_id=subject_id, property=property))
        return [ChangeEvent(**item) for item in body.get("events", [])]

    def get_subscription_offset(self, consumer_id: str) -> SubscriptionOffset:
        return self._request("GET", f"/v1/subscriptions/{urllib.parse.quote(consumer_id, safe='')}", model=SubscriptionOffset)

    def commit_subscription_offset(self, consumer_id: str, offset: int) -> SubscriptionOffset:
        return self._post(f"/v1/subscriptions/{urllib.parse.quote(consumer_id, safe='')}/commit", {"offset": offset}, SubscriptionOffset)

    def compute_causal_analytics(self, twin_id: str) -> CausalAnalytics:
        return self._request("GET", f"/v1/twins/{urllib.parse.quote(twin_id, safe='')}/causal/analytics", model=CausalAnalytics)

    def compute_causal_lineage(self, twin_id: str, max_depth: int = 5, as_of: Optional[datetime] = None, as_known: Optional[datetime] = None) -> list:
        query = _query(max_depth=max_depth, as_of=to_iso(as_of), as_known=to_iso(as_known))
        body = self._request("GET", f"/v1/twins/{urllib.parse.quote(twin_id, safe='')}/causal", query=query)
        return [CausalLink(**item) for item in body.get("causal_links", [])]

    def compute_impact(self, twin_id: str, max_depth: int = 5, predicate: Optional[str] = None, as_of: Optional[datetime] = None, as_known: Optional[datetime] = None) -> list:
        query = _query(max_depth=max_depth, predicate=predicate, as_of=to_iso(as_of), as_known=to_iso(as_known))
        body = self._request("GET", f"/v1/twins/{urllib.parse.quote(twin_id, safe='')}/impact", query=query)
        return [ImpactPath(**item) for item in body.get("impact_paths", [])]

    def list_conflicts(self, status: Optional[str] = None, limit: int = 100) -> list:
        body = self._request("GET", "/v1/conflicts", query=_query(status=status, limit=limit))
        return [ConflictRecord(**item) for item in body.get("conflicts", [])]

    def _assertion_dict(self, payload: StateAssertion) -> dict:
        return _prune({
            "subject_id": payload.subject_id,
            "property": payload.property,
            "value": payload.value,
            "state_kind": payload.state_kind,
            "event_time": to_iso(payload.event_time),
            "source": payload.source,
            "confidence": payload.confidence,
            "evidence_ref": payload.evidence_ref,
        })

    def _relation_dict(self, payload: RelationAssertion) -> dict:
        return _prune({
            "source_id": payload.source_id,
            "predicate": payload.predicate,
            "target_id": payload.target_id,
            "state_kind": payload.state_kind,
            "event_time": to_iso(payload.event_time),
            "valid_from": to_iso(payload.valid_from),
            "valid_to": to_iso(payload.valid_to),
            "source": payload.source,
            "confidence": payload.confidence,
            "evidence_ref": payload.evidence_ref,
        })

    def _reality_view(self, raw: dict) -> RealityView:
        items = []
        for item in raw.get("items", []):
            state = ResolvedState(**item["state"]) if item.get("state") else None
            items.append(RealityViewItem(property=item["property"], acceptable=item["acceptable"], state=state))
        return RealityView(tenant_id=raw.get("tenant_id", ""), twin_id=raw.get("twin_id", ""), purpose=raw.get("purpose", ""), status=raw.get("status", ""), items=items, authorization_ref=raw.get("authorization_ref"))

    def _post(self, path: str, body: dict, model: type, idempotency_key: str = "") -> Any:
        return self._request("POST", path, body=body, model=model, idempotency_key=idempotency_key)

    def _get(self, path: str, model: type) -> Any:
        return self._request("GET", path, model=model)

    def _request(self, method: str, path: str, query: Optional[dict] = None, body: Optional[dict] = None, model: Optional[type] = None, idempotency_key: str = "") -> Any:
        url = self.base_url + path
        if query:
            url += "?" + urllib.parse.urlencode(query)
        request = urllib.request.Request(url, method=method)
        request.add_header("Accept", "application/json")
        request.add_header("X-Tenant-ID", self.tenant_id)
        if idempotency_key:
            request.add_header("Idempotency-Key", idempotency_key)
        payload: Optional[bytes] = None
        if body is not None:
            payload = json.dumps(body).encode("utf-8")
            request.add_header("Content-Type", "application/json")
        try:
            with urllib.request.urlopen(request, data=payload, timeout=self.timeout) as response:
                raw = response.read()
                if not raw:
                    return None
                value = json.loads(raw.decode("utf-8"))
        except urllib.error.HTTPError as error:
            message = ""
            try:
                message = json.loads(error.read().decode("utf-8")).get("error", "")
            except Exception:
                pass
            raise APIError(error.code, message) from None
        if model is not None:
            return model(**value)
        return value


def _query(**values: Any) -> dict:
    return {key: value for key, value in values.items() if value is not None}
