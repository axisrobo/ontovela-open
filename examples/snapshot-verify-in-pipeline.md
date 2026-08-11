# Snapshot Verify in a Release Pipeline

Pin the reality used to approve a change, then prove it later.

```bash
# Approve against a pinned reality slice.
snapshot=$(curl -s -X POST http://localhost:8080/v1/twins/facility:WH-3/snapshots \
  -H 'X-Tenant-ID: acme')
id=$(echo "$snapshot" | jq -r .id)

# Before executing, verify the digest is intact.
curl -s -H 'X-Tenant-ID: acme' \
  http://localhost:8080/v1/snapshots/$id/verify

# Diff against a later snapshot to show what changed.
curl -s -H 'X-Tenant-ID: acme' \
  http://localhost:8080/v1/snapshots/$id/diff/$later_id
```

`verify` returns `valid: true` only when the manifest digest matches. The diff
also reports `policy_changed` when the resolution policy differs between
snapshots, so a plan signed under one policy is never silently compared under
another.
