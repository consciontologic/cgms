# CGMS UI asset notices

This package bundles 208 native illustration crops across First Light, Rain Glaze,
Ember Glaze and Drypoint, and
the single shared concealed back from the repository's
[`docs/design/cgmsart-decks`](../../../../docs/design/cgmsart-decks/README.md) exploration.
The files are copied without recoloring, repainting or replacing their source
pixels. [`manifest.json`](manifest.json) records copied asset paths, their source
crop references and SHA-256 checksums. Rank and suit labels are rendered live by
Flutter rather than embedded in those paintings.

The original deck's `production.json`, `atlas-crops.json` and `manifest.json`
retain the image-generation prompts, output IDs, source checksums, crop rectangles
and complete provenance. The illustrations were generated for the CGMS cgmsart
exploration using the repository's original concept as a style reference.
No new deck artwork is commissioned by this integration. The authority assigns
one of the four styles to each player for the whole match; authorized visible
cards use their current controller's public style. Artwork never determines
physical copy identity, ownership, rarity or game rules.
Human faces are confined to J/Q/K artwork; aces and numbers use objects and scenes.
All concealed cards use exactly the same back.

The owner accepted First Light as the screen UI's visual baseline on 2026-09-29
and all four styles for shared match presentation on 2026-10-01; see the
[UI baseline](../../../../docs/design/UI_BASELINE.md). That acceptance
does not change their recorded generation provenance or licensing status.
The native illustration crops are approximately 280 × 350 pixels. Enlarged web
inspection resamples them; these are not high-resolution print masters.
Inclusion does not claim exclusive rights or third-party clearance.

Original source code and vectors use the package's [MIT LICENSE](../LICENSE).
The generated artwork's exploration/provenance status is not replaced by that
source-code license.

The unmodified Barlow Condensed Bold and Atkinson Hyperlegible Next Regular/Bold
font binaries are bundled locally under SIL Open Font License 1.1. Keep their
full original notices, authors and contributor files on redistribution:

- [Barlow OFL](fonts/Barlow-OFL.txt)
- [Atkinson Hyperlegible Next OFL](fonts/AtkinsonHyperlegibleNext-OFL.txt)
- [Pinned upstream provenance and checksums](fonts/PROVENANCE.md)

The package font families retain their original names. Fonts require no system
installation, remote font service or runtime network request.

## Symbol fallback

Unmodified Noto Sans Symbols2 is bundled for missing suit/interface glyphs, preventing runtime font-service downloads. It is distributed under the [SIL Open Font License](fonts/NotoSansSymbols2-OFL.txt). [Font provenance](fonts/PROVENANCE.md#local-symbol-fallback) records the official pinned source and verified hashes. Barlow and Atkinson remain the primary typography.
