# Online economy and store boundaries

The sole completion checklist remains [ROADMAP](../planning/ROADMAP.md).
Decisions 0015–0018 supersede conflicting product and appearance assumptions in
older implementation reports. The platform research below was refreshed from
official sources on 2026-10-01 before SDK selection; it does not certify store
acceptance.

## Adopted products and eligibility

| Product | Earned dirt redemption | Ad behavior / activation |
|---|---|---|
| Ad-free | Never | 30 days of ad removal; the three-free-starts-per-UTC-day cap still applies unless separate unlimited access is active. |
| Daily | Yes, exactly one day | **Paid:** 24 hours of unlimited online starts and no ads from confirmed purchase; non-renewing. **Earned:** unlimited online starts with ads unless separate ad-free access is active. |
| Weekly | Yes, exactly seven days | **Paid:** unlimited online starts and no ads. **Earned:** unlimited online starts with ads unless separate ad-free access is active. |
| Monthly | Never | Paid monthly access provides unlimited online starts and no ads. |
| Yearly | Never | Paid annual access provides unlimited online starts and no ads. |

There are no cosmetic sales, cosmetic currency unlocks, sold dirt or new
promotion campaigns. All four card styles are assigned automatically for an
entire match, without entitlement checks; see [UI baseline](UI_BASELINE.md).
Live purchases remain unavailable. Decision 0018 approves the durations and
benefits above and Daily's non-renewing behavior. Prices and renewal rules for
the other paid products remain outside this local contract and must not be
inferred from provider fixtures.

[Decision 0015](../project/DECISION_LOG.md) explicitly supersedes decision 0013's
ad-free earned passes. Unlimited online access and ad removal are separate
benefits. New Daily/Weekly redemptions cannot create or extend ad-free access.
The migration preserves the original expiry of already-promised legacy ad-free
passes as a separate benefit; it never extends that expiry when new dirt is spent.

## Allowances, earning and redemption

Retain three free online starts per UTC date, once-per-start charging, one reward
per financially finalized non-nullified online game, and no automatic refunds or
rewards for unfinished games. Offline play earns no dirt. Decision 0014's adopted
reward is 10 dirt per eligible game and its price is 30 dirt per day. The implementation preserves that accepted per-day calculation: Daily costs 30
and Weekly costs 210. These are calculated from the existing rate, not a newly
chosen price or a claim that the owner separately approved a Weekly discount.

A lost start response recovers the original operation rather than charging again.
At 23:59 UTC the third free start exhausts that date's allowance; at 00:00 UTC a
new allowance applies without changing the older game's charge. An unfinished
game remains resumable without an invented timeout or outcome.

New redemptions accept only one or seven days. Add the duration to the later of
now or current pass expiry. A Daily redemption one day before expiry extends that
expiry by one day. Original receipts for retired two–six-day offers remain
recoverable and idempotent; their original duration/cost/expiry are not rewritten,
and they do not authorize a new purchase of a retired duration.

## Local implementation and privacy contract

`backend/internal/economy` owns product accounting separately from score, cash
and debt. Additive, checksum-pinned migrations preserve old receipts and legacy
benefits. `ReserveStartTx`, `RewardFinalizedTx` and `BuyPassTx` remain scoped to
the caller's transaction with stable operation identities and ordered account
locks. Nullified games earn nothing; replacement games are new starts. A failed
next-game reservation leaves the finalized game intact until admission succeeds.

`CGMS_ECONOMY_ENABLED=1`, or `economy_enabled: true`, enables the adopted policy
for new matches; existing economy matches retain their pinned policy. Startup
applies migrations even when new-match activation is off. The API's authenticated
account view exposes only the caller's benefits, remaining allowance and next UTC
reset. `POST /v1/economy/passes` retains `{operation_id, days}` and recovery by
`GET /v1/economy/passes/{operation}`. Unknown commits remain uncertain until the
original receipt is reconciled; a changed request body conflicts.

Successful redemption wakes automatic next-game admission. Public snapshots
expose only generic `admission_waiting`, never which account lacks access or its
balance. Private account/pass state is not added to shared game projections.

New cosmetic provider grants are rejected. Historical transactions remain
reconcilable for duplicate/restore/revocation handling without reopening cosmetic
sales or unlocking card styles. Existing internal promotion records are historical
compatibility data, with no public issuance/redemption channel. Verified provider
fixtures remain synthetic: pending/unknown status cannot grant, account/product
binding is immutable, revisions are ordered, and absolute expiry prevents double
grants. No live provider verification, native SDK, store sandbox or signing is
implemented by this change.

The local fixed-term Ad-free contract exposes private `ad_free_until` separately
from unlimited-play expiry. It combines independently valid paid/grandfathered
ad-free sources; spending earned dirt never extends it. Repeated verified
notifications and restores retain the original absolute expiry. Revoking the
standalone benefit does not remove an independently valid earned pass.

Decision 0018 fixes new standalone Ad-free at 30 × 24 hours and paid Daily at
24 hours from the original provider-verified confirmation instant. The trusted
internal `ConfirmedAt` field is private provider metadata, never accepted from
client JSON or returned in public game/account projections. Validate its exact
expiry rather than calculating a fresh expiry during restore or notification
processing. A future confirmation cannot activate benefits early. Daily is
non-renewing; extending either fixed-term offer requires a distinct verified
purchase identity. Historical unanchored Ad-free receipts keep their original
absolute expiry without gaining time. Weekly/Monthly/Yearly retain verified
provider period expiry and all grant both unlimited starts and ad removal; no
local calendar-duration or renewal policy is invented for their future adapters.

## Verification and outstanding decisions

Regression gates cover one/seven-day eligibility, legacy receipt recovery,
unlimited/ad-free separation, grandfathered expiry, no new cosmetic grants,
atomic allowance/reward/redemption, duplicate provider events and uncertainty.
Rendered browser acceptance must verify the final Flutter build against the real
local authority, including private account views, lost-response recovery and
phone/tablet/desktop layouts. Historical evidence below does not prove the new
policy; the current task report records the fresh executed checks.

The local Ad-free duration and paid Daily benefits are approved in decision 0018.
Real purchase prices, other renewal rules, provider adapters, native/store
acceptance and production remain outside this implementation. P03 bot work is
excluded. The [approved-term report](../reports/2026-10-01-approved-paid-access.md)
records the completed local P06 gate and its native/store limitations.

<a id="platform-applicability-matrix"></a>

## Platform applicability matrix — checked 2026-10-01

The official sources linked in each row were checked for this P06 research
slice. Platform facts and CGMS interpretations are separate; product names alone
do not select a store product type, renewal behavior or SDK.

| Concern | Apple App Store | Google Play | CGMS applicability / remaining boundary |
|---|---|---|---|
| Paid digital access | §3.1.1 requires IAP for paid in-app functionality; §3.1.1(a) has storefront-specific external-link exceptions. [Review guidelines](https://developer.apple.com/app-store/review/guidelines/) | Play-distributed paid digital features normally require Play Billing; exceptions and regional programmes have separate conditions. [Payments policy](https://support.google.com/googleplay/android-developer/answer/9858738?hl=en) | Ad removal and online access are digital features. Use the standard store path as the planning assumption; no external checkout or regional exception is enabled. |
| Standalone fixed-term Ad-free | Non-renewing subscriptions provide limited-duration service without automatic renewal; non-consumables do not expire. [Product types](https://developer.apple.com/help/app-store-connect/reference/in-app-purchases-and-subscriptions/in-app-purchase-types/) | Prepaid subscriptions do not auto-renew; one-time rental options also support limited-duration access. [Prepaid lifecycle](https://developer.android.com/google/play/billing/lifecycle/subscriptions), [rental options](https://developer.android.com/google/play/billing/one-time-product-multi-purchase-options-offers) | **Mapping inference:** evaluate a non-renewing/prepaid or suitable rental representation for the approved 30-day term. Do not silently turn Ad-free into a permanent unlock, recurring charge or unlimited-play bundle. No product mapping is selected here. |
| Paid Daily; paid Weekly/Monthly/Yearly | Auto-renewable periods must be at least seven days and provide ongoing value. [§3.1.2(a)](https://developer.apple.com/app-store/review/guidelines/) | Prepaid and auto-renewing subscription lifecycles differ; a top-up is a new purchase with its own acknowledgement. [Subscription lifecycle](https://developer.android.com/google/play/billing/lifecycle/subscriptions) | A one-day Apple auto-renewable product is incompatible. Paid Daily is an approved non-renewing 24-hour unlimited/ad-free pass. All paid Weekly/Monthly/Yearly access also removes ads; their names do not authorize auto-renewal, calendar arithmetic, prices or discounts. |
| Earned-only Daily/Weekly | The cited IAP rules do not establish a blanket earned-point exception for CGMS; provide the actual earning/redemption flow to App Review. [§3.1](https://developer.apple.com/app-store/review/guidelines/) | Earned or awarded points may be issued and exchanged for digital services without Play Billing; selling such points requires billing. [Payments FAQ](https://support.google.com/googleplay/android-developer/answer/10281818?hl=en) | Dirt is never sold. Only Daily/Weekly are dirt-redeemable and retain ads unless another ad-free benefit is active. Google's permission is not evidence of Apple acceptance. No cosmetic or promotional offering is added. |
| Restore and account binding | `appAccountToken` associates a transaction with the app's customer; non-renewing subscriptions require app-managed duration and cross-device restoration. [Account token](https://developer.apple.com/documentation/appstoreservernotifications/appaccounttoken), [subscription handling](https://developer.apple.com/documentation/storekit/handling-subscriptions-billing) | Verify purchases on the backend; `purchaseToken` is the durable unique key, not `orderId`. Verify attached account identifiers before granting. [Security](https://developer.android.com/google/play/billing/security), [integration](https://developer.android.com/google/play/billing/integrate) | Restore the original verified entitlement to its bound CGMS account without buying again or extending expiry. Keep app/product/environment binding and existing cross-account rejection. A local fixture is not proof of real-store restore. |
| Pending or interrupted payment | A pending result requires further customer action; a later successful transaction arrives through StoreKit updates. [Pending purchases](https://developer.apple.com/documentation/storekit/product/purchaseresult/pending) | Grant and acknowledge only after verified `PURCHASED`; reconcile purchases on resume after missed callbacks. [Integration](https://developer.android.com/google/play/billing/integrate) | Pending, unknown and a missing response do not authorize a grant or a replacement purchase. Preserve the existing immutable identity until authoritative reconciliation resolves it. |
| Duplicate, revoked and expired purchases | Deduplicate notifications with `notificationUUID`; transaction `revocationDate` identifies refunds/revocations. [Notification identity](https://developer.apple.com/documentation/appstoreservernotifications/notificationuuid), [revocation](https://developer.apple.com/documentation/storekit/transaction/revocationdate) | Deduplicate RTDN `messageId`, then query the Developer API for complete state; handle voided purchases. [RTDN contract](https://developer.android.com/google/play/billing/rtdn-reference), [security](https://developer.android.com/google/play/billing/security) | A transport notification is not a new purchase or an ordered entitlement revision. Reconcile verified state and expiry once; retries cannot grant twice. Revocation of one benefit must not erase separately valid earned or paid access. |
| Account deletion | Apps with account creation must offer in-app deletion, including automatically created guest accounts. Explain billing separately; deleting an app account does not cancel Apple renewal. [Deletion guidance](https://developer.apple.com/support/offering-account-deletion-in-your-app/) | Provide an in-app deletion path plus a functional external deletion-request URL; delete associated data and disclose legitimate retention. [Deletion requirements](https://support.google.com/googleplay/android-developer/answer/13327111?hl=en) | Audit guest and registered paths, retained fraud/accounting records and disclosures before release. Existing session revocation/anonymization alone does not prove store compliance; no public deletion website is deployed by this research. |
| Ads and privacy disclosures | App privacy declarations include integrated third-party practices. Tracking requires ATT where applicable. [App privacy](https://developer.apple.com/help/app-store-connect/manage-app-information/manage-app-privacy), [tracking guidance](https://developer.apple.com/app-store/user-privacy-and-data-use/) | Data safety must describe actual collection/sharing; the developer remains responsible for SDK behavior. [Data safety](https://support.google.com/googleplay/android-developer/answer/10787469?hl=en), [User Data](https://support.google.com/googleplay/android-developer/answer/10144311?hl=en) | “Ads remain active” describes entitlement policy, not installed advertising technology. Before choosing an ad SDK, inventory account/purchase identifiers, telemetry, retention and partners. No live ad delivery, tracking consent or privacy-label acceptance is claimed. |

### Consequences for the future native adapter

Two duration details matter before selecting an SDK. For an Apple non-renewing
product, restore must re-evaluate the approved fixed term: `currentEntitlements`
includes the latest non-renewing transaction even when finished, so its presence
alone cannot prove remaining time. See [current entitlements](https://developer.apple.com/documentation/storekit/transaction/currententitlements).
For a Google prepaid plan shorter than one week, acknowledgement is due within
half its duration; plans of a week or longer allow three days. The clock starts
when the purchase becomes `PURCHASED`, not while payment is pending. Thus a one-day
prepaid representation would imply a **12-hour acknowledgement deadline**
(calculation from the policy), not the generic three-day assumption. See
[prepaid lifecycle](https://developer.android.com/google/play/billing/lifecycle/subscriptions)
and [pending purchases](https://developer.android.com/google/play/billing/integrate#handle-pending-transactions).
These are integration constraints, not CGMS duration decisions.

For example, a lost purchase callback followed by a duplicate notification must
restore the same verified purchase and original expiry, with no second grant.
A later verified revocation removes only that purchase's benefit. These cases
belong in provider fixtures now and in signed store tests at R06N/R06; this
research does not claim that either store is connected to the current client.

Standalone Ad-free is 30 days; paid Daily is a non-renewing 24-hour unlimited/ad-free
pass. Paid Weekly/Monthly/Yearly are also unlimited and ad-free. Prices, renewal
consent where applicable, territory selection and product-to-store mapping must
be explicit before native activation. Current earned prices and match-long styles
are unchanged. SDK selection, provider
credentials/endpoints, signed sandbox purchases/restores, deletion-site delivery,
advertising integration and store submission remain deferred to the relevant
R06N/R06 gates. No broader roadmap gate is completed by this matrix alone.


## Retained historical verification

The [non-production report](../reports/2026-10-01-nonproduction-verification.md)
records the earlier one–six-day/ad-free earned-pass behavior, concurrency, once-only
rewards, recovery and provider fixtures. Those dated results remain historical,
not acceptance of decisions 0015–0016. Its raw browser archive was removed under
the owner's [directory cleanup instruction](../reports/2026-10-02-screenshot-retention.md).
The former cosmetic-catalogue and
promotion-channel blockers were removed from product scope, not completed by
implementing sales or promotions. ROADMAP retains all applicable unverified gates.
