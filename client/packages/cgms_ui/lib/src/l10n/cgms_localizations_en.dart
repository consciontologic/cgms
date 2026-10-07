// ignore: unused_import
import 'package:intl/intl.dart' as intl;
import 'cgms_localizations.dart';

// ignore_for_file: type=lint

/// The translations for English (`en`).
class CgmsLocalizationsEn extends CgmsLocalizations {
  CgmsLocalizationsEn([String locale = 'en']) : super(locale);

  @override
  String get concealedCard => 'Concealed card';

  @override
  String get drawPile => 'Draw pile';

  @override
  String get magnifyCard => 'Magnify card';

  @override
  String inspectCard(String card) {
    return 'Inspect $card';
  }

  @override
  String copyNumber(int number) {
    return 'Copy $number';
  }

  @override
  String cardIdentity(String rank, String suit, int copy) {
    return '$rank of $suit, copy $copy';
  }

  @override
  String get selected => 'Selected';

  @override
  String get hearts => 'Hearts';

  @override
  String get diamonds => 'Diamonds';

  @override
  String get clubs => 'Clubs';

  @override
  String get spades => 'Spades';

  @override
  String get startMatch => 'Start match';

  @override
  String get nextGame => 'Your next game';

  @override
  String get setupNotice =>
      'Choose a display name and the number of games in your match.';

  @override
  String get setupEyebrow => 'A GAME OF PUBLIC POWER & PRIVATE DOUBT';

  @override
  String get takeYourSeat => 'TAKE YOUR SEAT.';

  @override
  String get setupTagline =>
      'Build influence. Read the table.\nLeave nothing to chance.';

  @override
  String get editionName => 'cgmsart / FIRST LIGHT';

  @override
  String get setupFormTitle => 'A place at the table';

  @override
  String get setupDeckDescription =>
      'Four seats. Two decks. Every physical card matters.';

  @override
  String get yourName => 'Your name';

  @override
  String get nameHint => 'Enter a display name';

  @override
  String get nameRequired => 'Enter a name to take your seat.';

  @override
  String get matchGameCount => 'Games in the match';

  @override
  String get oneGame => '1 game';

  @override
  String get threeGames => '3 games';

  @override
  String get yourMove => 'YOUR MOVE.';

  @override
  String get atTheTable => 'AT THE TABLE.';

  @override
  String get loadingTable => 'Loading the table';

  @override
  String get tableLoadError => 'The table could not be loaded';

  @override
  String get disconnectedActionsPaused => 'Disconnected · actions paused';

  @override
  String get connectionExplanation =>
      'Your last known cards and pending decisions stay here. Actions remain paused until the connection is restored.';

  @override
  String get retryConnection => 'Retry connection';

  @override
  String get restoreConnection => 'Restore connection';

  @override
  String get publicTable => 'The public table';

  @override
  String get readTheirPosition => 'READ THEIR POSITION';

  @override
  String get scoresAndDebt => 'Scores & debt';

  @override
  String get yourPublicPosition => 'YOUR PUBLIC POSITION';

  @override
  String get yourHand => 'Your hand';

  @override
  String get privateHandEyebrow => 'ONLY YOU CAN SEE THESE';

  @override
  String get emptyHandTitle => 'Your hand is empty';

  @override
  String get emptyHandMessage =>
      'Public cards and scores remain available. There are no private cards to inspect.';

  @override
  String get tradeCards => 'Trade cards';

  @override
  String get endTurnAndScore => 'End turn & score';

  @override
  String get finishResponsesFirst =>
      'Finish the response window before ending your turn.';

  @override
  String get tableJournal => 'Table journal';

  @override
  String get eventEyebrow => 'YOUR VIEW OF EVENTS';

  @override
  String get noEvents => 'No events yet.';

  @override
  String get acceptOffer => 'Accept offer';

  @override
  String get declineOffer => 'Decline offer';

  @override
  String get tradeEyebrow => 'A VOLUNTARY EXCHANGE';

  @override
  String get makeADeal => 'MAKE A DEAL.';

  @override
  String get youGive => 'YOU GIVE';

  @override
  String get youReceive => 'YOU RECEIVE';

  @override
  String get tradeExplanation =>
      'Private offer terms · visible to the two parties.\nProposing reserves no cards. Acceptance opens responses; only resolution transfers the cards.';

  @override
  String get proposeTrade => 'Propose trade';

  @override
  String get withdrawOffer => 'Withdraw offer';

  @override
  String get resultRecorded => 'THE RESULT IS RECORDED';

  @override
  String get pointsHaveHistory => 'EVERY POINT HAS A HISTORY';

  @override
  String get matchScore => 'MATCH SCORE';

  @override
  String get thisGame => 'THIS GAME';

  @override
  String get availablePoints => 'AVAILABLE';

  @override
  String get debt => 'DEBT';

  @override
  String get scoreNoticeTitle => 'Score is not spending money';

  @override
  String get scoreNoticeMessage =>
      'Receipts repay the oldest outstanding debt first. Repayment does not deduct your score twice. At financial finalization, unspent current-game points close; recorded scores remain.';

  @override
  String get finishSettlement => 'Finish settlement';

  @override
  String get artworkCosmetic =>
      'Artwork is cosmetic. Rank, suit and availability come from the supplied player view.';

  @override
  String get closeInspection => 'Close inspection';

  @override
  String roundPhase(int round, int total, String phase) {
    return 'ROUND $round / $total  ·  $phase';
  }

  @override
  String exposedCardsOf(String name) {
    return '$name · exposed cards';
  }

  @override
  String cardCount(int count) {
    String _temp0 = intl.Intl.pluralLogic(
      count,
      locale: localeName,
      other: '$count cards',
      one: '1 card',
    );
    return '$_temp0';
  }

  @override
  String concealedCounts(int hand, int aces) {
    return '$hand in hand · $aces concealed aces';
  }

  @override
  String physicalCopy(int copy) {
    return 'Physical copy $copy of 2';
  }

  @override
  String get allCards => 'All';

  @override
  String get inspectSelection => 'Inspect selection';

  @override
  String get clearSelection => 'Clear selection';

  @override
  String get selectionOutsideFilter => 'Your selection is outside this filter.';

  @override
  String get handSelectionHint =>
      'Tap to select; double-tap to magnify. Review the action before playing.';

  @override
  String get noMatchingCards => 'No cards match this filter.';

  @override
  String get changeHandFilter => 'Choose another suit or All to see your hand.';

  @override
  String get yourExposedCards => 'Your exposed cards';

  @override
  String get jumpToCards => 'Jump to cards';

  @override
  String suitFilter(String suit, int count) {
    return '$suit ($count)';
  }

  @override
  String showingCards(int visible, int total) {
    return 'Showing $visible of $total cards';
  }

  @override
  String selectedCard(String card) {
    return 'Selected: $card';
  }

  @override
  String suitGroup(String suit, int count) {
    String _temp0 = intl.Intl.pluralLogic(
      count,
      locale: localeName,
      other: '$count cards',
      one: '1 card',
    );
    return '$suit · $_temp0';
  }

  @override
  String get scrollPublicTable => 'Scroll to compare all public positions';

  @override
  String get inspectReviewedCard => 'Inspect card';

  @override
  String inspectCombo(String name) {
    return 'Inspect $name';
  }

  @override
  String get handPatterns => 'Patterns in your hand';

  @override
  String get actionBoardPublic => 'TABLE · ALL OPEN CARDS';

  @override
  String get actionBoardYou => 'You';

  @override
  String actionBoardRound(int round, int total) {
    return 'Round $round / $total';
  }

  @override
  String actionBoardCounts(int hand, int aces) {
    return 'Hand $hand · Aces $aces';
  }

  @override
  String get actionBoardPrivate => 'Private';

  @override
  String get actionBoardEndTurn => 'End turn';

  @override
  String get actionBoardReviewPile => 'Review every card in this stack';

  @override
  String get actionBoardInspectHint =>
      'Double-tap, use the magnifier, or press I to magnify';

  @override
  String get actionBoardSelect => 'Select card';

  @override
  String actionBoardAces(int count) {
    return 'Aces · $count';
  }

  @override
  String get touchTable => 'Table';

  @override
  String get touchHand => 'Hand';

  @override
  String get touchActivity => 'Activity';

  @override
  String get touchNoPlayers => 'No players at the table';

  @override
  String get touchNoPlayersDetail =>
      'Player positions appear when a table is available.';

  @override
  String get touchNoPublicCards => 'No exposed cards yet';

  @override
  String get touchNoPublicCardsDetail =>
      'This player\'s concealed cards stay private.';

  @override
  String touchSeatSummary(int cards, int combos) {
    String _temp0 = intl.Intl.pluralLogic(
      combos,
      locale: localeName,
      other: '$combos combos',
      one: '1 combo',
    );
    return '$cards exposed · $_temp0';
  }

  @override
  String get touchMoveDetails => 'Move details';

  @override
  String get touchResponsePassed => 'Passed';

  @override
  String get touchResponseAwaiting => 'Awaiting response';

  @override
  String get touchResponseWaiting => 'Waiting';

  @override
  String get adaptiveTable => 'Table';

  @override
  String get adaptiveHand => 'Hand';

  @override
  String get adaptiveDecisions => 'Decisions';

  @override
  String get adaptiveCoup => 'Declare Coup now';

  @override
  String adaptiveFinances(String score, String cash, String debt) {
    return 'Score $score · Cash $cash · Debt $debt';
  }

  @override
  String get onlineCGMSOnline => 'CGMS · Online';

  @override
  String get onlineArtwork => 'Artwork';

  @override
  String get onlineRightToLeft => 'Right to left';

  @override
  String get online200Text => '200% text';

  @override
  String get onlineCheckOperation => 'Check operation';

  @override
  String get onlineRetryOriginalOperation => 'Retry original operation';

  @override
  String get onlineContinueAsGuest => 'Continue as guest';

  @override
  String get onlineRecoverRoomOperation => 'Recover room operation';

  @override
  String get onlineCreateRoom => 'Create room';

  @override
  String get onlineJoinRoom => 'Join room';

  @override
  String get onlineRefreshSeats => 'Refresh seats';

  @override
  String get onlineCreateInvitation => 'Create invitation';

  @override
  String get onlineExactSettlementIsProgressingOrdinaryActionsWait =>
      'Exact settlement is progressing. Ordinary actions wait.';

  @override
  String get onlineReturnToDecisions => 'Return to decisions';

  @override
  String get onlineClearSelection => 'Clear selection';

  @override
  String get onlinePassResponse => 'Pass response';

  @override
  String get onlineLoanIntactFormation => 'Loan this intact formation';

  @override
  String get onlineDoppelgangerSubstitution => 'Doppelganger substitution';

  @override
  String get onlineProtectASeriesOrFormation => 'Protect a series or formation';

  @override
  String get online => ' / ';

  @override
  String get onlineUsername => 'Username';

  @override
  String get onlinePassword => 'Password';

  @override
  String get onlinePlayers => 'Players';

  @override
  String get onlineGamesPerMatch => 'Games per match';

  @override
  String get onlineRoomCode => 'Room code';

  @override
  String get onlineInvitation => 'Invitation';

  @override
  String get onlineAction => 'Action';

  @override
  String get onlineSuit => 'Suit';

  @override
  String get onlineFormation => 'Formation';

  @override
  String get onlineSubstitutedRank => 'Substituted rank';

  @override
  String get onlineSubstitutedSuit => 'Substituted suit';

  @override
  String get onlineProtectedSeries => 'Protected series';

  @override
  String get onlineAccept => 'Accept';

  @override
  String get onlineAcceptOffer => 'Accept Offer';

  @override
  String get onlineAce => 'Ace';

  @override
  String get onlineAdvancementDefense => 'Advancement Defense';

  @override
  String get onlineAmount => 'Amount';

  @override
  String get onlineArray => 'Array';

  @override
  String get onlineAttach => 'Attach';

  @override
  String get onlineAttack => 'Attack';

  @override
  String get onlineAwardId => 'Award Id';

  @override
  String get onlineBaronAttack => 'Baron Attack';

  @override
  String get onlineBaronBombAttack => 'Baron Bomb Attack';

  @override
  String get onlineBarricadeSacrifice => 'Barricade Sacrifice';

  @override
  String get onlineBombAttack => 'Bomb Attack';

  @override
  String get onlineBoolean => 'Boolean';

  @override
  String get onlineCode => 'Code';

  @override
  String get onlineCompensation => 'Compensation';

  @override
  String get onlineCompensationResponse => 'Compensation Response';

  @override
  String get onlineCondition => 'Condition';

  @override
  String get onlineConfinement => 'Confinement';

  @override
  String get onlineConst => 'Const';

  @override
  String get onlineCoup => 'Coup';

  @override
  String get onlineDebtId => 'Debt Id';

  @override
  String get onlineDecision => 'Decision';

  @override
  String get onlineDecisionClosed => 'Decision Closed';

  @override
  String get onlineDecisionId => 'Decision Id';

  @override
  String get onlineDeclareOrdinary => 'Declare Ordinary';

  @override
  String get onlineDeclineOffer => 'Decline Offer';

  @override
  String get onlineDepartureChoice => 'Departure Choice';

  @override
  String get onlineDexterAssassination => 'Dexter Assassination';

  @override
  String get onlineEndTurn => 'End Turn';

  @override
  String get onlineExile => 'Exile';

  @override
  String get onlineExposeUnusedAce => 'Expose Unused Ace';

  @override
  String get onlineFate => 'Fate';

  @override
  String get onlineFinishSettlement => 'Finish Settlement';

  @override
  String get onlineForgive => 'Forgive';

  @override
  String get onlineFormationId => 'Formation Id';

  @override
  String get onlineInfiltrator => 'Infiltrator';

  @override
  String get onlineInfiltratorAttack => 'Infiltrator Attack';

  @override
  String get onlineInflation => 'Inflation';

  @override
  String get onlineInteger => 'Integer';

  @override
  String get onlineItems => 'Items';

  @override
  String get onlineJustice => 'Justice';

  @override
  String get onlineKidnapper => 'Kidnapper';

  @override
  String get onlineKidnapperOutcome => 'Kidnapper Outcome';

  @override
  String get onlineMainInflation => 'Main Inflation';

  @override
  String get onlineMode => 'Mode';

  @override
  String get onlineNegotiate => 'Negotiate';

  @override
  String get onlineNullify => 'Nullify';

  @override
  String get onlineNumericalDefense => 'Numerical Defense';

  @override
  String get onlineObject => 'Object';

  @override
  String get onlineOffer => 'Offer';

  @override
  String get onlineOfferId => 'Offer Id';

  @override
  String get onlineOpenFormation => 'Open Formation';

  @override
  String get onlineOpenSeries => 'Open Series';

  @override
  String get onlinePass => 'Pass';

  @override
  String get onlinePayment => 'Payment';

  @override
  String get onlinePendingActionId => 'Pending Action Id';

  @override
  String get onlinePonzi => 'Ponzi';

  @override
  String get onlinePrice => 'Price';

  @override
  String get onlinePromiseAccept => 'Promise Accept';

  @override
  String get onlinePromiseFinalAnswer => 'Promise Final Answer';

  @override
  String get onlinePromiseFinalOffer => 'Promise Final Offer';

  @override
  String get onlinePromiseOffer => 'Promise Offer';

  @override
  String get onlinePromisePay => 'Promise Pay';

  @override
  String get onlinePromiseRefuse => 'Promise Refuse';

  @override
  String get onlinePromiseId => 'Promise Id';

  @override
  String get onlineProperties => 'Properties';

  @override
  String get onlinePurchase => 'Purchase';

  @override
  String get onlineQuantity => 'Quantity';

  @override
  String get onlineRecipient => 'Recipient';

  @override
  String get onlineRequired => 'Required';

  @override
  String get onlineReturnLoan => 'Return Loan';

  @override
  String get onlineRevision => 'Revision';

  @override
  String get onlineRicherSacrifice => 'Richer Sacrifice';

  @override
  String get onlineRingleader => 'Ringleader';

  @override
  String get onlineSelection => 'Selection';

  @override
  String get onlineString => 'String';

  @override
  String get onlineTakeBack => 'Take Back';

  @override
  String get onlineTargetSeat => 'Target Seat';

  @override
  String get onlineTargets => 'Targets';

  @override
  String get onlineTerms => 'Terms';

  @override
  String get onlineThreshold => 'Threshold';

  @override
  String get onlineType => 'Type';

  @override
  String get onlineValue => 'Value';

  @override
  String get onlineVoluntaryTransfer => 'Voluntary Transfer';

  @override
  String get onlineWithdrawOffer => 'Withdraw Offer';

  @override
  String onlineQuotaStatus(String ability, String status) {
    return '$ability quota: $status';
  }

  @override
  String get onlineQuotaUsed => 'used';

  @override
  String get onlineQuotaAvailable => 'available';

  @override
  String get onlineRefreshConnection => 'Refresh connection';

  @override
  String get onlineDisplaySettings => 'Display settings';

  @override
  String get onlineOutcomeUnknownKeepThisPageOpenWhileReconciling =>
      'Outcome unknown. Keep this page open while reconciling.';

  @override
  String get onlineSignInOrContinueAsAGuest =>
      'Sign in or continue as a guest.';

  @override
  String get onlineTheTableChangedReviewYourSelectionBeforeSubmittingAgain =>
      'The table changed. Review your selection before submitting again.';

  @override
  String get onlineYourActionMayHaveSucceededCheckTheOriginalOperation =>
      'Your action may have succeeded. Check the original operation before continuing.';

  @override
  String get onlineThisActionIsNotPermittedWithTheseChoicesReview =>
      'This action is not permitted with these choices. Review the rules and current turn.';

  @override
  String get onlineThisSessionCannotAccessThatRoomOrMatch =>
      'This session cannot access that room or match.';

  @override
  String get onlineTheRoomCouldNotAcceptThisRequestCheckIts =>
      'The room could not accept this request. Check its seats and invitation.';

  @override
  String get onlineConnectionUnavailableRetryWhenTheAuthorityIsReachable =>
      'Connection unavailable. Retry when the authority is reachable.';

  @override
  String get onlineCreateAThreeOrFourPlayerRoomMatchLength =>
      'Create a three- or four-player room. Match length is fixed before play.';

  @override
  String get onlineEnterMatch => 'Enter match';

  @override
  String get onlineTheGameOrDecisionChangedClearThisDraftBefore =>
      'The game or decision changed. Clear this draft before choosing a new action.';

  @override
  String get onlineTheAuthorityValidatesYourChoicesAnOfferReservesNo =>
      'The authority validates your choices. An offer reserves no cards; acceptance begins ordered responses.';

  @override
  String get onlineChooseAPositiveQuantityPaymentResolvesAfterResponsesExcess =>
      'Choose a positive quantity. Payment resolves after responses; excess is not returned as change.';

  @override
  String get onlineSacrificedQueen => 'Sacrificed Queen';

  @override
  String get onlineNone => 'None';

  @override
  String get onlineRecipientSeat => 'Recipient seat';

  @override
  String get onlineProtectedFormation => 'Protected formation';

  @override
  String get onlineAmountNumerator => 'Amount numerator';

  @override
  String get onlineDenominator => 'Denominator';

  @override
  String get onlineAcceptPromise => 'Accept promise';

  @override
  String get onlineAcceptFinalOffer => 'Accept final offer';

  @override
  String get onlineRejectFinalOffer => 'Reject final offer';

  @override
  String get onlinePayPromise => 'Pay promise';

  @override
  String get onlineRefusePromise => 'Refuse promise';

  @override
  String get onlineProposeFinalAmount => 'Propose final amount';

  @override
  String get onlineForgiveDebt => 'Forgive debt';

  @override
  String get onlineYouOwe => 'You owe';

  @override
  String get onlineOwedToYou => 'Owed to you';

  @override
  String get onlineYouGive => 'You give';

  @override
  String get onlineYouReceive => 'You receive';

  @override
  String get onlineUsedThisTurn => 'used this turn';

  @override
  String get onlineUsedThisRound => 'used this round';

  @override
  String get onlineSelectCardsOnTableOrHand => 'select cards on table or hand';

  @override
  String get onlineConditionTheNamedAwardOccursForYouInThis =>
      'Condition: the named award occurs for you in this game.';

  @override
  String onlineRoomIdentity(String value1) {
    return 'Room $value1';
  }

  @override
  String onlineRoomConfiguration(String value1, String value2) {
    return '$value1 games · $value2 seats';
  }

  @override
  String onlineSeatNumber(String value1) {
    return 'Seat $value1';
  }

  @override
  String onlineGameProgress(String value1, String value2) {
    return 'Game $value1 of $value2';
  }

  @override
  String onlineOwnMatchScore(String value1) {
    return 'Your match score: $value1';
  }

  @override
  String onlineSubmitAction(String value1) {
    return 'Submit $value1';
  }

  @override
  String onlineOriginGame(String value1) {
    return 'Origin game $value1';
  }

  @override
  String onlineRanksBySeat(String value1) {
    return 'Match ranks by seat: $value1';
  }

  @override
  String onlinePrivateOfferStatus(String value1, String value2) {
    return 'Private offer · revision $value1 · $value2';
  }

  @override
  String get onlineConnectionLost =>
      'Connection lost · last confirmed table · actions paused';

  @override
  String get onlineUnrevealedCard => 'Unrevealed card';

  @override
  String onlineDecisionOwner(String value1, String value2, String value3) {
    return 'Decision owner: seat $value1 · $value2 · $value3';
  }

  @override
  String onlineEnding(String ending) {
    return 'Ending: $ending';
  }

  @override
  String onlineEndingDeclarer(String value1, String value2) {
    return 'Ending: $value1 · Declarer: seat $value2';
  }

  @override
  String onlinePurchaseCalculation(
    String value1,
    String value2,
    String value3,
    String value4,
    String value5,
  ) {
    return 'Unit price $value1 · Quantity $value2 · Effective payment $value3 · Cost $value4 · Excess $value5';
  }

  @override
  String onlineCombatCalculation(
    String value1,
    String value2,
    String value3,
    String value4,
    String value5,
    String value6,
    String value7,
  ) {
    return 'Attack $value1 + $value2 = $value3 · Defense $value4 + $value5 = $value6 · $value7';
  }

  @override
  String onlinePromiseParties(String value1, String value2, String value3) {
    return 'Promise from seat $value1 to seat $value2 · $value3';
  }

  @override
  String onlinePromiseTerms(String value1, String value2, String value3) {
    return '$value1 $value2 · Award $value3';
  }

  @override
  String onlinePromiseSettlement(String value1, String value2, String value3) {
    return 'Promised $value1 · Paid $value2 · Final offer $value3';
  }

  @override
  String onlineDebtParties(
    String value1,
    String value2,
    String value3,
    String value4,
  ) {
    return '$value1 $value2 · Debtor seat $value3 · Creditor seat $value4';
  }

  @override
  String onlineOfferParties(String value1, String value2, String value3) {
    return 'From seat $value1 to seat $value2 · $value3';
  }

  @override
  String onlineLastingEffect(
    String value1,
    String value2,
    String value3,
    String value4,
    String value5,
  ) {
    return '$value1 · owner $value2 · target $value3 · custodian $value4 · $value5';
  }

  @override
  String onlineCompensationDraw(String value1, String value2) {
    return 'Compensation draw · seat $value1 · $value2 remaining';
  }

  @override
  String onlineMatchCorrection(String value1, String value2) {
    return 'Nonspendable match correction · $value1 · origin $value2';
  }

  @override
  String onlineKnownOfferCard(String value1, String value2, String value3) {
    return '$value1 of $value2 · offer card $value3';
  }

  @override
  String get onlineMatchComplete => 'Match complete';

  @override
  String get onlineAvailable => 'available';

  @override
  String get onlineUsed => 'used';

  @override
  String get onlineHeld => 'held';

  @override
  String get onlineAbsent => 'none';

  @override
  String get onlineChoose => 'choose';

  @override
  String onlineOpeningAceStatus(String opening, String ace) {
    return 'Opening: $opening · Ace: $ace';
  }

  @override
  String onlinePrivilegesStatus(String compensation, String underground) {
    return 'Compensation: $compensation · Underground privilege: $underground';
  }

  @override
  String onlineSelectedCards(String label, String count, String instruction) {
    return '$label: $count cards · $instruction';
  }

  @override
  String onlineTransferCards(String label, String count) {
    return '$label: $count · select visible cards';
  }

  @override
  String onlineFormationCards(String count) {
    return 'Formation cards: $count';
  }

  @override
  String onlineSubstituteKings(String count) {
    return 'Substitute kings: $count of 2';
  }

  @override
  String onlineCompletedGame(String scores) {
    return 'Completed game · $scores';
  }

  @override
  String onlineSeatScore(String seat, String score) {
    return 'Seat $seat: $score';
  }

  @override
  String get onlineJack => 'Jack';

  @override
  String get onlineQueen => 'Queen';

  @override
  String get onlineKing => 'King';

  @override
  String get onlineRoomOwnerStarts =>
      'The room owner starts the match when everyone has joined.';

  @override
  String onlineRoomWaiting(String count) {
    return 'Waiting for $count more players.';
  }

  @override
  String get onlineTimingOwnTurn => 'This action is available on your turn.';

  @override
  String get onlineTimingPending =>
      'Wait for the current response or decision before an ordinary action.';

  @override
  String onlineTimingResponder(String seat) {
    return 'Seat $seat must respond first.';
  }

  @override
  String get onlineTimingEffect =>
      'Complete the current effect decision; this is not a response window.';

  @override
  String get onlineTimingNoResponse =>
      'There is no response window for this action.';

  @override
  String get onlineTimingNoDecision =>
      'There is no effect decision for you to complete.';

  @override
  String get economyTitle => 'Allowance and dirt';

  @override
  String get economyDisabled =>
      'Online allowances are not enabled on this server.';

  @override
  String get economyUnavailable =>
      'Could not refresh your allowance. Try again.';

  @override
  String get economyRefresh => 'Refresh allowance';

  @override
  String economyDirt(int amount) {
    return '$amount dirt';
  }

  @override
  String economyStarts(int count) {
    return '$count free online starts remaining today';
  }

  @override
  String economyReset(String time) {
    return 'Free starts reset at $time UTC.';
  }

  @override
  String economyReward(int amount) {
    return 'Earn $amount dirt once per finalized online game. Offline play earns no dirt; unfinished games do not refund starts.';
  }

  @override
  String get economyUnlimited => 'Unlimited online starts';

  @override
  String economyPassUntil(String time) {
    return 'Pass expiry: $time UTC';
  }

  @override
  String economyPremiumUntil(String time) {
    return 'Premium expiry: $time UTC';
  }

  @override
  String get economyDays => 'Earned access';

  @override
  String economyDaysValue(int days) {
    return '$days day(s)';
  }

  @override
  String economySpend(int cost) {
    return 'Spend $cost dirt';
  }

  @override
  String get economyPassHelp =>
      'Earned Daily and Weekly access allow unlimited online starts and keep ads unless separate ad-free access is active. Time extends from now or your current pass expiry, whichever is later.';

  @override
  String get economyDaily => 'Daily · 1 day';

  @override
  String get economyWeekly => 'Weekly · 7 days';

  @override
  String get economyNoAds => 'Ad-free access active';

  @override
  String economyAdFreeUntil(String time) {
    return 'Ad-free expiry: $time UTC';
  }

  @override
  String get economyAdsActive => 'Ads remain active';

  @override
  String get economyProductUnavailable =>
      'That pass is no longer offered. No dirt was spent. Choose Daily or Weekly access.';

  @override
  String get economyPaidUnavailable =>
      'Paid Ad-free, Daily, Weekly, Monthly and Yearly products are not available in this build.';

  @override
  String get economyAdFreeTerms =>
      'Ad-free: 30 days without ads. The 3 free online starts per day still apply unless you also have unlimited access.';

  @override
  String get economyPaidDailyTerms =>
      'Paid Daily: 24 hours of unlimited online starts and no ads from confirmed purchase. Does not renew automatically.';

  @override
  String get economyPaidLongerTerms =>
      'Paid Weekly, Monthly and Yearly: unlimited online starts and no ads.';

  @override
  String get economyUnknown =>
      'Your pass purchase has an unknown result. Recover it before buying another pass.';

  @override
  String get economyRecover => 'Recover pass purchase';

  @override
  String get economyInsufficient => 'Not enough dirt for this pass.';

  @override
  String economyConfirmed(int cost, String time) {
    return 'Pass confirmed · $cost dirt · expires $time UTC';
  }

  @override
  String get economyRecoveryUnavailable =>
      'Recovery storage is unavailable. Restore access before trying another purchase.';

  @override
  String get economyAdmissionWaiting =>
      'The next game is waiting for every player’s online access. Check your allowance; a player may need a pass, the UTC reset, or account access.';

  @override
  String get onlineNextGamePending =>
      'Game settled. Waiting for the next game.';

  @override
  String onlineCopyCode(String label) {
    return 'Copy $label';
  }

  @override
  String onlineCodeCopied(String label) {
    return '$label copied';
  }

  @override
  String onlineCodeValue(String label, String value) {
    return '$label: $value';
  }

  @override
  String get onlineClipboardUnavailable =>
      'Clipboard access is unavailable. The code remains visible; check your browser permissions and try Copy again.';

  @override
  String get onlineClubUsedThisTurn => 'Used this turn';

  @override
  String get onlineDropCopyLegend => 'Copy numbers in brackets.';

  @override
  String get onlineDropCostsOpeningNew => 'Uses this turn’s opening allowance.';

  @override
  String get onlineDropCostsOpeningExisting => 'No opening allowance spent.';

  @override
  String get onlineDropCostsFormation =>
      'Opening allowance unless earned Underground. No cards consumed.';

  @override
  String get onlineDropCostsAttach =>
      'No opening allowance. Royal needs its supported series.';

  @override
  String get onlineDropCostsRicher =>
      'Round allowance. On resolution discard King; draw up to two.';

  @override
  String get onlineDropCostsRingleader =>
      'Once per game. On resolution discard King; gain 10.';

  @override
  String get onlineDropCostsConfinement =>
      'Shared Ace round allowance. Discard Ace after resolution.';

  @override
  String get onlineDropCostsCompensation =>
      'Shared Ace round and game allowance. Ace becomes persistent effect.';

  @override
  String get onlineDropCostsInflation =>
      'Shared Ace round allowance. Ace becomes persistent effect; prevention discards it.';

  @override
  String get onlineDropCostsCombatResponse =>
      'Shared Ace round allowance. Reveal/discard Ace after combat, even if canceled.';

  @override
  String get onlineDropCostsPurchase =>
      'On resolution return exact payment and draw chosen quantity; canceled costs nothing.';

  @override
  String get onlineDropCostsReturnLoan =>
      'Return these borrowed cards. Alliance and earned privileges remain.';

  @override
  String get onlineDropCostsFate =>
      'Turn-cycle allowance. Return chosen King before reset shuffle.';

  @override
  String get onlineDropCostsJustice =>
      'Turn-cycle allowance. Return chosen Queen after resolution.';

  @override
  String get onlineDropCostsKidnapper =>
      'Own-turn allowance. Special Doppelganger theft returns all members on resolution.';

  @override
  String get onlineDropCostsPonzi =>
      'Turn-cycle allowance. Return all four Jacks after resolution.';

  @override
  String get onlineDropCostsDexter =>
      'Shared round allowance. On resolution return Diamond; discard target royal.';

  @override
  String get onlineDropCostsBarricade =>
      'Turn-cycle allowance. On resolution discard Queen; return five; draw up to five restricted cards.';

  @override
  String get onlineDropCostsNegotiator =>
      'Round allowance. Final commitment uses added Clubs, transfers Spades, and returns Jack.';

  @override
  String get onlineDropCostsTakeBack =>
      'Return formation cards; ability allowances remain spent.';

  @override
  String get onlineDropCostsAttack =>
      'Attacking Clubs become used at declaration, even if canceled.';

  @override
  String get onlineDropCostsBaron =>
      'Baron round allowance; no Baron card consumed.';

  @override
  String get onlineDropCostsBomb =>
      'On success discard Bomb; return Clubs only for ordinary Bomb attack.';

  @override
  String get onlineDropCostsInfiltrator =>
      'Return Jack at declaration; spend turn-cycle allowance.';

  @override
  String get onlineDropCostsExile =>
      'Spend Exile round allowance; no card consumed.';

  @override
  String get onlineDropCostsCombatAce =>
      'Shared Ace round allowance; reveal/discard Ace after combat, even if canceled.';

  @override
  String get onlineCostsTitle => 'Costs and timing';

  @override
  String get onlineCostsGeneral =>
      'Illegal declarations spend nothing. Accepted allowances stay spent after cancellation; support is checked again before unresolved effects commit.';

  @override
  String get onlineCostsAttack =>
      'Own turn; choose unused exposed Clubs and a legal target. Physical Clubs become used at legal declaration, including if the attack is canceled. Exchange costs apply only after successful resolution.';

  @override
  String get onlineCostsCombatAce =>
      'If using a numerical or Advancement Ace, its shared round allowance is spent at declaration and the Ace is revealed and discarded after combat, including cancellation.';

  @override
  String get onlineCostsBaron =>
      'Baron also spends its round allowance at declaration. No Baron card is consumed; cancellation keeps the round allowance and Club usage spent.';

  @override
  String get onlineCostsBomb =>
      'Successful ordinary Bomb destruction returns the attacking Clubs and discards the Bomb. Baron discards the Bomb without returning Clubs. Cancellation removes neither.';

  @override
  String get onlineCostsInfiltrator =>
      'Return the Infiltrator jack at legal declaration; cancellation does not return it or its turn-cycle allowance.';

  @override
  String get onlineCostsExile =>
      'Exile spends its round allowance at the accepted Barricade bypass. It has no card cost; cancellation does not refund the allowance.';

  @override
  String get onlineCostsOpening =>
      'Own turn: normally open one new number series or open/reconfigure one combo. Earned Underground permits unlimited combo openings/reconfigurations; it does not waive Ace or attack prerequisites.';

  @override
  String get onlineCostsFormation =>
      'Opening moves the selected cards into a formation; it consumes no cards. Passive abilities need their support. Doppelganger pairs and replacement slots stay bound when support is lost.';

  @override
  String get onlineCostsTakeBack =>
      'Own turn: combos may be taken back. Number series and attachments cannot be taken back. Taking back cards never refreshes an ability allowance.';

  @override
  String get onlineCostsAttach =>
      'Own turn: add to an existing series or restore a historically opened suit without a new opening allowance. Attached royals need their supporting number series to function.';

  @override
  String get onlineCostsFate =>
      'Own turn; turn-cycle allowance spent at activation. Return one chosen Fate king immediately before the reset shuffle. Cancellation before resolution keeps the king but spends the allowance.';

  @override
  String get onlineCostsCoup =>
      'Any eligible boundary, including pending decisions and while confined. Requires historical Clubs, both physical Hearts queens and both physical Diamonds queens, all unrestricted, and no alliance. No allowance or consumable cost; no response window. Earlier committed effects remain.';

  @override
  String get onlineCostsJustice =>
      'Own turn; turn-cycle allowance spent at activation. Return one chosen Justice queen after the selection sequence, even with zero captures. Cancellation before resolution keeps the queen but spends the allowance.';

  @override
  String get onlineCostsCode =>
      'Own turn; turn-cycle allowance spent at activation, with no card payment. Resolution replaces the governing Code and settings. Cancellation keeps the old Code and spends the allowance.';

  @override
  String get onlineCostsKidnapper =>
      'Own turn; one activation per own turn, committed at acceptance. Normal theft consumes no combo cards. Special Doppelganger theft returns the entire Kidnapper formation on resolution. Cancellation keeps those cards but spends the allowance.';

  @override
  String get onlineCostsPonzi =>
      'Any idle boundary; turn-cycle allowance spent at activation, with no card payment yet. Return all four jacks after resolution, even when nothing is stolen. Cancellation keeps the jacks but spends the turn-cycle allowance.';

  @override
  String get onlineCostsRingleader =>
      'Own turn; once per game, spent at activation. Discard the supported king and award 10 together on resolution. Cancellation keeps the king but spends the game allowance.';

  @override
  String get onlineCostsRicher =>
      'Own turn; round allowance spent at activation. Discard the supported king and draw up to two on resolution. Cancellation keeps the king but spends the allowance.';

  @override
  String get onlineCostsBarricade =>
      'Own turn; turn-cycle allowance spent at activation. On resolution discard the supported queen, return five other selected cards and draw up to five; those draws are restricted until next round. Cancellation keeps all selected cards.';

  @override
  String get onlineCostsNegotiator =>
      'Incoming club-attack response; round allowance spent only on a legal activation. At the final decision, added Clubs become used, selected Spades transfer, the jack returns and the attack is canceled together. Abort before commitment keeps those costs unpaid; the accepted allowance and original Club usage stay spent.';

  @override
  String get onlineCostsDexter =>
      'Own turn; supported Dexter and prepayment Diamond total at least 20. Spends the round allowance shared with prevention. Return the selected number Diamond and discard the target royal together on resolution. Cancellation keeps both cards but spends the allowance.';

  @override
  String get onlineCostsDexterPrevention =>
      'Inside hostile Ace resolution; supported Dexter and prepayment Diamond total 1–10. A legal payable prevention spends the shared round allowance. Return the chosen Diamond and prevent the effect together; the hostile Ace is discarded and its allowance stays spent.';

  @override
  String get onlineCostsConfinement =>
      'Authorized response to the pending actor; spends the shared Ace round allowance at acceptance. Discard the Ace when resolution finishes, including Dexter prevention. Coup cannot be stopped.';

  @override
  String get onlineCostsCombatResponse =>
      'Incoming-combat response; spends the shared Ace round allowance and commits its modifier. Reveal and discard the Ace after combat resolves or is canceled. Dexter prevention removes the modifier, not its usage or discard.';

  @override
  String get onlineCostsInflation =>
      'Own main action or authorized incoming-attack response; spends the shared Ace round allowance. A duplicate active target is illegal and spends nothing. Resolution puts the Ace in its public effect zone; Dexter prevention discards it. Earlier cancellation keeps the Ace but spends the allowance.';

  @override
  String get onlineCostsCompensation =>
      'Own main action or authorized incoming-attack response before losses; spends both the shared Ace round allowance and game activation at acceptance. Resolution puts the Ace in its public effect zone. Cancellation keeps the Ace but spends both allowances.';

  @override
  String get onlineCostsPurchase =>
      'Own idle boundary; no separate purchase allowance and no card payment at acceptance. Revalidate exact payment, quantity and supply after responses, then return, shuffle and draw together. Cancellation keeps every payment card and draws nothing.';

  @override
  String get onlineCostsOffer =>
      'A proposal reserves no cards or allowances and does not block play. Acceptance names its exact revision in the originating turn. Transfers and any loan alliance happen together only after successful resolution; an accepted Ace transfer spends its supplier’s round allowance even if canceled.';

  @override
  String get onlineCostsReturnLoan =>
      'The recipient voluntarily returns only borrowed cards they still control. An intact formation returns intact; surviving individual cards return exposed and unassigned without rebuilding it. The game-long alliance and previously earned privileges remain.';

  @override
  String get onlineCostsDecision =>
      'Complete only the current authorized effect input. Earlier committed costs and allowances remain spent; unresolved payments and movement commit only with the completed effect.';

  @override
  String get onlineCostsPass =>
      'Passing uses this response opportunity without playing a card. Navigation, disconnection and silence never submit a pass.';

  @override
  String get onlineWelcomeTitle => 'A table worth turning.';

  @override
  String get onlineWelcomeBody =>
      'Build your hand. Make your move. Every deal changes the table.';

  @override
  String get onlineWelcomeCards => 'Illustrated CGMS cards';

  @override
  String get onlineEntryTitle => 'Take your seat';

  @override
  String get onlineEntryBody =>
      'Continue as a guest, or sign in to your account.';

  @override
  String get onlineAccountSignIn => 'Sign in or register';

  @override
  String get onlineYourRoom => 'Your room';

  @override
  String get onlineRoomJoined => 'Joined';

  @override
  String get onlineRoomEmptySeat => 'Waiting for a player';

  @override
  String get onlineSetupTitle => 'Gather your table';

  @override
  String get onlineSetupBody =>
      'Play against bots, create a room for your group, or join with a private invitation.';

  @override
  String get onlineJoinHelp =>
      'Use the room code and invitation shared by the room owner.';

  @override
  String get onlineActionFamilyPlay => 'Open & build';

  @override
  String get onlineActionFamilyAttack => 'Attack & defend';

  @override
  String get onlineActionFamilyAbilities => 'Abilities';

  @override
  String get onlineActionFamilyDeals => 'Trades & loans';

  @override
  String get onlineActionFamilyFinance => 'Points & settlement';

  @override
  String get onlineActionFamilyTurn => 'Turn & decisions';

  @override
  String get onlineActionPrepare => 'Prepare your action';

  @override
  String get onlineReconnectHelp =>
      'Your last confirmed table is still visible. Reconnecting automatically; no action is passed or resent.';

  @override
  String onlineSessionConnecting(int attempt, int limit) {
    return 'Connecting to the table… attempt $attempt of $limit';
  }

  @override
  String get onlineRecoveryExhausted =>
      'Connection is still unavailable after bounded recovery. Use Refresh connection to try again.';

  @override
  String get onlineInvalidResponse =>
      'The server returned an unreadable response. Use Refresh connection to try again.';

  @override
  String get onlineReconnectStopped =>
      'Your last confirmed table is still visible. Use Refresh connection to retry; no action is passed or resent.';

  @override
  String get onlinePlayAgainstBots => 'Play against bots';

  @override
  String get onlineBotDifficulty => 'Bot difficulty';

  @override
  String get onlineBotBeginner => 'Beginner';

  @override
  String get onlineBotStandard => 'Standard';

  @override
  String get onlineBotAdvanced => 'Advanced';

  @override
  String get onlineBot => 'Bot';

  @override
  String get onlineServerProgress => 'Preparing the next move…';

  @override
  String get onlineBotSetupHelp =>
      'Bots fill the other seats. Choose a difficulty, then start the match from your room.';

  @override
  String onlineBotSeat(String difficulty) {
    return 'Bot · $difficulty';
  }

  @override
  String get onlineRoomGamesRange =>
      'Enter a whole number of games from 1 to 100.';

  @override
  String get onlineRoundDeparture => 'Round complete · Choose whether to stay';

  @override
  String get onlineRoundWaiting => 'Round complete · Waiting for players';

  @override
  String get onlineRoundChoiceHelp =>
      'Choose whether to stay for the next round. Leaving ends this game for you and removes positive current-game score and available points; debts remain. You can rejoin the next game.';

  @override
  String get onlineDepartureRecorded =>
      'Your choice is saved. Waiting for other players.';

  @override
  String get onlineStayInGame => 'Stay in game';

  @override
  String get onlineLeaveGame => 'Leave game';

  @override
  String get activateCard => 'Card actions';

  @override
  String get cardRoleSource => 'Source card';

  @override
  String get cardRoleTarget => 'Target card';

  @override
  String get cardRolePayment => 'Payment card';

  @override
  String get cardRoleModifier => 'Modifier card';

  @override
  String get cardGestureHint =>
      'Tap or Space to select. Double tap or Enter for card actions. I to inspect.';

  @override
  String get chooseDestination => 'Choose destination';

  @override
  String get inspectPublicCards => 'Inspect public cards';

  @override
  String get inspectSelectedCards => 'Inspect selected cards';

  @override
  String get onlineCardUnavailable => 'Card no longer available';

  @override
  String get onlineGameOptions => 'Game options';

  @override
  String get onlineRecords => 'Offers, promises & scores';

  @override
  String get onlineCardGestures => 'Card gestures';

  @override
  String get onlinePendingDecision => 'Pending decision';

  @override
  String get onlineResultsSettlement => 'Results & settlement';

  @override
  String get onlineCardInvisible => 'Card no longer visible';

  @override
  String get onlineChooseMove => 'Choose a move';

  @override
  String get onlineCloseCardContext => 'Close card context';

  @override
  String get onlineNoMoveDestination =>
      'No move at this destination. Select other cards or a destination on the board.';

  @override
  String get onlineCoupCardConfirmation =>
      'Declare Coup with both Hearts Queens and both Diamond Queens.';

  @override
  String get onlineChooseBoard => 'Select cards or a destination';

  @override
  String get onlineNoCardResponse =>
      'No card response is available. Pass to continue.';

  @override
  String get onlineCardResponseHelp =>
      'Double tap a matching card to respond, or choose Pass response.';

  @override
  String get onlineKidnapperExposedChoice =>
      'Choose an exposed royal for Kidnapper';

  @override
  String get onlineResolveAuthorizedStage =>
      'Resolve only this authorized decision stage.';

  @override
  String get onlineContinueWithoutCard => 'Continue without spending a card';

  @override
  String get onlineJusticeRecordedOutcome =>
      'Request the recorded concealed outcome. No hidden candidate is selected.';

  @override
  String get onlineJusticeConcealedChoice => 'Choose concealed Justice outcome';

  @override
  String get onlineConsumedQueen => 'Consumed Queen';

  @override
  String get onlineModifier => 'Modifier';

  @override
  String get onlineDiamondPayment => 'Diamond payment';

  @override
  String get onlineBarricadeOrder => 'Queen first, then five returned cards';

  @override
  String get onlineDoppelgangerPair => 'Doppelganger pair';

  @override
  String get onlineComboMembers => 'Combo members';

  @override
  String get onlineDoppelgangerFormation => 'Doppelganger formation';

  @override
  String get onlineReplacementRank => 'Replacement rank';

  @override
  String get onlineReplacementSuit => 'Replacement suit';

  @override
  String get onlineChoosePublicTarget => 'Choose your public target';

  @override
  String get onlineInfiltratorCommit =>
      'Infiltrator · return the attached Jack at declaration';

  @override
  String get onlineExileCommit => 'Exile · spend this round’s bypass';

  @override
  String get onlinePrivateTrade => 'Private card trade';

  @override
  String get onlineProposalRevision => 'Proposal revision';

  @override
  String get onlineNewOffer => 'New offer';

  @override
  String get onlinePrivateOfferTerms =>
      'Only named parties see these terms. Offering reserves no cards. Acceptance uses this exact revision.';

  @override
  String get onlineVictoryThreshold => 'Victory threshold';

  @override
  String get onlineNumberDiamondThreshold => '13 number Diamonds';

  @override
  String get onlineRoyalThreshold => '15 Royals';

  @override
  String get onlineAceCombatValue => 'Ace combat value';

  @override
  String get onlineAdvancementValue => 'Advancement · +10';

  @override
  String get onlineDiamondBonusRange => 'Diamond bonus (1–10)';

  @override
  String get onlinePurchasePriceRange => 'Purchase price (5–10)';

  @override
  String get onlineJackRank => 'Jack';

  @override
  String get onlineQueenRank => 'Queen';

  @override
  String get onlineKingRank => 'King';

  @override
  String get onlineCardGestureHelp =>
      'Tap to select exact cards; tap additional cards to build a bundle. Double tap or use keyboard activation for the selected card action. Drag to preview the exact move, costs and destination; release a complete legal move to submit it once. Choices appear only for missing parameters or ambiguity. Inspect cards independently. Escape cancels an unsubmitted selection.';

  @override
  String onlineReviewBeforeConfirm(String action) {
    return 'Review $action before confirming.';
  }

  @override
  String onlineChooseRoleBoard(String role) {
    return 'Choose $role';
  }

  @override
  String onlineResolveKind(String kind) {
    return 'Choose how to resolve $kind';
  }

  @override
  String onlineSourceCards(String cards) {
    return 'Source cards · $cards';
  }

  @override
  String onlineFormationSummary(String formation) {
    return 'Formation · $formation';
  }

  @override
  String onlineConfirmAction(String action) {
    return 'Confirm $action';
  }

  @override
  String onlineSeriesSummary(String suit) {
    return 'Series · $suit';
  }

  @override
  String onlineComboSummary(String combo) {
    return 'Combo · $combo';
  }

  @override
  String onlineProtectionSummary(String target) {
    return 'Protection · $target';
  }

  @override
  String onlineAttachTo(String suit) {
    return 'Attach to $suit';
  }

  @override
  String onlineOfferReplacement(String seat, String revision) {
    return 'Replace offer to seat $seat · revision $revision';
  }

  @override
  String onlineCommittedClubs(String cards) {
    return 'Committed Clubs · $cards';
  }

  @override
  String onlineFrozenTargets(String cards) {
    return 'Frozen targets · $cards';
  }

  @override
  String onlineCombatSeats(String attacker, String defender) {
    return 'Seat $attacker attacks seat $defender';
  }

  @override
  String onlinePendingSeatAction(String actor, String action) {
    return 'Seat $actor · $action';
  }

  @override
  String onlinePendingSeatTarget(String actor, String action, String target) {
    return 'Seat $actor · $action → Seat $target';
  }

  @override
  String onlineCardRoleSummary(String role, String cards) {
    return '$role · $cards';
  }

  @override
  String get onlineChooseSupportedSeries => 'Choose a supported series';

  @override
  String get onlineChooseTargetCard => 'Choose the target card';

  @override
  String get onlineChooseDiamondPayment => 'Select a Diamond payment';

  @override
  String get onlineChooseAttackingClubs => 'Select attacking Clubs';

  @override
  String get onlineChooseAce => 'Select an Ace modifier';

  @override
  String get onlineChooseProtection => 'Choose a protection target';

  @override
  String get onlineChoosePayment => 'Select Spade payment cards';

  @override
  String get onlineChooseFormationMembers => 'Select exact combo members';

  @override
  String get onlineChooseKings => 'Select the two Doppelganger kings';

  @override
  String get onlineChooseRequestedCards => 'Choose requested cards';

  @override
  String onlineReleaseMove(String action) {
    return 'Release to submit $action';
  }

  @override
  String onlineReleaseChoice(String action) {
    return 'Release to choose $action';
  }

  @override
  String onlineDropSeat(int seat) {
    return 'Seat $seat';
  }

  @override
  String get onlineLoanDropTerms =>
      'No cards requested. Exact revision acceptance and successful responses resolve the loan and exclusive alliance together.';

  @override
  String get turnDrawToHand => 'Turn draw: one card added to your hand.';

  @override
  String get turnDrawToConcealedAce =>
      'Turn draw: one card added to your concealed Ace area.';

  @override
  String turnDrawBySeat(int seat) {
    return 'Seat $seat drew one card.';
  }
}
