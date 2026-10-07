#!/usr/bin/env python3
"""Independent, offline acceptance checks for the art exploration outputs."""
from __future__ import annotations

import argparse
import hashlib
import json
from pathlib import Path
import xml.etree.ElementTree as ET

from PIL import Image, ImageChops, ImageDraw, ImageFont

ROOT = Path(__file__).resolve().parent
EDITION_IDS = ["first-light", "ember-glaze", "rain-glaze", "drypoint"]
SUITS = ["hearts", "diamonds", "clubs", "spades"]
RANKS = ["A", *map(str, range(2, 11)), "J", "Q", "K"]
COLORS = {"hearts": "#F5A4BE", "diamonds": "#FFD275", "clubs": "#83DCD3", "spades": "#B6B2FF"}


def digest(path: Path) -> str:
    return hashlib.sha256(path.read_bytes()).hexdigest()


def contrast(a: str, b: str) -> float:
    def luminance(hex_value: str) -> float:
        values = [int(hex_value[i:i + 2], 16) / 255 for i in (1, 3, 5)]
        linear = [v / 12.92 if v <= .04045 else ((v + .055) / 1.055) ** 2.4 for v in values]
        return sum(v * c for v, c in zip(linear, [.2126, .7152, .0722]))
    high, low = sorted([luminance(a), luminance(b)], reverse=True)
    return (high + .05) / (low + .05)


def verify(root: Path = ROOT) -> dict:
    manifest = json.loads((root / "manifest.json").read_text())
    cards = manifest["cards"]
    assert manifest["editions"] == EDITION_IDS
    assert manifest["ranks"] == RANKS and manifest["suits"] == SUITS
    assert manifest["card_size"] == [720, 1008]
    assert len(cards) == 208, f"Expected 208 cards, found {len(cards)}"
    expected = {(e, s, r) for e in EDITION_IDS for s in SUITS for r in RANKS}
    assert {(c["edition"], c["suit"], c["rank"]) for c in cards} == expected
    assert len({c["face"] for c in cards}) == 208
    assert len({c["artwork"] for c in cards}) == 208
    assert {c["back"] for c in cards} == {"shared/back.png"}
    assert len(list((root / "cards").glob("*/*/face.png"))) == 208
    assert len(list((root / "cards").glob("*/*/artwork.png"))) == 208
    assert len(list((root / "overlays").glob("*.png"))) == 52
    assert len(list((root / "overlays").glob("*.svg"))) == 52
    back = Image.open(root / "shared/back.png").convert("RGBA")
    assert back.size == (720, 1008)
    alpha = back.getchannel("A")
    assert alpha.getpixel((719, 0)) == 0 and alpha.getpixel((300, 0)) == 255
    seen_art = set()
    source_cache = {}
    rank_font = ImageFont.truetype(str(root / "fonts/BarlowCondensed-Bold.ttf"), 124)
    for card in cards:
        face_path = root / card["face"]
        art_path = root / card["artwork"]
        face = Image.open(face_path).convert("RGBA")
        art = Image.open(art_path).convert("RGB")
        overlay = Image.open(root / card["overlay"]).convert("RGBA")
        assert face.size == (720, 1008) and overlay.size == face.size
        assert ImageChops.difference(face.getchannel("A"), alpha).getbbox() is None
        # Outer twenty pixels are identical across every face and the shared back.
        for box in [(0, 0, 720, 20), (0, 988, 720, 1008), (0, 0, 20, 1008), (700, 0, 720, 1008)]:
            assert ImageChops.difference(face.crop(box).convert("RGB"), back.crop(box).convert("RGB")).getbbox() is None, card["face"]
        for box in [(20, 24, 146, 266), (574, 742, 700, 984)]:
            assert ImageChops.difference(face.crop(box).convert("RGB"), overlay.crop(box).convert("RGB")).getbbox() is None
        # The actual rank pixels must match this card's rank, independently of
        # filenames, manifest labels and SVG text. Keep the suit outside this box.
        for box, baseline in [((20, 24, 146, 174), 128), ((574, 742, 700, 892), 128)]:
            reference = Image.new("RGBA", (126, 150), "#1D2335")
            ImageDraw.Draw(reference).text((63, baseline), card["rank"], font=rank_font, fill="#F5F1EA", anchor="ms")
            assert ImageChops.difference(reference.convert("RGB"), overlay.crop(box).convert("RGB")).getbbox() is None, "Raster rank differs from identity"
        expected_color = tuple(int(COLORS[card["suit"]][i:i + 2], 16) for i in (1, 3, 5)) + (255,)
        assert overlay.getpixel((83, 213)) == expected_color
        assert overlay.getpixel((637, 931)) == expected_color
        assert card["label"] == {"rank": card["rank"], "suit": card["suit"], "font": "Barlow Condensed Bold", "color": "#F5F1EA", "suit_color": COLORS[card["suit"]]}
        assert card["overlay"] == f"overlays/{card['suit']}-{card['rank']}.png"
        assert digest(root / card["overlay"]) == card["overlay_sha256"]
        assert digest(face_path) == card["face_sha256"]
        assert digest(art_path) == card["artwork_sha256"]
        assert card["artwork_sha256"] not in seen_art, "An illustration is reused verbatim"
        seen_art.add(card["artwork_sha256"])
        assert list(art.size) == card["native_artwork_size"]
        source_path = root / card["source_atlas"]
        if source_path not in source_cache:
            source_cache[source_path] = (Image.open(source_path).convert("RGB"), digest(source_path))
        source, sha = source_cache[source_path]
        assert sha == card["source_sha256"]
        assert list(source.size) == card["source_size"]
        assert ImageChops.difference(art, source.crop(card["source_bbox"])).getbbox() is None, "Illustration pixels altered"
        svg = ET.parse(root / card["overlay_svg"])
        texts = [el.text for el in svg.iter() if el.tag.endswith("text")]
        assert texts == [card["rank"], card["rank"]]
        assert "edition" not in (root / card["overlay_svg"]).read_text()
    for edition in EDITION_IDS:
        sheet = Image.open(root / f"sheets/{edition}-52.png")
        assert sheet.size == (5592, 2700)
        for suit in SUITS:
            assert Image.open(root / f"sheets/{edition}-{suit}.png").size == (2800, 2520)
        per_edition = [(c["suit"], c["rank"]) for c in cards if c["edition"] == edition]
        assert per_edition == [(s, r) for s in SUITS for r in RANKS]
    assert Image.open(root / "previews/comparison.png").size == (2620, 4510)
    assert Image.open(root / "previews/artwork-comparison.png").size == (2620, 4510)
    assert Image.open(root / "previews/mobile-390.png").size == (390, 844)
    for size in [(96, 134), (120, 168)]:
        for card in cards:
            thumb = root / "previews/mobile" / f"{size[0]}" / card["edition"] / f"{card['suit']}-{card['rank']}.png"
            assert Image.open(thumb).size == size
    html = (root / "index.html").read_text()
    assert 'alt="Hidden card"' in html
    assert 'id="concealed"' in html
    concealed = html.split('id="concealed"', 1)[1].split("</section>", 1)[0]
    assert "shared/back.png" in concealed
    assert all(token not in concealed for token in ["data-rank", "data-suit", "data-edition", "cards/"])
    assert all((root / "fonts" / name).is_file() for name in ["Barlow-OFL.txt", "AtkinsonHyperlegibleNext-OFL.txt", "PROVENANCE.md"])
    assert contrast("#F5F1EA", "#1D2335") >= 4.5
    assert all(contrast(color, "#1D2335") >= 3 for color in COLORS.values())
    assert digest(root / "shared/back.png") == manifest["shared_back"]["sha256"]
    return {"cards": len(cards), "unique_native_illustrations": len(seen_art), "source_atlases": len(source_cache), "shared_label_overlays": 52, "contact_sheets": 4, "suit_sheets": 16, "rank_contrast": round(contrast("#F5F1EA", "#1D2335"), 2), "suit_contrasts": {s: round(contrast(c, "#1D2335"), 2) for s, c in COLORS.items()}, "result": "PASS", "limits": "Human visual review is required for faces, visual consistency and subjective legibility; this is not Flutter/runtime/accessibility certification."}


if __name__ == "__main__":
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--root", type=Path, default=ROOT)
    parser.add_argument("--expect-incomplete", action="store_true", help="Before assembly, prove missing outputs are rejected without changing recovery state.")
    args = parser.parse_args()
    if args.expect_incomplete:
        try:
            verify(args.root)
        except (AssertionError, FileNotFoundError) as error:
            print(f"EXPECTED REJECTION before assembly: {error}")
        else:
            raise AssertionError("Expected incomplete outputs to fail verification")
    else:
        print(json.dumps(verify(args.root), indent=2))
