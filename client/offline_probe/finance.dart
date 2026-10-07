// Independent, SDK-only financial comparison candidate. Synthetic decoded
// fixtures only: this is not a save-file loader or an online command boundary.
import 'dart:convert';

class Exact implements Comparable<Exact> {
  final BigInt n, d;
  Exact._(this.n, this.d);
  factory Exact(BigInt n, BigInt d) {
    if (d <= BigInt.zero)
      throw const FormatException('positive denominator required');
    final g = n.gcd(d);
    return Exact._(n ~/ g, d ~/ g);
  }
  static final zero = Exact(BigInt.zero, BigInt.one),
      one = Exact(BigInt.one, BigInt.one);
  static Exact integer(int n) => Exact(BigInt.from(n), BigInt.one);
  static Exact parse(dynamic value) {
    if (value is! Map ||
        value.length != 2 ||
        value['numerator'] is! String ||
        value['denominator'] is! String)
      throw const FormatException('rational object required');
    final pattern = RegExp(r'^(0|-[1-9][0-9]*|[1-9][0-9]*)$');
    final n = value['numerator'] as String, d = value['denominator'] as String;
    if (!pattern.hasMatch(n) || !pattern.hasMatch(d))
      throw const FormatException('decimal integer required');
    final a = Exact(BigInt.parse(n), BigInt.parse(d));
    if (a.n.toString() != n || a.d.toString() != d)
      throw const FormatException('reduced rational required');
    return a;
  }

  Exact add(Exact b) => Exact(n * b.d + b.n * d, d * b.d);
  Exact sub(Exact b) => Exact(n * b.d - b.n * d, d * b.d);
  Exact half() => Exact(n, d * BigInt.two);
  @override
  int compareTo(Exact b) => (n * b.d).compareTo(b.n * d);
  bool get positive => n > BigInt.zero;
  Map<String, String> toJson() => {
    'numerator': n.toString(),
    'denominator': d.toString(),
  };
}

Map<String, dynamic> copyMap(Map value) =>
    jsonDecode(jsonEncode(value)) as Map<String, dynamic>;
void require(bool condition) {
  if (!condition)
    throw const FormatException('unsupported or invalid financial fixture');
}

List<dynamic> items(Map l, String key) => (l[key] as List?) ?? [];
Exact amount(dynamic x) => Exact.parse(x);
void addAt(Map l, String field, int seat, Exact value) {
  l[field][seat] = amount(l[field][seat]).add(value).toJson();
}

void append(Map l, String field, Map value) {
  l[field] = [...items(l, field), value];
}

const financeSchema = 'cgms-offline-finance-probe-v1';

// Candidate validation covers accepted fixture structure and provenance. It does
// not authenticate a historical ledger or validate physical-card eligibility.
void validateLedger(Map l) {
  const fields = {
    'game_id',
    'phase',
    'scores',
    'cash',
    'debts',
    'charges',
    'corrections',
    'completed',
    'departed',
    'finished',
    'operations',
    'next_sequence',
  };
  require(l.length == fields.length && l.keys.every(fields.contains));
  require(l['game_id'] is String && (l['game_id'] as String).isNotEmpty);
  final scores = l['scores'] as List;
  final seats = scores.length;
  require(seats == 3 || seats == 4);
  require(['playing', 'settlement', 'finalized', 'void'].contains(l['phase']));
  for (final field in ['cash', 'departed', 'finished']) {
    require(l[field] is List && (l[field] as List).length == seats);
  }
  for (final field in ['departed', 'finished']) {
    require((l[field] as List).every((v) => v is bool));
  }
  for (final a in scores) {
    amount(a);
  }
  for (final a in l['cash']) {
    final x = amount(a);
    require(
      x.compareTo(Exact.zero) >= 0 &&
          (l['phase'] != 'finalized' || !x.positive),
    );
  }
  bool validSeat(dynamic s) => s is int && s >= 0 && s < seats;
  final games = <String, int>{};
  for (final r in items(l, 'completed')) {
    require(
      r['game_id'] is String &&
          r['game_id'] != '' &&
          !games.containsKey(r['game_id']) &&
          (r['scores'] as List).length == seats,
    );
    for (final a in r['scores']) {
      amount(a);
    }
    games[r['game_id']] = games.length;
    if (r['game_id'] == l['game_id']) {
      require(
        l['phase'] == 'finalized' &&
            identical(r, items(l, 'completed').last) &&
            jsonEncode(r['scores']) == jsonEncode(scores),
      );
    }
  }
  if (l['phase'] == 'finalized') {
    require(games.containsKey(l['game_id']));
  } else {
    games[l['game_id']] = games.length;
  }
  final charges = <String, Map>{};
  for (final c in items(l, 'charges')) {
    require(
      c['id'] is String &&
          c['id'] != '' &&
          !charges.containsKey(c['id']) &&
          games.containsKey(c['game_id']) &&
          validSeat(c['debtor']) &&
          (c['creditor'] == -1 || validSeat(c['creditor'])) &&
          c['debtor'] != c['creditor'] &&
          amount(c['amount']).positive,
    );
    charges[c['id']] = c;
  }
  var sequence = 0;
  final debts = <String, Map>{};
  final used = <String>{};
  for (final d in items(l, 'debts')) {
    require(
      d['id'] is String &&
          d['id'] != '' &&
          !debts.containsKey(d['id']) &&
          d['sequence'] is int &&
          d['sequence'] > sequence &&
          validSeat(d['debtor']) &&
          (d['creditor'] == -1 || validSeat(d['creditor'])) &&
          d['debtor'] != d['creditor'],
    );
    final c = charges[d['charge_id']];
    require(c != null && used.add(d['charge_id']));
    require(
      c!['game_id'] == d['game_id'] &&
          c['debtor'] == d['debtor'] &&
          c['creditor'] == d['creditor'] &&
          amount(d['remaining']).compareTo(Exact.zero) >= 0 &&
          amount(c['amount']).compareTo(amount(d['remaining'])) >= 0,
    );
    sequence = d['sequence'];
    debts[d['id']] = d;
  }
  require(used.length == charges.length && sequence == l['next_sequence']);
  final operations = <String>{};
  for (final id in items(l, 'operations')) {
    require(id is String && id.isNotEmpty && operations.add(id));
  }
  final corrections = <String>{}, corrected = <String, Exact>{};
  for (final c in items(l, 'corrections')) {
    final d = debts[c['debt_id']];
    require(
      d != null &&
          c['charge_id'] == d['charge_id'] &&
          validSeat(c['debtor']) &&
          c['debtor'] == d['debtor'] &&
          amount(c['amount']).positive &&
          operations.contains(c['operation_id']) &&
          corrections.add(c['operation_id']) &&
          games.containsKey(c['game_id']),
    );
    require(
      c['match_level'] == (c['game_id'] != d!['game_id']) &&
          games[c['game_id']]! >= games[d['game_id']]!,
    );
    final sum = (corrected[d['id']] ?? Exact.zero).add(amount(c['amount']));
    corrected[d['id']] = sum;
    require(
      sum
              .add(amount(d['remaining']))
              .compareTo(amount(charges[d['charge_id']]!['amount'])) <=
          0,
    );
  }
}

// Unbatched Q10 oracle: fixed-seat gross receipts, then FIFO creditor receipts.
// A work-budget refusal rejects the whole synthetic trace; it never forgives
// debt or returns a partially committed ledger. Production continuation is open.
void settle(Map l, List receipts) {
  require(receipts.length <= 64);
  final queue = <Map<String, dynamic>>[];
  for (var i = 0; i < receipts.length; i++) {
    final r = receipts[i];
    require(
      r['seat'] is int &&
          r['seat'] >= 0 &&
          r['seat'] < (l['scores'] as List).length &&
          amount(r['amount']).compareTo(Exact.zero) >= 0,
    );
    queue.add({...Map<String, dynamic>.from(r), 'order': i});
  }
  queue.sort((a, b) {
    final c = (a['seat'] as int).compareTo(b['seat']);
    return c != 0 ? c : (a['order'] as int).compareTo(b['order']);
  });
  var work = 0;
  for (var head = 0; head < queue.length; head++) {
    require(++work <= 4096);
    final seat = queue[head]['seat'] as int;
    var remaining = amount(queue[head]['amount']);
    addAt(l, 'scores', seat, remaining);
    for (final debt in items(l, 'debts')) {
      if (debt['debtor'] != seat || !remaining.positive) continue;
      final owed = amount(debt['remaining']);
      if (!owed.positive) continue;
      require(++work <= 4096);
      final paid = owed.compareTo(remaining) < 0 ? owed : remaining;
      debt['remaining'] = owed.sub(paid).toJson();
      remaining = remaining.sub(paid);
      if (debt['creditor'] >= 0)
        queue.add({'seat': debt['creditor'], 'amount': paid.toJson()});
    }
    addAt(l, 'cash', seat, remaining);
  }
}

void charge(Map l, String id, int debtor, int creditor, Exact a) {
  final seats = (l['scores'] as List).length;
  require(
    id.isNotEmpty &&
        debtor >= 0 &&
        debtor < seats &&
        creditor >= -1 &&
        creditor < seats &&
        creditor != debtor &&
        a.positive &&
        l['phase'] == 'playing',
  );
  for (final c in items(l, 'charges')) {
    if (c['id'] == id) {
      require(
        c['debtor'] == debtor &&
            c['creditor'] == creditor &&
            c['game_id'] == l['game_id'] &&
            amount(c['amount']).compareTo(a) == 0,
      );
      return;
    }
  }
  addAt(l, 'scores', debtor, Exact.zero.sub(a));
  l['next_sequence']++;
  append(l, 'charges', {
    'creditor': creditor,
    'id': id,
    'game_id': l['game_id'],
    'debtor': debtor,
    'amount': a.toJson(),
  });
  append(l, 'debts', {
    'id': id,
    'charge_id': id,
    'game_id': l['game_id'],
    'debtor': debtor,
    'creditor': creditor,
    'remaining': a.toJson(),
    'sequence': l['next_sequence'],
  });
  final cash = amount(l['cash'][debtor]);
  if (cash.positive) {
    l['cash'][debtor] = Exact.zero.toJson();
    addAt(l, 'scores', debtor, Exact.zero.sub(cash));
    settle(l, [
      {'seat': debtor, 'amount': cash.toJson()},
    ]);
  }
}

void apply(Map l, Map c, int limit) {
  const fields = {
    'start',
    'game_id',
    'kind',
    'id',
    'seat',
    'other',
    'partner',
    'amount',
    'debt',
    'consent',
    'receipts',
    'next_game',
  };
  require(
    c.keys.every(fields.contains) &&
        c['game_id'] == l['game_id'] &&
        items(l, 'debts').length <= 64,
  );
  final seats = (l['scores'] as List).length;
  final seat = c['seat'] as int;
  final other = c['other'] as int;
  bool valid(int s) => s >= 0 && s < seats;
  switch (c['kind']) {
    case 'charge':
      charge(l, c['id'], seat, other, amount(c['amount']));
      break;
    case 'receipt':
      require(l['phase'] == 'playing');
      settle(l, items(c, 'receipts'));
      break;
    case 'ponzi':
      final partner = c['partner'] as int;
      require(
        l['phase'] == 'playing' &&
            valid(seat) &&
            valid(other) &&
            seat != other &&
            (partner == -1 ||
                (valid(partner) && partner != seat && partner != other)),
      );
      final a = amount(l['cash'][seat]);
      l['cash'][seat] = Exact.zero.toJson();
      addAt(l, 'scores', seat, Exact.zero.sub(a));
      settle(
        l,
        partner == -1
            ? [
                {'seat': other, 'amount': a.toJson()},
              ]
            : [
                {'seat': other, 'amount': a.half().toJson()},
                {'seat': partner, 'amount': a.half().toJson()},
              ],
      );
      break;
    case 'forgive':
      final a = amount(c['amount']);
      require(
        c['id'] is String &&
            c['id'] != '' &&
            a.positive &&
            ['playing', 'settlement'].contains(l['phase']),
      );
      for (final correction in items(l, 'corrections')) {
        if (correction['operation_id'] == c['id']) {
          require(
            correction['debt_id'] == c['debt'] &&
                amount(correction['amount']).compareTo(a) == 0,
          );
          return;
        }
      }
      require(!items(l, 'operations').contains(c['id']));
      final matches = items(
        l,
        'debts',
      ).where((d) => d['id'] == c['debt']).toList();
      require(matches.length == 1);
      final debt = matches.single;
      require(amount(debt['remaining']).compareTo(a) >= 0);
      final consent = items(c, 'consent');
      require(
        debt['creditor'] >= 0
            ? consent.contains(debt['creditor'])
            : List.generate(seats, (s) => s).every(consent.contains),
      );
      debt['remaining'] = amount(debt['remaining']).sub(a).toJson();
      final cross = debt['game_id'] != l['game_id'];
      append(l, 'corrections', {
        'operation_id': c['id'],
        'debt_id': c['debt'],
        'charge_id': debt['charge_id'],
        'game_id': l['game_id'],
        'debtor': debt['debtor'],
        'amount': a.toJson(),
        'match_level': cross,
      });
      l['operations'] = [...items(l, 'operations'), c['id']];
      if (!cross) addAt(l, 'scores', debt['debtor'], a);
      break;
    case 'coup':
      require(l['phase'] == 'playing' && valid(seat) && !l['departed'][seat]);
      for (var s = 0; s < seats; s++) {
        if (l['departed'][s]) continue;
        l['scores'][s] = Exact.zero.toJson();
        l['cash'][s] = Exact.zero.toJson();
        for (final d in items(l, 'debts')) {
          if (d['debtor'] == s && d['game_id'] == l['game_id'])
            addAt(l, 'scores', s, Exact.zero.sub(amount(d['remaining'])));
        }
      }
      if (l['corrections'] != null)
        l['corrections'] = items(
          l,
          'corrections',
        ).where((c) => c['match_level'] || l['departed'][c['debtor']]).toList();
      for (var s = 0; s < seats; s++) {
        if (s != seat && !l['departed'][s])
          charge(l, '${l['game_id']}/coup/$s', s, -1, Exact.integer(50));
      }
      settle(l, [
        {'seat': seat, 'amount': Exact.integer(50).toJson()},
      ]);
      l['phase'] = 'settlement';
      break;
    case 'close':
      require(l['phase'] == 'playing');
      for (final r in items(c, 'receipts')) {
        require(
          r['seat'] is int && valid(r['seat']) && !l['departed'][r['seat']],
        );
      }
      settle(l, items(c, 'receipts'));
      l['phase'] = 'settlement';
      break;
    case 'finish':
      require(l['phase'] == 'settlement' && valid(seat));
      l['finished'][seat] = true;
      break;
    case 'finalize':
      require(
        l['phase'] == 'settlement' &&
            items(l, 'completed').length < limit &&
            (l['finished'] as List).every((f) => f == true),
      );
      append(l, 'completed', {
        'game_id': l['game_id'],
        'scores': [...l['scores']],
      });
      l['cash'] = List.generate(seats, (_) => Exact.zero.toJson());
      l['phase'] = 'finalized';
      break;
    case 'nullify':
      require(c['start'] is Map && l['phase'] != 'finalized');
      final start = copyMap(c['start']);
      validateLedger(start);
      require(
        start['phase'] == 'playing' &&
            start['game_id'] == l['game_id'] &&
            (start['scores'] as List).length == seats &&
            List.generate(seats, (s) => s).every(items(c, 'consent').contains),
      );
      l.clear();
      l.addAll(start);
      l['phase'] = 'void';
      break;
    case 'next':
      require(
        ['finalized', 'void'].contains(l['phase']) &&
            items(l, 'completed').length < limit &&
            c['next_game'] is String &&
            c['next_game'] != '' &&
            c['next_game'] != l['game_id'] &&
            !items(l, 'completed').any((r) => r['game_id'] == c['next_game']),
      );
      l['game_id'] = c['next_game'];
      l['phase'] = 'playing';
      for (final f in ['cash', 'scores']) {
        l[f] = List.generate(seats, (_) => Exact.zero.toJson());
      }
      for (final f in ['departed', 'finished']) {
        l[f] = List.filled(seats, false);
      }
      break;
    default:
      throw const FormatException('unsupported financial operation');
  }
}

Map<String, dynamic> frame(Map l, int limit) {
  final seats = (l['scores'] as List).length;
  final totals = List.filled(seats, Exact.zero);
  for (final r in items(l, 'completed')) {
    for (var s = 0; s < seats; s++) {
      totals[s] = totals[s].add(amount(r['scores'][s]));
    }
  }
  if (!['finalized', 'void'].contains(l['phase'])) {
    for (var s = 0; s < seats; s++) {
      totals[s] = totals[s].add(amount(l['scores'][s]));
    }
  }
  for (final c in items(l, 'corrections')) {
    if (c['match_level'])
      totals[c['debtor']] = totals[c['debtor']].add(amount(c['amount']));
  }
  final observations = <Object>[], decisions = <Object?>[];
  for (var s = 0; s < seats; s++) {
    final legal = <Object>[];
    if (l['phase'] == 'settlement' && !l['finished'][s])
      legal.add({'game_id': l['game_id'], 'seat': s, 'kind': 'finish'});
    final observation = {
      'game_id': l['game_id'],
      'seat': s,
      'phase': l['phase'],
      'score': l['scores'][s],
      'cash': l['cash'][s],
      'debts': items(
        l,
        'debts',
      ).where((d) => d['debtor'] == s || d['creditor'] == s).toList(),
      'legal': legal,
    };
    observations.add(observation);
    decisions.add(legal.isEmpty ? null : legal.first);
  }
  return copyMap({
    'ledger': l,
    'totals': totals.map((a) => a.toJson()).toList(),
    'ranks': totals
        .map((a) => 1 + totals.where((b) => b.compareTo(a) > 0).length)
        .toList(),
    'complete':
        l['phase'] == 'finalized' && items(l, 'completed').length == limit,
    'observations': observations,
    'decisions': decisions,
  });
}

Map<String, dynamic> runFinance(Map<String, dynamic> request) {
  var step = -1;
  var error = 'invalid_request';
  try {
    require(
      request.length == 4 &&
          request.keys.every(
            {'schema', 'ledger', 'game_limit', 'commands'}.contains,
          ) &&
          request['schema'] == financeSchema &&
          request['game_limit'] is int &&
          request['game_limit'] >= 1 &&
          request['game_limit'] <= 3 &&
          request['commands'] is List &&
          (request['commands'] as List).length <= 64,
    );
    final limit = request['game_limit'] as int;
    final l = copyMap(request['ledger'] as Map);
    error = 'invalid_state';
    validateLedger(l);
    require(
      (items(l, 'completed').length != limit || l['phase'] == 'finalized') &&
          items(l, 'completed').length <= limit &&
          items(l, 'debts').length <= 64,
    );
    final frames = <Object>[frame(l, limit)];
    error = 'command_rejected';
    for (final c in request['commands']) {
      step++;
      apply(l, c as Map, limit);
      validateLedger(l);
      frames.add(frame(l, limit));
    }
    return {'schema': financeSchema, 'failed_step': -1, 'frames': frames};
  } on FormatException {
    return {'schema': financeSchema, 'error': error, 'failed_step': step};
  } on TypeError {
    return {'schema': financeSchema, 'error': error, 'failed_step': step};
  } on RangeError {
    return {'schema': financeSchema, 'error': error, 'failed_step': step};
  }
}
