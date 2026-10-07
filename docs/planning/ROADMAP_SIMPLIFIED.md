# CGMS development roadmap — plain-language guide

This is an explanatory companion, not a second checklist. All task IDs,
completion status, dependencies and evidence live only in [ROADMAP.md](ROADMAP.md).

The simulator, rules/persistence and online API phases have recorded local
verification. The Flutter host integrates authoritative online gameplay and retains
local UI v1 scenarios. Existing evidence is preserved; current browser verification
is recorded by P04, while native acceptance remains a separate final release gate.

Develop and test the client through Flutter web now, using `client/`
and its `packages/cgms_ui`. P04 covers distinct desktop, phone and tablet
experiences, each with its own composition, navigation, interactions and
presentation components. Sharing logic/state/cards/theme is encouraged; resizing
one desktop interface does not count. Phone/tablet web must faithfully represent
the appearance and journeys intended for native apps. [UI baseline v2](../design/UI_BASELINE.md)
defines width classes, browser viewport/transition tests and acceptance criteria.

P01 compares offline engines on the host and records a provisional recommendation;
P02 develops saved-game, tutorial and bot contracts. Neither native devices nor
P01 completion blocks P04 web development. P03 keeps its existing experiment
resource constraint. Product/economy contracts and operational work proceed with
host/browser tests; production web gameplay still requires online authority.

In the final release phase, R06N requires Android/iOS packaging and real phone/
tablet execution, matched engine performance/memory measurements, final engine
selection, airplane-mode saves, lifecycle/recovery, native input/accessibility,
platform behavior and real store sandbox checks. Compare native appearance and
journeys to accepted phone/tablet web references. R06 then completes policy,
signing/security and store release verification; R07 is the full release rehearsal.
Missing native infrastructure blocks release acceptance, not earlier web work.
Browser tests and Android compilation cannot substitute for these native gates.
