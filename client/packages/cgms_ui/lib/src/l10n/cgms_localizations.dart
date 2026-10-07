import 'dart:async';

import 'package:flutter/foundation.dart';
import 'package:flutter/widgets.dart';
import 'package:flutter_localizations/flutter_localizations.dart';
import 'package:intl/intl.dart' as intl;

import 'cgms_localizations_en.dart';

// ignore_for_file: type=lint

/// Callers can lookup localized strings with an instance of CgmsLocalizations
/// returned by `CgmsLocalizations.of(context)`.
///
/// Applications need to include `CgmsLocalizations.delegate()` in their app's
/// `localizationDelegates` list, and the locales they support in the app's
/// `supportedLocales` list. For example:
///
/// ```dart
/// import 'l10n/cgms_localizations.dart';
///
/// return MaterialApp(
///   localizationsDelegates: CgmsLocalizations.localizationsDelegates,
///   supportedLocales: CgmsLocalizations.supportedLocales,
///   home: MyApplicationHome(),
/// );
/// ```
///
/// ## Update pubspec.yaml
///
/// Please make sure to update your pubspec.yaml to include the following
/// packages:
///
/// ```yaml
/// dependencies:
///   # Internationalization support.
///   flutter_localizations:
///     sdk: flutter
///   intl: any # Use the pinned version from flutter_localizations
///
///   # Rest of dependencies
/// ```
///
/// ## iOS Applications
///
/// iOS applications define key application metadata, including supported
/// locales, in an Info.plist file that is built into the application bundle.
/// To configure the locales supported by your app, you’ll need to edit this
/// file.
///
/// First, open your project’s ios/Runner.xcworkspace Xcode workspace file.
/// Then, in the Project Navigator, open the Info.plist file under the Runner
/// project’s Runner folder.
///
/// Next, select the Information Property List item, select Add Item from the
/// Editor menu, then select Localizations from the pop-up menu.
///
/// Select and expand the newly-created Localizations item then, for each
/// locale your application supports, add a new item and select the locale
/// you wish to add from the pop-up menu in the Value field. This list should
/// be consistent with the languages listed in the CgmsLocalizations.supportedLocales
/// property.
abstract class CgmsLocalizations {
  CgmsLocalizations(String locale)
    : localeName = intl.Intl.canonicalizedLocale(locale.toString());

  final String localeName;

  static CgmsLocalizations of(BuildContext context) {
    return Localizations.of<CgmsLocalizations>(context, CgmsLocalizations)!;
  }

  static const LocalizationsDelegate<CgmsLocalizations> delegate =
      _CgmsLocalizationsDelegate();

  /// A list of this localizations delegate along with the default localizations
  /// delegates.
  ///
  /// Returns a list of localizations delegates containing this delegate along with
  /// GlobalMaterialLocalizations.delegate, GlobalCupertinoLocalizations.delegate,
  /// and GlobalWidgetsLocalizations.delegate.
  ///
  /// Additional delegates can be added by appending to this list in
  /// MaterialApp. This list does not have to be used at all if a custom list
  /// of delegates is preferred or required.
  static const List<LocalizationsDelegate<dynamic>> localizationsDelegates =
      <LocalizationsDelegate<dynamic>>[
        delegate,
        GlobalMaterialLocalizations.delegate,
        GlobalCupertinoLocalizations.delegate,
        GlobalWidgetsLocalizations.delegate,
      ];

  /// A list of this localizations delegate's supported locales.
  static const List<Locale> supportedLocales = <Locale>[Locale('en')];

  /// No description provided for @concealedCard.
  ///
  /// In en, this message translates to:
  /// **'Concealed card'**
  String get concealedCard;

  /// No description provided for @drawPile.
  ///
  /// In en, this message translates to:
  /// **'Draw pile'**
  String get drawPile;

  /// No description provided for @magnifyCard.
  ///
  /// In en, this message translates to:
  /// **'Magnify card'**
  String get magnifyCard;

  /// No description provided for @inspectCard.
  ///
  /// In en, this message translates to:
  /// **'Inspect {card}'**
  String inspectCard(String card);

  /// No description provided for @copyNumber.
  ///
  /// In en, this message translates to:
  /// **'Copy {number}'**
  String copyNumber(int number);

  /// No description provided for @cardIdentity.
  ///
  /// In en, this message translates to:
  /// **'{rank} of {suit}, copy {copy}'**
  String cardIdentity(String rank, String suit, int copy);

  /// No description provided for @selected.
  ///
  /// In en, this message translates to:
  /// **'Selected'**
  String get selected;

  /// No description provided for @hearts.
  ///
  /// In en, this message translates to:
  /// **'Hearts'**
  String get hearts;

  /// No description provided for @diamonds.
  ///
  /// In en, this message translates to:
  /// **'Diamonds'**
  String get diamonds;

  /// No description provided for @clubs.
  ///
  /// In en, this message translates to:
  /// **'Clubs'**
  String get clubs;

  /// No description provided for @spades.
  ///
  /// In en, this message translates to:
  /// **'Spades'**
  String get spades;

  /// No description provided for @startMatch.
  ///
  /// In en, this message translates to:
  /// **'Start match'**
  String get startMatch;

  /// No description provided for @nextGame.
  ///
  /// In en, this message translates to:
  /// **'Your next game'**
  String get nextGame;

  /// No description provided for @setupNotice.
  ///
  /// In en, this message translates to:
  /// **'Choose a display name and the number of games in your match.'**
  String get setupNotice;

  /// No description provided for @setupEyebrow.
  ///
  /// In en, this message translates to:
  /// **'A GAME OF PUBLIC POWER & PRIVATE DOUBT'**
  String get setupEyebrow;

  /// No description provided for @takeYourSeat.
  ///
  /// In en, this message translates to:
  /// **'TAKE YOUR SEAT.'**
  String get takeYourSeat;

  /// No description provided for @setupTagline.
  ///
  /// In en, this message translates to:
  /// **'Build influence. Read the table.\nLeave nothing to chance.'**
  String get setupTagline;

  /// No description provided for @editionName.
  ///
  /// In en, this message translates to:
  /// **'cgmsart / FIRST LIGHT'**
  String get editionName;

  /// No description provided for @setupFormTitle.
  ///
  /// In en, this message translates to:
  /// **'A place at the table'**
  String get setupFormTitle;

  /// No description provided for @setupDeckDescription.
  ///
  /// In en, this message translates to:
  /// **'Four seats. Two decks. Every physical card matters.'**
  String get setupDeckDescription;

  /// No description provided for @yourName.
  ///
  /// In en, this message translates to:
  /// **'Your name'**
  String get yourName;

  /// No description provided for @nameHint.
  ///
  /// In en, this message translates to:
  /// **'Enter a display name'**
  String get nameHint;

  /// No description provided for @nameRequired.
  ///
  /// In en, this message translates to:
  /// **'Enter a name to take your seat.'**
  String get nameRequired;

  /// No description provided for @matchGameCount.
  ///
  /// In en, this message translates to:
  /// **'Games in the match'**
  String get matchGameCount;

  /// No description provided for @oneGame.
  ///
  /// In en, this message translates to:
  /// **'1 game'**
  String get oneGame;

  /// No description provided for @threeGames.
  ///
  /// In en, this message translates to:
  /// **'3 games'**
  String get threeGames;

  /// No description provided for @yourMove.
  ///
  /// In en, this message translates to:
  /// **'YOUR MOVE.'**
  String get yourMove;

  /// No description provided for @atTheTable.
  ///
  /// In en, this message translates to:
  /// **'AT THE TABLE.'**
  String get atTheTable;

  /// No description provided for @loadingTable.
  ///
  /// In en, this message translates to:
  /// **'Loading the table'**
  String get loadingTable;

  /// No description provided for @tableLoadError.
  ///
  /// In en, this message translates to:
  /// **'The table could not be loaded'**
  String get tableLoadError;

  /// No description provided for @disconnectedActionsPaused.
  ///
  /// In en, this message translates to:
  /// **'Disconnected · actions paused'**
  String get disconnectedActionsPaused;

  /// No description provided for @connectionExplanation.
  ///
  /// In en, this message translates to:
  /// **'Your last known cards and pending decisions stay here. Actions remain paused until the connection is restored.'**
  String get connectionExplanation;

  /// No description provided for @retryConnection.
  ///
  /// In en, this message translates to:
  /// **'Retry connection'**
  String get retryConnection;

  /// No description provided for @restoreConnection.
  ///
  /// In en, this message translates to:
  /// **'Restore connection'**
  String get restoreConnection;

  /// No description provided for @publicTable.
  ///
  /// In en, this message translates to:
  /// **'The public table'**
  String get publicTable;

  /// No description provided for @readTheirPosition.
  ///
  /// In en, this message translates to:
  /// **'READ THEIR POSITION'**
  String get readTheirPosition;

  /// No description provided for @scoresAndDebt.
  ///
  /// In en, this message translates to:
  /// **'Scores & debt'**
  String get scoresAndDebt;

  /// No description provided for @yourPublicPosition.
  ///
  /// In en, this message translates to:
  /// **'YOUR PUBLIC POSITION'**
  String get yourPublicPosition;

  /// No description provided for @yourHand.
  ///
  /// In en, this message translates to:
  /// **'Your hand'**
  String get yourHand;

  /// No description provided for @privateHandEyebrow.
  ///
  /// In en, this message translates to:
  /// **'ONLY YOU CAN SEE THESE'**
  String get privateHandEyebrow;

  /// No description provided for @emptyHandTitle.
  ///
  /// In en, this message translates to:
  /// **'Your hand is empty'**
  String get emptyHandTitle;

  /// No description provided for @emptyHandMessage.
  ///
  /// In en, this message translates to:
  /// **'Public cards and scores remain available. There are no private cards to inspect.'**
  String get emptyHandMessage;

  /// No description provided for @tradeCards.
  ///
  /// In en, this message translates to:
  /// **'Trade cards'**
  String get tradeCards;

  /// No description provided for @endTurnAndScore.
  ///
  /// In en, this message translates to:
  /// **'End turn & score'**
  String get endTurnAndScore;

  /// No description provided for @finishResponsesFirst.
  ///
  /// In en, this message translates to:
  /// **'Finish the response window before ending your turn.'**
  String get finishResponsesFirst;

  /// No description provided for @tableJournal.
  ///
  /// In en, this message translates to:
  /// **'Table journal'**
  String get tableJournal;

  /// No description provided for @eventEyebrow.
  ///
  /// In en, this message translates to:
  /// **'YOUR VIEW OF EVENTS'**
  String get eventEyebrow;

  /// No description provided for @noEvents.
  ///
  /// In en, this message translates to:
  /// **'No events yet.'**
  String get noEvents;

  /// No description provided for @acceptOffer.
  ///
  /// In en, this message translates to:
  /// **'Accept offer'**
  String get acceptOffer;

  /// No description provided for @declineOffer.
  ///
  /// In en, this message translates to:
  /// **'Decline offer'**
  String get declineOffer;

  /// No description provided for @tradeEyebrow.
  ///
  /// In en, this message translates to:
  /// **'A VOLUNTARY EXCHANGE'**
  String get tradeEyebrow;

  /// No description provided for @makeADeal.
  ///
  /// In en, this message translates to:
  /// **'MAKE A DEAL.'**
  String get makeADeal;

  /// No description provided for @youGive.
  ///
  /// In en, this message translates to:
  /// **'YOU GIVE'**
  String get youGive;

  /// No description provided for @youReceive.
  ///
  /// In en, this message translates to:
  /// **'YOU RECEIVE'**
  String get youReceive;

  /// No description provided for @tradeExplanation.
  ///
  /// In en, this message translates to:
  /// **'Private offer terms · visible to the two parties.\nProposing reserves no cards. Acceptance opens responses; only resolution transfers the cards.'**
  String get tradeExplanation;

  /// No description provided for @proposeTrade.
  ///
  /// In en, this message translates to:
  /// **'Propose trade'**
  String get proposeTrade;

  /// No description provided for @withdrawOffer.
  ///
  /// In en, this message translates to:
  /// **'Withdraw offer'**
  String get withdrawOffer;

  /// No description provided for @resultRecorded.
  ///
  /// In en, this message translates to:
  /// **'THE RESULT IS RECORDED'**
  String get resultRecorded;

  /// No description provided for @pointsHaveHistory.
  ///
  /// In en, this message translates to:
  /// **'EVERY POINT HAS A HISTORY'**
  String get pointsHaveHistory;

  /// No description provided for @matchScore.
  ///
  /// In en, this message translates to:
  /// **'MATCH SCORE'**
  String get matchScore;

  /// No description provided for @thisGame.
  ///
  /// In en, this message translates to:
  /// **'THIS GAME'**
  String get thisGame;

  /// No description provided for @availablePoints.
  ///
  /// In en, this message translates to:
  /// **'AVAILABLE'**
  String get availablePoints;

  /// No description provided for @debt.
  ///
  /// In en, this message translates to:
  /// **'DEBT'**
  String get debt;

  /// No description provided for @scoreNoticeTitle.
  ///
  /// In en, this message translates to:
  /// **'Score is not spending money'**
  String get scoreNoticeTitle;

  /// No description provided for @scoreNoticeMessage.
  ///
  /// In en, this message translates to:
  /// **'Receipts repay the oldest outstanding debt first. Repayment does not deduct your score twice. At financial finalization, unspent current-game points close; recorded scores remain.'**
  String get scoreNoticeMessage;

  /// No description provided for @finishSettlement.
  ///
  /// In en, this message translates to:
  /// **'Finish settlement'**
  String get finishSettlement;

  /// No description provided for @artworkCosmetic.
  ///
  /// In en, this message translates to:
  /// **'Artwork is cosmetic. Rank, suit and availability come from the supplied player view.'**
  String get artworkCosmetic;

  /// No description provided for @closeInspection.
  ///
  /// In en, this message translates to:
  /// **'Close inspection'**
  String get closeInspection;

  /// No description provided for @roundPhase.
  ///
  /// In en, this message translates to:
  /// **'ROUND {round} / {total}  ·  {phase}'**
  String roundPhase(int round, int total, String phase);

  /// No description provided for @exposedCardsOf.
  ///
  /// In en, this message translates to:
  /// **'{name} · exposed cards'**
  String exposedCardsOf(String name);

  /// No description provided for @cardCount.
  ///
  /// In en, this message translates to:
  /// **'{count, plural, one{1 card} other{{count} cards}}'**
  String cardCount(int count);

  /// No description provided for @concealedCounts.
  ///
  /// In en, this message translates to:
  /// **'{hand} in hand · {aces} concealed aces'**
  String concealedCounts(int hand, int aces);

  /// No description provided for @physicalCopy.
  ///
  /// In en, this message translates to:
  /// **'Physical copy {copy} of 2'**
  String physicalCopy(int copy);

  /// No description provided for @allCards.
  ///
  /// In en, this message translates to:
  /// **'All'**
  String get allCards;

  /// No description provided for @inspectSelection.
  ///
  /// In en, this message translates to:
  /// **'Inspect selection'**
  String get inspectSelection;

  /// No description provided for @clearSelection.
  ///
  /// In en, this message translates to:
  /// **'Clear selection'**
  String get clearSelection;

  /// No description provided for @selectionOutsideFilter.
  ///
  /// In en, this message translates to:
  /// **'Your selection is outside this filter.'**
  String get selectionOutsideFilter;

  /// No description provided for @handSelectionHint.
  ///
  /// In en, this message translates to:
  /// **'Tap to select; double-tap to magnify. Review the action before playing.'**
  String get handSelectionHint;

  /// No description provided for @noMatchingCards.
  ///
  /// In en, this message translates to:
  /// **'No cards match this filter.'**
  String get noMatchingCards;

  /// No description provided for @changeHandFilter.
  ///
  /// In en, this message translates to:
  /// **'Choose another suit or All to see your hand.'**
  String get changeHandFilter;

  /// No description provided for @yourExposedCards.
  ///
  /// In en, this message translates to:
  /// **'Your exposed cards'**
  String get yourExposedCards;

  /// No description provided for @jumpToCards.
  ///
  /// In en, this message translates to:
  /// **'Jump to cards'**
  String get jumpToCards;

  /// No description provided for @suitFilter.
  ///
  /// In en, this message translates to:
  /// **'{suit} ({count})'**
  String suitFilter(String suit, int count);

  /// No description provided for @showingCards.
  ///
  /// In en, this message translates to:
  /// **'Showing {visible} of {total} cards'**
  String showingCards(int visible, int total);

  /// No description provided for @selectedCard.
  ///
  /// In en, this message translates to:
  /// **'Selected: {card}'**
  String selectedCard(String card);

  /// No description provided for @suitGroup.
  ///
  /// In en, this message translates to:
  /// **'{suit} · {count, plural, one{1 card} other{{count} cards}}'**
  String suitGroup(String suit, int count);

  /// No description provided for @scrollPublicTable.
  ///
  /// In en, this message translates to:
  /// **'Scroll to compare all public positions'**
  String get scrollPublicTable;

  /// No description provided for @inspectReviewedCard.
  ///
  /// In en, this message translates to:
  /// **'Inspect card'**
  String get inspectReviewedCard;

  /// No description provided for @inspectCombo.
  ///
  /// In en, this message translates to:
  /// **'Inspect {name}'**
  String inspectCombo(String name);

  /// No description provided for @handPatterns.
  ///
  /// In en, this message translates to:
  /// **'Patterns in your hand'**
  String get handPatterns;

  /// No description provided for @actionBoardPublic.
  ///
  /// In en, this message translates to:
  /// **'TABLE · ALL OPEN CARDS'**
  String get actionBoardPublic;

  /// No description provided for @actionBoardYou.
  ///
  /// In en, this message translates to:
  /// **'You'**
  String get actionBoardYou;

  /// No description provided for @actionBoardRound.
  ///
  /// In en, this message translates to:
  /// **'Round {round} / {total}'**
  String actionBoardRound(int round, int total);

  /// No description provided for @actionBoardCounts.
  ///
  /// In en, this message translates to:
  /// **'Hand {hand} · Aces {aces}'**
  String actionBoardCounts(int hand, int aces);

  /// No description provided for @actionBoardPrivate.
  ///
  /// In en, this message translates to:
  /// **'Private'**
  String get actionBoardPrivate;

  /// No description provided for @actionBoardEndTurn.
  ///
  /// In en, this message translates to:
  /// **'End turn'**
  String get actionBoardEndTurn;

  /// No description provided for @actionBoardReviewPile.
  ///
  /// In en, this message translates to:
  /// **'Review every card in this stack'**
  String get actionBoardReviewPile;

  /// No description provided for @actionBoardInspectHint.
  ///
  /// In en, this message translates to:
  /// **'Double-tap, use the magnifier, or press I to magnify'**
  String get actionBoardInspectHint;

  /// No description provided for @actionBoardSelect.
  ///
  /// In en, this message translates to:
  /// **'Select card'**
  String get actionBoardSelect;

  /// No description provided for @actionBoardAces.
  ///
  /// In en, this message translates to:
  /// **'Aces · {count}'**
  String actionBoardAces(int count);

  /// No description provided for @touchTable.
  ///
  /// In en, this message translates to:
  /// **'Table'**
  String get touchTable;

  /// No description provided for @touchHand.
  ///
  /// In en, this message translates to:
  /// **'Hand'**
  String get touchHand;

  /// No description provided for @touchActivity.
  ///
  /// In en, this message translates to:
  /// **'Activity'**
  String get touchActivity;

  /// No description provided for @touchNoPlayers.
  ///
  /// In en, this message translates to:
  /// **'No players at the table'**
  String get touchNoPlayers;

  /// No description provided for @touchNoPlayersDetail.
  ///
  /// In en, this message translates to:
  /// **'Player positions appear when a table is available.'**
  String get touchNoPlayersDetail;

  /// No description provided for @touchNoPublicCards.
  ///
  /// In en, this message translates to:
  /// **'No exposed cards yet'**
  String get touchNoPublicCards;

  /// No description provided for @touchNoPublicCardsDetail.
  ///
  /// In en, this message translates to:
  /// **'This player\'s concealed cards stay private.'**
  String get touchNoPublicCardsDetail;

  /// No description provided for @touchSeatSummary.
  ///
  /// In en, this message translates to:
  /// **'{cards} exposed · {combos, plural, one{1 combo} other{{combos} combos}}'**
  String touchSeatSummary(int cards, int combos);

  /// No description provided for @touchMoveDetails.
  ///
  /// In en, this message translates to:
  /// **'Move details'**
  String get touchMoveDetails;

  /// No description provided for @touchResponsePassed.
  ///
  /// In en, this message translates to:
  /// **'Passed'**
  String get touchResponsePassed;

  /// No description provided for @touchResponseAwaiting.
  ///
  /// In en, this message translates to:
  /// **'Awaiting response'**
  String get touchResponseAwaiting;

  /// No description provided for @touchResponseWaiting.
  ///
  /// In en, this message translates to:
  /// **'Waiting'**
  String get touchResponseWaiting;

  /// No description provided for @adaptiveTable.
  ///
  /// In en, this message translates to:
  /// **'Table'**
  String get adaptiveTable;

  /// No description provided for @adaptiveHand.
  ///
  /// In en, this message translates to:
  /// **'Hand'**
  String get adaptiveHand;

  /// No description provided for @adaptiveDecisions.
  ///
  /// In en, this message translates to:
  /// **'Decisions'**
  String get adaptiveDecisions;

  /// No description provided for @adaptiveCoup.
  ///
  /// In en, this message translates to:
  /// **'Declare Coup now'**
  String get adaptiveCoup;

  /// No description provided for @adaptiveFinances.
  ///
  /// In en, this message translates to:
  /// **'Score {score} · Cash {cash} · Debt {debt}'**
  String adaptiveFinances(String score, String cash, String debt);

  /// No description provided for @onlineCGMSOnline.
  ///
  /// In en, this message translates to:
  /// **'CGMS · Online'**
  String get onlineCGMSOnline;

  /// No description provided for @onlineArtwork.
  ///
  /// In en, this message translates to:
  /// **'Artwork'**
  String get onlineArtwork;

  /// No description provided for @onlineRightToLeft.
  ///
  /// In en, this message translates to:
  /// **'Right to left'**
  String get onlineRightToLeft;

  /// No description provided for @online200Text.
  ///
  /// In en, this message translates to:
  /// **'200% text'**
  String get online200Text;

  /// No description provided for @onlineCheckOperation.
  ///
  /// In en, this message translates to:
  /// **'Check operation'**
  String get onlineCheckOperation;

  /// No description provided for @onlineRetryOriginalOperation.
  ///
  /// In en, this message translates to:
  /// **'Retry original operation'**
  String get onlineRetryOriginalOperation;

  /// No description provided for @onlineContinueAsGuest.
  ///
  /// In en, this message translates to:
  /// **'Continue as guest'**
  String get onlineContinueAsGuest;

  /// No description provided for @onlineRecoverRoomOperation.
  ///
  /// In en, this message translates to:
  /// **'Recover room operation'**
  String get onlineRecoverRoomOperation;

  /// No description provided for @onlineCreateRoom.
  ///
  /// In en, this message translates to:
  /// **'Create room'**
  String get onlineCreateRoom;

  /// No description provided for @onlineJoinRoom.
  ///
  /// In en, this message translates to:
  /// **'Join room'**
  String get onlineJoinRoom;

  /// No description provided for @onlineRefreshSeats.
  ///
  /// In en, this message translates to:
  /// **'Refresh seats'**
  String get onlineRefreshSeats;

  /// No description provided for @onlineCreateInvitation.
  ///
  /// In en, this message translates to:
  /// **'Create invitation'**
  String get onlineCreateInvitation;

  /// No description provided for @onlineExactSettlementIsProgressingOrdinaryActionsWait.
  ///
  /// In en, this message translates to:
  /// **'Exact settlement is progressing. Ordinary actions wait.'**
  String get onlineExactSettlementIsProgressingOrdinaryActionsWait;

  /// No description provided for @onlineReturnToDecisions.
  ///
  /// In en, this message translates to:
  /// **'Return to decisions'**
  String get onlineReturnToDecisions;

  /// No description provided for @onlineClearSelection.
  ///
  /// In en, this message translates to:
  /// **'Clear selection'**
  String get onlineClearSelection;

  /// No description provided for @onlinePassResponse.
  ///
  /// In en, this message translates to:
  /// **'Pass response'**
  String get onlinePassResponse;

  /// No description provided for @onlineLoanIntactFormation.
  ///
  /// In en, this message translates to:
  /// **'Loan this intact formation'**
  String get onlineLoanIntactFormation;

  /// No description provided for @onlineDoppelgangerSubstitution.
  ///
  /// In en, this message translates to:
  /// **'Doppelganger substitution'**
  String get onlineDoppelgangerSubstitution;

  /// No description provided for @onlineProtectASeriesOrFormation.
  ///
  /// In en, this message translates to:
  /// **'Protect a series or formation'**
  String get onlineProtectASeriesOrFormation;

  /// No description provided for @online.
  ///
  /// In en, this message translates to:
  /// **' / '**
  String get online;

  /// No description provided for @onlineUsername.
  ///
  /// In en, this message translates to:
  /// **'Username'**
  String get onlineUsername;

  /// No description provided for @onlinePassword.
  ///
  /// In en, this message translates to:
  /// **'Password'**
  String get onlinePassword;

  /// No description provided for @onlinePlayers.
  ///
  /// In en, this message translates to:
  /// **'Players'**
  String get onlinePlayers;

  /// No description provided for @onlineGamesPerMatch.
  ///
  /// In en, this message translates to:
  /// **'Games per match'**
  String get onlineGamesPerMatch;

  /// No description provided for @onlineRoomCode.
  ///
  /// In en, this message translates to:
  /// **'Room code'**
  String get onlineRoomCode;

  /// No description provided for @onlineInvitation.
  ///
  /// In en, this message translates to:
  /// **'Invitation'**
  String get onlineInvitation;

  /// No description provided for @onlineAction.
  ///
  /// In en, this message translates to:
  /// **'Action'**
  String get onlineAction;

  /// No description provided for @onlineSuit.
  ///
  /// In en, this message translates to:
  /// **'Suit'**
  String get onlineSuit;

  /// No description provided for @onlineFormation.
  ///
  /// In en, this message translates to:
  /// **'Formation'**
  String get onlineFormation;

  /// No description provided for @onlineSubstitutedRank.
  ///
  /// In en, this message translates to:
  /// **'Substituted rank'**
  String get onlineSubstitutedRank;

  /// No description provided for @onlineSubstitutedSuit.
  ///
  /// In en, this message translates to:
  /// **'Substituted suit'**
  String get onlineSubstitutedSuit;

  /// No description provided for @onlineProtectedSeries.
  ///
  /// In en, this message translates to:
  /// **'Protected series'**
  String get onlineProtectedSeries;

  /// No description provided for @onlineAccept.
  ///
  /// In en, this message translates to:
  /// **'Accept'**
  String get onlineAccept;

  /// No description provided for @onlineAcceptOffer.
  ///
  /// In en, this message translates to:
  /// **'Accept Offer'**
  String get onlineAcceptOffer;

  /// No description provided for @onlineAce.
  ///
  /// In en, this message translates to:
  /// **'Ace'**
  String get onlineAce;

  /// No description provided for @onlineAdvancementDefense.
  ///
  /// In en, this message translates to:
  /// **'Advancement Defense'**
  String get onlineAdvancementDefense;

  /// No description provided for @onlineAmount.
  ///
  /// In en, this message translates to:
  /// **'Amount'**
  String get onlineAmount;

  /// No description provided for @onlineArray.
  ///
  /// In en, this message translates to:
  /// **'Array'**
  String get onlineArray;

  /// No description provided for @onlineAttach.
  ///
  /// In en, this message translates to:
  /// **'Attach'**
  String get onlineAttach;

  /// No description provided for @onlineAttack.
  ///
  /// In en, this message translates to:
  /// **'Attack'**
  String get onlineAttack;

  /// No description provided for @onlineAwardId.
  ///
  /// In en, this message translates to:
  /// **'Award Id'**
  String get onlineAwardId;

  /// No description provided for @onlineBaronAttack.
  ///
  /// In en, this message translates to:
  /// **'Baron Attack'**
  String get onlineBaronAttack;

  /// No description provided for @onlineBaronBombAttack.
  ///
  /// In en, this message translates to:
  /// **'Baron Bomb Attack'**
  String get onlineBaronBombAttack;

  /// No description provided for @onlineBarricadeSacrifice.
  ///
  /// In en, this message translates to:
  /// **'Barricade Sacrifice'**
  String get onlineBarricadeSacrifice;

  /// No description provided for @onlineBombAttack.
  ///
  /// In en, this message translates to:
  /// **'Bomb Attack'**
  String get onlineBombAttack;

  /// No description provided for @onlineBoolean.
  ///
  /// In en, this message translates to:
  /// **'Boolean'**
  String get onlineBoolean;

  /// No description provided for @onlineCode.
  ///
  /// In en, this message translates to:
  /// **'Code'**
  String get onlineCode;

  /// No description provided for @onlineCompensation.
  ///
  /// In en, this message translates to:
  /// **'Compensation'**
  String get onlineCompensation;

  /// No description provided for @onlineCompensationResponse.
  ///
  /// In en, this message translates to:
  /// **'Compensation Response'**
  String get onlineCompensationResponse;

  /// No description provided for @onlineCondition.
  ///
  /// In en, this message translates to:
  /// **'Condition'**
  String get onlineCondition;

  /// No description provided for @onlineConfinement.
  ///
  /// In en, this message translates to:
  /// **'Confinement'**
  String get onlineConfinement;

  /// No description provided for @onlineConst.
  ///
  /// In en, this message translates to:
  /// **'Const'**
  String get onlineConst;

  /// No description provided for @onlineCoup.
  ///
  /// In en, this message translates to:
  /// **'Coup'**
  String get onlineCoup;

  /// No description provided for @onlineDebtId.
  ///
  /// In en, this message translates to:
  /// **'Debt Id'**
  String get onlineDebtId;

  /// No description provided for @onlineDecision.
  ///
  /// In en, this message translates to:
  /// **'Decision'**
  String get onlineDecision;

  /// No description provided for @onlineDecisionClosed.
  ///
  /// In en, this message translates to:
  /// **'Decision Closed'**
  String get onlineDecisionClosed;

  /// No description provided for @onlineDecisionId.
  ///
  /// In en, this message translates to:
  /// **'Decision Id'**
  String get onlineDecisionId;

  /// No description provided for @onlineDeclareOrdinary.
  ///
  /// In en, this message translates to:
  /// **'Declare Ordinary'**
  String get onlineDeclareOrdinary;

  /// No description provided for @onlineDeclineOffer.
  ///
  /// In en, this message translates to:
  /// **'Decline Offer'**
  String get onlineDeclineOffer;

  /// No description provided for @onlineDepartureChoice.
  ///
  /// In en, this message translates to:
  /// **'Departure Choice'**
  String get onlineDepartureChoice;

  /// No description provided for @onlineDexterAssassination.
  ///
  /// In en, this message translates to:
  /// **'Dexter Assassination'**
  String get onlineDexterAssassination;

  /// No description provided for @onlineEndTurn.
  ///
  /// In en, this message translates to:
  /// **'End Turn'**
  String get onlineEndTurn;

  /// No description provided for @onlineExile.
  ///
  /// In en, this message translates to:
  /// **'Exile'**
  String get onlineExile;

  /// No description provided for @onlineExposeUnusedAce.
  ///
  /// In en, this message translates to:
  /// **'Expose Unused Ace'**
  String get onlineExposeUnusedAce;

  /// No description provided for @onlineFate.
  ///
  /// In en, this message translates to:
  /// **'Fate'**
  String get onlineFate;

  /// No description provided for @onlineFinishSettlement.
  ///
  /// In en, this message translates to:
  /// **'Finish Settlement'**
  String get onlineFinishSettlement;

  /// No description provided for @onlineForgive.
  ///
  /// In en, this message translates to:
  /// **'Forgive'**
  String get onlineForgive;

  /// No description provided for @onlineFormationId.
  ///
  /// In en, this message translates to:
  /// **'Formation Id'**
  String get onlineFormationId;

  /// No description provided for @onlineInfiltrator.
  ///
  /// In en, this message translates to:
  /// **'Infiltrator'**
  String get onlineInfiltrator;

  /// No description provided for @onlineInfiltratorAttack.
  ///
  /// In en, this message translates to:
  /// **'Infiltrator Attack'**
  String get onlineInfiltratorAttack;

  /// No description provided for @onlineInflation.
  ///
  /// In en, this message translates to:
  /// **'Inflation'**
  String get onlineInflation;

  /// No description provided for @onlineInteger.
  ///
  /// In en, this message translates to:
  /// **'Integer'**
  String get onlineInteger;

  /// No description provided for @onlineItems.
  ///
  /// In en, this message translates to:
  /// **'Items'**
  String get onlineItems;

  /// No description provided for @onlineJustice.
  ///
  /// In en, this message translates to:
  /// **'Justice'**
  String get onlineJustice;

  /// No description provided for @onlineKidnapper.
  ///
  /// In en, this message translates to:
  /// **'Kidnapper'**
  String get onlineKidnapper;

  /// No description provided for @onlineKidnapperOutcome.
  ///
  /// In en, this message translates to:
  /// **'Kidnapper Outcome'**
  String get onlineKidnapperOutcome;

  /// No description provided for @onlineMainInflation.
  ///
  /// In en, this message translates to:
  /// **'Main Inflation'**
  String get onlineMainInflation;

  /// No description provided for @onlineMode.
  ///
  /// In en, this message translates to:
  /// **'Mode'**
  String get onlineMode;

  /// No description provided for @onlineNegotiate.
  ///
  /// In en, this message translates to:
  /// **'Negotiate'**
  String get onlineNegotiate;

  /// No description provided for @onlineNullify.
  ///
  /// In en, this message translates to:
  /// **'Nullify'**
  String get onlineNullify;

  /// No description provided for @onlineNumericalDefense.
  ///
  /// In en, this message translates to:
  /// **'Numerical Defense'**
  String get onlineNumericalDefense;

  /// No description provided for @onlineObject.
  ///
  /// In en, this message translates to:
  /// **'Object'**
  String get onlineObject;

  /// No description provided for @onlineOffer.
  ///
  /// In en, this message translates to:
  /// **'Offer'**
  String get onlineOffer;

  /// No description provided for @onlineOfferId.
  ///
  /// In en, this message translates to:
  /// **'Offer Id'**
  String get onlineOfferId;

  /// No description provided for @onlineOpenFormation.
  ///
  /// In en, this message translates to:
  /// **'Open Formation'**
  String get onlineOpenFormation;

  /// No description provided for @onlineOpenSeries.
  ///
  /// In en, this message translates to:
  /// **'Open Series'**
  String get onlineOpenSeries;

  /// No description provided for @onlinePass.
  ///
  /// In en, this message translates to:
  /// **'Pass'**
  String get onlinePass;

  /// No description provided for @onlinePayment.
  ///
  /// In en, this message translates to:
  /// **'Payment'**
  String get onlinePayment;

  /// No description provided for @onlinePendingActionId.
  ///
  /// In en, this message translates to:
  /// **'Pending Action Id'**
  String get onlinePendingActionId;

  /// No description provided for @onlinePonzi.
  ///
  /// In en, this message translates to:
  /// **'Ponzi'**
  String get onlinePonzi;

  /// No description provided for @onlinePrice.
  ///
  /// In en, this message translates to:
  /// **'Price'**
  String get onlinePrice;

  /// No description provided for @onlinePromiseAccept.
  ///
  /// In en, this message translates to:
  /// **'Promise Accept'**
  String get onlinePromiseAccept;

  /// No description provided for @onlinePromiseFinalAnswer.
  ///
  /// In en, this message translates to:
  /// **'Promise Final Answer'**
  String get onlinePromiseFinalAnswer;

  /// No description provided for @onlinePromiseFinalOffer.
  ///
  /// In en, this message translates to:
  /// **'Promise Final Offer'**
  String get onlinePromiseFinalOffer;

  /// No description provided for @onlinePromiseOffer.
  ///
  /// In en, this message translates to:
  /// **'Promise Offer'**
  String get onlinePromiseOffer;

  /// No description provided for @onlinePromisePay.
  ///
  /// In en, this message translates to:
  /// **'Promise Pay'**
  String get onlinePromisePay;

  /// No description provided for @onlinePromiseRefuse.
  ///
  /// In en, this message translates to:
  /// **'Promise Refuse'**
  String get onlinePromiseRefuse;

  /// No description provided for @onlinePromiseId.
  ///
  /// In en, this message translates to:
  /// **'Promise Id'**
  String get onlinePromiseId;

  /// No description provided for @onlineProperties.
  ///
  /// In en, this message translates to:
  /// **'Properties'**
  String get onlineProperties;

  /// No description provided for @onlinePurchase.
  ///
  /// In en, this message translates to:
  /// **'Purchase'**
  String get onlinePurchase;

  /// No description provided for @onlineQuantity.
  ///
  /// In en, this message translates to:
  /// **'Quantity'**
  String get onlineQuantity;

  /// No description provided for @onlineRecipient.
  ///
  /// In en, this message translates to:
  /// **'Recipient'**
  String get onlineRecipient;

  /// No description provided for @onlineRequired.
  ///
  /// In en, this message translates to:
  /// **'Required'**
  String get onlineRequired;

  /// No description provided for @onlineReturnLoan.
  ///
  /// In en, this message translates to:
  /// **'Return Loan'**
  String get onlineReturnLoan;

  /// No description provided for @onlineRevision.
  ///
  /// In en, this message translates to:
  /// **'Revision'**
  String get onlineRevision;

  /// No description provided for @onlineRicherSacrifice.
  ///
  /// In en, this message translates to:
  /// **'Richer Sacrifice'**
  String get onlineRicherSacrifice;

  /// No description provided for @onlineRingleader.
  ///
  /// In en, this message translates to:
  /// **'Ringleader'**
  String get onlineRingleader;

  /// No description provided for @onlineSelection.
  ///
  /// In en, this message translates to:
  /// **'Selection'**
  String get onlineSelection;

  /// No description provided for @onlineString.
  ///
  /// In en, this message translates to:
  /// **'String'**
  String get onlineString;

  /// No description provided for @onlineTakeBack.
  ///
  /// In en, this message translates to:
  /// **'Take Back'**
  String get onlineTakeBack;

  /// No description provided for @onlineTargetSeat.
  ///
  /// In en, this message translates to:
  /// **'Target Seat'**
  String get onlineTargetSeat;

  /// No description provided for @onlineTargets.
  ///
  /// In en, this message translates to:
  /// **'Targets'**
  String get onlineTargets;

  /// No description provided for @onlineTerms.
  ///
  /// In en, this message translates to:
  /// **'Terms'**
  String get onlineTerms;

  /// No description provided for @onlineThreshold.
  ///
  /// In en, this message translates to:
  /// **'Threshold'**
  String get onlineThreshold;

  /// No description provided for @onlineType.
  ///
  /// In en, this message translates to:
  /// **'Type'**
  String get onlineType;

  /// No description provided for @onlineValue.
  ///
  /// In en, this message translates to:
  /// **'Value'**
  String get onlineValue;

  /// No description provided for @onlineVoluntaryTransfer.
  ///
  /// In en, this message translates to:
  /// **'Voluntary Transfer'**
  String get onlineVoluntaryTransfer;

  /// No description provided for @onlineWithdrawOffer.
  ///
  /// In en, this message translates to:
  /// **'Withdraw Offer'**
  String get onlineWithdrawOffer;

  /// No description provided for @onlineQuotaStatus.
  ///
  /// In en, this message translates to:
  /// **'{ability} quota: {status}'**
  String onlineQuotaStatus(String ability, String status);

  /// No description provided for @onlineQuotaUsed.
  ///
  /// In en, this message translates to:
  /// **'used'**
  String get onlineQuotaUsed;

  /// No description provided for @onlineQuotaAvailable.
  ///
  /// In en, this message translates to:
  /// **'available'**
  String get onlineQuotaAvailable;

  /// No description provided for @onlineRefreshConnection.
  ///
  /// In en, this message translates to:
  /// **'Refresh connection'**
  String get onlineRefreshConnection;

  /// No description provided for @onlineDisplaySettings.
  ///
  /// In en, this message translates to:
  /// **'Display settings'**
  String get onlineDisplaySettings;

  /// No description provided for @onlineOutcomeUnknownKeepThisPageOpenWhileReconciling.
  ///
  /// In en, this message translates to:
  /// **'Outcome unknown. Keep this page open while reconciling.'**
  String get onlineOutcomeUnknownKeepThisPageOpenWhileReconciling;

  /// No description provided for @onlineSignInOrContinueAsAGuest.
  ///
  /// In en, this message translates to:
  /// **'Sign in or continue as a guest.'**
  String get onlineSignInOrContinueAsAGuest;

  /// No description provided for @onlineTheTableChangedReviewYourSelectionBeforeSubmittingAgain.
  ///
  /// In en, this message translates to:
  /// **'The table changed. Review your selection before submitting again.'**
  String get onlineTheTableChangedReviewYourSelectionBeforeSubmittingAgain;

  /// No description provided for @onlineYourActionMayHaveSucceededCheckTheOriginalOperation.
  ///
  /// In en, this message translates to:
  /// **'Your action may have succeeded. Check the original operation before continuing.'**
  String get onlineYourActionMayHaveSucceededCheckTheOriginalOperation;

  /// No description provided for @onlineThisActionIsNotPermittedWithTheseChoicesReview.
  ///
  /// In en, this message translates to:
  /// **'This action is not permitted with these choices. Review the rules and current turn.'**
  String get onlineThisActionIsNotPermittedWithTheseChoicesReview;

  /// No description provided for @onlineThisSessionCannotAccessThatRoomOrMatch.
  ///
  /// In en, this message translates to:
  /// **'This session cannot access that room or match.'**
  String get onlineThisSessionCannotAccessThatRoomOrMatch;

  /// No description provided for @onlineTheRoomCouldNotAcceptThisRequestCheckIts.
  ///
  /// In en, this message translates to:
  /// **'The room could not accept this request. Check its seats and invitation.'**
  String get onlineTheRoomCouldNotAcceptThisRequestCheckIts;

  /// No description provided for @onlineConnectionUnavailableRetryWhenTheAuthorityIsReachable.
  ///
  /// In en, this message translates to:
  /// **'Connection unavailable. Retry when the authority is reachable.'**
  String get onlineConnectionUnavailableRetryWhenTheAuthorityIsReachable;

  /// No description provided for @onlineCreateAThreeOrFourPlayerRoomMatchLength.
  ///
  /// In en, this message translates to:
  /// **'Create a three- or four-player room. Match length is fixed before play.'**
  String get onlineCreateAThreeOrFourPlayerRoomMatchLength;

  /// No description provided for @onlineEnterMatch.
  ///
  /// In en, this message translates to:
  /// **'Enter match'**
  String get onlineEnterMatch;

  /// No description provided for @onlineTheGameOrDecisionChangedClearThisDraftBefore.
  ///
  /// In en, this message translates to:
  /// **'The game or decision changed. Clear this draft before choosing a new action.'**
  String get onlineTheGameOrDecisionChangedClearThisDraftBefore;

  /// No description provided for @onlineTheAuthorityValidatesYourChoicesAnOfferReservesNo.
  ///
  /// In en, this message translates to:
  /// **'The authority validates your choices. An offer reserves no cards; acceptance begins ordered responses.'**
  String get onlineTheAuthorityValidatesYourChoicesAnOfferReservesNo;

  /// No description provided for @onlineChooseAPositiveQuantityPaymentResolvesAfterResponsesExcess.
  ///
  /// In en, this message translates to:
  /// **'Choose a positive quantity. Payment resolves after responses; excess is not returned as change.'**
  String get onlineChooseAPositiveQuantityPaymentResolvesAfterResponsesExcess;

  /// No description provided for @onlineSacrificedQueen.
  ///
  /// In en, this message translates to:
  /// **'Sacrificed Queen'**
  String get onlineSacrificedQueen;

  /// No description provided for @onlineNone.
  ///
  /// In en, this message translates to:
  /// **'None'**
  String get onlineNone;

  /// No description provided for @onlineRecipientSeat.
  ///
  /// In en, this message translates to:
  /// **'Recipient seat'**
  String get onlineRecipientSeat;

  /// No description provided for @onlineProtectedFormation.
  ///
  /// In en, this message translates to:
  /// **'Protected formation'**
  String get onlineProtectedFormation;

  /// No description provided for @onlineAmountNumerator.
  ///
  /// In en, this message translates to:
  /// **'Amount numerator'**
  String get onlineAmountNumerator;

  /// No description provided for @onlineDenominator.
  ///
  /// In en, this message translates to:
  /// **'Denominator'**
  String get onlineDenominator;

  /// No description provided for @onlineAcceptPromise.
  ///
  /// In en, this message translates to:
  /// **'Accept promise'**
  String get onlineAcceptPromise;

  /// No description provided for @onlineAcceptFinalOffer.
  ///
  /// In en, this message translates to:
  /// **'Accept final offer'**
  String get onlineAcceptFinalOffer;

  /// No description provided for @onlineRejectFinalOffer.
  ///
  /// In en, this message translates to:
  /// **'Reject final offer'**
  String get onlineRejectFinalOffer;

  /// No description provided for @onlinePayPromise.
  ///
  /// In en, this message translates to:
  /// **'Pay promise'**
  String get onlinePayPromise;

  /// No description provided for @onlineRefusePromise.
  ///
  /// In en, this message translates to:
  /// **'Refuse promise'**
  String get onlineRefusePromise;

  /// No description provided for @onlineProposeFinalAmount.
  ///
  /// In en, this message translates to:
  /// **'Propose final amount'**
  String get onlineProposeFinalAmount;

  /// No description provided for @onlineForgiveDebt.
  ///
  /// In en, this message translates to:
  /// **'Forgive debt'**
  String get onlineForgiveDebt;

  /// No description provided for @onlineYouOwe.
  ///
  /// In en, this message translates to:
  /// **'You owe'**
  String get onlineYouOwe;

  /// No description provided for @onlineOwedToYou.
  ///
  /// In en, this message translates to:
  /// **'Owed to you'**
  String get onlineOwedToYou;

  /// No description provided for @onlineYouGive.
  ///
  /// In en, this message translates to:
  /// **'You give'**
  String get onlineYouGive;

  /// No description provided for @onlineYouReceive.
  ///
  /// In en, this message translates to:
  /// **'You receive'**
  String get onlineYouReceive;

  /// No description provided for @onlineUsedThisTurn.
  ///
  /// In en, this message translates to:
  /// **'used this turn'**
  String get onlineUsedThisTurn;

  /// No description provided for @onlineUsedThisRound.
  ///
  /// In en, this message translates to:
  /// **'used this round'**
  String get onlineUsedThisRound;

  /// No description provided for @onlineSelectCardsOnTableOrHand.
  ///
  /// In en, this message translates to:
  /// **'select cards on table or hand'**
  String get onlineSelectCardsOnTableOrHand;

  /// No description provided for @onlineConditionTheNamedAwardOccursForYouInThis.
  ///
  /// In en, this message translates to:
  /// **'Condition: the named award occurs for you in this game.'**
  String get onlineConditionTheNamedAwardOccursForYouInThis;

  /// No description provided for @onlineRoomIdentity.
  ///
  /// In en, this message translates to:
  /// **'Room {value1}'**
  String onlineRoomIdentity(String value1);

  /// No description provided for @onlineRoomConfiguration.
  ///
  /// In en, this message translates to:
  /// **'{value1} games · {value2} seats'**
  String onlineRoomConfiguration(String value1, String value2);

  /// No description provided for @onlineSeatNumber.
  ///
  /// In en, this message translates to:
  /// **'Seat {value1}'**
  String onlineSeatNumber(String value1);

  /// No description provided for @onlineGameProgress.
  ///
  /// In en, this message translates to:
  /// **'Game {value1} of {value2}'**
  String onlineGameProgress(String value1, String value2);

  /// No description provided for @onlineOwnMatchScore.
  ///
  /// In en, this message translates to:
  /// **'Your match score: {value1}'**
  String onlineOwnMatchScore(String value1);

  /// No description provided for @onlineSubmitAction.
  ///
  /// In en, this message translates to:
  /// **'Submit {value1}'**
  String onlineSubmitAction(String value1);

  /// No description provided for @onlineOriginGame.
  ///
  /// In en, this message translates to:
  /// **'Origin game {value1}'**
  String onlineOriginGame(String value1);

  /// No description provided for @onlineRanksBySeat.
  ///
  /// In en, this message translates to:
  /// **'Match ranks by seat: {value1}'**
  String onlineRanksBySeat(String value1);

  /// No description provided for @onlinePrivateOfferStatus.
  ///
  /// In en, this message translates to:
  /// **'Private offer · revision {value1} · {value2}'**
  String onlinePrivateOfferStatus(String value1, String value2);

  /// No description provided for @onlineConnectionLost.
  ///
  /// In en, this message translates to:
  /// **'Connection lost · last confirmed table · actions paused'**
  String get onlineConnectionLost;

  /// No description provided for @onlineUnrevealedCard.
  ///
  /// In en, this message translates to:
  /// **'Unrevealed card'**
  String get onlineUnrevealedCard;

  /// No description provided for @onlineDecisionOwner.
  ///
  /// In en, this message translates to:
  /// **'Decision owner: seat {value1} · {value2} · {value3}'**
  String onlineDecisionOwner(String value1, String value2, String value3);

  /// No description provided for @onlineEnding.
  ///
  /// In en, this message translates to:
  /// **'Ending: {ending}'**
  String onlineEnding(String ending);

  /// No description provided for @onlineEndingDeclarer.
  ///
  /// In en, this message translates to:
  /// **'Ending: {value1} · Declarer: seat {value2}'**
  String onlineEndingDeclarer(String value1, String value2);

  /// No description provided for @onlinePurchaseCalculation.
  ///
  /// In en, this message translates to:
  /// **'Unit price {value1} · Quantity {value2} · Effective payment {value3} · Cost {value4} · Excess {value5}'**
  String onlinePurchaseCalculation(
    String value1,
    String value2,
    String value3,
    String value4,
    String value5,
  );

  /// No description provided for @onlineCombatCalculation.
  ///
  /// In en, this message translates to:
  /// **'Attack {value1} + {value2} = {value3} · Defense {value4} + {value5} = {value6} · {value7}'**
  String onlineCombatCalculation(
    String value1,
    String value2,
    String value3,
    String value4,
    String value5,
    String value6,
    String value7,
  );

  /// No description provided for @onlinePromiseParties.
  ///
  /// In en, this message translates to:
  /// **'Promise from seat {value1} to seat {value2} · {value3}'**
  String onlinePromiseParties(String value1, String value2, String value3);

  /// No description provided for @onlinePromiseTerms.
  ///
  /// In en, this message translates to:
  /// **'{value1} {value2} · Award {value3}'**
  String onlinePromiseTerms(String value1, String value2, String value3);

  /// No description provided for @onlinePromiseSettlement.
  ///
  /// In en, this message translates to:
  /// **'Promised {value1} · Paid {value2} · Final offer {value3}'**
  String onlinePromiseSettlement(String value1, String value2, String value3);

  /// No description provided for @onlineDebtParties.
  ///
  /// In en, this message translates to:
  /// **'{value1} {value2} · Debtor seat {value3} · Creditor seat {value4}'**
  String onlineDebtParties(
    String value1,
    String value2,
    String value3,
    String value4,
  );

  /// No description provided for @onlineOfferParties.
  ///
  /// In en, this message translates to:
  /// **'From seat {value1} to seat {value2} · {value3}'**
  String onlineOfferParties(String value1, String value2, String value3);

  /// No description provided for @onlineLastingEffect.
  ///
  /// In en, this message translates to:
  /// **'{value1} · owner {value2} · target {value3} · custodian {value4} · {value5}'**
  String onlineLastingEffect(
    String value1,
    String value2,
    String value3,
    String value4,
    String value5,
  );

  /// No description provided for @onlineCompensationDraw.
  ///
  /// In en, this message translates to:
  /// **'Compensation draw · seat {value1} · {value2} remaining'**
  String onlineCompensationDraw(String value1, String value2);

  /// No description provided for @onlineMatchCorrection.
  ///
  /// In en, this message translates to:
  /// **'Nonspendable match correction · {value1} · origin {value2}'**
  String onlineMatchCorrection(String value1, String value2);

  /// No description provided for @onlineKnownOfferCard.
  ///
  /// In en, this message translates to:
  /// **'{value1} of {value2} · offer card {value3}'**
  String onlineKnownOfferCard(String value1, String value2, String value3);

  /// No description provided for @onlineMatchComplete.
  ///
  /// In en, this message translates to:
  /// **'Match complete'**
  String get onlineMatchComplete;

  /// No description provided for @onlineAvailable.
  ///
  /// In en, this message translates to:
  /// **'available'**
  String get onlineAvailable;

  /// No description provided for @onlineUsed.
  ///
  /// In en, this message translates to:
  /// **'used'**
  String get onlineUsed;

  /// No description provided for @onlineHeld.
  ///
  /// In en, this message translates to:
  /// **'held'**
  String get onlineHeld;

  /// No description provided for @onlineAbsent.
  ///
  /// In en, this message translates to:
  /// **'none'**
  String get onlineAbsent;

  /// No description provided for @onlineChoose.
  ///
  /// In en, this message translates to:
  /// **'choose'**
  String get onlineChoose;

  /// No description provided for @onlineOpeningAceStatus.
  ///
  /// In en, this message translates to:
  /// **'Opening: {opening} · Ace: {ace}'**
  String onlineOpeningAceStatus(String opening, String ace);

  /// No description provided for @onlinePrivilegesStatus.
  ///
  /// In en, this message translates to:
  /// **'Compensation: {compensation} · Underground privilege: {underground}'**
  String onlinePrivilegesStatus(String compensation, String underground);

  /// No description provided for @onlineSelectedCards.
  ///
  /// In en, this message translates to:
  /// **'{label}: {count} cards · {instruction}'**
  String onlineSelectedCards(String label, String count, String instruction);

  /// No description provided for @onlineTransferCards.
  ///
  /// In en, this message translates to:
  /// **'{label}: {count} · select visible cards'**
  String onlineTransferCards(String label, String count);

  /// No description provided for @onlineFormationCards.
  ///
  /// In en, this message translates to:
  /// **'Formation cards: {count}'**
  String onlineFormationCards(String count);

  /// No description provided for @onlineSubstituteKings.
  ///
  /// In en, this message translates to:
  /// **'Substitute kings: {count} of 2'**
  String onlineSubstituteKings(String count);

  /// No description provided for @onlineCompletedGame.
  ///
  /// In en, this message translates to:
  /// **'Completed game · {scores}'**
  String onlineCompletedGame(String scores);

  /// No description provided for @onlineSeatScore.
  ///
  /// In en, this message translates to:
  /// **'Seat {seat}: {score}'**
  String onlineSeatScore(String seat, String score);

  /// No description provided for @onlineJack.
  ///
  /// In en, this message translates to:
  /// **'Jack'**
  String get onlineJack;

  /// No description provided for @onlineQueen.
  ///
  /// In en, this message translates to:
  /// **'Queen'**
  String get onlineQueen;

  /// No description provided for @onlineKing.
  ///
  /// In en, this message translates to:
  /// **'King'**
  String get onlineKing;

  /// No description provided for @onlineRoomOwnerStarts.
  ///
  /// In en, this message translates to:
  /// **'The room owner starts the match when everyone has joined.'**
  String get onlineRoomOwnerStarts;

  /// No description provided for @onlineRoomWaiting.
  ///
  /// In en, this message translates to:
  /// **'Waiting for {count} more players.'**
  String onlineRoomWaiting(String count);

  /// No description provided for @onlineTimingOwnTurn.
  ///
  /// In en, this message translates to:
  /// **'This action is available on your turn.'**
  String get onlineTimingOwnTurn;

  /// No description provided for @onlineTimingPending.
  ///
  /// In en, this message translates to:
  /// **'Wait for the current response or decision before an ordinary action.'**
  String get onlineTimingPending;

  /// No description provided for @onlineTimingResponder.
  ///
  /// In en, this message translates to:
  /// **'Seat {seat} must respond first.'**
  String onlineTimingResponder(String seat);

  /// No description provided for @onlineTimingEffect.
  ///
  /// In en, this message translates to:
  /// **'Complete the current effect decision; this is not a response window.'**
  String get onlineTimingEffect;

  /// No description provided for @onlineTimingNoResponse.
  ///
  /// In en, this message translates to:
  /// **'There is no response window for this action.'**
  String get onlineTimingNoResponse;

  /// No description provided for @onlineTimingNoDecision.
  ///
  /// In en, this message translates to:
  /// **'There is no effect decision for you to complete.'**
  String get onlineTimingNoDecision;

  /// No description provided for @economyTitle.
  ///
  /// In en, this message translates to:
  /// **'Allowance and dirt'**
  String get economyTitle;

  /// No description provided for @economyDisabled.
  ///
  /// In en, this message translates to:
  /// **'Online allowances are not enabled on this server.'**
  String get economyDisabled;

  /// No description provided for @economyUnavailable.
  ///
  /// In en, this message translates to:
  /// **'Could not refresh your allowance. Try again.'**
  String get economyUnavailable;

  /// No description provided for @economyRefresh.
  ///
  /// In en, this message translates to:
  /// **'Refresh allowance'**
  String get economyRefresh;

  /// No description provided for @economyDirt.
  ///
  /// In en, this message translates to:
  /// **'{amount} dirt'**
  String economyDirt(int amount);

  /// No description provided for @economyStarts.
  ///
  /// In en, this message translates to:
  /// **'{count} free online starts remaining today'**
  String economyStarts(int count);

  /// No description provided for @economyReset.
  ///
  /// In en, this message translates to:
  /// **'Free starts reset at {time} UTC.'**
  String economyReset(String time);

  /// No description provided for @economyReward.
  ///
  /// In en, this message translates to:
  /// **'Earn {amount} dirt once per finalized online game. Offline play earns no dirt; unfinished games do not refund starts.'**
  String economyReward(int amount);

  /// No description provided for @economyUnlimited.
  ///
  /// In en, this message translates to:
  /// **'Unlimited online starts'**
  String get economyUnlimited;

  /// No description provided for @economyPassUntil.
  ///
  /// In en, this message translates to:
  /// **'Pass expiry: {time} UTC'**
  String economyPassUntil(String time);

  /// No description provided for @economyPremiumUntil.
  ///
  /// In en, this message translates to:
  /// **'Premium expiry: {time} UTC'**
  String economyPremiumUntil(String time);

  /// No description provided for @economyDays.
  ///
  /// In en, this message translates to:
  /// **'Earned access'**
  String get economyDays;

  /// No description provided for @economyDaysValue.
  ///
  /// In en, this message translates to:
  /// **'{days} day(s)'**
  String economyDaysValue(int days);

  /// No description provided for @economySpend.
  ///
  /// In en, this message translates to:
  /// **'Spend {cost} dirt'**
  String economySpend(int cost);

  /// No description provided for @economyPassHelp.
  ///
  /// In en, this message translates to:
  /// **'Earned Daily and Weekly access allow unlimited online starts and keep ads unless separate ad-free access is active. Time extends from now or your current pass expiry, whichever is later.'**
  String get economyPassHelp;

  /// No description provided for @economyDaily.
  ///
  /// In en, this message translates to:
  /// **'Daily · 1 day'**
  String get economyDaily;

  /// No description provided for @economyWeekly.
  ///
  /// In en, this message translates to:
  /// **'Weekly · 7 days'**
  String get economyWeekly;

  /// No description provided for @economyNoAds.
  ///
  /// In en, this message translates to:
  /// **'Ad-free access active'**
  String get economyNoAds;

  /// No description provided for @economyAdFreeUntil.
  ///
  /// In en, this message translates to:
  /// **'Ad-free expiry: {time} UTC'**
  String economyAdFreeUntil(String time);

  /// No description provided for @economyAdsActive.
  ///
  /// In en, this message translates to:
  /// **'Ads remain active'**
  String get economyAdsActive;

  /// No description provided for @economyProductUnavailable.
  ///
  /// In en, this message translates to:
  /// **'That pass is no longer offered. No dirt was spent. Choose Daily or Weekly access.'**
  String get economyProductUnavailable;

  /// No description provided for @economyPaidUnavailable.
  ///
  /// In en, this message translates to:
  /// **'Paid Ad-free, Daily, Weekly, Monthly and Yearly products are not available in this build.'**
  String get economyPaidUnavailable;

  /// No description provided for @economyAdFreeTerms.
  ///
  /// In en, this message translates to:
  /// **'Ad-free: 30 days without ads. The 3 free online starts per day still apply unless you also have unlimited access.'**
  String get economyAdFreeTerms;

  /// No description provided for @economyPaidDailyTerms.
  ///
  /// In en, this message translates to:
  /// **'Paid Daily: 24 hours of unlimited online starts and no ads from confirmed purchase. Does not renew automatically.'**
  String get economyPaidDailyTerms;

  /// No description provided for @economyPaidLongerTerms.
  ///
  /// In en, this message translates to:
  /// **'Paid Weekly, Monthly and Yearly: unlimited online starts and no ads.'**
  String get economyPaidLongerTerms;

  /// No description provided for @economyUnknown.
  ///
  /// In en, this message translates to:
  /// **'Your pass purchase has an unknown result. Recover it before buying another pass.'**
  String get economyUnknown;

  /// No description provided for @economyRecover.
  ///
  /// In en, this message translates to:
  /// **'Recover pass purchase'**
  String get economyRecover;

  /// No description provided for @economyInsufficient.
  ///
  /// In en, this message translates to:
  /// **'Not enough dirt for this pass.'**
  String get economyInsufficient;

  /// No description provided for @economyConfirmed.
  ///
  /// In en, this message translates to:
  /// **'Pass confirmed · {cost} dirt · expires {time} UTC'**
  String economyConfirmed(int cost, String time);

  /// No description provided for @economyRecoveryUnavailable.
  ///
  /// In en, this message translates to:
  /// **'Recovery storage is unavailable. Restore access before trying another purchase.'**
  String get economyRecoveryUnavailable;

  /// No description provided for @economyAdmissionWaiting.
  ///
  /// In en, this message translates to:
  /// **'The next game is waiting for every player’s online access. Check your allowance; a player may need a pass, the UTC reset, or account access.'**
  String get economyAdmissionWaiting;

  /// No description provided for @onlineNextGamePending.
  ///
  /// In en, this message translates to:
  /// **'Game settled. Waiting for the next game.'**
  String get onlineNextGamePending;

  /// No description provided for @onlineCopyCode.
  ///
  /// In en, this message translates to:
  /// **'Copy {label}'**
  String onlineCopyCode(String label);

  /// No description provided for @onlineCodeCopied.
  ///
  /// In en, this message translates to:
  /// **'{label} copied'**
  String onlineCodeCopied(String label);

  /// No description provided for @onlineCodeValue.
  ///
  /// In en, this message translates to:
  /// **'{label}: {value}'**
  String onlineCodeValue(String label, String value);

  /// No description provided for @onlineClipboardUnavailable.
  ///
  /// In en, this message translates to:
  /// **'Clipboard access is unavailable. The code remains visible; check your browser permissions and try Copy again.'**
  String get onlineClipboardUnavailable;

  /// No description provided for @onlineClubUsedThisTurn.
  ///
  /// In en, this message translates to:
  /// **'Used this turn'**
  String get onlineClubUsedThisTurn;

  /// No description provided for @onlineDropCopyLegend.
  ///
  /// In en, this message translates to:
  /// **'Copy numbers in brackets.'**
  String get onlineDropCopyLegend;

  /// No description provided for @onlineDropCostsOpeningNew.
  ///
  /// In en, this message translates to:
  /// **'Uses this turn’s opening allowance.'**
  String get onlineDropCostsOpeningNew;

  /// No description provided for @onlineDropCostsOpeningExisting.
  ///
  /// In en, this message translates to:
  /// **'No opening allowance spent.'**
  String get onlineDropCostsOpeningExisting;

  /// No description provided for @onlineDropCostsFormation.
  ///
  /// In en, this message translates to:
  /// **'Opening allowance unless earned Underground. No cards consumed.'**
  String get onlineDropCostsFormation;

  /// No description provided for @onlineDropCostsAttach.
  ///
  /// In en, this message translates to:
  /// **'No opening allowance. Royal needs its supported series.'**
  String get onlineDropCostsAttach;

  /// No description provided for @onlineDropCostsRicher.
  ///
  /// In en, this message translates to:
  /// **'Round allowance. On resolution discard King; draw up to two.'**
  String get onlineDropCostsRicher;

  /// No description provided for @onlineDropCostsRingleader.
  ///
  /// In en, this message translates to:
  /// **'Once per game. On resolution discard King; gain 10.'**
  String get onlineDropCostsRingleader;

  /// No description provided for @onlineDropCostsConfinement.
  ///
  /// In en, this message translates to:
  /// **'Shared Ace round allowance. Discard Ace after resolution.'**
  String get onlineDropCostsConfinement;

  /// No description provided for @onlineDropCostsCompensation.
  ///
  /// In en, this message translates to:
  /// **'Shared Ace round and game allowance. Ace becomes persistent effect.'**
  String get onlineDropCostsCompensation;

  /// No description provided for @onlineDropCostsInflation.
  ///
  /// In en, this message translates to:
  /// **'Shared Ace round allowance. Ace becomes persistent effect; prevention discards it.'**
  String get onlineDropCostsInflation;

  /// No description provided for @onlineDropCostsCombatResponse.
  ///
  /// In en, this message translates to:
  /// **'Shared Ace round allowance. Reveal/discard Ace after combat, even if canceled.'**
  String get onlineDropCostsCombatResponse;

  /// No description provided for @onlineDropCostsPurchase.
  ///
  /// In en, this message translates to:
  /// **'On resolution return exact payment and draw chosen quantity; canceled costs nothing.'**
  String get onlineDropCostsPurchase;

  /// No description provided for @onlineDropCostsReturnLoan.
  ///
  /// In en, this message translates to:
  /// **'Return these borrowed cards. Alliance and earned privileges remain.'**
  String get onlineDropCostsReturnLoan;

  /// No description provided for @onlineDropCostsFate.
  ///
  /// In en, this message translates to:
  /// **'Turn-cycle allowance. Return chosen King before reset shuffle.'**
  String get onlineDropCostsFate;

  /// No description provided for @onlineDropCostsJustice.
  ///
  /// In en, this message translates to:
  /// **'Turn-cycle allowance. Return chosen Queen after resolution.'**
  String get onlineDropCostsJustice;

  /// No description provided for @onlineDropCostsKidnapper.
  ///
  /// In en, this message translates to:
  /// **'Own-turn allowance. Special Doppelganger theft returns all members on resolution.'**
  String get onlineDropCostsKidnapper;

  /// No description provided for @onlineDropCostsPonzi.
  ///
  /// In en, this message translates to:
  /// **'Turn-cycle allowance. Return all four Jacks after resolution.'**
  String get onlineDropCostsPonzi;

  /// No description provided for @onlineDropCostsDexter.
  ///
  /// In en, this message translates to:
  /// **'Shared round allowance. On resolution return Diamond; discard target royal.'**
  String get onlineDropCostsDexter;

  /// No description provided for @onlineDropCostsBarricade.
  ///
  /// In en, this message translates to:
  /// **'Turn-cycle allowance. On resolution discard Queen; return five; draw up to five restricted cards.'**
  String get onlineDropCostsBarricade;

  /// No description provided for @onlineDropCostsNegotiator.
  ///
  /// In en, this message translates to:
  /// **'Round allowance. Final commitment uses added Clubs, transfers Spades, and returns Jack.'**
  String get onlineDropCostsNegotiator;

  /// No description provided for @onlineDropCostsTakeBack.
  ///
  /// In en, this message translates to:
  /// **'Return formation cards; ability allowances remain spent.'**
  String get onlineDropCostsTakeBack;

  /// No description provided for @onlineDropCostsAttack.
  ///
  /// In en, this message translates to:
  /// **'Attacking Clubs become used at declaration, even if canceled.'**
  String get onlineDropCostsAttack;

  /// No description provided for @onlineDropCostsBaron.
  ///
  /// In en, this message translates to:
  /// **'Baron round allowance; no Baron card consumed.'**
  String get onlineDropCostsBaron;

  /// No description provided for @onlineDropCostsBomb.
  ///
  /// In en, this message translates to:
  /// **'On success discard Bomb; return Clubs only for ordinary Bomb attack.'**
  String get onlineDropCostsBomb;

  /// No description provided for @onlineDropCostsInfiltrator.
  ///
  /// In en, this message translates to:
  /// **'Return Jack at declaration; spend turn-cycle allowance.'**
  String get onlineDropCostsInfiltrator;

  /// No description provided for @onlineDropCostsExile.
  ///
  /// In en, this message translates to:
  /// **'Spend Exile round allowance; no card consumed.'**
  String get onlineDropCostsExile;

  /// No description provided for @onlineDropCostsCombatAce.
  ///
  /// In en, this message translates to:
  /// **'Shared Ace round allowance; reveal/discard Ace after combat, even if canceled.'**
  String get onlineDropCostsCombatAce;

  /// No description provided for @onlineCostsTitle.
  ///
  /// In en, this message translates to:
  /// **'Costs and timing'**
  String get onlineCostsTitle;

  /// No description provided for @onlineCostsGeneral.
  ///
  /// In en, this message translates to:
  /// **'Illegal declarations spend nothing. Accepted allowances stay spent after cancellation; support is checked again before unresolved effects commit.'**
  String get onlineCostsGeneral;

  /// No description provided for @onlineCostsAttack.
  ///
  /// In en, this message translates to:
  /// **'Own turn; choose unused exposed Clubs and a legal target. Physical Clubs become used at legal declaration, including if the attack is canceled. Exchange costs apply only after successful resolution.'**
  String get onlineCostsAttack;

  /// No description provided for @onlineCostsCombatAce.
  ///
  /// In en, this message translates to:
  /// **'If using a numerical or Advancement Ace, its shared round allowance is spent at declaration and the Ace is revealed and discarded after combat, including cancellation.'**
  String get onlineCostsCombatAce;

  /// No description provided for @onlineCostsBaron.
  ///
  /// In en, this message translates to:
  /// **'Baron also spends its round allowance at declaration. No Baron card is consumed; cancellation keeps the round allowance and Club usage spent.'**
  String get onlineCostsBaron;

  /// No description provided for @onlineCostsBomb.
  ///
  /// In en, this message translates to:
  /// **'Successful ordinary Bomb destruction returns the attacking Clubs and discards the Bomb. Baron discards the Bomb without returning Clubs. Cancellation removes neither.'**
  String get onlineCostsBomb;

  /// No description provided for @onlineCostsInfiltrator.
  ///
  /// In en, this message translates to:
  /// **'Return the Infiltrator jack at legal declaration; cancellation does not return it or its turn-cycle allowance.'**
  String get onlineCostsInfiltrator;

  /// No description provided for @onlineCostsExile.
  ///
  /// In en, this message translates to:
  /// **'Exile spends its round allowance at the accepted Barricade bypass. It has no card cost; cancellation does not refund the allowance.'**
  String get onlineCostsExile;

  /// No description provided for @onlineCostsOpening.
  ///
  /// In en, this message translates to:
  /// **'Own turn: normally open one new number series or open/reconfigure one combo. Earned Underground permits unlimited combo openings/reconfigurations; it does not waive Ace or attack prerequisites.'**
  String get onlineCostsOpening;

  /// No description provided for @onlineCostsFormation.
  ///
  /// In en, this message translates to:
  /// **'Opening moves the selected cards into a formation; it consumes no cards. Passive abilities need their support. Doppelganger pairs and replacement slots stay bound when support is lost.'**
  String get onlineCostsFormation;

  /// No description provided for @onlineCostsTakeBack.
  ///
  /// In en, this message translates to:
  /// **'Own turn: combos may be taken back. Number series and attachments cannot be taken back. Taking back cards never refreshes an ability allowance.'**
  String get onlineCostsTakeBack;

  /// No description provided for @onlineCostsAttach.
  ///
  /// In en, this message translates to:
  /// **'Own turn: add to an existing series or restore a historically opened suit without a new opening allowance. Attached royals need their supporting number series to function.'**
  String get onlineCostsAttach;

  /// No description provided for @onlineCostsFate.
  ///
  /// In en, this message translates to:
  /// **'Own turn; turn-cycle allowance spent at activation. Return one chosen Fate king immediately before the reset shuffle. Cancellation before resolution keeps the king but spends the allowance.'**
  String get onlineCostsFate;

  /// No description provided for @onlineCostsCoup.
  ///
  /// In en, this message translates to:
  /// **'Any eligible boundary, including pending decisions and while confined. Requires historical Clubs, both physical Hearts queens and both physical Diamonds queens, all unrestricted, and no alliance. No allowance or consumable cost; no response window. Earlier committed effects remain.'**
  String get onlineCostsCoup;

  /// No description provided for @onlineCostsJustice.
  ///
  /// In en, this message translates to:
  /// **'Own turn; turn-cycle allowance spent at activation. Return one chosen Justice queen after the selection sequence, even with zero captures. Cancellation before resolution keeps the queen but spends the allowance.'**
  String get onlineCostsJustice;

  /// No description provided for @onlineCostsCode.
  ///
  /// In en, this message translates to:
  /// **'Own turn; turn-cycle allowance spent at activation, with no card payment. Resolution replaces the governing Code and settings. Cancellation keeps the old Code and spends the allowance.'**
  String get onlineCostsCode;

  /// No description provided for @onlineCostsKidnapper.
  ///
  /// In en, this message translates to:
  /// **'Own turn; one activation per own turn, committed at acceptance. Normal theft consumes no combo cards. Special Doppelganger theft returns the entire Kidnapper formation on resolution. Cancellation keeps those cards but spends the allowance.'**
  String get onlineCostsKidnapper;

  /// No description provided for @onlineCostsPonzi.
  ///
  /// In en, this message translates to:
  /// **'Any idle boundary; turn-cycle allowance spent at activation, with no card payment yet. Return all four jacks after resolution, even when nothing is stolen. Cancellation keeps the jacks but spends the turn-cycle allowance.'**
  String get onlineCostsPonzi;

  /// No description provided for @onlineCostsRingleader.
  ///
  /// In en, this message translates to:
  /// **'Own turn; once per game, spent at activation. Discard the supported king and award 10 together on resolution. Cancellation keeps the king but spends the game allowance.'**
  String get onlineCostsRingleader;

  /// No description provided for @onlineCostsRicher.
  ///
  /// In en, this message translates to:
  /// **'Own turn; round allowance spent at activation. Discard the supported king and draw up to two on resolution. Cancellation keeps the king but spends the allowance.'**
  String get onlineCostsRicher;

  /// No description provided for @onlineCostsBarricade.
  ///
  /// In en, this message translates to:
  /// **'Own turn; turn-cycle allowance spent at activation. On resolution discard the supported queen, return five other selected cards and draw up to five; those draws are restricted until next round. Cancellation keeps all selected cards.'**
  String get onlineCostsBarricade;

  /// No description provided for @onlineCostsNegotiator.
  ///
  /// In en, this message translates to:
  /// **'Incoming club-attack response; round allowance spent only on a legal activation. At the final decision, added Clubs become used, selected Spades transfer, the jack returns and the attack is canceled together. Abort before commitment keeps those costs unpaid; the accepted allowance and original Club usage stay spent.'**
  String get onlineCostsNegotiator;

  /// No description provided for @onlineCostsDexter.
  ///
  /// In en, this message translates to:
  /// **'Own turn; supported Dexter and prepayment Diamond total at least 20. Spends the round allowance shared with prevention. Return the selected number Diamond and discard the target royal together on resolution. Cancellation keeps both cards but spends the allowance.'**
  String get onlineCostsDexter;

  /// No description provided for @onlineCostsDexterPrevention.
  ///
  /// In en, this message translates to:
  /// **'Inside hostile Ace resolution; supported Dexter and prepayment Diamond total 1–10. A legal payable prevention spends the shared round allowance. Return the chosen Diamond and prevent the effect together; the hostile Ace is discarded and its allowance stays spent.'**
  String get onlineCostsDexterPrevention;

  /// No description provided for @onlineCostsConfinement.
  ///
  /// In en, this message translates to:
  /// **'Authorized response to the pending actor; spends the shared Ace round allowance at acceptance. Discard the Ace when resolution finishes, including Dexter prevention. Coup cannot be stopped.'**
  String get onlineCostsConfinement;

  /// No description provided for @onlineCostsCombatResponse.
  ///
  /// In en, this message translates to:
  /// **'Incoming-combat response; spends the shared Ace round allowance and commits its modifier. Reveal and discard the Ace after combat resolves or is canceled. Dexter prevention removes the modifier, not its usage or discard.'**
  String get onlineCostsCombatResponse;

  /// No description provided for @onlineCostsInflation.
  ///
  /// In en, this message translates to:
  /// **'Own main action or authorized incoming-attack response; spends the shared Ace round allowance. A duplicate active target is illegal and spends nothing. Resolution puts the Ace in its public effect zone; Dexter prevention discards it. Earlier cancellation keeps the Ace but spends the allowance.'**
  String get onlineCostsInflation;

  /// No description provided for @onlineCostsCompensation.
  ///
  /// In en, this message translates to:
  /// **'Own main action or authorized incoming-attack response before losses; spends both the shared Ace round allowance and game activation at acceptance. Resolution puts the Ace in its public effect zone. Cancellation keeps the Ace but spends both allowances.'**
  String get onlineCostsCompensation;

  /// No description provided for @onlineCostsPurchase.
  ///
  /// In en, this message translates to:
  /// **'Own idle boundary; no separate purchase allowance and no card payment at acceptance. Revalidate exact payment, quantity and supply after responses, then return, shuffle and draw together. Cancellation keeps every payment card and draws nothing.'**
  String get onlineCostsPurchase;

  /// No description provided for @onlineCostsOffer.
  ///
  /// In en, this message translates to:
  /// **'A proposal reserves no cards or allowances and does not block play. Acceptance names its exact revision in the originating turn. Transfers and any loan alliance happen together only after successful resolution; an accepted Ace transfer spends its supplier’s round allowance even if canceled.'**
  String get onlineCostsOffer;

  /// No description provided for @onlineCostsReturnLoan.
  ///
  /// In en, this message translates to:
  /// **'The recipient voluntarily returns only borrowed cards they still control. An intact formation returns intact; surviving individual cards return exposed and unassigned without rebuilding it. The game-long alliance and previously earned privileges remain.'**
  String get onlineCostsReturnLoan;

  /// No description provided for @onlineCostsDecision.
  ///
  /// In en, this message translates to:
  /// **'Complete only the current authorized effect input. Earlier committed costs and allowances remain spent; unresolved payments and movement commit only with the completed effect.'**
  String get onlineCostsDecision;

  /// No description provided for @onlineCostsPass.
  ///
  /// In en, this message translates to:
  /// **'Passing uses this response opportunity without playing a card. Navigation, disconnection and silence never submit a pass.'**
  String get onlineCostsPass;

  /// No description provided for @onlineWelcomeTitle.
  ///
  /// In en, this message translates to:
  /// **'A table worth turning.'**
  String get onlineWelcomeTitle;

  /// No description provided for @onlineWelcomeBody.
  ///
  /// In en, this message translates to:
  /// **'Build your hand. Make your move. Every deal changes the table.'**
  String get onlineWelcomeBody;

  /// No description provided for @onlineWelcomeCards.
  ///
  /// In en, this message translates to:
  /// **'Illustrated CGMS cards'**
  String get onlineWelcomeCards;

  /// No description provided for @onlineEntryTitle.
  ///
  /// In en, this message translates to:
  /// **'Take your seat'**
  String get onlineEntryTitle;

  /// No description provided for @onlineEntryBody.
  ///
  /// In en, this message translates to:
  /// **'Continue as a guest, or sign in to your account.'**
  String get onlineEntryBody;

  /// No description provided for @onlineAccountSignIn.
  ///
  /// In en, this message translates to:
  /// **'Sign in or register'**
  String get onlineAccountSignIn;

  /// No description provided for @onlineYourRoom.
  ///
  /// In en, this message translates to:
  /// **'Your room'**
  String get onlineYourRoom;

  /// No description provided for @onlineRoomJoined.
  ///
  /// In en, this message translates to:
  /// **'Joined'**
  String get onlineRoomJoined;

  /// No description provided for @onlineRoomEmptySeat.
  ///
  /// In en, this message translates to:
  /// **'Waiting for a player'**
  String get onlineRoomEmptySeat;

  /// No description provided for @onlineSetupTitle.
  ///
  /// In en, this message translates to:
  /// **'Gather your table'**
  String get onlineSetupTitle;

  /// No description provided for @onlineSetupBody.
  ///
  /// In en, this message translates to:
  /// **'Play against bots, create a room for your group, or join with a private invitation.'**
  String get onlineSetupBody;

  /// No description provided for @onlineJoinHelp.
  ///
  /// In en, this message translates to:
  /// **'Use the room code and invitation shared by the room owner.'**
  String get onlineJoinHelp;

  /// No description provided for @onlineActionFamilyPlay.
  ///
  /// In en, this message translates to:
  /// **'Open & build'**
  String get onlineActionFamilyPlay;

  /// No description provided for @onlineActionFamilyAttack.
  ///
  /// In en, this message translates to:
  /// **'Attack & defend'**
  String get onlineActionFamilyAttack;

  /// No description provided for @onlineActionFamilyAbilities.
  ///
  /// In en, this message translates to:
  /// **'Abilities'**
  String get onlineActionFamilyAbilities;

  /// No description provided for @onlineActionFamilyDeals.
  ///
  /// In en, this message translates to:
  /// **'Trades & loans'**
  String get onlineActionFamilyDeals;

  /// No description provided for @onlineActionFamilyFinance.
  ///
  /// In en, this message translates to:
  /// **'Points & settlement'**
  String get onlineActionFamilyFinance;

  /// No description provided for @onlineActionFamilyTurn.
  ///
  /// In en, this message translates to:
  /// **'Turn & decisions'**
  String get onlineActionFamilyTurn;

  /// No description provided for @onlineActionPrepare.
  ///
  /// In en, this message translates to:
  /// **'Prepare your action'**
  String get onlineActionPrepare;

  /// No description provided for @onlineReconnectHelp.
  ///
  /// In en, this message translates to:
  /// **'Your last confirmed table is still visible. Reconnecting automatically; no action is passed or resent.'**
  String get onlineReconnectHelp;

  /// No description provided for @onlineSessionConnecting.
  ///
  /// In en, this message translates to:
  /// **'Connecting to the table… attempt {attempt} of {limit}'**
  String onlineSessionConnecting(int attempt, int limit);

  /// No description provided for @onlineRecoveryExhausted.
  ///
  /// In en, this message translates to:
  /// **'Connection is still unavailable after bounded recovery. Use Refresh connection to try again.'**
  String get onlineRecoveryExhausted;

  /// No description provided for @onlineInvalidResponse.
  ///
  /// In en, this message translates to:
  /// **'The server returned an unreadable response. Use Refresh connection to try again.'**
  String get onlineInvalidResponse;

  /// No description provided for @onlineReconnectStopped.
  ///
  /// In en, this message translates to:
  /// **'Your last confirmed table is still visible. Use Refresh connection to retry; no action is passed or resent.'**
  String get onlineReconnectStopped;

  /// No description provided for @onlinePlayAgainstBots.
  ///
  /// In en, this message translates to:
  /// **'Play against bots'**
  String get onlinePlayAgainstBots;

  /// No description provided for @onlineBotDifficulty.
  ///
  /// In en, this message translates to:
  /// **'Bot difficulty'**
  String get onlineBotDifficulty;

  /// No description provided for @onlineBotBeginner.
  ///
  /// In en, this message translates to:
  /// **'Beginner'**
  String get onlineBotBeginner;

  /// No description provided for @onlineBotStandard.
  ///
  /// In en, this message translates to:
  /// **'Standard'**
  String get onlineBotStandard;

  /// No description provided for @onlineBotAdvanced.
  ///
  /// In en, this message translates to:
  /// **'Advanced'**
  String get onlineBotAdvanced;

  /// No description provided for @onlineBot.
  ///
  /// In en, this message translates to:
  /// **'Bot'**
  String get onlineBot;

  /// No description provided for @onlineServerProgress.
  ///
  /// In en, this message translates to:
  /// **'Preparing the next move…'**
  String get onlineServerProgress;

  /// No description provided for @onlineBotSetupHelp.
  ///
  /// In en, this message translates to:
  /// **'Bots fill the other seats. Choose a difficulty, then start the match from your room.'**
  String get onlineBotSetupHelp;

  /// No description provided for @onlineBotSeat.
  ///
  /// In en, this message translates to:
  /// **'Bot · {difficulty}'**
  String onlineBotSeat(String difficulty);

  /// No description provided for @onlineRoomGamesRange.
  ///
  /// In en, this message translates to:
  /// **'Enter a whole number of games from 1 to 100.'**
  String get onlineRoomGamesRange;

  /// No description provided for @onlineRoundDeparture.
  ///
  /// In en, this message translates to:
  /// **'Round complete · Choose whether to stay'**
  String get onlineRoundDeparture;

  /// No description provided for @onlineRoundWaiting.
  ///
  /// In en, this message translates to:
  /// **'Round complete · Waiting for players'**
  String get onlineRoundWaiting;

  /// No description provided for @onlineRoundChoiceHelp.
  ///
  /// In en, this message translates to:
  /// **'Choose whether to stay for the next round. Leaving ends this game for you and removes positive current-game score and available points; debts remain. You can rejoin the next game.'**
  String get onlineRoundChoiceHelp;

  /// No description provided for @onlineDepartureRecorded.
  ///
  /// In en, this message translates to:
  /// **'Your choice is saved. Waiting for other players.'**
  String get onlineDepartureRecorded;

  /// No description provided for @onlineStayInGame.
  ///
  /// In en, this message translates to:
  /// **'Stay in game'**
  String get onlineStayInGame;

  /// No description provided for @onlineLeaveGame.
  ///
  /// In en, this message translates to:
  /// **'Leave game'**
  String get onlineLeaveGame;

  /// No description provided for @activateCard.
  ///
  /// In en, this message translates to:
  /// **'Card actions'**
  String get activateCard;

  /// No description provided for @cardRoleSource.
  ///
  /// In en, this message translates to:
  /// **'Source card'**
  String get cardRoleSource;

  /// No description provided for @cardRoleTarget.
  ///
  /// In en, this message translates to:
  /// **'Target card'**
  String get cardRoleTarget;

  /// No description provided for @cardRolePayment.
  ///
  /// In en, this message translates to:
  /// **'Payment card'**
  String get cardRolePayment;

  /// No description provided for @cardRoleModifier.
  ///
  /// In en, this message translates to:
  /// **'Modifier card'**
  String get cardRoleModifier;

  /// No description provided for @cardGestureHint.
  ///
  /// In en, this message translates to:
  /// **'Tap or Space to select. Double tap or Enter for card actions. I to inspect.'**
  String get cardGestureHint;

  /// No description provided for @chooseDestination.
  ///
  /// In en, this message translates to:
  /// **'Choose destination'**
  String get chooseDestination;

  /// No description provided for @inspectPublicCards.
  ///
  /// In en, this message translates to:
  /// **'Inspect public cards'**
  String get inspectPublicCards;

  /// No description provided for @inspectSelectedCards.
  ///
  /// In en, this message translates to:
  /// **'Inspect selected cards'**
  String get inspectSelectedCards;

  /// No description provided for @onlineCardUnavailable.
  ///
  /// In en, this message translates to:
  /// **'Card no longer available'**
  String get onlineCardUnavailable;

  /// No description provided for @onlineGameOptions.
  ///
  /// In en, this message translates to:
  /// **'Game options'**
  String get onlineGameOptions;

  /// No description provided for @onlineRecords.
  ///
  /// In en, this message translates to:
  /// **'Offers, promises & scores'**
  String get onlineRecords;

  /// No description provided for @onlineCardGestures.
  ///
  /// In en, this message translates to:
  /// **'Card gestures'**
  String get onlineCardGestures;

  /// No description provided for @onlinePendingDecision.
  ///
  /// In en, this message translates to:
  /// **'Pending decision'**
  String get onlinePendingDecision;

  /// No description provided for @onlineResultsSettlement.
  ///
  /// In en, this message translates to:
  /// **'Results & settlement'**
  String get onlineResultsSettlement;

  /// No description provided for @onlineCardInvisible.
  ///
  /// In en, this message translates to:
  /// **'Card no longer visible'**
  String get onlineCardInvisible;

  /// No description provided for @onlineChooseMove.
  ///
  /// In en, this message translates to:
  /// **'Choose a move'**
  String get onlineChooseMove;

  /// No description provided for @onlineCloseCardContext.
  ///
  /// In en, this message translates to:
  /// **'Close card context'**
  String get onlineCloseCardContext;

  /// No description provided for @onlineNoMoveDestination.
  ///
  /// In en, this message translates to:
  /// **'No move at this destination. Select other cards or a destination on the board.'**
  String get onlineNoMoveDestination;

  /// No description provided for @onlineCoupCardConfirmation.
  ///
  /// In en, this message translates to:
  /// **'Declare Coup with both Hearts Queens and both Diamond Queens.'**
  String get onlineCoupCardConfirmation;

  /// No description provided for @onlineChooseBoard.
  ///
  /// In en, this message translates to:
  /// **'Select cards or a destination'**
  String get onlineChooseBoard;

  /// No description provided for @onlineNoCardResponse.
  ///
  /// In en, this message translates to:
  /// **'No card response is available. Pass to continue.'**
  String get onlineNoCardResponse;

  /// No description provided for @onlineCardResponseHelp.
  ///
  /// In en, this message translates to:
  /// **'Double tap a matching card to respond, or choose Pass response.'**
  String get onlineCardResponseHelp;

  /// No description provided for @onlineKidnapperExposedChoice.
  ///
  /// In en, this message translates to:
  /// **'Choose an exposed royal for Kidnapper'**
  String get onlineKidnapperExposedChoice;

  /// No description provided for @onlineResolveAuthorizedStage.
  ///
  /// In en, this message translates to:
  /// **'Resolve only this authorized decision stage.'**
  String get onlineResolveAuthorizedStage;

  /// No description provided for @onlineContinueWithoutCard.
  ///
  /// In en, this message translates to:
  /// **'Continue without spending a card'**
  String get onlineContinueWithoutCard;

  /// No description provided for @onlineJusticeRecordedOutcome.
  ///
  /// In en, this message translates to:
  /// **'Request the recorded concealed outcome. No hidden candidate is selected.'**
  String get onlineJusticeRecordedOutcome;

  /// No description provided for @onlineJusticeConcealedChoice.
  ///
  /// In en, this message translates to:
  /// **'Choose concealed Justice outcome'**
  String get onlineJusticeConcealedChoice;

  /// No description provided for @onlineConsumedQueen.
  ///
  /// In en, this message translates to:
  /// **'Consumed Queen'**
  String get onlineConsumedQueen;

  /// No description provided for @onlineModifier.
  ///
  /// In en, this message translates to:
  /// **'Modifier'**
  String get onlineModifier;

  /// No description provided for @onlineDiamondPayment.
  ///
  /// In en, this message translates to:
  /// **'Diamond payment'**
  String get onlineDiamondPayment;

  /// No description provided for @onlineBarricadeOrder.
  ///
  /// In en, this message translates to:
  /// **'Queen first, then five returned cards'**
  String get onlineBarricadeOrder;

  /// No description provided for @onlineDoppelgangerPair.
  ///
  /// In en, this message translates to:
  /// **'Doppelganger pair'**
  String get onlineDoppelgangerPair;

  /// No description provided for @onlineComboMembers.
  ///
  /// In en, this message translates to:
  /// **'Combo members'**
  String get onlineComboMembers;

  /// No description provided for @onlineDoppelgangerFormation.
  ///
  /// In en, this message translates to:
  /// **'Doppelganger formation'**
  String get onlineDoppelgangerFormation;

  /// No description provided for @onlineReplacementRank.
  ///
  /// In en, this message translates to:
  /// **'Replacement rank'**
  String get onlineReplacementRank;

  /// No description provided for @onlineReplacementSuit.
  ///
  /// In en, this message translates to:
  /// **'Replacement suit'**
  String get onlineReplacementSuit;

  /// No description provided for @onlineChoosePublicTarget.
  ///
  /// In en, this message translates to:
  /// **'Choose your public target'**
  String get onlineChoosePublicTarget;

  /// No description provided for @onlineInfiltratorCommit.
  ///
  /// In en, this message translates to:
  /// **'Infiltrator · return the attached Jack at declaration'**
  String get onlineInfiltratorCommit;

  /// No description provided for @onlineExileCommit.
  ///
  /// In en, this message translates to:
  /// **'Exile · spend this round’s bypass'**
  String get onlineExileCommit;

  /// No description provided for @onlinePrivateTrade.
  ///
  /// In en, this message translates to:
  /// **'Private card trade'**
  String get onlinePrivateTrade;

  /// No description provided for @onlineProposalRevision.
  ///
  /// In en, this message translates to:
  /// **'Proposal revision'**
  String get onlineProposalRevision;

  /// No description provided for @onlineNewOffer.
  ///
  /// In en, this message translates to:
  /// **'New offer'**
  String get onlineNewOffer;

  /// No description provided for @onlinePrivateOfferTerms.
  ///
  /// In en, this message translates to:
  /// **'Only named parties see these terms. Offering reserves no cards. Acceptance uses this exact revision.'**
  String get onlinePrivateOfferTerms;

  /// No description provided for @onlineVictoryThreshold.
  ///
  /// In en, this message translates to:
  /// **'Victory threshold'**
  String get onlineVictoryThreshold;

  /// No description provided for @onlineNumberDiamondThreshold.
  ///
  /// In en, this message translates to:
  /// **'13 number Diamonds'**
  String get onlineNumberDiamondThreshold;

  /// No description provided for @onlineRoyalThreshold.
  ///
  /// In en, this message translates to:
  /// **'15 Royals'**
  String get onlineRoyalThreshold;

  /// No description provided for @onlineAceCombatValue.
  ///
  /// In en, this message translates to:
  /// **'Ace combat value'**
  String get onlineAceCombatValue;

  /// No description provided for @onlineAdvancementValue.
  ///
  /// In en, this message translates to:
  /// **'Advancement · +10'**
  String get onlineAdvancementValue;

  /// No description provided for @onlineDiamondBonusRange.
  ///
  /// In en, this message translates to:
  /// **'Diamond bonus (1–10)'**
  String get onlineDiamondBonusRange;

  /// No description provided for @onlinePurchasePriceRange.
  ///
  /// In en, this message translates to:
  /// **'Purchase price (5–10)'**
  String get onlinePurchasePriceRange;

  /// No description provided for @onlineJackRank.
  ///
  /// In en, this message translates to:
  /// **'Jack'**
  String get onlineJackRank;

  /// No description provided for @onlineQueenRank.
  ///
  /// In en, this message translates to:
  /// **'Queen'**
  String get onlineQueenRank;

  /// No description provided for @onlineKingRank.
  ///
  /// In en, this message translates to:
  /// **'King'**
  String get onlineKingRank;

  /// No description provided for @onlineCardGestureHelp.
  ///
  /// In en, this message translates to:
  /// **'Tap to select exact cards; tap additional cards to build a bundle. Double tap or use keyboard activation for the selected card action. Drag to preview the exact move, costs and destination; release a complete legal move to submit it once. Choices appear only for missing parameters or ambiguity. Inspect cards independently. Escape cancels an unsubmitted selection.'**
  String get onlineCardGestureHelp;

  /// No description provided for @onlineReviewBeforeConfirm.
  ///
  /// In en, this message translates to:
  /// **'Review {action} before confirming.'**
  String onlineReviewBeforeConfirm(String action);

  /// No description provided for @onlineChooseRoleBoard.
  ///
  /// In en, this message translates to:
  /// **'Choose {role}'**
  String onlineChooseRoleBoard(String role);

  /// No description provided for @onlineResolveKind.
  ///
  /// In en, this message translates to:
  /// **'Choose how to resolve {kind}'**
  String onlineResolveKind(String kind);

  /// No description provided for @onlineSourceCards.
  ///
  /// In en, this message translates to:
  /// **'Source cards · {cards}'**
  String onlineSourceCards(String cards);

  /// No description provided for @onlineFormationSummary.
  ///
  /// In en, this message translates to:
  /// **'Formation · {formation}'**
  String onlineFormationSummary(String formation);

  /// No description provided for @onlineConfirmAction.
  ///
  /// In en, this message translates to:
  /// **'Confirm {action}'**
  String onlineConfirmAction(String action);

  /// No description provided for @onlineSeriesSummary.
  ///
  /// In en, this message translates to:
  /// **'Series · {suit}'**
  String onlineSeriesSummary(String suit);

  /// No description provided for @onlineComboSummary.
  ///
  /// In en, this message translates to:
  /// **'Combo · {combo}'**
  String onlineComboSummary(String combo);

  /// No description provided for @onlineProtectionSummary.
  ///
  /// In en, this message translates to:
  /// **'Protection · {target}'**
  String onlineProtectionSummary(String target);

  /// No description provided for @onlineAttachTo.
  ///
  /// In en, this message translates to:
  /// **'Attach to {suit}'**
  String onlineAttachTo(String suit);

  /// No description provided for @onlineOfferReplacement.
  ///
  /// In en, this message translates to:
  /// **'Replace offer to seat {seat} · revision {revision}'**
  String onlineOfferReplacement(String seat, String revision);

  /// No description provided for @onlineCommittedClubs.
  ///
  /// In en, this message translates to:
  /// **'Committed Clubs · {cards}'**
  String onlineCommittedClubs(String cards);

  /// No description provided for @onlineFrozenTargets.
  ///
  /// In en, this message translates to:
  /// **'Frozen targets · {cards}'**
  String onlineFrozenTargets(String cards);

  /// No description provided for @onlineCombatSeats.
  ///
  /// In en, this message translates to:
  /// **'Seat {attacker} attacks seat {defender}'**
  String onlineCombatSeats(String attacker, String defender);

  /// No description provided for @onlinePendingSeatAction.
  ///
  /// In en, this message translates to:
  /// **'Seat {actor} · {action}'**
  String onlinePendingSeatAction(String actor, String action);

  /// No description provided for @onlinePendingSeatTarget.
  ///
  /// In en, this message translates to:
  /// **'Seat {actor} · {action} → Seat {target}'**
  String onlinePendingSeatTarget(String actor, String action, String target);

  /// No description provided for @onlineCardRoleSummary.
  ///
  /// In en, this message translates to:
  /// **'{role} · {cards}'**
  String onlineCardRoleSummary(String role, String cards);

  /// No description provided for @onlineChooseSupportedSeries.
  ///
  /// In en, this message translates to:
  /// **'Choose a supported series'**
  String get onlineChooseSupportedSeries;

  /// No description provided for @onlineChooseTargetCard.
  ///
  /// In en, this message translates to:
  /// **'Choose the target card'**
  String get onlineChooseTargetCard;

  /// No description provided for @onlineChooseDiamondPayment.
  ///
  /// In en, this message translates to:
  /// **'Select a Diamond payment'**
  String get onlineChooseDiamondPayment;

  /// No description provided for @onlineChooseAttackingClubs.
  ///
  /// In en, this message translates to:
  /// **'Select attacking Clubs'**
  String get onlineChooseAttackingClubs;

  /// No description provided for @onlineChooseAce.
  ///
  /// In en, this message translates to:
  /// **'Select an Ace modifier'**
  String get onlineChooseAce;

  /// No description provided for @onlineChooseProtection.
  ///
  /// In en, this message translates to:
  /// **'Choose a protection target'**
  String get onlineChooseProtection;

  /// No description provided for @onlineChoosePayment.
  ///
  /// In en, this message translates to:
  /// **'Select Spade payment cards'**
  String get onlineChoosePayment;

  /// No description provided for @onlineChooseFormationMembers.
  ///
  /// In en, this message translates to:
  /// **'Select exact combo members'**
  String get onlineChooseFormationMembers;

  /// No description provided for @onlineChooseKings.
  ///
  /// In en, this message translates to:
  /// **'Select the two Doppelganger kings'**
  String get onlineChooseKings;

  /// No description provided for @onlineChooseRequestedCards.
  ///
  /// In en, this message translates to:
  /// **'Choose requested cards'**
  String get onlineChooseRequestedCards;

  /// No description provided for @onlineReleaseMove.
  ///
  /// In en, this message translates to:
  /// **'Release to submit {action}'**
  String onlineReleaseMove(String action);

  /// No description provided for @onlineReleaseChoice.
  ///
  /// In en, this message translates to:
  /// **'Release to choose {action}'**
  String onlineReleaseChoice(String action);

  /// No description provided for @onlineDropSeat.
  ///
  /// In en, this message translates to:
  /// **'Seat {seat}'**
  String onlineDropSeat(int seat);

  /// No description provided for @onlineLoanDropTerms.
  ///
  /// In en, this message translates to:
  /// **'No cards requested. Exact revision acceptance and successful responses resolve the loan and exclusive alliance together.'**
  String get onlineLoanDropTerms;

  /// No description provided for @turnDrawToHand.
  ///
  /// In en, this message translates to:
  /// **'Turn draw: one card added to your hand.'**
  String get turnDrawToHand;

  /// No description provided for @turnDrawToConcealedAce.
  ///
  /// In en, this message translates to:
  /// **'Turn draw: one card added to your concealed Ace area.'**
  String get turnDrawToConcealedAce;

  /// No description provided for @turnDrawBySeat.
  ///
  /// In en, this message translates to:
  /// **'Seat {seat} drew one card.'**
  String turnDrawBySeat(int seat);
}

class _CgmsLocalizationsDelegate
    extends LocalizationsDelegate<CgmsLocalizations> {
  const _CgmsLocalizationsDelegate();

  @override
  Future<CgmsLocalizations> load(Locale locale) {
    return SynchronousFuture<CgmsLocalizations>(
      lookupCgmsLocalizations(locale),
    );
  }

  @override
  bool isSupported(Locale locale) =>
      <String>['en'].contains(locale.languageCode);

  @override
  bool shouldReload(_CgmsLocalizationsDelegate old) => false;
}

CgmsLocalizations lookupCgmsLocalizations(Locale locale) {
  // Lookup logic when only language code is specified.
  switch (locale.languageCode) {
    case 'en':
      return CgmsLocalizationsEn();
  }

  throw FlutterError(
    'CgmsLocalizations.delegate failed to load unsupported locale "$locale". This is likely '
    'an issue with the localizations generation tool. Please file an issue '
    'on GitHub with a reproducible sample app and the gen-l10n configuration '
    'that was used.',
  );
}
