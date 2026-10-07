# Screenshot retention and push investigation — 2026-10-02

The owner's final instruction removes **every file and subdirectory under
`client/screenshots/` except `cgmsart-phone.png`**, including JSON reports, logs,
copied harnesses and its README. This replaces both earlier selective retention
and the later image-only cleanup policy. The earlier image sweeps and final
directory cleanup are distinguished below; the archive was not relocated.
The phone reference remains byte-identical: **273,544 bytes**, SHA-256
`2c116c49cc67e707d7dd6d45e12aa2063aa014c5edc3aa3416a8a36eea18bc8b`.
Artwork, the twenty source mockups, test fixtures and application assets are not
screenshots and remain intact.

The available evidence does **not establish a screenshot-caused push failure**.
No push was attempted. No oversized-file rejection or failed-push message was
found in project tracking or available safe-run logs. Deleting working-tree images
reduces future tree contents; it does not remove earlier Git blobs.

## Inventory and disposition

The [per-file inventory](2026-10-02-screenshot-inventory.csv) records the original
**1,806 indexed screenshot images / 232,375,977 bytes**, their index status,
SHA-256, purpose and references. **264 images / 39,128,350 bytes** were staged
additions from preceding work. Of the original inventory, **1,805 screenshots /
232,102,433 bytes** are now removed and only the phone reference remains. The
first selective cleanup removed 262 images / 31,627,881 bytes; the owner's later
instruction removed the other **1,543 images / 200,474,552 bytes**. These are exact
uncompressed file totals, not measured Git-transfer savings.

| Original screenshot group | Original count / bytes | Images now retained |
|---|---:|---:|
| Root widget renders and historical online acceptance | 12 / 3,083,727 | 1 phone render |
| `design-partner-20261001` | 687 / 80,793,064 | 0 |
| `nonproduction-20261001` | 122 / 16,399,358 | 0 |
| `nonproduction-recheck-20261001` | 216 / 28,797,740 | 0 |
| `styles-economy-20261001` | 253 / 29,579,823 | 0 |
| `p06-local-20261001` | 90 / 11,718,778 | 0 |
| `p06-approved-20261001` | 162 / 22,875,137 | 0 |
| `phone-card-proportions-20261001` | 192 / 27,327,932 | 0 |
| `card-art-gestures-20261001` | 72 / 11,800,418 | 0 |

The first expanded workspace scan found **987 ignored browser captures** in
`client/.browser-tools/`, which were removed. After all final browser checks and
visual reviews completed, the final sweep removed **817 additional ignored
captures / 224,921,443 bytes**. Together, **1,804 ignored browser captures** were
removed; their paths are included in the inventory. **996 captures / 255,051,164
bytes** have exact recorded sizes and hashes; 808 earlier ignored-capture removals
have exact paths/counts but no retained byte
sizes or hashes because the initial sweep hit an old root-owned directory before
flushing its inventory. These unknown values are explicitly marked, not estimated.
The deletion script now writes its complete plan before removing files. The image
sweep inventory contained **3,610 unique paths**: the retained phone reference and
**3,609 removed screenshots**. The recorded removed bytes total **487,153,597**,
plus the unrecorded sizes of those 808 ignored captures. No complete byte total
is claimed for those unknown files.

No screenshot images were found elsewhere in the workspace after excluding actual
source artwork, mockups, fixtures and application/browser assets. In particular,
`client/.browser-tools/build-focus-diagnostic/assets/` contains copied application
artwork, not screenshots; those 209 image assets were preserved.

The final directory cleanup removed its JSON reports, manifests, logs and copied
harnesses as well. Existing summaries under `docs/reports/` remain, including
failed-run outcomes and their later resolutions. The [adaptive repair report](2026-10-02-adaptive-startup-recovery.md)
records the Firefox sequence's verified closure; deletion does not change the
historical failure result. Paths and hashes in the inventory identify removed
files; capture-time counts must not be interpreted as current file availability.
Historical report links into the deleted archive are now plain text, with an
explicit archive-removal notice on each affected report.

## Git history and unresolved push diagnosis

Read-only `git rev-list --objects --all` plus `git cat-file --batch-check` examined
**4,428 reachable blobs / 976,286,522 uncompressed bytes**. There are no blobs at
or above 50 MiB. The largest blob is required artwork:
`docs/design/cgmsart-decks/sheets/first-light-52.png`, **17,590,390 bytes**, introduced
by `2626c38` (`docs(cgms): replace Civic Ledger with illustrated cgmsart direction`).
An earlier artwork version remains at
`docs/design/afterimage-decks/sheets/first-light-52.png`, **17,589,724 bytes**.
The largest historical screenshot blob is the now-removed desktop widget render,
**837,701 bytes**. History contains **2,050 screenshot-path blobs / 213,134,872
uncompressed bytes**, including older versions and paths.

The committed retired screenshots remain reachable through:

| Commit | Retired screenshot paths |
|---|---|
| `dd2858c` | Relocated screenshot archives and root online/widget captures |
| `0775248` | `client/screenshots/nonproduction-recheck-20261001/**` |
| `6248bb5` | `client/screenshots/design-partner-20261001/**` |
| `90eef62` | `client/screenshots/styles-economy-20261001/** and P06 archives` |

At inspection, `HEAD` and the **locally cached** `origin/main` both named
`90eef62`, with ahead/behind counts `0/0`. A subsequent noninteractive, read-only
`git ls-remote origin HEAD` succeeded and returned
`90eef625e34564261392a7712f983b764a591df6`, confirming that the remote HEAD was
reachable and matched at that time. This does not test push authorization,
server receive policy or a future transfer. No remote refs or history changed.
The object store reported 853.17 MiB in packs and 91.96 MiB loose;
neither figure is a measured next-push payload. Working-tree deletion will not
shrink published history or prove that a push succeeds.

If the next human-run `make git` still fails, preserve its exact rejection and
identify whether the problem is authentication, transport, server policy or pack
size. Agents may inspect that evidence, but AGENTS.md prohibits agent commits,
pushes, force-pushes and rewriting published history. Any eventual repository
history migration requires a separately planned, human-owned operation; none was
performed here. No historical oversized-file blocker was demonstrated, so no
history rewrite is proposed as a required fix.

## Prevention and verification

`client/.gitignore` excludes every file and subdirectory beneath
`client/screenshots/` with only `cgmsart-phone.png` excepted. Mockups, artwork,
fixtures and application assets outside the capture directory are unaffected.
Browser runs write temporary captures to ignored `.browser-tools/` directories;
after visual review, remove those images and keep lightweight verification output
under `docs/reports/`, never in the reserved screenshot directory.

The first removal attempt encountered root-owned directories left by old WebKit
container runs. Both failure logs were inspected. Ownership was repaired on only
eleven workspace capture directories using the already-installed `alpine:3.24.2`
image; no workstation package, global configuration or persistent service data was
changed. The final removal run passed and its recovery record was marked resolved.

Verification preserves the phone hash, checks that it is the directory's sole
entry, compares protected files outside that directory, checks ignore rules,
and validates that Markdown links no longer target deleted files. New captures
created during subsequent browser verification receive a final removal sweep.
Screenshot removal changes no application behavior. Subsequent application and
browser verification is recorded separately in the
[shared vertical layout report](2026-10-02-shared-vertical-layout.md).

The initial final-tree verification passed: the sole screenshot was the exact
phone reference; all **1,907 protected evidence/artwork/mockup files** matched
their previous hashes; all **2,793 inventory paths** had the recorded disposition;
17 ignore assertions and owned-document screenshot-link checks passed.
`git diff --check` also passed. Evidence:
`screenshot-owner-verify--20261002T044443Z-524401`.

The final sweep ran only after the coordinating agent finished every browser run,
reviewed the final phone/tablet/desktop, inspector, standings and style captures,
and stopped the owned QA fixtures. Its inventory was flushed before deletion;
all 817 new captures had recorded sizes and hashes. At that checkpoint, no
screenshot images remained in browser output directories, and the phone reference
was the only screenshot in the workspace, excluding source artwork, mockups,
fixtures and application/browser distribution assets.

Final deletion evidence: `screenshot-owner-final-sweep--20261002T055352Z-706079`.
Final verification passed with **1,907 protected hashes**, **3,610 inventory
dispositions**, **22 ignore assertions**, and all **7 screenshot links across
158 Markdown files** valid. Evidence:
`screenshot-owner-final-verify-repaired--20261002T055648Z-711758`.
The first final verification found 262 obsolete `retired generated image`
disposition labels from the earlier selective cleanup; those files were already
absent. After inspecting that failure, their disposition was updated to the
owner-requested removal wording while preserving their paths, hashes, sizes,
purposes and references, and the unchanged verification assertions passed.

Deletion evidence: `screenshot-owner-delete-final--20261002T044055Z-517123`.
Permission diagnosis: `screenshot-owner-delete--20261002T043851Z-512528` and
`screenshot-owner-delete-repaired--20261002T044022Z-515870`; both were inspected
before repairs and resumption. Read-only remote evidence from the earlier audit:
`screenshot-remote-readonly--20261001T215111Z-3960454` (exit 0).

## Final directory-only cleanup

The subsequent instruction removed **660 remaining non-image files / 2,965,966
bytes** and **250 subdirectories**, leaving only the byte-identical phone
reference. This includes 215 logs, 191 JSON reports/manifests, 117 command records,
116 exit records, 15 text files, four copied browser harnesses, the directory
README and one extensionless record. These were archived outputs, not active
application or regression-test sources. None were relocated.

Their sizes, SHA-256 values and dispositions were appended to the existing
inventory before removal, bringing it to **4,270 unique paths**: 3,609 previously
removed images, 660 removed archive files and the retained phone reference.
The known removed-byte total is **490,119,563**, plus the previously disclosed
808 ignored image files with unrecorded sizes. The final cleanup preserves
**1,248 previously protected files outside `client/screenshots/`**; the earlier
1,907-file check predates the user's explicit removal of 659 protected archive
records inside that directory.

Inventory: `screenshot-directory-final-inventory--20261002T073731Z-916628`.
Deletion: `screenshot-directory-cleanup--20261002T073853Z-919405`.
Directory, hash, inventory and ignore verification passed:
`screenshot-directory-verify--20261002T074222Z-926834` (1,248 external hashes,
4,270 inventory rows and 16 ignore assertions). Historical report links into the
removed archive were replaced with plain capture-time descriptions; the surviving
reports explicitly disclose that the raw archives are no longer available.

The subsequent compact-quadrant verification generated **809 temporary captures /
173,736,401 bytes** in ignored browser artifact directories. All were inventoried
before removal after visual review, bringing the existing inventory to **5,079
unique paths**. No captures remain in those artifact directories; the sole phone
reference and all 1,248 protected external hashes remain unchanged. Compact case
outcomes and failure history remain in the
[quadrant verification manifest](2026-10-02-quadrant-verification.json).
Cleanup evidence: `quadrant-final-capture-cleanup--20261002T090315Z-1123233`.

The public-size follow-up generated **248 temporary captures / 61,526,125
bytes**, including regression and inspected fixture-failure captures. Their
size/hash inventory was flushed before deletion after final visual review.
The inventory now contains **5,327 unique paths**. No browser capture images
remain; the sole phone reference and all 1,248 protected external hashes still
match. Case outcomes and measurements remain in the
[public-growth manifest](2026-10-02-public-growth-verification.json).
Cleanup evidence: `public-growth-capture-cleanup--20261002T173156Z-2014563`.

The full-width follow-up generated **226 temporary captures / 67,513,510
bytes**. They were inventoried before deletion after visual review, including
the failing regression captures. The inventory now contains **5,553 unique
paths**. No browser capture images remain; the phone reference and all 1,248
protected external hashes are unchanged. Lightweight case outcomes, measurements
and failure diagnoses remain in the
[full-width manifest](2026-10-02-public-width-verification.json).
Cleanup evidence: `fullwidth-capture-cleanup--20261002T180804Z-2092698`.
