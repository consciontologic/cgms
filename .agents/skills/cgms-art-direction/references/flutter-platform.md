# Flutter and integration contract — shared vertical board

Use the [UI baseline](../../../../docs/design/UI_BASELINE.md) and existing
[portable package contract](../../../../client/INTEGRATION.md). The
owner-revised baseline is the presentation authority; the app retains historical evidence. Earlier proposed ThemeExtension APIs,
widget inventories, layout algorithms and alternate navigation examples are
retired; they do not require replacing working components.

## Existing shared implementation

`CgmsTheme.dark()` and `CgmsColors` supply the theme. `CgmsActionBoard` renders
`TableViewData` and narrow callbacks, with host-owned action/header/footer slots.
The host adapts rules/state; renderer code does not infer permissions. Reuse the
package's assets, vector suit marks, ARB localization, bundled fonts, semantics
and tests when integrating it into another client. See [components](components.md)
and the [package README](../../../../client/packages/cgms_ui/README.md).

The online route uses the Go authority and `CgmsAdaptiveTable`; the separate
local-demo route demonstrates prepared branches with the same adaptive table.
Preserve that distinction: a demo reconnect simulation is not network evidence.
Use the dated real-authority browser reports for online acceptance and the
separate engine/offline suites for their stated scope. New authoritative behavior
needs its own implementation and tests while keeping the current baseline. Compatibility views
remain available to existing consumers; they do not become the default adaptive
composition.

## Adaptive composition and testing

Preserve the same vertical board across desktop, phone and tablet, enlarging
items with available width. Use the baseline width/viewport/transition matrix;
phone/tablet web must faithfully represent intended native appearance/journeys.
Share state and action callbacks above presentation switching. Never reset or
repeat commands on resize/navigation. Retain v1 fit evidence as historical only.
Browser tests run now; native packaging, lifecycle, performance, accessibility
and device/platform gates run in R06N before R06/R07 release acceptance.

## Input, semantics, and text

Preserve current behavior: focused buttons activate with Enter/Space; focused
cards also support **I** for independent inspection. Touch/click selects
where actionable; double tap inspects independently. Long press does neither.
Multi-card piles expose separately identified cards through review controls.
Hover is supplementary. No global Enter/victory shortcut, drag-only action,
implicit pass on Escape or modal dismissal, or focus loss after unrelated updates.

Expose semantic role, ownership, physical copy, selected/used status and permitted
actions from the same viewer-filtered projection as visible content. Combine
redundant icon/text meaning and exclude decorative art. Do not encode concealed
identity in a back, asset choice, DOM, tooltip, analytics label or focus order.
Public hand/Ace counts remain separate; hidden eligible-card breakdowns stay private.

Keep focus visible independently of selection. Restore focus after dismissible
inspection and make disabled explanations accessible. Announce meaningful
turn/result/connection changes without repeating the entire board. A legally
available Coup must remain immediate under the game-aware overlay policy, with
no new confirmation or presentation delay. Full screen-reader/device conformance
remains a validation gate, not a consequence of UI approval.

Inherit `TextScaler`, respect platform accessibility settings, and use growing
content rather than global no-scaling or `FittedBox` on essential facts. Test
long localized labels, exact fractions, minus signs and physical-copy labels.
Do not round authoritative values to make columns fit. Keep package ARB and
generated labels together; custom host slots own their own localization.

## Motion and feedback

Retain the current reduced-motion setting and platform preference handling.
Do not introduce travel, parallax, shake, looping texture or animation-driven
state changes. An animation must not commit a transaction, resolve a response,
delay Coup or obscure a durable result. Audio/haptics are optional supplements.
No countdowns, automatic passes or pressure signals absent from the rules.

## Web loading and navigation; future PWA work

Preserve the existing bootstrap, local resources and Router-based navigation.
Back/Forward, direct entry, screen changes and refresh are navigation events,
not game commands. URLs contain no credentials, hands or private terms. Closing
a browser cannot reliably implement quitting, surrender or resignation. Keep
honest loading/error/retry states and do not auto-reload a live decision to repair
an illustration. An installed shell is not proof of offline game capability.

For a separately authorized deployment/cache task, specify ownership, versions,
expiry, update/recovery behavior and actual cached asset scope. Cache only
approved public shell/rules/fonts/art; do not persist private game data under a
generic offline feature. Test cold/failed startup, stale assets, refresh and
service-worker updates before claiming support. A warm shell and an uncached
first visit have different offline behavior. Updates wait for a safe chosen
point after uncertain action acceptance is reconciled.

## Connectivity and game boundaries

Keep last-confirmed authorized state visible during stale/disconnected states;
disable ordinary mutations and distinguish rejected from outcome-unknown.
Reconcile the original operation before any resubmit; never invent success,
automatically replay a request, pass a required response, forfeit or advance a
turn because connectivity changed. Restore effect/decision IDs, required actor,
stage and committed random results. Confined/departed eligibility skips are
rule decisions; eligible disconnected seats remain awaited.

Effect choices in Negotiator/Justice/Dexter are durable inputs, not extra response
windows; accepted activations cannot be withdrawn. Nonblocking offers carry
party-visible terms, revision and origin turn separately from committed
response-bearing transfers. Alliances/loan privileges begin only on success.
Fixed-seat Compensation queues resume committed progress; purchase requests
carry explicit quantity/payment and must not partially fulfill client-side.

Bind every game-scoped command, including settlement, to its immutable game ID.
Fresh games and nullified replacements have new IDs; do not retarget queued
commands. Duplicate lookup returns the original result without new effects.
Victory admission uses current eligibility/phase rather than unrelated stale
aggregate versions; ordinary selected commands retain their required checks.

Board closure stops gameplay immediately, while the financial allow-list can
remain active until explicit settlement decisions/finalization. No timeout is
refusal. Preserve exact debt/receipt ordering, rational shares, persistent effect
ownership, restricted-card markers, public Ace custody and nonspendable match
corrections in [gameplay contracts](gameplay-layout.md). Financial finalization
closes game cash; a new game starts at zero available cash with the required
scores/debts carried. Widgets display these results rather than recompute them.

## Verification boundaries

Use [validation](validation.md) and the current app suites
for relevant regression checks. Record real rendered/input evidence at supported
sizes and honest scope limits. Online browser journeys, host offline bots and
native acceptance have separate roadmap gates and evidence; one cannot prove the
others. Native screen-reader/device testing and production recovery remain
release gates.

Framework reference material: [themes](https://docs.flutter.dev/cookbook/design/themes),
[input/accessibility](https://docs.flutter.dev/ui/adaptive-responsive/input),
[Semantics](https://api.flutter.dev/flutter/widgets/Semantics-class.html),
[text scaling](https://docs.flutter.dev/release/breaking-changes/android-14-nonlinear-text-scaling-migration),
[web initialization](https://docs.flutter.dev/platform-integration/web/initialization),
[navigation](https://docs.flutter.dev/ui/navigation) and
[accessibility testing](https://docs.flutter.dev/ui/accessibility/accessibility-testing).
Verify relevant APIs against the pinned SDK when changing code; these links do
not replace the current package or grant runtime validation claims.
