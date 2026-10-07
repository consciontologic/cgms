// Package game contains deterministic CGMS domain operations. Adapters own I/O.
package game

import "fmt"

type Suit string

const (
	Hearts   Suit = "hearts"
	Clubs    Suit = "clubs"
	Spades   Suit = "spades"
	Diamonds Suit = "diamonds"
)

type Card struct {
	ID   string `json:"id"`
	Deck int    `json:"deck"`
	Suit Suit   `json:"suit"`
	Rank int    `json:"rank"`
}

func Deck() []Card {
	cards := make([]Card, 0, 104)
	for deck := 1; deck <= 2; deck++ {
		for _, suit := range []Suit{Hearts, Clubs, Spades, Diamonds} {
			for rank := 1; rank <= 13; rank++ {
				cards = append(cards, Card{fmt.Sprintf("deck-%d-%s-%02d", deck, suit, rank), deck, suit, rank})
			}
		}
	}
	return cards
}
