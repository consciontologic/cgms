import 'dart:convert';

// Wire field names from docs/api/online.openapi.json. Authority validates legality.
const commandFields = <String, List<String>>{
  'end-turn': [],
  'coup': [],
  'nullify': [],
  'finish-settlement': [],
  'declare-ordinary': ['threshold'],
  'departure-choice': ['accept'],
  'promise-offer': [
    'promise_id',
    'recipient',
    'award_id',
    'mode',
    'amount',
    'condition',
  ],
  'promise-accept': ['promise_id'],
  'promise-final-offer': ['promise_id', 'amount'],
  'promise-final-answer': ['promise_id', 'accept'],
  'promise-pay': ['promise_id'],
  'promise-refuse': ['promise_id'],
  'voluntary-transfer': ['recipient', 'amount'],
  'forgive': ['debt_id', 'amount'],
  'open-series': ['suit', 'selection'],
  'open-formation': ['formation', 'formation_id'],
  'take-back': ['selection', 'formation_id'],
  'attach': ['selection', 'suit'],
  'ringleader': ['selection', 'ace', 'value'],
  'richer-sacrifice': ['selection'],
  'barricade-sacrifice': ['selection'],
  'dexter-assassination': ['selection', 'targets'],
  'fate': ['selection', 'target_seat', 'formation_id'],
  'code': ['formation_id', 'value', 'price'],
  'compensation': ['ace'],
  'main-inflation': ['ace', 'target_seat'],
  'kidnapper': ['formation_id', 'target_seat'],
  'kidnapper-outcome': ['selection', 'pending_action_id'],
  'offer': ['offer_id', 'revision', 'terms'],
  'withdraw-offer': ['offer_id', 'revision'],
  'decline-offer': ['offer_id', 'revision'],
  'accept-offer': ['offer_id', 'revision'],
  'purchase': ['quantity', 'payment'],
  'ponzi': ['value'],
  'attack': [
    'selection',
    'targets',
    'ace',
    'value',
    'suit',
    'exile',
    'infiltrator',
  ],
  'infiltrator-attack': [
    'selection',
    'targets',
    'ace',
    'value',
    'suit',
    'exile',
    'infiltrator',
  ],
  'baron-attack': [
    'selection',
    'targets',
    'ace',
    'value',
    'suit',
    'exile',
    'infiltrator',
  ],
  'bomb-attack': [
    'selection',
    'targets',
    'ace',
    'value',
    'suit',
    'exile',
    'infiltrator',
  ],
  'baron-bomb-attack': [
    'selection',
    'targets',
    'ace',
    'value',
    'suit',
    'exile',
    'infiltrator',
  ],
  'justice': ['ace'],
  'decision-closed': ['pending_action_id', 'decision_id'],
  'pass': ['pending_action_id'],
  'compensation-response': ['pending_action_id', 'ace'],
  'numerical-defense': ['pending_action_id', 'ace', 'value'],
  'advancement-defense': ['pending_action_id', 'ace'],
  'confinement': ['pending_action_id', 'ace'],
  'inflation': ['pending_action_id', 'ace'],
  'negotiate': ['pending_action_id'],
  'decision': [
    'pending_action_id',
    'decision_id',
    'selection',
    'ace',
    'value',
    'suit',
    'target_seat',
  ],
  'return-loan': ['selection', 'target_seat'],
  'expose-unused-ace': ['ace'],
};

/// Coarse public timing only. Holdings, costs, quotas and detailed legality stay
/// with the authority; an available composer is never a promise of acceptance.
enum CommandTiming {
  available,
  ownTurn,
  pending,
  otherResponder,
  effectInput,
  noResponse,
  noDecision,
  automatic,
}

CommandTiming commandTiming(
  String type,
  Map<String, dynamic> board,
  Map<String, dynamic> online,
) {
  // Coup is admitted by current eligibility at the server's atomic boundary,
  // including pending decisions/draws and confinement. Never apply the coarse
  // ordinary-action timing lock to this victory intent.
  if (type == 'coup') return CommandTiming.available;
  if (online['automatic_pending'] == true || online['server_pending'] == true) {
    return CommandTiming.automatic;
  }
  final window = board['window_id'] != null;
  final decision = board['decision_id'] != null;
  final effect = (board['decision_kind'] as String? ?? '').isNotEmpty;
  final owner = board['required_actor'] == board['seat'];
  const responses = {
    'pass',
    'compensation-response',
    'numerical-defense',
    'advancement-defense',
    'confinement',
    'inflation',
    'negotiate',
  };
  if (responses.contains(type)) {
    if (!window) return CommandTiming.noResponse;
    if (effect || decision) return CommandTiming.effectInput;
    return owner ? CommandTiming.available : CommandTiming.otherResponder;
  }
  if (type == 'decision' || type == 'decision-closed') {
    return window && decision && owner
        ? CommandTiming.available
        : CommandTiming.noDecision;
  }
  const ordinary = {
    'end-turn',
    'open-series',
    'open-formation',
    'take-back',
    'attach',
    'ringleader',
    'richer-sacrifice',
    'barricade-sacrifice',
    'dexter-assassination',
    'fate',
    'code',
    'compensation',
    'main-inflation',
    'kidnapper',
    'purchase',
    'attack',
    'infiltrator-attack',
    'baron-attack',
    'bomb-attack',
    'baron-bomb-attack',
    'justice',
  };
  // An intact formation loan may be offered on either party's turn.
  // Ordinary victory can likewise be established by another seat's transfer.
  if ((type == 'offer' ||
          type == 'accept-offer' ||
          type == 'declare-ordinary') &&
      window)
    return CommandTiming.pending;
  if (ordinary.contains(type)) {
    if (window) return CommandTiming.pending;
    if (board['active'] != board['seat']) return CommandTiming.ownTurn;
  }
  return CommandTiming.available;
}

// Closed structural shapes prevent accidental persistence of credentials or faces.
final _payloadSchemas =
    jsonDecode(
          r'''{"Payload":{"type":"object","properties":{"pending_action_id":{"type":"string"},"decision_id":{"type":"string"},"selection":{"type":"array","items":{"type":"string"}},"targets":{"type":"array","items":{"type":"string"}},"ace":{"type":"string"},"value":{"type":"integer"},"quantity":{"type":"integer"},"payment":{"type":"array","items":{"type":"string"}},"suit":{"type":"string"},"formation":{"$ref":"#/components/schemas/Formation"},"formation_id":{"type":"string"},"offer_id":{"type":"string"},"revision":{"type":"integer"},"terms":{"$ref":"#/components/schemas/ProposalTerms"},"target_seat":{"type":"integer"},"price":{"type":"integer"},"exile":{"type":"boolean"},"infiltrator":{"type":"boolean"},"promise_id":{"type":"string"},"recipient":{"type":"integer"},"award_id":{"type":"string"},"mode":{"type":"string"},"amount":{"$ref":"#/components/schemas/Amount"},"accept":{"type":"boolean"},"debt_id":{"type":"string"},"threshold":{"type":"string"},"condition":{"$ref":"#/components/schemas/Condition"}},"required":[],"additionalProperties":false},"Formation":{"type":"object","properties":{"kind":{"type":"string"},"cards":{"type":"array","items":{"$ref":"#/components/schemas/CardHandle"}},"substitute":{"type":"object","properties":{"kings":{"type":"array","items":{"$ref":"#/components/schemas/CardHandle"}},"slot":{"type":"object","properties":{"rank":{"type":"integer"},"suit":{"type":"string"}},"required":["rank","suit"],"additionalProperties":false}},"required":["kings","slot"],"additionalProperties":false},"protection":{"type":"object","properties":{"series":{"type":"string"},"formation":{"type":"string"}},"required":[],"additionalProperties":false}},"required":["kind","cards"],"additionalProperties":false},"ProposalTerms":{"type":"object","properties":{"to":{"type":"integer","minimum":1,"maximum":4},"give":{"type":"array","items":{"$ref":"#/components/schemas/CardHandle"}},"receive":{"type":"array","items":{"$ref":"#/components/schemas/CardHandle"}},"loan":{"type":"boolean"}},"required":["to"],"additionalProperties":false},"Amount":{"type":"object","properties":{"numerator":{"type":"string","pattern":"^-?(0|[1-9][0-9]*)$"},"denominator":{"type":"string","pattern":"^[1-9][0-9]*$"}},"required":["numerator","denominator"],"additionalProperties":false},"Condition":{"type":"object","properties":{"kind":{"const":"award-occurrence"},"game_id":{"type":"string"},"award_id":{"type":"string"},"payer":{"type":"integer","minimum":1,"maximum":4}},"required":["kind","game_id","award_id","payer"],"additionalProperties":false},"CardHandle":{"type":"string","pattern":"^h_[A-Za-z0-9_-]{43}$","description":"Opaque actor/game/current-view capability. History handles preserve knowledge but cannot authorize current selection."}}''',
        )
        as Map<String, dynamic>;

void validateCommandPayload(String type, Map<String, dynamic> payload) {
  final allowed = commandFields[type];
  if (allowed == null ||
      payload.keys.any((key) => !allowed.contains(key)) ||
      !_matches(payload, _payloadSchemas['Payload'] as Map<String, dynamic>)) {
    throw ArgumentError('INVALID_COMMAND');
  }
}

bool _matches(dynamic value, Map<String, dynamic> schema) {
  final reference = schema[r'$ref'];
  if (reference is String)
    return _matches(
      value,
      _payloadSchemas[reference.split('/').last] as Map<String, dynamic>,
    );
  switch (schema['type']) {
    case 'object':
      if (value is! Map) return false;
      final properties = schema['properties'] as Map<String, dynamic>;
      if (value.keys.any((key) => !properties.containsKey(key))) return false;
      if ((schema['required'] as List? ?? []).any(
        (key) => !value.containsKey(key),
      ))
        return false;
      return value.entries.every(
        (entry) => _matches(
          entry.value,
          properties[entry.key] as Map<String, dynamic>,
        ),
      );
    case 'array':
      return value is List &&
          value.every(
            (item) => _matches(item, schema['items'] as Map<String, dynamic>),
          );
    case 'string':
      return value is String;
    case 'integer':
      return value is int;
    case 'boolean':
      return value is bool;
  }
  return schema.containsKey('const') && value == schema['const'];
}
