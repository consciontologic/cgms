---
name: cgms-deterministic-replay
description: "Design, implement or diagnose CGMS deterministic seed streams, serialization, decision transcripts and replay divergence."
---

# CGMS deterministic replay

Use for S03–S04/S09 replay work; implementation requires an implementation task.
Read the [CLI manifest contract](../../../sims/README.md),
[research decisions](../../../sims/docs/phase-0-research.md) and
[format specification](../../../sims/docs/phase-0-format-spec.md).

1. Pin algorithm/version, seed derivation, byte encoding, sampling/shuffling and
   canonical serialization before recording vectors. A root seed alone is not
   reproducibility. Reject unsupported compatibility versions instead of silently
   interpreting old bytes with new rules. Preserve effective config bytes/hash,
   source/build provenance, input hashes and every override.
2. Separate match/game/chance/bot/analysis streams. Fix seed identity from logical
   work IDs, never worker ID, completion order or timestamps. Paired treatments
   need common exogenous seeds despite different config hashes. Sort unordered
   collections explicitly and version the total order and random draw counters.
3. Record accepted commands and rejected diagnostics separately; include immutable
   game/effect/decision IDs, actor, stage, options, selected inputs and committed
   random outcomes. Exact replay consumes that transcript; re-running a bot is a
   different test. Resume never resamples a committed Justice selection or charges
   a previously resolved response again.
4. Verify artifact hashes/compatibility before replay, then compare each state and
   event digest. At the first mismatch preserve index, input identity, expected and
   actual digest, prior verified checkpoint, version/hash differences and a private
   canonical field diff. Do not continue and report only the final score mismatch.
5. Test known PRNG/sampling/config vectors, equal effective configs, worker counts,
   uninterrupted versus resumed runs and deliberate transcript corruption. Include
   nullified replacement game IDs, N06 corrections, T01 revisions, T02 draw cursors
   and resumable settlement where supported; unsupported cases stay explicit.
6. Keep seeds, future random state, deck order and omniscient traces restricted.
   Shareable Markdown uses an explicit observation projection. Never infer online
   cryptographic draw suitability from offline deterministic replay success.

Acceptance evidence identifies exact bytes/vectors and the first divergence in
negative tests. Fail on map-order hashes, JSON floating-point rationals, stale
commands retargeted to a replacement game, or resume that repeats prior expensive
settlement. Follow [execution/recovery instructions](../../instructions/PHASE_0_SIMULATION.md)
for evidence publication and cancellation.
