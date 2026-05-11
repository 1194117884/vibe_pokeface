package doudizhu

import (
	"math/rand"
	"sort"
)

// Card represents a playing card in a standard 54-card deck (52 suited cards + 2 jokers).
// Card ID encoding: 0-51 = standard cards, 52=small joker, 53=big joker
// 0-12: ♠3-♠2, 13-25: ♥3-♥2, 26-38: ♣3-♣2, 39-51: ♦3-♦2
type Card struct {
	ID int `json:"id"`
}

var suitChars = []string{"♠", "♥", "♣", "♦"}
var rankChars = []string{"3", "4", "5", "6", "7", "8", "9", "10", "J", "Q", "K", "A", "2"}

const handSize = 17

// Suit returns the suit index of the card: 0=♠, 1=♥, 2=♣, 3=♦, 4=joker.
func (c Card) Suit() int {
	if c.ID >= 52 {
		return 4 // joker suit
	}
	return c.ID / 13
}

// Rank returns the rank value of the card.
// Standard cards: 3→3, 4→4, ..., K→13, A→14, 2→15. Jokers: small→16, big→17.
func (c Card) Rank() int {
	if c.ID == 52 {
		return 16 // small joker
	}
	if c.ID == 53 {
		return 17 // big joker
	}
	return (c.ID % 13) + 3 // 3→3, 4→4, ..., 2→15
}

// Display returns a human-readable string representation of the card.
func (c Card) Display() string {
	if c.ID == 52 {
		return "🃏"
	}
	if c.ID == 53 {
		return "👑"
	}
	return suitChars[c.Suit()] + rankChars[c.ID%13]
}

// NewDeck creates and returns a new 54-card deck (52 standard cards + 2 jokers).
func NewDeck() []Card {
	cards := make([]Card, 54)
	for i := 0; i < 54; i++ {
		cards[i] = Card{ID: i}
	}
	return cards
}

// Shuffle randomizes the order of cards in the deck using Fisher-Yates shuffle.
func Shuffle(deck []Card) {
	rand.Shuffle(len(deck), func(i, j int) {
		deck[i], deck[j] = deck[j], deck[i]
	})
}

// SortCards sorts cards in descending rank order (highest first).
// Cards of the same rank are ordered by suit: ♠, ♥, ♣, ♦.
func SortCards(cards []Card) {
	sort.Slice(cards, func(i, j int) bool {
		ri, rj := cards[i].Rank(), cards[j].Rank()
		if ri != rj {
			return ri > rj // higher rank first
		}
		return cards[i].Suit() < cards[j].Suit()
	})
}

// Deal splits a 54-card deck into three 17-card hands and 3 remaining cards.
// The returned slices are copies and do not alias the input deck.
func Deal(deck []Card) (hand1, hand2, hand3, remaining []Card) {
	hand1 = make([]Card, handSize)
	copy(hand1, deck[0:handSize])
	hand2 = make([]Card, handSize)
	copy(hand2, deck[handSize:2*handSize])
	hand3 = make([]Card, handSize)
	copy(hand3, deck[2*handSize:3*handSize])
	remaining = make([]Card, 3)
	copy(remaining, deck[51:54])
	return
}
