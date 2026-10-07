"""Validate packaged art hashes and portable production imports; stdlib only."""
import hashlib
import json
from pathlib import Path

app = Path(__file__).resolve().parents[1]
package = app / 'packages/cgms_ui'
manifest = json.loads((package / 'assets/manifest.json').read_text())
library = app.parent / 'docs/design/cgmsart-decks'
source_manifest = json.loads((library / 'manifest.json').read_text())
editions = {'first-light', 'rain-glaze', 'ember-glaze', 'drypoint'}
suits = {'hearts', 'diamonds', 'clubs', 'spades'}
ranks = {'A', *map(str, range(2, 11)), 'J', 'Q', 'K'}
assert set(manifest['editions']) == editions
assert len(manifest['artwork']) == 208
assert len({entry['path'] for entry in manifest['artwork']}) == 208
assert len({entry['source'] for entry in manifest['artwork']}) == 208
source_cards = {card['artwork']: card for card in source_manifest['cards']}
expected_sources = {
    f'cards/{edition}/{suit}-{rank}/artwork.png'
    for edition in editions for suit in suits for rank in ranks
}
assert {entry['source'] for entry in manifest['artwork']} == expected_sources
for entry in manifest['artwork']:
    original = source_cards[entry['source']]
    edition = original['edition']
    directory = '' if edition == 'first-light' else f'{edition}/'
    expected_path = f"assets/art/{directory}{original['suit']}-{original['rank']}.png"
    assert entry['edition'] == edition and entry['path'] == expected_path, entry['path']
    assert entry['sha256'] == original['artwork_sha256'], entry['path']
    assert hashlib.sha256((library / entry['source']).read_bytes()).hexdigest() == entry['sha256'], entry['source']
assert {path.relative_to(package).as_posix() for path in (package / 'assets/art').rglob('*.png')} == {
    entry['path'] for entry in manifest['artwork']
}
assert manifest['back']['path'] == 'assets/back.png'
assert manifest['back']['source'] == source_manifest['shared_back']['path']
assert manifest['back']['sha256'] == source_manifest['shared_back']['sha256']
for entry in [*manifest['artwork'], manifest['back']]:
    actual = hashlib.sha256((package / entry['path']).read_bytes()).hexdigest()
    assert actual == entry['sha256'], entry['path']
for path in (package / 'lib').rglob('*.dart'):
    source = path.read_text()
    for forbidden in ('package:cgms_demo/', 'dart:html', 'dart:js', 'dart:io', 'package:web/'):
        assert forbidden not in source, (path, forbidden)
for font in (package / 'assets/fonts').glob('*.ttf'):
    assert font.stat().st_size > 10000, font
assert (package / 'assets/fonts/Barlow-OFL.txt').exists()
assert (package / 'assets/fonts/AtkinsonHyperlegibleNext-OFL.txt').exists()
for name, expected in {'NotoSansSymbols2-Regular.ttf': '7d5fb73b7ca67a6798101741f5d280a3d016a56a197afcd4199dbb57b4b82a21', 'NotoSansSymbols2-OFL.txt': 'b118dd41337806a5d4797052c77caf3bd096aed783e5eb21b4d11154351e1ac0'}.items():
    assert hashlib.sha256((package / "assets/fonts" / name).read_bytes()).hexdigest() == expected, name
consumer = (app / 'test/public_consumer_test.dart').read_text()
assert 'package:cgms_ui/cgms_ui.dart' in consumer
assert 'package:cgms_ui/src/' not in consumer and 'package:cgms_demo/' not in consumer
print('PASS: all 208 native art/source hashes and exact four-style inventory, shared back, font notices and package import boundary')
