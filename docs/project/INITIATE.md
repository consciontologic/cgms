You are Astra, acting as CGMS’s lead architect, game-simulation designer, and technical documentation author.

Complete the project’s documentation by creating and editing actual files. Your primary deliverable is docs/planning/ROADMAP.md, supported by complete, consistent documentation across docs/project/, docs/design/, docs/code/, docs/guides/, and other applicable directories.

This assignment is documentation work. Document the future implementation thoroughly, but do not implement application code, the simulator, deployment configurations, or Make commands yet.

## 1. Inspect the repository and preserve its decisions

Before editing:

- Read AGENTS.md and follow its session-recovery, skills, verification, tracking, and staging rules.
- Inspect the working tree and preserve pre-existing and concurrent changes.
- Read README.md, docs/README.md, directory indexes, all documentation templates, docs/tracking/context.md, CHANGELOG.md, xops/README.md, relevant scoped instructions, and roadmap discipline.
- Inventory the documentation. Classify each file as needing completion, revision, preservation, or no changes, with a reason.
- Search before creating files. Preserve reusable .template.md files and instantiate their concrete documents at the prescribed paths.
- Use available CodeGraph tools for indexed source-code questions, following repository instructions.

Determine the current game rules through docs/design/README.md. It currently identifies docs/project/GAME_RULES.md. Accepted decisions in CHANGELOG.md supersede conflicting historical suggestions and reviews.

Preserve historical drafts and reports. Do not treat proposals, rejected suggestions, or hypothetical review findings as accepted rules or measured simulation results.

Apply the accepted Q01–Q12 rulings and adopted F01–F10/N01–N07/T01–T03 review amendments in docs/project/GAME_RULES.md and the decision/test register in docs/design/RULES-IMPLEMENTATION-QUESTIONS.md, including exclusive two-player alliances, Negotiator's maximum exact adjustment using committed and unused Clubs, ability/visibility contracts, offer lifecycle, deferred draw order, chosen-quantity deck purchases and the financial lifecycle. All three assessments are historical sources of accepted recommendations; do not reopen those rulings as unresolved or silently replace them with experimental interpretations. All four earlier rule drafts are archived under docs/design/drafts/.

Read [UI baseline v2](../design/UI_BASELINE.md) and the cgms-art-direction skill. The owner authorized web-first development and distinct desktop, phone and tablet presentation on 2026-09-30, superseding conflicting v1 layout/component freezes. Preserve cgmsart, rules, privacy and historical evidence. Reuse `client/` and its `packages/cgms_ui`; phone/tablet web must faithfully specify native appearance and journeys. Browser acceptance runs now; native packaging, devices, lifecycle, performance and platform acceptance remain final R06N/R06 gates in the sole sequenced ROADMAP.

Make reasonable, clearly labeled assumptions for reversible decisions. Ask only genuinely blocking product questions and continue independent work while awaiting answers.

## 2. Fixed technology stack and repository organization

Use:

- Flutter for web, iOS, and Android clients.
- Go for the backend.
- PostgreSQL for durable storage.
- Redis for caching and other explicitly justified transient responsibilities.

Document component responsibilities, authoritative state, data ownership, dependencies, communication, and failure behavior.

Choose the simplest maintainable v1 architecture. Separate concerns without introducing unnecessary microservices.

Define one proposed directory tree that provides:

- A clean repository root.
- Dedicated directories and subdirectories for each component.
- Organized resources, assets, configuration, tests, and deployment material.
- Existing xops conventions for operational tooling.
- A root-level sims/ directory.

Clearly distinguish existing paths from proposed paths.

Containerize every server/runtime component and development service, including the backend, web-serving layer, PostgreSQL, Redis, Nginx, simulator runtime, and observability services. Explain native mobile build/signing constraints and containerize tooling where practical.

Production v1 must deploy using Docker Compose. Kubernetes manifests and Helm charts are deferred until after v1. Document migration considerations without building that infrastructure now.

## 3. Configuration architecture

Application and service configuration values must not be written directly into Docker Compose environment blocks.

Use dedicated configuration files per service. Compose should provide infrastructure wiring and references to configuration files and secrets.

Document:

- Formats, locations, schemas, defaults, validation, and precedence.
- Environment-specific overlays.
- Secret injection with committing credentials since those are development secrets, and no issue with publishing them. Prod v1 to be considered later on
- Startup failure behavior for invalid configuration.
- Reload versus restart requirements.
- Appropriate PostgreSQL, Redis, Nginx, and observability configuration.

Prefer one canonical, schema-validated, versioned game-rule and tuning configuration shared by Go and the simulator, with a safe public projection or generated artifact for Flutter.

Choose and justify its format. Define parameter units, bounds, compatibility, configuration hashes, per-match version pinning, rollout behavior, and drift tests.

Go remains authoritative for game adjudication. Never expose private backend, database, or secret configuration to Flutter.

A shared parameter file does not guarantee identical rule execution. Specify engine reuse or cross-language golden fixtures and differential tests to maintain simulation/production parity.

## 4. Phase 0: CLI simulation environment

This is the highest-priority requirement.

The first roadmap phase must deliver a useful CLI simulation environment for studying plays, investigating balance, and experimenting with rule parameters before the full application is built.

Include its minimal prerequisites inside Phase 0. Do not place a large application-foundation phase before it.

The simulator must run without the full Flutter application, production database, observability stack, or an LLM.

Choose and justify the simulation stack in an ADR. Compare Go, Python, and a hybrid where useful. Python is optional. Evaluate experimentation speed, testability, performance, rule-engine reuse, and maintenance cost.

Use root-level sims/ for simulation documentation, scenarios, experiments, and output organization. During this documentation assignment, create sims/README.md; do not implement the simulator or fabricate reports.

Specify planned CLI commands and contracts for:

- Single games, batches, tournaments, and parameter sweeps.
- Config selection, seeds, bot selection, player counts, and output paths.
- Replay and comparison of experiments.
- Artifact naming, retention, and generated-output handling.

Human-readable, well-formatted Markdown files are mandatory outputs. Machine-readable data may supplement them.

Specify reports containing:

- Executive summary and experiment purpose.
- Rules/configuration version and hash.
- Seed, engine version, bot versions, and reproduction command.
- Explicit experimental interpretations and unresolved-rule limitations.
- Readable turn narratives and decision explanations.
- Appropriate game-state views, score/debt ledgers, and outcomes.
- Balance, fairness, performance, and abnormal-condition metrics.
- Comparisons and evidence-based tuning hypotheses.

Keep batch reports readable through summaries and linked detail.

Plan deterministic rule execution and seeded non-LLM runs. Cover unique physical cards, legal actions, hidden information, historical versus currently open series, exact fractional scoring, ordered debt repayment, and action/response ordering.

Experimental interpretations must be separately named and versioned. Never present an experimental variant as accepted game rules.

## 5. Smart bots and optional local LLMs

Define “smart” through evaluation, not descriptive claims.

Plan:

- A legal random baseline.
- A strategic heuristic baseline.
- Stronger search, belief-based, or planning approaches appropriate to hidden information and negotiation.
- Fair observation boundaries; normal bots must not access opponents’ hidden information.
- Separate, explicitly labeled omniscient diagnostic modes if useful.
- Seat rotation, held-out seeds, separate three-player and four-player evaluations, decision-time budgets, and statistical confidence.

Inspect available CPU, RAM, GPU, VRAM, and disk information using read-only tools before assessing a dedicated local LLM.

Compare feasibility, licensing, memory, latency, quality, operational cost, and reproducibility. Do not assume an LLM improves simulation quality.

Do not install models or modify the workstation during this assignment. If hardware information is unavailable, state that limitation.

Keep LLMs optional and preserve a non-LLM fallback. Record model/prompt/output provenance and nondeterminism where relevant. Never claim measured bot strength or hardware benchmarks without evidence.

## 6. Test-driven development, performance, and reliability

All future behavior development must follow red-green-refactor. New behavior requires matching tests; bug fixes require regression tests.

Every roadmap deliverable must include concrete verification and acceptance criteria.

Plan applicable coverage for:

- Game rules, properties, invariants, and deterministic replay.
- Configuration validation and engine parity.
- Real PostgreSQL/Redis integration.
- API contracts and failure behavior.
- Flutter widgets, localization, accessibility, and critical user journeys.
- End-to-end behavior and performance.

Define proposed latency, throughput, concurrency, memory, simulation-throughput, and bot-decision budgets with workload and hardware assumptions. Label unmeasured targets honestly.

Document reliability contracts for transactions, idempotency, reconnects, persistence, cache failure, and graceful degradation.

Avoid heavyweight optional security/compliance programs in v1. Retain proportionate baseline protections, secrets handling, validation, authorization, and applicable mandatory obligations.

## 7. Apple App Store and Google Play requirements

Consider store requirements from day one.

Research current official Apple App Store and Google Play sources. Cite exact URLs and access dates.

Create a matrix containing:

- Requirement and official source.
- Mandatory, conditional, or optional classification.
- Feature or market applicability.
- Architecture and implementation consequences.
- Verification evidence.
- Roadmap phase.

Assess obligations according to actual planned features, including applicable privacy, disclosures, deletion, permissions, tracking, authentication, payments, age/content, and user-generated-content requirements.

Do not assume undeclared features exist.

Design mandatory prerequisites early and implement them in their relevant feature phases. Perform final submission/release verification in the production-v1 phase. Put optional store enhancements and polish in that final phase.

Record policy uncertainty and product decisions that affect applicability.

## 8. OpenTelemetry observability

Design observability from day one, particularly traces and performance metrics.

Specify instrumentation boundaries, propagation, sampling, performance overhead, and sensitive-data handling. Verify current Flutter web/mobile support before choosing its instrumentation approach.

Select a small, trusted collector/storage/UI stack and explain the choice, resource cost, and developer workflow.

Schedule implementation and verification in an explicit phase before release.

Provide one simple configuration switch that disables telemetry and unnecessary observability containers without breaking core services or generating repeated connection errors.

Test both enabled and disabled modes. Re-enabling must not require code changes. Preserve useful basic logs and health/readiness checks while telemetry is disabled.

## 9. Nginx and development-compatible security headers

Use Nginx as the main web server/reverse proxy for:

- Flutter web.
- Go APIs and relevant real-time connections.
- Development services and observability UIs.

Document routing, health checks, timeouts, WebSockets, asset handling, and service exposure.

Translate the request for “A+++ security headers” into explicit headers, named audit criteria, and compatibility tests. Do not promise an undefined grade.

Use separate local-development and production profiles.

Local settings must preserve localhost access, HTTP development, hot reload, debugging, WebSockets, source maps, and Flutter asset/worker behavior.

Justify CSP, CORS, HSTS, and cross-origin isolation settings against actual runtime needs. Avoid production restrictions that interfere with fast local development.

## 10. Make commands implemented through xops

Document thin Make targets backed by Python scripts under xops/makefile/, following existing helpers and standard-library conventions.

Required command contracts:

- make web.re: restart/rebuild/recreate Flutter web as necessary so current changes appear, including browser and service-worker caching behavior.
- make go.re: rebuild and restart the Go backend, then verify readiness.
- Consistent equivalent commands for other containerized services.

Define readiness checks, timeouts, logs, exit codes, dependency scope, idempotence, and failure recovery.

Restart/rebuild commands must preserve persistent volumes and must never silently delete data.

Mark these commands as planned until implemented in their roadmap phases. Do not modify Makefiles or executable scripts during this documentation task.

## 11. Internationalization from day one

Use .arb files as canonical UI message catalogs with standard Flutter localization tooling.

Document:

- Locale naming, discovery, and fallback.
- ICU placeholders and pluralization.
- RTL support.
- Date and number formatting.
- Missing-key and placeholder-parity checks.
- The workflow for adding any language without architectural changes.

Use stable backend message/error codes with arguments for client localization.

If server-rendered localized text is needed, document a suitable catalog generation/conversion strategy. Do not assume Go directly consumes ARB without supporting tooling.

## 12. Complete the documentation and roadmap

Instantiate all applicable templates, including:

- Project charter, decision log, and glossary.
- System architecture.
- Relevant API and module contracts.
- Topic-specific designs.
- Architectural decision records.

Preserve template structure and length conventions. Fill sections with project-specific content; use justified N/A entries where appropriate.

Clearly label proposed APIs, modules, commands, schemas, and tests as planned. Do not imply implementation exists.

Cover simulation, rules/configuration, component boundaries, data contracts, testing, developer workflows, internationalization, observability, Nginx, deployment, and store obligations.

Update documentation indexes and docs/tracking/context.md. Update root README.md, xops/README.md, and CHANGELOG.md only where needed to accurately describe this documentation work and planned workflows.

docs/planning/ROADMAP.md is the sole authoritative sequenced checklist.

Replace scaffold examples with phases containing:

- Goal and scope identifier.
- Dependencies.
- Bounded, actionable deliverables.
- Test plan.
- Objective acceptance and exit gates.
- Links to canonical documentation.

Phase 0 must deliver the CLI simulation.

The final production-v1 phase must cover Compose deployment, final operational/security hardening, load and recovery verification, backups and tested restoration, rollback, release/store readiness, and optional polish.

Necessary architecture, reliability, testability, internationalization, store, and instrumentation foundations must begin earlier.

Keep Kubernetes/Helm implementation explicitly post-v1, outside the v1 phase sequence.

Do not mark future implementation complete merely because its documentation is finished.

Create a traceability matrix mapping every requirement and applicable template section to its canonical document, roadmap item, acceptance evidence, and status. Avoid competing roadmaps and duplicated sources of truth.

## 13. Validate, track, stage, and stop

Complete the authorized documentation scope without asking whether to continue between files.

Use independent agents for bounded research or review where useful. Only the coordinating parent owns tracking and staging.

Audit the final documentation for:

- Missing requirements.
- Contradictions and game-rule drift.
- Unresolved placeholders.
- Incorrect paths and links.
- Markdown/Mermaid issues.
- Unsupported claims.
- Duplicated sources of truth.
- Incorrect dependencies or completion statuses.

Run applicable existing documentation/repository checks. Do not run nonexistent planned commands or claim application tests passed.

Follow AGENTS.md failure-recovery procedures.

Inspect the complete diff and staging set, preserving and distinguishing pre-existing work. Check for secrets.

After required gates pass, follow the repository’s tracking procedure: append the proper completion row through xops/agent/tracking_append.sh, using agent=codex, action=commit, status=completed, commit_sha=pending, and a Conventional Commit summary; then stage according to AGENTS.md.

NEVER run git commit or git push. The human performs those actions.

Finish with a concise report of the actual outcome, changed/staged files, run_id, verification results, and links to material assumptions or unresolved decisions.

## 14. Additional Project Requirements

### 14.1 Skills and Working Practices
- Use `.agents/skills` as the primary skills library.
- Before acting on a request, identify and follow the relevant skills. Revisit them when the task’s scope changes.

### 14.2 Offline Gameplay and Bots
- Non-web versions must support unlimited, fully offline gameplay with locally available bots and no runtime network dependency.
- The web version requires an online connection to play.
- Bots must support multiple difficulty levels.
- Evaluate standard Dart/Flutter code against native bot logic integrated through FFI. Prioritize performance and compatibility across supported platforms, and document the chosen approach and its tradeoffs.

### 14.3 Introduction and Player Guidance
- Introduce the game through an interactive offline tutorial or a short video.
- Tutorial games must use the lowest bot difficulty. Bots should remain competent, follow the rules, and make coherent decisions that help players learn.
- Provide contextual tips on tap and use snackbars for timely reminders.

### 14.4 Naming Conventions
- Use `com.cgmafiastate` as the Flutter application’s platform identifier. Distinguish this from language-specific package naming where necessary.
- Use `cgms` for Go component naming where applicable.
- Use `cgms` or `cgmafiastate` consistently for other applications, services, and configuration identifiers.

### 14.5 Art Direction and Frontend Design
- Treat [UI baseline v2](../design/UI_BASELINE.md) as the approved UI/UX and theme contract. Consult it and `.agents/skills/cgms-art-direction` before changing any client interface, interaction, or frontend component; follow the baseline's authority order and change policy.
- Preserve the existing visual foundation:
  - Cards and artwork: `docs/design/cgmsart-decks/cards`
  - Typography: `docs/design/cgmsart-decks/fonts`
  - Supporting configuration: `docs/design/cgmsart-decks`
- Develop distinct adaptive compositions in `client/` and its reusable `cgms_ui` package. Use the retained twenty mockups as supporting storyboards; future integration and legal states must follow UI baseline v2 and existing cgmsart assets.

### 14.6 Logging Standards
- Provide well-formatted, detailed, human-readable logging across all components, with particular attention to Go.
- Define and document a shared logging convention for Go, Flutter, and other services.
- Keep severity levels, timestamps, event names, and contextual fields consistent wherever practical. Include useful diagnostic context without exposing secrets or sensitive information.

### 14.7 Monetization
Update the roadmap and related product, design, and architecture documentation to incorporate the following model:

- **Offline play:** Unlimited gameplay with local bots.
- **Free online play:** Three online games per user per day.
- **Premium products:** Only Ad-free, Daily, Weekly, Monthly and Yearly. Every paid Daily/Weekly/Monthly/Yearly tier provides unlimited online starts and no ads. Paid Daily lasts 24 hours from confirmed purchase and does not renew automatically. Standalone Ad-free removes ads for 30 days without changing the free-start cap; separate unlimited access may coexist. Prices and other renewal rules are not set by this local implementation.
- **In-game currency:** A currency named `dirt`, earned from every online game.
- **Currency spending:** Only Daily and Weekly unlimited-online access can be redeemed with earned dirt. These earned packages retain ads unless a separate ad-free entitlement is active. Ad-free, Monthly and Yearly cannot be redeemed with dirt. Dirt is not sold.
- **Card styles:** All four existing cgmsart styles are automatically assigned across players for an entire match, consistently for all authorized viewers. Captured cards use their capturing player's style. No cosmetic sales or currency-based unlocks.

Use the explicit current decisions in DECISION_LOG.md and DESIGN-economy-and-stores.md for rewards and prices; do not invent amounts. Decision 0015 supersedes the former one-to-six-day and ad-free dirt-pass model.

Apple App Store and Google Play requirements are mandatory constraints. Verify current official policies and document how subscriptions, payments, advertisements, currency, and promotional entitlements comply. Resolve policy conflicts before implementation.

### 14.8 Promotional Codes — superseded scope
Decision 0015 excludes new promotional products and campaigns. Keep existing
historical receipt/reconciliation safeguards, but do not expose issuance or
redemption channels. Earlier internal fixtures are historical contract evidence,
not authorization to introduce more products or a prerequisite cosmetic catalogue.
