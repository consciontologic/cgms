package game

import (
	"encoding/json"
	"testing"
)

func TestExtendedMetadataRejectsMalformedCheckpoint(t *testing.T) {
	cases := map[string]func(*State){
		"formation-controller": func(s *State) {
			s.Formations = []FormationRecord{{ID: "f", Controller: 99, Spec: FormationSpec{Kind: "justice", Cards: []string{"deck-1-hearts-12"}}}}
		},
		"formation-unknown-card": func(s *State) {
			s.Formations = []FormationRecord{{ID: "f", Controller: 1, Spec: FormationSpec{Kind: "justice", Cards: []string{"missing"}}}}
		},
		"proposal-party": func(s *State) {
			s.Proposals = []Proposal{{ID: "p", GameID: s.GameID, Revision: 1, Turn: 1, From: 99, Terms: ProposalTerms{To: 2, Give: []string{"deck-1-hearts-02"}}, Status: "offered"}}
		},
		"proposal-status": func(s *State) {
			s.Proposals = []Proposal{{ID: "p", GameID: s.GameID, Revision: 1, Turn: 1, From: 1, Terms: ProposalTerms{To: 2, Give: []string{"deck-1-hearts-02"}}, Status: "fabricated"}}
		},
		"opening-kind": func(s *State) {
			s.Pending = &PendingAction{ID: "p", Kind: "opening", Actor: 1, Opening: &OpeningIntent{Kind: "fabricated"}}
		},
		"ability-target": func(s *State) {
			s.Pending = &PendingAction{ID: "p", Kind: "ability", Actor: 1, Ability: &AbilityIntent{Kind: "main-inflation", TargetSeat: 99, AceID: "deck-1-spades-01"}}
		},
		"ability-card": func(s *State) {
			s.Pending = &PendingAction{ID: "p", Kind: "ability", Actor: 1, Ability: &AbilityIntent{Kind: "ringleader", Cards: []string{"missing"}}}
		},
	}
	for name, mutate := range cases {
		t.Run(name, func(t *testing.T) {
			s, _ := NewState(3, "metadata")
			mutate(&s)
			data, _ := json.Marshal(s)
			var restored State
			if e := json.Unmarshal(data, &restored); e != nil {
				t.Fatal(e)
			}
			if restored.Validate() == nil {
				t.Fatal("malformed serialized checkpoint accepted")
			}
		})
	}
}

func FuzzExtendedCheckpointValidation(f *testing.F) {
	s, _ := NewState(3, "fuzz-metadata")
	data, _ := json.Marshal(s)
	f.Add(data)
	s.Pending = &PendingAction{ID: "opening", Kind: "opening", Actor: 1, Opening: &OpeningIntent{Kind: "open-series", Suit: Hearts}}
	data, _ = json.Marshal(s)
	f.Add(data)
	s.Pending = nil
	s.Formations = []FormationRecord{{ID: "historical", Controller: 1, Spec: FormationSpec{Kind: "justice", Cards: []string{"deck-1-hearts-12"}}}}
	data, _ = json.Marshal(s)
	f.Add(data)
	f.Fuzz(func(t *testing.T, data []byte) {
		if len(data) > 100000 {
			return
		}
		var s State
		if json.Unmarshal(data, &s) != nil {
			return
		}
		_ = s.Validate()
	})
}

func TestExtendedMetadataKeepsBrokenHistoricalFormation(t *testing.T) {
	s, _ := NewState(3, "historical")
	// Historical records can survive all members leaving their original owner.
	s.Formations = []FormationRecord{{ID: "former-justice", Controller: 1, Spec: FormationSpec{Kind: "justice", Cards: []string{"deck-1-hearts-12", "deck-1-spades-12", "deck-1-clubs-12"}}}}
	if e := s.Validate(); e != nil {
		t.Fatal("historical record wrongly requires functioning custody", e)
	}
	s.Proposals = []Proposal{{ID: "old-offer", GameID: s.GameID, Turn: 1, Revision: 1, From: 1, Terms: ProposalTerms{To: 2, Give: []string{"deck-1-hearts-12"}}, Status: "expired"}}
	if e := s.Validate(); e != nil {
		t.Fatal("expired proposal wrongly requires current custody", e)
	}
	s.Proposals[0].Status = "accepted"
	s.Pending = &PendingAction{ID: "accepted", Kind: "transfer", Actor: 3, ProposalID: "old-offer"}
	if s.Validate() == nil {
		t.Fatal("forged acceptor identity accepted")
	}
}
