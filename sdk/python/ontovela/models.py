"""Dataclasses mirroring the ONTOVELA v0.1 public wire contract."""

from __future__ import annotations

from dataclasses import dataclass, field, asdict
from datetime import datetime
from typing import Any, Optional


def to_iso(value: Optional[datetime]) -> Optional[str]:
    if value is None:
        return None
    return value.astimezone().isoformat().replace("+00:00", "Z")


@dataclass
class TwinInput:
    id: str
    type_ref: str
    lifecycle: Optional[str] = None
    external_bindings: Optional[Any] = None


@dataclass
class Twin(TwinInput):
    tenant_id: str = ""
    created_at: Optional[str] = None


@dataclass
class SourceBindingInput:
    id: str
    source: str
    property: str
    authority_rank: int = 0
    max_lag_seconds: int = 0


@dataclass
class SourceBinding(SourceBindingInput):
    tenant_id: str = ""
    created_at: Optional[str] = None


@dataclass
class StateAssertionInput:
    subject_id: str
    property: str
    value: Any
    state_kind: str
    event_time: datetime
    source: str
    evidence_ref: str
    confidence: Optional[float] = None


@dataclass
class StateAssertion(StateAssertionInput):
    id: str = ""
    tenant_id: str = ""
    system_time: Optional[str] = None


@dataclass
class RelationAssertionInput:
    source_id: str
    predicate: str
    target_id: str
    state_kind: str
    event_time: datetime
    valid_from: datetime
    source: str
    evidence_ref: str
    valid_to: Optional[datetime] = None
    confidence: Optional[float] = None


@dataclass
class RelationAssertion(RelationAssertionInput):
    id: str = ""
    tenant_id: str = ""
    system_time: Optional[str] = None


@dataclass
class ResolvedState:
    tenant_id: str
    subject_id: str
    property: str
    status: str
    freshness: str
    resolution_policy: str
    value: Any = None
    state_kind: Optional[str] = None
    source: Optional[str] = None
    confidence: Optional[float] = None
    supporting_assertion_id: Optional[str] = None
    conflicting_assertion_ids: list = field(default_factory=list)
    event_time: Optional[str] = None


@dataclass
class RequiredState:
    property: str
    max_age_seconds: int


@dataclass
class RealityViewRequest:
    twin_id: str
    purpose: str
    required_state: list = field(default_factory=list)


@dataclass
class RealityViewItem:
    property: str
    acceptable: bool
    state: ResolvedState


@dataclass
class RealityView:
    tenant_id: str
    twin_id: str
    purpose: str
    status: str
    items: list = field(default_factory=list)


@dataclass
class Snapshot:
    id: str
    tenant_id: str
    subject_id: str
    resolution_policy: str
    digest: str
    states: list = field(default_factory=list)
    relations: list = field(default_factory=list)
    as_of: Optional[str] = None
    as_known: Optional[str] = None
    created_at: Optional[str] = None


@dataclass
class SnapshotDiff:
    from_snapshot_id: str
    to_snapshot_id: str
    added_properties: list = field(default_factory=list)
    removed_properties: list = field(default_factory=list)
    changed_properties: list = field(default_factory=list)
    added_relation_ids: list = field(default_factory=list)
    removed_relation_ids: list = field(default_factory=list)


@dataclass
class ChangeEvent:
    offset: int
    tenant_id: str
    kind: str
    subject_id: str
    payload: Any
    occurred_at: Optional[str] = None
    property: Optional[str] = None


@dataclass
class SubscriptionOffset:
    tenant_id: str
    consumer_id: str
    committed_offset: int
    updated_at: Optional[str] = None


@dataclass
class ConflictRecord:
    tenant_id: str
    subject_id: str
    property: str
    status: str
    assertion_ids: list
    resolution_policy: str
    updated_at: Optional[str] = None


@dataclass
class SourceHeartbeat:
    tenant_id: str
    source: str
    last_heartbeat_at: Optional[str] = None


@dataclass
class TwinType:
    type_ref: str
    description: str
    properties: list = field(default_factory=list)
    relations: list = field(default_factory=list)


@dataclass
class ImpactPath:
    subject_id: str
    depth: int
    relation_id: str
    predicate: str
    target_id: str
    state_kind: str
    source: str
    evidence_ref: str


def to_dict(dataclass_instance: Any) -> dict:
    return asdict(dataclass_instance)
