package bots

import (
	"errors"
	"fmt"
	"github.com/metaphy6/cgms/backend/internal/game"
	"github.com/metaphy6/cgms/backend/internal/randomstream"
	"slices"
	"sort"
	"strings"
)

const GameplayFinanceVersion = "authorized-finance-v1"

type GameplayForgivenessRequest struct {
	DebtID   string      `json:"debt_id"`
	Debtor   int         `json:"debtor"`
	Amount   game.Amount `json:"amount"`
	Consents []int       `json:"consents"`
}
type GameplayFinanceObservation struct {
	ForgivenessRequests []GameplayForgivenessRequest `json:"forgiveness_requests,omitempty"`
	GameID              string                       `json:"game_id"`
	Seat                int                          `json:"seat"` // One-based throughout this bot interface.
	Seats               int                          `json:"seats"`
	Phase               string                       `json:"phase"`
	Cash                game.Amount                  `json:"cash"`
	Score               game.Amount                  `json:"score"`
	Debts               []game.FinancialDebt         `json:"debts"`    // Party-only; underlying debt seats remain zero-based.
	Promises            []game.Promise               `json:"promises"` // Party-only; underlying promise seats remain zero-based.
	Finished            bool                         `json:"finished"`
	Waiting             string                       `json:"waiting,omitempty"`
	MenuVersion         string                       `json:"menu_version"`
	Legal               []game.LifecycleOperation    `json:"legal"`
}
type GameplayFinanceDecision struct {
	Operation   game.LifecycleOperation `json:"operation"`
	Reasons     []string                `json:"reasons"`
	Operations  int                     `json:"operations"`
	Policy      string                  `json:"policy"`
	MenuVersion string                  `json:"menu_version"`
}

func ProjectGameplayFinance(l game.Lifecycle, seat int) (GameplayFinanceObservation, error) {
	if l.Validate() != nil || seat < 1 || seat > len(l.Board.Players) {
		return GameplayFinanceObservation{}, errors.New("invalid financial observer")
	}
	own := seat - 1
	o := GameplayFinanceObservation{GameID: l.Board.GameID, Seat: seat, Seats: len(l.Board.Players), Phase: l.Ledger.Phase, Cash: l.Ledger.Cash[own], Score: l.Ledger.Scores[own], Finished: l.Ledger.Finished[own], MenuVersion: GameplayFinanceVersion}
	for _, d := range l.Ledger.Debts {
		if d.Debtor == own || d.Creditor == own {
			o.Debts = append(o.Debts, d)
		}
	}
	for _, request := range l.ForgivenessConsents {
		for _, debt := range l.Ledger.Debts {
			if debt.ID == request.DebtID && debt.Creditor == -1 && slices.Contains(request.Seats, debt.Debtor+1) {
				o.ForgivenessRequests = append(o.ForgivenessRequests, GameplayForgivenessRequest{debt.ID, debt.Debtor + 1, request.Amount, slices.Clone(request.Seats)})
			}
		}
	}
	for _, p := range l.Promises.Promises {
		if p.Payer == own || p.Recipient == own {
			o.Promises = append(o.Promises, p)
		}
	}
	if l.Automatic != nil {
		o.Waiting = "automatic-settlement"
		return o, nil
	}
	if l.Board.Pending != nil {
		o.Waiting = "board-action"
		return o, nil
	}
	if o.Finished {
		return o, nil
	}
	revision := game.Digest(o)
	seen := map[string]bool{}
	add := func(op game.LifecycleOperation) {
		op.Actor = seat
		op.GameID = o.GameID
		op.ID = "finance-menu/" + revision + "/" + game.Digest(op)
		key := game.Digest(op)
		if seen[key] {
			return
		}
		if _, e := l.ApplyOperation(op); e == nil {
			o.Legal = append(o.Legal, op)
			seen[key] = true
		}
	}
	for _, p := range o.Promises {
		if p.Recipient == own {
			if p.Status == "offered" {
				add(game.LifecycleOperation{Kind: "promise-accept", PromiseID: p.ID})
			}
			if p.Status == "final-offered" {
				add(game.LifecycleOperation{Kind: "promise-final-answer", PromiseID: p.ID, Accept: true})
				add(game.LifecycleOperation{Kind: "promise-final-answer", PromiseID: p.ID, Accept: false})
			}
		}
		if p.Payer == own && p.Triggered {
			add(game.LifecycleOperation{Kind: "promise-pay", PromiseID: p.ID})
			add(game.LifecycleOperation{Kind: "promise-refuse", PromiseID: p.ID})
			if p.Status == "accepted" {
				amounts := []game.Amount{game.IntAmount(0), p.Promised.Mul(financeFraction(2, 5)), o.Cash}
				for _, a := range amounts {
					if a.Cmp(p.Promised) < 0 {
						add(game.LifecycleOperation{Kind: "promise-final-offer", PromiseID: p.ID, Amount: a})
					}
				}
			}
		}
	}
	for _, d := range o.Debts {
		if d.Creditor == own || d.Creditor == -1 && d.Debtor == own {
			already := false
			for _, request := range o.ForgivenessRequests {
				if request.DebtID == d.ID && request.Amount.Cmp(d.Remaining) == 0 && slices.Contains(request.Consents, seat) {
					already = true
				}
			}
			if already {
				continue
			}
			add(game.LifecycleOperation{Kind: "forgive", DebtID: d.ID, Amount: d.Remaining, Seats: []int{seat}})
		}
	}
	for _, request := range o.ForgivenessRequests {
		if !slices.Contains(request.Consents, seat) {
			add(game.LifecycleOperation{Kind: "forgive", DebtID: request.DebtID, Amount: request.Amount})
		}
	}
	if o.Phase == "playing" {
		for recipient := 1; recipient <= o.Seats; recipient++ {
			if recipient == seat {
				continue
			}
			for _, award := range []string{fmt.Sprintf("%s/ordinary-end/seat-%d", o.GameID, seat), fmt.Sprintf("%s/coup/seat-%d", o.GameID, seat)} {
				id := fmt.Sprintf("promise-menu/%s/%d/%s", revision, recipient, award)
				add(game.LifecycleOperation{Kind: "promise-offer", PromiseID: id, Recipient: recipient, AwardID: award, Mode: "percentage", Amount: financeFraction(2, 5)})
			}
		}
	}
	if o.Cash.Sign() > 0 {
		for recipient := 1; recipient <= o.Seats; recipient++ {
			if recipient != seat {
				add(game.LifecycleOperation{Kind: "voluntary-transfer", Recipient: recipient, Amount: o.Cash})
			}
		}
	}
	add(game.LifecycleOperation{Kind: "finish-settlement"})
	sort.Slice(o.Legal, func(i, j int) bool { return game.Digest(o.Legal[i]) < game.Digest(o.Legal[j]) })
	return o, nil
}
func ChooseGameplayFinancial(policy string, o GameplayFinanceObservation, r *randomstream.Stream, budget int) (GameplayFinanceDecision, error) {
	d := GameplayFinanceDecision{Policy: policy, MenuVersion: GameplayFinanceVersion}
	policy = gameplayScoringPolicy(policy)
	if !opportunityPolicy(policy) && policy != "gameplay-random@v2" && policy != "economic@v1" && policy != "pressure@v1" && policy != "gameplay-random@v1" && policy != "legal-random@v1" && policy != "heuristic@v1" {
		return d, errors.New("unknown financial policy")
	}
	if len(o.Legal) == 0 {
		return d, ErrUnsupported
	}
	if budget < len(o.Legal) {
		return d, ErrBudget
	}
	menu := slices.Clone(o.Legal)
	for _, op := range menu {
		if op.Actor != o.Seat || op.GameID != o.GameID {
			return d, errors.New("unauthorized financial operation")
		}
		d.Operations++
	}
	sort.Slice(menu, func(i, j int) bool { return financeActionKey(o, menu[i]) < financeActionKey(o, menu[j]) })
	pick := 0
	if policy == "gameplay-random@v1" || policy == "gameplay-random@v2" || policy == "legal-random@v1" {
		if r == nil {
			return d, errors.New("missing financial stream")
		}
		var j uint64
		var ops int
		var e error
		if policy == "gameplay-random@v2" {
			var index int
			index, ops, e = sampleFinancialKinds(menu, r.Uint64, budget-d.Operations)
			j = uint64(index)
		} else {
			j, ops, e = boundedSample(r.Uint64, uint64(len(menu)), budget-d.Operations)
		}
		if e != nil {
			return d, e
		}
		d.Operations += ops
		pick = int(j)
		d.Reasons = []string{"explicit uniform choice among authorized finite financial operations; no inferred consent or refusal"}
		if policy == "gameplay-random@v2" {
			d.Reasons = []string{"uniform authorized operation kind, then uniform entry within that kind; explicit decline has equal class probability; no inferred consent or refusal"}
		}
	} else {
		if budget < 2*len(menu) {
			return d, ErrBudget
		}
		best := -100000
		for i, op := range menu {
			score := -100
			reason := "avoid unsolicited gifts and promise churn"
			switch op.Kind {
			case "departure-choice":
				score = -200
				if !op.Accept {
					score = 0
				}
				reason = "explicitly remain in the game at departure boundary"
			case "decline-financial-opportunity":
				score = 0
				reason = "explicitly decline this optional financial opportunity"
			case "finish-settlement":
				score = 0
				reason = "explicitly finish after required financial decisions resolve"
			case "promise-accept":
				score = 100
				reason = "accept offered incoming reward without creating a debt"
			case "promise-pay":
				score = 200
				reason = "honor an affordable triggered commitment after existing debts"
				if policy == "pressure@v1" {
					score = 30
				}
			case "promise-final-offer":
				score = 150
				reason = "offer a reduced exact payable amount instead of silent refusal"
				if op.Amount.Cmp(o.Cash) > 0 {
					score = -200
					reason = "avoid an unaffordable final offer"
				}
				if op.Amount.Sign() == 0 {
					score = 100
				}
				if policy == "pressure@v1" {
					score = 20
				}
			case "promise-final-answer":
				if op.Accept {
					score = 100
					reason = "explicitly accept a reduced incoming payment; gross receipts repay own debts"
				} else {
					score = -10
					reason = "reject reduced offer only as an explicit policy choice"
				}
			case "promise-refuse":
				score = 50
				reason = "explicit refusal only when selected from the legal menu; records reputation consequence"
				if policy == "pressure@v1" {
					score = 80
					reason = "opportunistic pressure policy preserves spendable points at explicit reputation cost"
				}
			case "forgive":
				score = -50
				reason = "retain player receivable absent an alliance benefit model"
			}
			if op.Kind == "promise-final-offer" {
				for _, debt := range o.Debts {
					if debt.Debtor == o.Seat-1 && debt.Remaining.Sign() > 0 {
						score = -200
						reason = "do not offer a voluntary payment while older debts prohibit payment"
					}
				}
			}
			if opportunityPolicy(policy) && policy != "opportunity-no-finance@v1" {
				score, reason = opportunityFinanceScore(o, op, score, reason)
			}
			d.Operations++
			if score > best {
				best = score
				pick = i
				d.Reasons = []string{fmt.Sprintf("%s: priority %d", reason, score), "own cash and party-only debts/promises; no opponent hidden state", "automatic repayments precede voluntary transfer; forgiveness generates no cash"}
			}
		}
	}
	d.Operation = menu[pick]
	d.Operation.Seats = slices.Clone(menu[pick].Seats)
	return d, nil
}

func financeFraction(a, b int64) game.Amount {
	v, _ := game.IntAmount(a).Quo(game.IntAmount(b))
	return v
}

func financeActionKey(o GameplayFinanceObservation, op game.LifecycleOperation) string {
	var promise *game.Promise
	op.ID = ""
	op.GameID = ""
	op.AwardID = strings.TrimPrefix(op.AwardID, o.GameID)
	if op.PromiseID != "" {
		for _, p := range o.Promises {
			if p.ID == op.PromiseID {
				op.AwardID = strings.TrimPrefix(p.AwardID, o.GameID)
				op.Recipient = p.Recipient + 1
				op.Mode = p.Mode
				v := p
				v.ID = ""
				v.AwardID = strings.TrimPrefix(v.AwardID, o.GameID)
				promise = &v
				break
			}
		}
		op.PromiseID = ""
	}
	if op.DebtID != "" {
		for _, d := range o.Debts {
			if d.ID == op.DebtID {
				op.Recipient = d.Debtor + 1
				op.DebtID = fmt.Sprintf("%d/%d/%d", d.Debtor, d.Creditor, d.Sequence)
				break
			}
		}
	}
	return game.Digest(struct {
		Operation game.LifecycleOperation
		Promise   *game.Promise
	}{op, promise})
}

// sampleFinancialKinds preserves menu order within each class and sorts class
// names. Each class has probability 1/K regardless of its number of offers.
func sampleFinancialKinds(menu []game.LifecycleOperation, next func() uint64, budget int) (int, int, error) {
	if len(menu) == 0 {
		return 0, 0, ErrUnsupported
	}
	if budget < len(menu) {
		return 0, 0, ErrBudget
	}
	classes := map[string][]int{}
	for i, op := range menu {
		classes[op.Kind] = append(classes[op.Kind], i)
	}
	kinds := make([]string, 0, len(classes))
	for kind := range classes {
		kinds = append(kinds, kind)
	}
	sort.Strings(kinds)
	used := len(menu)
	class, ops, err := boundedSample(next, uint64(len(kinds)), budget-used)
	used += ops
	if err != nil {
		return 0, used, err
	}
	members := classes[kinds[class]]
	entry, ops, err := boundedSample(next, uint64(len(members)), budget-used)
	used += ops
	if err != nil {
		return 0, used, err
	}
	return members[entry], used, nil
}
