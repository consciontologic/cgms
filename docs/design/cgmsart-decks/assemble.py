#!/usr/bin/env python3
"""Reproduce static cgmsart review artifacts from immutable generated atlases.

Illustration pixels are only cropped and resized for layout. Native crops remain
unmodified; lettering, suit vectors and the fixed frame are separate layers.
Requires Pillow; no network, image synthesis, app code or game-state handling.
"""
from __future__ import annotations

import argparse
import hashlib
import html
import json
from pathlib import Path

from PIL import Image, ImageDraw, ImageFont, ImageOps

ROOT = Path(__file__).resolve().parent
SIZE = (720, 1008)
ART_BOX = (24, 24, 696, 984)
SILHOUETTE = [(0, 0), (690, 0), (719, 30), (719, 1007), (0, 1007)]
BG = "#101322"
SURFACE = "#1D2335"
TEXT = "#F5F1EA"
MUTED = "#BAC4D6"
COLORS = {"hearts": "#F5A4BE", "diamonds": "#FFD275", "clubs": "#83DCD3", "spades": "#B6B2FF"}
BASE_CROPS = {
    "hearts": {"x": [0, 290, 579, 869, 1159, 1448], "y": [0, 355, 699, 1086]},
    "diamonds": {"x": [0, 288, 580, 869, 1160, 1448], "y": [0, 360, 723, 1086]},
    "clubs": {"x": [0, 287, 580, 870, 1161, 1448], "y": [0, 361, 726, 1086]},
    "spades": {"x": [0, 289, 580, 869, 1159, 1448], "y": [0, 360, 720, 1086]},
}
SUIT_PATHS = {
    "hearts": "M8 15L26 4L50 25L74 4L92 15L96 43L50 92L4 43Z",
    "diamonds": "M50 2L92 50L50 98L8 50Z",
    "spades": "M50 3L88 40L95 59L86 77L66 80L54 70L64 96L36 96L46 70L34 80L14 77L5 59L12 40Z",
}
SUIT_POINTS = {
    "hearts": [(8, 15), (26, 4), (50, 25), (74, 4), (92, 15), (96, 43), (50, 92), (4, 43)],
    "diamonds": [(50, 2), (92, 50), (50, 98), (8, 50)],
    "spades": [(50, 3), (88, 40), (95, 59), (86, 77), (66, 80), (54, 70), (64, 96), (36, 96), (46, 70), (34, 80), (14, 77), (5, 59), (12, 40)],
}


def sha(path: Path) -> str:
    return hashlib.sha256(path.read_bytes()).hexdigest()


def font(size: int, bold: bool = False, rank: bool = False) -> ImageFont.FreeTypeFont:
    name = "BarlowCondensed-Bold.ttf" if rank else f"AtkinsonHyperlegibleNext-{'Bold' if bold else 'Regular'}.ttf"
    return ImageFont.truetype(str(ROOT / "fonts" / name), size)


def save(image: Image.Image, relative: str) -> str:
    path = ROOT / relative
    path.parent.mkdir(parents=True, exist_ok=True)
    image.save(path, compress_level=6)
    return relative


def write(relative: str, content: str) -> None:
    path = ROOT / relative
    path.parent.mkdir(parents=True, exist_ok=True)
    path.write_text(content, encoding="utf-8")


def frame() -> Image.Image:
    layer = Image.new("RGBA", SIZE, (0, 0, 0, 0))
    draw = ImageDraw.Draw(layer)
    draw.polygon(SILHOUETTE, fill=SURFACE)
    draw.line(SILHOUETTE + [SILHOUETTE[0]], fill="#8595B0", width=5)
    draw.line([(12, 12), (685, 12), (707, 36), (707, 995), (12, 995), (12, 12)], fill="#414B65", width=2)
    return layer


def suit_icon(suit: str, size: int = 72) -> Image.Image:
    scale = 4
    layer = Image.new("RGBA", (size * scale, size * scale), (0, 0, 0, 0))
    draw = ImageDraw.Draw(layer)
    unit = size * scale / 100
    color = COLORS[suit]
    if suit == "clubs":
        for box in [(30, 2, 70, 44), (5, 31, 49, 76), (51, 31, 95, 76)]:
            draw.ellipse(tuple(round(v * unit) for v in box), fill=color)
        draw.polygon([(round(x * unit), round(y * unit)) for x, y in [(43, 37), (57, 37), (57, 73), (67, 96), (33, 96), (43, 73)]], fill=color)
    else:
        draw.polygon([(round(x * unit), round(y * unit)) for x, y in SUIT_POINTS[suit]], fill=color)
    return layer.resize((size, size), Image.Resampling.LANCZOS)


def suit_svg(suit: str) -> str:
    if suit == "clubs":
        return '<ellipse cx="50" cy="23" rx="20" ry="21"/><ellipse cx="27" cy="53.5" rx="22" ry="22.5"/><ellipse cx="73" cy="53.5" rx="22" ry="22.5"/><path d="M43 37H57V73L67 96H33L43 73Z"/>'
    return f'<path d="{SUIT_PATHS[suit]}"/>'


def overlay(rank: str, suit: str) -> Image.Image:
    layer = Image.new("RGBA", SIZE, (0, 0, 0, 0))
    draw = ImageDraw.Draw(layer)
    for box in [(20, 24, 145, 265), (574, 742, 699, 983)]:
        draw.rectangle(box, fill=SURFACE)
    for center, baseline, icon_y in [(83, 152, 177), (637, 870, 895)]:
        draw.text((center, baseline), rank, font=font(124, rank=True), fill=TEXT, anchor="ms")
        layer.alpha_composite(suit_icon(suit), (center - 36, icon_y))
    return layer


def write_overlay(rank: str, suit: str) -> dict:
    path = f"overlays/{suit}-{rank}.png"
    save(overlay(rank, suit), path)
    groups = "".join(f'<text x="{x}" y="{y}" text-anchor="middle">{rank}</text><g transform="translate({x - 36} {iy}) scale(.72)" fill="{COLORS[suit]}">{suit_svg(suit)}</g>' for x, y, iy in [(83, 152, 177), (637, 870, 895)])
    svg_path = f"overlays/{suit}-{rank}.svg"
    write(svg_path, f'''<svg xmlns="http://www.w3.org/2000/svg" width="720" height="1008" viewBox="0 0 720 1008" role="img" aria-label="{rank} of {suit}">
<style>@font-face{{font-family:Barlow Condensed;src:url(../fonts/BarlowCondensed-Bold.ttf)}}text{{font-family:Barlow Condensed;font-weight:700;font-size:124px;fill:{TEXT}}}</style>
<rect x="20" y="24" width="126" height="242" fill="{SURFACE}"/><rect x="574" y="742" width="126" height="242" fill="{SURFACE}"/>{groups}</svg>''')
    return {"overlay": path, "overlay_svg": svg_path, "overlay_sha256": sha(ROOT / path)}


def crop_spec(edition: str, suit: str, source_size: tuple[int, int]) -> dict:
    contract_path = ROOT / "atlas-crops.json"
    contract = json.loads(contract_path.read_text()) if contract_path.exists() else {}
    key = f"{edition}/{suit}"
    for candidate in [contract.get(key), contract.get("atlases", {}).get(key), contract.get(f"sources/{key}.png"), contract.get(edition, {}).get(suit)]:
        if candidate:
            return candidate
    if edition == "first-light" and source_size == (1448, 1086):
        return BASE_CROPS[suit]
    return {"x": [round(source_size[0] * n / 5) for n in range(6)], "y": [round(source_size[1] * n / 3) for n in range(4)]}


def compose(art: Image.Image, labels: Image.Image | None = None) -> Image.Image:
    canvas = frame()
    width, height = ART_BOX[2] - ART_BOX[0], ART_BOX[3] - ART_BOX[1]
    # Cover crops only for the review face. artwork.png always preserves all pixels
    # in the native atlas crop; no color operations or synthetic texture are used.
    resized = ImageOps.fit(art.convert("RGBA"), (width, height), Image.Resampling.LANCZOS)
    canvas.alpha_composite(resized, ART_BOX[:2])
    if labels is not None:
        canvas.alpha_composite(labels)
    return canvas


def build_cards(production: dict, editions: list[dict]) -> list[dict]:
    cards = []
    layers = {(s, r): write_overlay(r, s) for s in production["suits"] for r in production["ranks"]}
    for edition in editions:
        for suit in production["suits"]:
            source_path = ROOT / "sources" / edition["id"] / f"{suit}.png"
            source = Image.open(source_path).convert("RGB")
            spec = crop_spec(edition["id"], suit, source.size)
            assert len(spec["x"]) == 6 and len(spec["y"]) == 4
            assert spec["x"][0] == spec["y"][0] == 0
            assert (spec["x"][-1], spec["y"][-1]) == source.size
            assert all(a + 6 < b for axis in ("x", "y") for a, b in zip(spec[axis], spec[axis][1:]))
            for index, rank in enumerate(production["ranks"]):
                x, y = index % 5, index // 5
                box = [spec["x"][x] + 3, spec["y"][y] + 3, spec["x"][x + 1] - 3, spec["y"][y + 1] - 3]
                art = source.crop(box)
                stem = f"cards/{edition['id']}/{suit}-{rank}"
                art_path = save(art, f"{stem}/artwork.png")
                face_path = save(compose(art, overlay(rank, suit)), f"{stem}/face.png")
                record = {"edition": edition["id"], "suit": suit, "rank": rank, "subject": production["subjects"][suit][index], "face": face_path, "artwork": art_path, **layers[suit, rank], "back": "shared/back.png", "label": {"rank": rank, "suit": suit, "font": "Barlow Condensed Bold", "color": TEXT, "suit_color": COLORS[suit]}, "source_atlas": source_path.relative_to(ROOT).as_posix(), "source_size": list(source.size), "source_bbox": box, "native_artwork_size": list(art.size), "source_sha256": sha(source_path), "artwork_sha256": sha(ROOT / art_path), "face_sha256": sha(ROOT / face_path), "layout": {"art_box": list(ART_BOX), "art_fit": "centered cover; native unmodified crop delivered separately", "rank_baselines": [[83, 152], [637, 870]], "rank_font_pixels": 124}, "usage": "Cosmetic exploratory artwork; no gameplay value or permanent combo role added", "shipping_approval": "exploration only"}
                cards.append(record)
        print(f"Assembled {edition['name']}: 52 unique identities", flush=True)
    return cards


def new_sheet(size: tuple[int, int], title: str, subtitle: str) -> Image.Image:
    sheet = Image.new("RGB", size, BG)
    draw = ImageDraw.Draw(sheet)
    draw.text((64, 38), "CGMS / cgmsart", fill=MUTED, font=font(26, True))
    draw.text((60, 83), title, fill=TEXT, font=font(62, rank=True))
    draw.text((64, 168), subtitle, fill=MUTED, font=font(24))
    return sheet


def place_card(sheet: Image.Image, path: str, xy: tuple[int, int], size: tuple[int, int]) -> None:
    card = Image.open(ROOT / path).convert("RGBA").resize(size, Image.Resampling.LANCZOS)
    sheet.paste(card, xy, card)


def contact_sheets(production: dict, cards: list[dict], editions: list[dict]) -> None:
    for edition in editions:
        chosen = [c for c in cards if c["edition"] == edition["id"]]
        sheet = new_sheet((5592, 2700), f"{edition['name']} / complete 52-card deck", "Hearts · Diamonds · Clubs · Spades   /   A · 2–10 · J · Q · K   /   Cosmetic edition · No jokers")
        draw = ImageDraw.Draw(sheet)
        for index, card in enumerate(chosen):
            col, row = index % 13, index // 13
            place_card(sheet, card["face"], (64 + col * 424, 300 + row * 590), (400, 560))
        for row, suit in enumerate(production["suits"]):
            draw.text((64, 264 + row * 590), suit.upper(), fill=COLORS[suit], font=font(22, True))
        save(sheet, f"sheets/{edition['id']}-52.png")
        for suit in production["suits"]:
            sheet = new_sheet((2800, 2520), f"{edition['name']} / {suit.title()}", "A · 2 · 3 · 4 · 5  /  6 · 7 · 8 · 9 · 10  /  J · Q · K     —     13 unique cards")
            suit_cards = [c for c in chosen if c["suit"] == suit]
            for index, card in enumerate(suit_cards):
                place_card(sheet, card["face"], (60 + (index % 5) * 540, 256 + (index // 5) * 744), (520, 728))
            draw = ImageDraw.Draw(sheet)
            draw.text((1680, 1840), "ONE SUIT. THIRTEEN COMPOSITIONS.", fill=MUTED, font=font(24, True))
            draw.text((1680, 1890), "Human faces only on J, Q and K.\nLabels are separate vector/raster layers.\nArtwork is cosmetic; game rules are unchanged.", fill=MUTED, font=font(22), spacing=12)
            save(sheet, f"sheets/{edition['id']}-{suit}.png")


def comparison(production: dict, cards: list[dict]) -> None:
    subjects = [("hearts", "A"), ("clubs", "8"), ("hearts", "Q"), ("diamonds", "J"), ("spades", "K")]
    sheet = new_sheet((2620, 4510), "Four editions / same card, same identity", "Compare reflected light and pigment treatment. Geometry, subjects, labels and suit colors stay fixed.")
    draw = ImageDraw.Draw(sheet)
    lookup = {(c["edition"], c["suit"], c["rank"]): c for c in cards}
    for col, edition in enumerate(production["editions"]):
        draw.text((64 + col * 642, 235), edition["name"], fill=TEXT, font=font(38, True))
        for row, (suit, rank) in enumerate(subjects):
            x, y = 64 + col * 642, 334 + row * 818
            draw.text((x, y - 36), f"{rank} OF {suit.upper()}", fill=COLORS[suit], font=font(22, True))
            place_card(sheet, lookup[edition["id"], suit, rank]["face"], (x, y), (540, 756))
    save(sheet, "previews/comparison.png")
    artwork_sheet = new_sheet((2620, 4510), "Four editions / illustration comparison", "Same source subjects · Separate from rank and suit labels · Enlarged review crops, not native-resolution detail")
    draw = ImageDraw.Draw(artwork_sheet)
    for col, edition in enumerate(production["editions"]):
        draw.text((64 + col * 642, 235), edition["name"], fill=TEXT, font=font(38, True))
        for row, (suit, rank) in enumerate(subjects):
            x, y = 64 + col * 642, 334 + row * 818
            draw.text((x, y - 36), f"{rank} OF {suit.upper()}", fill=COLORS[suit], font=font(22, True))
            art = Image.open(ROOT / lookup[edition["id"], suit, rank]["artwork"])
            enlarged = ImageOps.contain(art, (540, 756), Image.Resampling.LANCZOS)
            artwork_sheet.paste(enlarged, (x, y))
    save(artwork_sheet, "previews/artwork-comparison.png")


def closeups(production: dict, cards: list[dict]) -> None:
    lookup = {(c["edition"], c["suit"], c["rank"]): c for c in cards}
    for suit in production["suits"]:
        sheet = new_sheet((2800, 1640), f"{suit.title()} / artwork material study", "First Light · A / 8 / J / Q / K   —   Top: enlarged review crop. Bottom: native source pixels, without labels.")
        draw = ImageDraw.Draw(sheet)
        for col, rank in enumerate(["A", "8", "J", "Q", "K"]):
            card = lookup["first-light", suit, rank]
            art = Image.open(ROOT / card["artwork"])
            x = 60 + col * 540
            draw.text((x, 240), f"{rank} / {suit}", fill=COLORS[suit], font=font(30, True))
            enlarged = ImageOps.contain(art, (510, 670), Image.Resampling.LANCZOS)
            sheet.paste(enlarged, (x, 290))
            draw.text((x, 992), f"Native: {art.width} × {art.height}px", fill=MUTED, font=font(24))
            sheet.paste(art, (x, 1040))
        save(sheet, f"previews/closeup-{suit}.png")


def mobile(cards: list[dict]) -> None:
    for card in cards:
        for width, height in [(96, 134), (120, 168)]:
            image = Image.open(ROOT / card["face"]).resize((width, height), Image.Resampling.LANCZOS)
            save(image, f"previews/mobile/{width}/{card['edition']}/{card['suit']}-{card['rank']}.png")
    sheet = new_sheet((1300, 670), "Mobile playing-size checks", "Exact output pixels at 100% view. Static face readability only; no game UI or runtime approval.")
    draw = ImageDraw.Draw(sheet)
    picks = [("hearts", "A"), ("diamonds", "10"), ("clubs", "8"), ("spades", "J"), ("hearts", "Q"), ("spades", "K")]
    lookup = {(c["edition"], c["suit"], c["rank"]): c for c in cards}
    for row, (width, height) in enumerate([(96, 134), (120, 168)]):
        y = 258 + row * 190
        draw.text((64, y + 20), f"{width} × {height}px", fill=MUTED, font=font(22, True))
        for col, (suit, rank) in enumerate(picks):
            place_card(sheet, lookup["first-light", suit, rank]["face"], (282 + col * 142, y), (width, height))
    save(sheet, "previews/mobile-sizes.png")
    stage = Image.new("RGB", (390, 844), BG)
    draw = ImageDraw.Draw(stage)
    draw.text((22, 24), "cgmsart", font=font(32, rank=True), fill=TEXT)
    draw.text((22, 70), "Mobile scale / static art proof", font=font(18), fill=MUTED)
    draw.text((22, 126), "96 × 134px", font=font(18, True), fill=TEXT)
    for col, (suit, rank) in enumerate(picks[:3]):
        place_card(stage, lookup["first-light", suit, rank]["face"], (22 + col * 124, 166), (96, 134))
    draw.text((22, 338), "120 × 168px", font=font(18, True), fill=TEXT)
    for col, (suit, rank) in enumerate(picks[4:]):
        place_card(stage, lookup["first-light", suit, rank]["face"], (22 + col * 148, 376), (120, 168))
    draw.text((22, 582), "One shared concealed presentation", font=font(17, True), fill=TEXT)
    for col in range(3):
        place_card(stage, "shared/back.png", (22 + col * 124, 622), (96, 134))
    draw.text((22, 794), "Art review only · No gameplay controls", font=font(16), fill=MUTED)
    save(stage, "previews/mobile-390.png")


def shared_back() -> dict:
    path = ROOT / "sources/shared-back.png"
    source = Image.open(path).convert("RGB")
    contract_path = ROOT / "atlas-crops.json"
    contract = json.loads(contract_path.read_text()) if contract_path.exists() else {}
    bbox = contract.get("shared-back", {}).get("bbox", [0, 0, source.width, source.height])
    back = compose(source.crop(bbox))
    save(back, "shared/back.png")
    proof = new_sheet((1400, 1450), "One shared back / every edition", "Identical silhouette and edge. No rank, suit, edition, ownership or copy markings.")
    place_card(proof, "shared/back.png", (310, 310), SIZE)
    save(proof, "previews/shared-back.png")
    save(frame(), "shared/frame.png")
    write("shared/template.svg", '''<svg xmlns="http://www.w3.org/2000/svg" width="720" height="1008" viewBox="0 0 720 1008">
<title>cgmsart card template: fixed geometry; insert art and a separate identity overlay</title>
<defs><clipPath id="card"><path d="M0 0H690L719 30V1007H0Z"/></clipPath></defs>
<g clip-path="url(#card)"><path d="M0 0H690L719 30V1007H0Z" fill="#1D2335" stroke="#8595B0" stroke-width="5"/><path d="M12 12H685L707 36V995H12Z" fill="none" stroke="#414B65" stroke-width="2"/>
<rect id="artwork-slot" x="24" y="24" width="672" height="960" fill="none"/>
<g id="identity-overlay-slot"/></g></svg>''')
    return {"path": "shared/back.png", "sha256": sha(ROOT / "shared/back.png"), "source": "sources/shared-back.png", "source_sha256": sha(path), "source_bbox": bbox, "source_size": list(source.size), "reuse": "Exactly one back for all ranks, suits, editions and physical copies"}


def gallery(production: dict, cards: list[dict]) -> None:
    options = "".join(f'<option value="{e["id"]}">{e["name"]}</option>' for e in production["editions"])
    panels = []
    for edition in production["editions"]:
        figures = []
        for card in [c for c in cards if c["edition"] == edition["id"]]:
            figures.append(f'<figure><a href="{card["face"]}"><img class="face" loading="lazy" src="{card["face"]}" alt="{card["rank"]} of {card["suit"]}, {edition["name"]}"><img class="art" loading="lazy" src="{card["artwork"]}" alt="{html.escape(card["subject"])}"></a><figcaption>{card["rank"]} of {card["suit"]}</figcaption></figure>')
        links = " · ".join(f'<a href="sheets/{edition["id"]}-{s}.png">{s.title()} sheet</a>' for s in production["suits"])
        hidden = "" if edition["id"] == "first-light" else " hidden"
        panels.append(f'<section class="edition" id="{edition["id"]}"{hidden}><h2>{edition["name"]} / 52 cards</h2><p><a href="sheets/{edition["id"]}-52.png">Full deck contact sheet</a> · {links}</p><div class="grid">{"".join(figures)}</div></section>')
    write("index.html", f'''<!doctype html>
<html lang="en"><meta charset="utf-8"><meta name="viewport" content="width=device-width,initial-scale=1"><title>CGMS cgmsart / Four sibling decks</title>
<style>@font-face{{font-family:Atkinson;src:url(fonts/AtkinsonHyperlegibleNext-Regular.ttf)}}@font-face{{font-family:Barlow;src:url(fonts/BarlowCondensed-Bold.ttf);font-weight:700}}
*{{box-sizing:border-box}}body{{margin:0;background:{BG};color:{TEXT};font:18px/1.5 Atkinson,sans-serif}}main{{max-width:1560px;margin:auto;padding:28px}}h1,h2{{font-family:Barlow,sans-serif;line-height:1.1}}h1{{font-size:clamp(44px,6vw,80px);max-width:850px;margin:24px 0}}h2{{font-size:36px}}a{{color:#83DCD3}}p{{max-width:980px}}.muted{{color:{MUTED}}}.hero{{width:100%;max-width:1000px}}.controls{{display:flex;flex-wrap:wrap;gap:18px;align-items:center;position:sticky;top:0;background:{BG};padding:16px 0;z-index:2}}select,button{{font:inherit;color:{TEXT};background:{SURFACE};border:1px solid #8595B0;min-height:48px;padding:8px 16px}}:focus-visible{{outline:3px solid #DFFF7A;outline-offset:3px}}.grid{{display:grid;grid-template-columns:repeat(auto-fill,minmax(150px,1fr));gap:24px 16px}}figure{{margin:0}}figure img{{width:100%;height:auto;display:block}}figcaption{{font-size:16px;padding:8px 0}}.art{{display:none}}body.art-only .face{{display:none}}body.art-only .art{{display:block;aspect-ratio:5/7;object-fit:contain;background:{SURFACE}}}[hidden]{{display:none!important}}.samples{{display:flex;gap:20px;flex-wrap:wrap;align-items:start}}.samples img{{max-width:100%;height:auto}}#concealed img{{width:96px;height:134px}}section{{margin-top:48px}}footer{{margin:64px 0 20px;color:{MUTED}}}.nav{{display:flex;gap:20px;flex-wrap:wrap}}@media(max-width:500px){{main{{padding:20px}}.grid{{grid-template-columns:repeat(3,minmax(0,1fr));gap:16px 10px}}figcaption{{font-size:14px}}}}
</style><main><p class="muted">CGMS / cgmsart / ART EXPLORATION</p><h1>Four editions.<br>One painted card family.</h1><p>First Light, Ember Glaze, Rain Glaze and Drypoint. Four cosmetic 52-card editions, each using the same rank and suit order. The game still uses two decks without jokers.</p><p class="muted">720 × 1008px face exports; native illustration crops are approximately 280 × 350px. High-resolution review layouts enlarge the generated source; they do not create new painted detail. This is a static art gallery, not a playable game or production asset approval.</p>
<nav class="nav"><a href="README.md">Art direction, variation matrix & provenance</a><a href="previews/comparison.png">Same-card comparison</a><a href="previews/artwork-comparison.png">Illustration-only comparison</a><a href="manifest.json">Asset manifest</a></nav>
<section><h2>Controlled variation</h2><a href="previews/comparison.png"><img class="hero" src="previews/comparison.png" alt="The same Ace of Hearts, 8 of Clubs, Queen of Hearts, Jack of Diamonds and King of Spades across all four editions"></a></section>
<div class="controls"><label>Edition <select id="edition">{options}</select></label><button id="art-toggle" type="button" aria-pressed="false">Show artwork only</button></div>
{''.join(panels)}
<section><h2>Material close-ups</h2><p>Native crop and enlarged review detail are labelled separately. <a href="previews/closeup-hearts.png">Hearts</a> · <a href="previews/closeup-diamonds.png">Diamonds</a> · <a href="previews/closeup-clubs.png">Clubs</a> · <a href="previews/closeup-spades.png">Spades</a></p><a href="previews/closeup-hearts.png"><img class="hero" loading="lazy" src="previews/closeup-hearts.png" alt="Hearts artwork closeups, enlarged and at native size"></a></section>
<section><h2>Mobile playing size</h2><p>View at 100% zoom. These are static readability proofs; Flutter rendering and accessibility remain untested.</p><div class="samples"><img src="previews/mobile-390.png" width="390" height="844" alt="390 pixel wide mobile scale preview with 96 by 134 and 120 by 168 pixel cards"><a href="previews/mobile-sizes.png">Open exact-size multi-card proof</a></div></section>
<section id="concealed"><h2>One shared back</h2><p>Every concealed position uses this same image and neutral description.</p><div class="samples"><img src="shared/back.png" alt="Hidden card"><img src="shared/back.png" alt="Hidden card"><img src="shared/back.png" alt="Hidden card"><img src="shared/back.png" alt="Hidden card"></div><p><a href="previews/shared-back.png">Inspect shared back</a></p></section>
<footer>Original generated illustrations. Barlow Condensed and Atkinson Hyperlegible Next are bundled under SIL OFL 1.1; <a href="fonts/PROVENANCE.md">font provenance</a>. Labels are delivered separately from the native illustrations.</footer></main>
<script>const select=document.getElementById('edition');select.addEventListener('change',()=>{{document.querySelectorAll('.edition').forEach(section=>section.hidden=section.id!==select.value)}});document.getElementById('art-toggle').addEventListener('click',event=>{{const active=document.body.classList.toggle('art-only');event.currentTarget.setAttribute('aria-pressed',String(active));event.currentTarget.textContent=active?'Show finished cards':'Show artwork only'}});</script></html>''')


def main() -> None:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--sample", action="store_true", help="Build First Light cards and sheets before all variant atlases arrive.")
    args = parser.parse_args()
    production = json.loads((ROOT / "production.json").read_text())
    editions = production["editions"][:1] if args.sample else production["editions"]
    required = [ROOT / "sources" / edition["id"] / f"{suit}.png" for edition in editions for suit in production["suits"]]
    if not args.sample:
        required.append(ROOT / "sources/shared-back.png")
    missing = [str(path.relative_to(ROOT)) for path in required if not path.is_file()]
    if missing:
        parser.error("Required source images are not ready: " + ", ".join(missing))
    cards = build_cards(production, editions)
    contact_sheets(production, cards, editions)
    if args.sample:
        print("First Light sample ready; full manifest is intentionally not written until all sources exist.")
        return
    back = shared_back()
    comparison(production, cards)
    closeups(production, cards)
    mobile(cards)
    gallery(production, cards)
    manifest = {"schema_version": 1, "purpose": "Reviewable art exploration, four cosmetic variants of one 52-card deck; CGMS two-deck rules unchanged", "editions": [e["id"] for e in editions], "suits": production["suits"], "ranks": production["ranks"], "card_size": list(SIZE), "shared_back": back, "cards": cards, "production_limits": ["Native atlas crops are approximately 280 × 350 pixels; larger exports are layout upscales", "Subject consistency and face-only-on-royals require human visual review", "No Flutter integration, runtime tests, print certification or production shipping approval"]}
    write("manifest.json", json.dumps(manifest, indent=2) + "\n")
    print("Complete: 208 cards, 52 shared overlays, 4 deck sheets, 16 suit sheets, comparison, closeups, mobile proofs and gallery.")


if __name__ == "__main__":
    main()
