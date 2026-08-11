# Supply-Chain Counterfactual Reference

Model a port closure, supplier failure, or demand spike while keeping real state
and branch state strictly separated.

## Flow

1. **Real state**: orders, inventory, logistics, supplier, and policy twins
   hold real `observed` / `reported` state in ONTOVELA.
2. **Branch seed**: `POST /v1/twins/{facility}/snapshots` creates a signed
   reality slice for a PEIRAVELA branch.
3. **Possible world**: PEIRAVELA writes branch scenarios as `simulated`
   assertions. ONTOVELA resolution and snapshots exclude them, so the real
   operational view never leaks the counterfactual.
4. **Compare**: causal lineage and impact paths show which commitments change
   under each branch; risk and recovery time are compared without contaminating
   the real world.

## Reality Integrity notes

- A `simulated` claim can hold the highest authority and still never resolve.
- Only real execution effects return as `observed` / `reported` claims.
- The signed snapshot digest makes the branch input auditable.
