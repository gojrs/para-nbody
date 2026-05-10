# Milestones

## Milestone 0 — Project Orientation

Goal: establish workflow, comparison strategy, and shared terminology.

Deliverables:

- Confirm `grav-charge` is the reference/traditional engine.
- Confirm `para-nbody` is the experimental engine.
- Add contribution rules.
- Add comparison plan.
- Identify first deterministic simulation scenario.

Done when:

- Documentation exists.
- The first code milestone is clearly defined.

## Milestone 1 — Deterministic Baseline Scenario

Goal: define one small simulation scenario that both engines can run.

Candidate scenario:

- Two bodies
- Fixed timestep
- Deterministic initial positions and velocities
- JSON-serializable result
- Numeric tolerance for comparison

Done when:

- Input dataset exists.
- Expected baseline output is documented or generated.
- `go test ./...` passes.