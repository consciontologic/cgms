# 📜 Charter: CGMS

## Vision

Make Card Game Mafia State a reproducible, expressive negotiation game that people can learn, play with friends, and investigate through transparent simulations.

## Scope (in)

- A standalone CLI simulator with deterministic accepted-rule scenarios, fair bots, replay and readable experiment reports.
- Three- and four-player online games on Flutter web, iOS and Android, with durable Go adjudication and confidential seat views.
- Unlimited native offline play with difficulty levels, competent introductory bots and contextual guidance.
- Localized, accessible interfaces using the established cgmsart assets; free online access and only Ad-free, Daily, Weekly, Monthly and Yearly premium products. Earned dirt can redeem Daily/Weekly access with ads retained. Card styles are automatically assigned for each match, without sales or unlocks.

## Scope (out)

- Kubernetes/Helm and unnecessary service separation before v1: the deployment target is Docker Compose.
- Mandatory LLMs or trusted uploads of offline results: simulation must work independently and offline play cannot grant online rewards.
- Runtime implementation in this documentation slice: future work remains in the sole [roadmap](../planning/ROADMAP.md).

## Audience

- **Primary:** people playing or learning a three- or four-player negotiation card game.
- **Secondary:** game designers evaluating rules/bots and contributors implementing shared contracts.
- **Operators:** the project maintainer; production ownership and staffing are not yet assigned.

## Success criteria

Reproduce implemented rule traces with identical state/events, conserve all 104 cards, preserve authorized observations and exact ledgers, and recover committed actions without duplication. Shipping requires the roadmap's client, load, restore and release gates; balance, latency and production readiness are currently unverified.

## Non-goals & explicit trade-offs

Accepted untimed play may wait indefinitely for an eligible disconnected participant. Exact accounting and private information take precedence over convenient approximations. The proposed single-host design needs measured limits and recovery evidence, not a high-availability claim.

## Constraints

[GAME_RULES.md](GAME_RULES.md) governs gameplay; the [project brief](INITIATE.md) supplies product constraints. Use Flutter, Go, PostgreSQL, justified transient Redis and Compose; preserve ARB localization and cgmsart assets. Keep real credentials/private state out of committed files and public views. Commerce amounts, abandoned-game economy, store applicability and technical proposals still need their named decisions and validation before release.

## Development and acceptance sequence

Flutter web is the current development/testing platform under [UI baseline v2](../design/UI_BASELINE.md).
Distinct desktop, phone and tablet presentations share logic/state and cgmsart;
phone/tablet web must faithfully specify native appearance and user journeys.
Use the existing host/package. Native product commitments above remain required;
Android/iOS packaging, device, lifecycle, performance and platform acceptance are
final pre-release R06N/R06 gates. [ROADMAP](../planning/ROADMAP.md) is the only
sequenced checklist; unavailable native infrastructure does not block web work.
