// Independent SDK-only narrow board candidate. No FFI, process, fixture or oracle
// access. Explicit chance permutations are inputs; unsupported rules reject.
import 'dart:convert';
import 'finance.dart' as finance;

String canonical(dynamic value) {
  if (value is Map) {
    final keys = value.keys.cast<String>().toList()..sort();
    return '{${keys.map((k) => '${jsonEncode(k)}:${canonical(value[k])}').join(',')}}';
  }
  if (value is List) return '[${value.map(canonical).join(',')}]';
  return jsonEncode(value);
}

String sha256(String text) {
  const k = [
    0x428a2f98,
    0x71374491,
    0xb5c0fbcf,
    0xe9b5dba5,
    0x3956c25b,
    0x59f111f1,
    0x923f82a4,
    0xab1c5ed5,
    0xd807aa98,
    0x12835b01,
    0x243185be,
    0x550c7dc3,
    0x72be5d74,
    0x80deb1fe,
    0x9bdc06a7,
    0xc19bf174,
    0xe49b69c1,
    0xefbe4786,
    0x0fc19dc6,
    0x240ca1cc,
    0x2de92c6f,
    0x4a7484aa,
    0x5cb0a9dc,
    0x76f988da,
    0x983e5152,
    0xa831c66d,
    0xb00327c8,
    0xbf597fc7,
    0xc6e00bf3,
    0xd5a79147,
    0x06ca6351,
    0x14292967,
    0x27b70a85,
    0x2e1b2138,
    0x4d2c6dfc,
    0x53380d13,
    0x650a7354,
    0x766a0abb,
    0x81c2c92e,
    0x92722c85,
    0xa2bfe8a1,
    0xa81a664b,
    0xc24b8b70,
    0xc76c51a3,
    0xd192e819,
    0xd6990624,
    0xf40e3585,
    0x106aa070,
    0x19a4c116,
    0x1e376c08,
    0x2748774c,
    0x34b0bcb5,
    0x391c0cb3,
    0x4ed8aa4a,
    0x5b9cca4f,
    0x682e6ff3,
    0x748f82ee,
    0x78a5636f,
    0x84c87814,
    0x8cc70208,
    0x90befffa,
    0xa4506ceb,
    0xbef9a3f7,
    0xc67178f2,
  ];
  final h = [
    0x6a09e667,
    0xbb67ae85,
    0x3c6ef372,
    0xa54ff53a,
    0x510e527f,
    0x9b05688c,
    0x1f83d9ab,
    0x5be0cd19,
  ];
  final bytes = utf8.encode(text).toList();
  final bits = bytes.length * 8;
  bytes.add(128);
  while (bytes.length % 64 != 56) {
    bytes.add(0);
  }
  for (var i = 7; i >= 0; i--) {
    bytes.add((bits >> (i * 8)) & 255);
  }
  int rotate(int x, int n) => ((x >> n) | (x << (32 - n))) & 0xffffffff;
  for (var off = 0; off < bytes.length; off += 64) {
    final w = List.filled(64, 0);
    for (var i = 0; i < 16; i++) {
      for (var j = 0; j < 4; j++) {
        w[i] = (w[i] << 8) | bytes[off + i * 4 + j];
      }
    }
    for (var i = 16; i < 64; i++) {
      final x = w[i - 15], y = w[i - 2];
      w[i] =
          (w[i - 16] +
              (rotate(x, 7) ^ rotate(x, 18) ^ (x >> 3)) +
              w[i - 7] +
              (rotate(y, 17) ^ rotate(y, 19) ^ (y >> 10))) &
          0xffffffff;
    }
    var a = h[0],
        b = h[1],
        c = h[2],
        d = h[3],
        e = h[4],
        f = h[5],
        g = h[6],
        v = h[7];
    for (var i = 0; i < 64; i++) {
      final t1 =
          (v +
              (rotate(e, 6) ^ rotate(e, 11) ^ rotate(e, 25)) +
              ((e & f) ^ ((~e) & g)) +
              k[i] +
              w[i]) &
          0xffffffff;
      final t2 =
          ((rotate(a, 2) ^ rotate(a, 13) ^ rotate(a, 22)) +
              ((a & b) ^ (a & c) ^ (b & c))) &
          0xffffffff;
      v = g;
      g = f;
      f = e;
      e = (d + t1) & 0xffffffff;
      d = c;
      c = b;
      b = a;
      a = (t1 + t2) & 0xffffffff;
    }
    final values = [a, b, c, d, e, f, g, v];
    for (var i = 0; i < 8; i++) {
      h[i] = (h[i] + values[i]) & 0xffffffff;
    }
  }
  return h.map((v) => v.toRadixString(16).padLeft(8, '0')).join();
}

Map<String, dynamic> copy(Map value) =>
    jsonDecode(jsonEncode(value)) as Map<String, dynamic>;
void require(bool value) {
  if (!value)
    throw const FormatException('unsupported or invalid cardplay transition');
}

List list(Map m, String key) => (m[key] as List?) ?? [];
const suits = ['hearts', 'clubs', 'spades', 'diamonds'];
Map card(Map board, String id) =>
    list(board, 'cards').singleWhere((c) => c['card']['id'] == id) as Map;
bool publicZone(String zone) =>
    !['draw', 'hand', 'concealed-ace'].contains(zone);
void move(Map b, List ids, int seat, String zone) {
  for (final id in ids) {
    final c = card(b, id);
    if (publicZone(c['zone']) || publicZone(zone)) {
      if (!list(b, 'public_history').contains(id))
        b['public_history'] = [...list(b, 'public_history'), id];
    }
    c['zone'] = zone;
    c['controller'] = seat;
    c['allocation'] = '';
    b['draw_order'] = list(b, 'draw_order').where((v) => v != id).toList();
    if (zone == 'draw') {
      b['draw_order'].add(id);
      b['needs_shuffle'] = true;
    }
  }
}

void shuffle(Map b, List order) {
  final want = list(
    b,
    'cards',
  ).where((c) => c['zone'] == 'draw').map((c) => c['card']['id']).toSet();
  require(
    order.length == want.length &&
        order.toSet().length == order.length &&
        order.every(want.contains),
  );
  b['draw_order'] = [...order];
  b['needs_shuffle'] = false;
}

void draw(Map b, int seat, int count, bool expose) {
  require(b['phase'] == 'playing' && b['needs_shuffle'] == false);
  for (var i = 0; i < count && list(b, 'draw_order').isNotEmpty; i++) {
    final id = b['draw_order'][0], c = card(b, id)['card'];
    move(
      b,
      [id],
      seat,
      c['rank'] == 1
          ? 'concealed-ace'
          : expose && c['suit'] == 'diamonds' && c['rank'] <= 10
          ? 'series'
          : 'hand',
    );
  }
}

Map newBoard(int n, String id) => {
  'schema': 'cgms-state-v1',
  'game_id': id,
  'round': 1,
  'turn': 1,
  'active': 1,
  'phase': 'playing',
  'version': 0,
  'commands': <String, dynamic>{},
  'needs_shuffle': false,
  'draw_queue': null,
  'loans': null,
  'decisions': null,
  'bindings': null,
  'effects': null,
  'order': null,
  'public_history': null,
  'cards': [
    for (var deck = 1; deck <= 2; deck++)
      for (final suit in suits)
        for (var rank = 1; rank <= 13; rank++)
          {
            'card': {
              'id': 'deck-$deck-$suit-${rank.toString().padLeft(2, '0')}',
              'deck': deck,
              'suit': suit,
              'rank': rank,
            },
            'zone': 'draw',
            'controller': 0,
            'available_from_round': 0,
            'allocation': '',
            'used_turn': 0,
          },
  ],
  'draw_order': [
    for (var deck = 1; deck <= 2; deck++)
      for (final suit in suits)
        for (var rank = 1; rank <= 13; rank++)
          'deck-$deck-$suit-${rank.toString().padLeft(2, '0')}',
  ],
  'players': [
    for (var seat = 1; seat <= n; seat++)
      {
        'seat': seat,
        'history': null,
        'opening_used': false,
        'departed': false,
        'confined': false,
        'connected': true,
        'underground': false,
        'ally': 0,
        'ace_round': 0,
        'compensation_used': false,
        'quotas': <String, dynamic>{},
      },
  ],
};
Map newLifecycle(Map board, Map ledger) => {
  'schema': 'cgms-lifecycle-v1',
  'commands': <String, dynamic>{},
  'board': board,
  'ledger': ledger,
  'promises': {
    'game_id': board['game_id'],
    'seats': list(board, 'players').length,
    'phase': 'playing',
    'promises': null,
    'awards': null,
    'payments': null,
    'finished': List.filled(list(board, 'players').length, false),
  },
  'turn_index': 0,
  'turn_started': false,
  'round_closed': false,
  'ending': '',
  'declarer': 0,
  'code_cursor': 0,
};
void initiative(Map b, List order) {
  final players = list(b, 'players');
  final scores = <int, int>{for (final p in players) p['seat']: 0};
  for (final c in list(b, 'cards')) {
    if (c['zone'] == 'series' && c['card']['suit'] == 'diamonds')
      scores[c['controller']] =
          scores[c['controller']]! + (c['card']['rank'] as int);
  }
  List<int> stableSort(List<int> seats, Map<int, int> values) {
    final original = [...seats];
    return [...seats]..sort((a, b) {
      final diff = values[b]!.compareTo(values[a]!);
      return diff == 0
          ? original.indexOf(a).compareTo(original.indexOf(b))
          : diff;
    });
  }

  final supply = [...list(b, 'draw_order')], used = <String>[];
  var cursor = 0;
  List<int> resolve(List<int> group) {
    if (group.length < 2) return group;
    final values = <int, int>{};
    for (final seat in group) {
      while (cursor < supply.length) {
        final id = supply[cursor++] as String;
        used.add(id);
        final rank = card(b, id)['card']['rank'] as int;
        if (rank >= 2 && rank <= 10) {
          values[seat] = rank;
          break;
        }
      }
      if (!values.containsKey(seat)) {
        final start = (b['round'] - 1) % players.length + 1;
        return [...group]..sort(
          (a, c) => ((a - start + players.length) % players.length as int)
              .compareTo((c - start + players.length) % players.length as int),
        );
      }
    }
    final sorted = stableSort(group, values);
    final result = <int>[];
    for (var i = 0; i < sorted.length;) {
      var j = i + 1;
      while (j < sorted.length && values[sorted[i]] == values[sorted[j]]) {
        j++;
      }
      result.addAll(resolve(sorted.sublist(i, j)));
      i = j;
    }
    return result;
  }

  final seats = stableSort(
        players.map((p) => p['seat'] as int).toList(),
        scores,
      ),
      result = <int>[];
  for (var i = 0; i < seats.length;) {
    var j = i + 1;
    while (j < seats.length && scores[seats[i]] == scores[seats[j]]) {
      j++;
    }
    result.addAll(resolve(seats.sublist(i, j)));
    i = j;
  }
  if (used.isNotEmpty) {
    move(b, used, 0, 'draw');
    shuffle(b, order);
  }
  b['order'] = result;
  b['active'] = result.first;
}

Map observe(Map b, int seat) {
  final history = [...list(b, 'public_history')];
  final cards = <dynamic>[];
  for (final c in list(b, 'cards')) {
    final id = c['card']['id'];
    if (publicZone(c['zone']) && !history.contains(id)) history.add(id);
    if (c['controller'] == seat || publicZone(c['zone'])) cards.add(c);
  }
  final pending = b['pending'];
  return {
    'game_id': b['game_id'],
    'seat': seat,
    'round': b['round'],
    'turn': b['turn'],
    'active': b['active'],
    'phase': b['phase'],
    'version': b['version'],
    'menu_coverage': 'narrow-single-pair-combat-v1',
    'public_history': history.isEmpty ? null : history,
    'cards': cards.isEmpty ? null : cards,
    'players': [
      for (final p in list(b, 'players'))
        {
          'seat': p['seat'],
          'hand_count': list(b, 'cards')
              .where((c) => c['controller'] == p['seat'] && c['zone'] == 'hand')
              .length,
          'ace_count': list(b, 'cards')
              .where(
                (c) =>
                    c['controller'] == p['seat'] &&
                    c['zone'] == 'concealed-ace',
              )
              .length,
          'confined': p['confined'],
          'departed': p['departed'],
          'history': p['history'],
        },
    ],
    'effects': b['effects'],
    'required_actor': pending == null
        ? 0
        : pending['responders'][pending['cursor']],
    'legal': null,
    if (pending != null) 'window_id': pending['id'],
  };
}

Map? choose(Map o, bool used, bool started, bool closed) {
  if (o['phase'] != 'playing' || !started || closed) return null;
  final seat = o['seat'];
  Map decision(String kind, [String suit = '', String window = '']) => {
    'kind': kind,
    'actor': seat,
    'suit': suit,
    'window_id': window,
  };
  if (o.containsKey('window_id'))
    return o['required_actor'] == seat
        ? decision('pass', '', o['window_id'])
        : null;
  if (o['active'] != seat) return null;
  for (final suit in suits) {
    if (used && !list(o['players'][seat - 1], 'history').contains(suit))
      continue;
    for (final c in list(o, 'cards')) {
      if (c['controller'] == seat &&
          c['zone'] == 'hand' &&
          c['card']['suit'] == suit &&
          c['card']['rank'] >= 2 &&
          c['card']['rank'] <= 10 &&
          c['available_from_round'] <= o['round'])
        return decision('open-series', suit);
    }
  }
  return decision('end-turn');
}

Map cardplayFrame(Map match, List? events) {
  final l = match['game'] as Map, b = l['board'] as Map;
  final financeFrame = finance.frame(l['ledger'], match['game_limit']);
  final observations = <dynamic>[], decisions = <dynamic>[];
  for (var seat = 1; seat <= list(b, 'players').length; seat++) {
    final o = observe(b, seat),
        used = b['players'][seat - 1]['opening_used'] as bool;
    observations.add({
      'board': o,
      'ledger': financeFrame['observations'][seat - 1],
      'opening_used': used,
      'turn_started': l['turn_started'],
      'round_closed': l['round_closed'],
      'ending': l['ending'],
    });
    decisions.add(choose(o, used, l['turn_started'], l['round_closed']));
  }
  return {
    'match': match,
    'events': events,
    'observations': observations,
    'decisions': decisions,
  };
}

List? stepCardplay(Map m, Map step) {
  m['command_history'] = [...list(m, 'command_history')];
  final l = m['game'] as Map,
      b = l['board'] as Map,
      ledger = l['ledger'] as Map;
  require(step['game_id'] == b['game_id']);
  final n = list(b, 'players').length;
  List? events;
  switch (step['kind']) {
    case 'deal':
      require(list(b, 'draw_order').length == 104);
      shuffle(b, list(step, 'order'));
      for (var seat = 1; seat <= n; seat++) {
        draw(b, seat, 20, true);
        require(
          list(b, 'cards').any(
            (c) =>
                c['controller'] == seat &&
                c['zone'] == 'series' &&
                c['card']['suit'] == 'diamonds',
          ),
        );
        b['players'][seat - 1]['history'] = ['diamonds'];
      }
      break;
    case 'initiative':
      require(list(b, 'order').isEmpty);
      initiative(b, list(step, 'order'));
      break;
    case 'begin-turn':
      require(
        !l['turn_started'] &&
            !l['round_closed'] &&
            l['ending'] == '' &&
            !b.containsKey('pending'),
      );
      final seat = b['active'] as int;
      b['players'][seat - 1]['opening_used'] = false;
      b['players'][seat - 1]['confined'] = false;
      draw(b, seat, 1, false);
      l['turn_started'] = true;
      break;
    case 'command':
      final c = step['command'] as Map;
      require(c['game_id'] == b['game_id'] && c['id'] != '');
      final digest = sha256(canonical(c));
      final commands = b['commands'] as Map;
      if (commands.containsKey(c['id'])) {
        require(commands[c['id']] == digest);
        return null;
      }
      require(l['turn_started'] && b['phase'] == 'playing');
      final actor = c['actor'] as int;
      if (c['kind'] == 'open-series') {
        require(actor == b['active'] && !b.containsKey('pending'));
        final p = b['players'][actor - 1] as Map, suit = c['suit'];
        final old = list(p, 'history').contains(suit);
        require(old || !p['opening_used']);
        require(
          list(b, 'cards').any(
            (v) =>
                v['controller'] == actor &&
                v['zone'] == 'hand' &&
                v['card']['suit'] == suit &&
                v['card']['rank'] >= 2 &&
                v['card']['rank'] <= 10,
          ),
        );
        b['pending'] = {
          'id': c['id'],
          'kind': 'opening',
          'actor': actor,
          'responders': [
            for (var offset = 1; offset < n; offset++)
              (actor - 1 + offset) % n + 1,
          ],
          'opening': {
            'kind': 'open-series',
            'suit': suit,
            'prior_used': p['opening_used'],
          },
          'justice_queen': '',
          'defender': 0,
          'clubs': null,
          'targets': null,
          'cursor': 0,
          'attack_modifier': 0,
          'defense_modifier': 0,
          'aces': null,
          'decision': null,
        };
        if (!old) p['opening_used'] = true;
      } else if (c['kind'] == 'pass') {
        final p = b['pending'] as Map?;
        require(
          p != null &&
              p['id'] == c['window_id'] &&
              p['responders'][p['cursor']] == actor,
        );
        p!['cursor']++;
        if (p['cursor'] == list(p, 'responders').length) {
          final owner = p['actor'] as int, suit = p['opening']['suit'];
          final player = b['players'][owner - 1] as Map;
          final ids = list(b, 'cards')
              .where(
                (v) =>
                    v['controller'] == owner &&
                    v['zone'] == 'hand' &&
                    v['card']['suit'] == suit &&
                    v['card']['rank'] >= 2 &&
                    v['card']['rank'] <= 10,
              )
              .map((v) => v['card']['id'])
              .toList();
          require(ids.isNotEmpty);
          move(b, ids, owner, 'series');
          if (!list(player, 'history').contains(suit)) {
            player['history'] = [...list(player, 'history'), suit];
            player['opening_used'] = true;
          }
          b.remove('pending');
          events = [
            {
              'kind': 'open-series',
              'actor': owner,
              'amount': finance.Exact.zero.toJson(),
            },
          ];
        }
      } else {
        throw const FormatException('unsupported card action');
      }
      b['version']++;
      commands[c['id']] = digest;
      break;
    case 'end-turn':
      require(
        l['turn_started'] && !b.containsKey('pending') && l['ending'] == '',
      );
      l['turn_started'] = false;
      if (l['turn_index'] + 1 < list(b, 'order').length) {
        l['turn_index']++;
        b['active'] = b['order'][l['turn_index']];
        b['turn']++;
      } else {
        l['round_closed'] = true;
        // This profile has no attached kings, Underground, Code or debt, hence no
        // round receipts. Ordinary ending still scores real dealt/played holdings.
        if (b['round'] == 13) {
          final awards = <dynamic>[];
          for (var seat = 1; seat <= n; seat++) {
            var score = 0;
            for (final c in list(b, 'cards')) {
              if (c['controller'] != seat) continue;
              final face = c['card'];
              if (face['suit'] == 'diamonds' &&
                  face['rank'] >= 2 &&
                  face['rank'] <= 10)
                score += 5 + (face['rank'] as int);
              if (face['suit'] == 'spades' && face['rank'] >= 11) score += 10;
            }
            final a = finance.Exact.integer(score).toJson();
            awards.add({'seat': seat - 1, 'amount': a});
            l['promises']['awards'] = [
              ...list(l['promises'], 'awards'),
              {
                'id': '${b['game_id']}/ordinary-end/seat-$seat',
                'payer': seat - 1,
                'amount': a,
              },
            ];
          }
          finance.settle(ledger, awards);
          ledger['phase'] = 'settlement';
          l['promises']['phase'] = 'settlement';
          b['phase'] = 'settlement';
          b['effects'] = null;
          b['draw_queue'] = null;
          b.remove('proposals');
          for (final c in list(b, 'cards')) {
            c['available_from_round'] = 0;
          }
          l['ending'] = 'round-limit';
        }
      }
      break;
    case 'begin-round':
      require(l['round_closed'] && l['ending'] == '' && b['round'] < 13);
      b['round']++;
      b['turn']++;
      initiative(b, list(step, 'order'));
      l['turn_index'] = 0;
      l['round_closed'] = false;
      l.remove('departure_choices');
      l.remove('departures_resolved');
      l['code_cursor'] = 0;
      break;
    case 'finish':
      require(l['ending'] != '' && ledger['phase'] == 'settlement');
      final seat = (step['seat'] as int) - 1;
      require(seat >= 0 && seat < n);
      ledger['finished'][seat] = true;
      l['promises']['finished'][seat] = true;
      break;
    case 'finalize':
      require(
        ledger['phase'] == 'settlement' &&
            (ledger['finished'] as List).every((v) => v == true),
      );
      finance.append(ledger, 'completed', {
        'game_id': ledger['game_id'],
        'scores': [...ledger['scores']],
      });
      ledger['cash'] = List.generate(n, (_) => finance.Exact.zero.toJson());
      ledger['phase'] = 'finalized';
      l['promises']['phase'] = 'finalized';
      break;
    case 'next':
      require(
        ledger['phase'] == 'finalized' &&
            list(ledger, 'completed').length < m['game_limit'] &&
            !list(m, 'instances').contains(step['next_game']),
      );
      final id = step['next_game'] as String;
      m['command_history'] = [
        ...list(m, 'command_history'),
        {
          'game_id': b['game_id'],
          'card_commands': copy(b['commands']),
          'lifecycle_commands': copy(l['commands']),
        },
      ];
      final next = copy(ledger);
      next['game_id'] = id;
      next['phase'] = 'playing';
      for (final k in ['scores', 'cash']) {
        next[k] = List.generate(n, (_) => finance.Exact.zero.toJson());
      }
      for (final k in ['finished', 'departed']) {
        next[k] = List.filled(n, false);
      }
      m['game'] = newLifecycle(newBoard(n, id), next);
      m['start_ledger'] = copy(next);
      m['instances'] = [...list(m, 'instances'), id];
      m.remove('nullification_consents');
      break;
    default:
      throw const FormatException('unsupported lifecycle operation');
  }
  return events;
}

void validateProfile(Map m) {
  final l = m['game'] as Map,
      b = l['board'] as Map,
      ledger = l['ledger'] as Map;
  finance.validateLedger(ledger);
  require(
    m['game_limit'] >= 1 &&
        m['game_limit'] <= 3 &&
        list(m, 'nullification_consents').isEmpty,
  );
  require(
    b['schema'] == 'cgms-state-v1' &&
        b['game_id'] == ledger['game_id'] &&
        b['round'] >= 1 &&
        b['round'] <= 13,
  );
  for (final field in [
    'formations',
    'proposals',
    'loans',
    'effects',
    'bindings',
    'draw_queue',
    'decisions',
  ]) {
    require(list(b, field).isEmpty);
  }
  for (final field in ['automatic', 'promise_payment', 'payment_prior']) {
    require(l[field] == null);
  }
  for (final field in [
    'pending_receipts',
    'departure_choices',
    'forgiveness_consents',
  ]) {
    require(list(l, field).isEmpty);
  }
  require(
    b['code'] == null &&
        b['needs_shuffle'] == false &&
        l['end_turn_pending'] != true &&
        (l['automatic_purpose'] == null || l['automatic_purpose'] == '') &&
        ['', 'round-limit'].contains(l['ending']),
  );
  for (final field in ['debts', 'charges', 'corrections', 'operations']) {
    require(list(ledger, field).isEmpty);
  }
  require(
    list(l['promises'], 'promises').isEmpty &&
        list(l['promises'], 'payments').isEmpty,
  );
  final n = list(b, 'players').length;
  require(n == 3 || n == 4);
  for (final p in list(b, 'players')) {
    require(
      p['confined'] == false &&
          p['departed'] == false &&
          p['underground'] == false &&
          p['ally'] == 0 &&
          p['ace_round'] == 0 &&
          p['compensation_used'] == false &&
          (p['quotas'] as Map).isEmpty,
    );
  }
  final catalog = {
    for (final c in list(newBoard(n, 'catalog'), 'cards'))
      c['card']['id']: c['card'],
  };
  final seen = <String>{}, draws = <String>{};
  for (final c in list(b, 'cards')) {
    final face = c['card'] as Map, id = face['id'] as String;
    require(seen.add(id) && canonical(catalog[id]) == canonical(face));
    require(
      c['available_from_round'] == 0 &&
          c['used_turn'] == 0 &&
          c['allocation'] == '' &&
          ['draw', 'hand', 'concealed-ace', 'series'].contains(c['zone']),
    );
    if (c['zone'] == 'draw') {
      require(c['controller'] == 0);
      draws.add(id);
    } else {
      require(c['controller'] >= 1 && c['controller'] <= n);
    }
    require(c['zone'] != 'concealed-ace' || face['rank'] == 1);
    require(c['zone'] != 'hand' || face['rank'] != 1);
    require(c['zone'] != 'series' || (face['rank'] >= 2 && face['rank'] <= 10));
  }
  require(
    seen.length == 104 &&
        list(b, 'draw_order').length == draws.length &&
        list(b, 'draw_order').toSet().length == draws.length &&
        list(b, 'draw_order').every(draws.contains),
  );
  final pending = b['pending'];
  if (pending != null) {
    require(
      pending['kind'] == 'opening' &&
          pending['opening']['kind'] == 'open-series' &&
          pending['decision'] == null,
    );
  }
}

const cardplaySchema = 'cgms-offline-cardplay-v1';
Map<String, dynamic> runCardplay(
  Map<String, dynamic> request, {
  void Function(int, Map)? onFrame,
}) {
  Map<String, dynamic> fail(String code, int step) => {
    'schema': cardplaySchema,
    'error': code,
    'failed_step': step,
  };
  if (request['schema'] != cardplaySchema ||
      request.keys.any((k) => !['schema', 'match', 'steps'].contains(k)) ||
      list(request, 'steps').length > 2048)
    return fail('invalid_request', -1);
  var index = -1;
  try {
    final m = copy(request['match'] as Map);
    validateProfile(m);
    require(m['game_limit'] >= 1 && m['game_limit'] <= 3);
    final digests = <String>[];
    Map? last;
    List? events;
    for (index = -1; index < list(request, 'steps').length; index++) {
      if (index >= 0) events = stepCardplay(m, request['steps'][index]);
      last = cardplayFrame(m, events);
      digests.add(sha256(canonical(last)));
      onFrame?.call(index + 1, last);
    }
    return {
      'schema': cardplaySchema,
      'failed_step': -1,
      'digests': digests,
      'final': last,
    };
  } catch (_) {
    return fail(index < 0 ? 'invalid_state' : 'command_rejected', index);
  }
}
