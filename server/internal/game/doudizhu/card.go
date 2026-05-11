package doudizhu

import (
	"math/rand"
	"sort"
	"time"
)

// Card ID encoding: 0-51 = standard cards, 52=small joker, 53=big joker
// 0-12: ♠3-♠2, 13-25: ♥3-♥2, 26-38: ♣3-♣2, 39-51: ♦3-♦2
type Card struct {
	ID int `json:"id"`
}

var suitChars = []string{"♠", "♥", "♣", "♦"}
var rankChars = []string{"3", "4", "5", "6", "7", "8", "9", "10", "J", "Q", "K", "A", "2"}

func (c Card) Suit() int {
	if c.ID >= 52 {
		return 4 // joker suit
	}
	return c.ID / 13
}

func (c Card) Rank() int {
	if c.ID == 52 {
		return 16 // small joker
	}
	if c.ID == 53 {
		return 17 // big joker
	}
	return (c.ID % 13) + 3 // 3→3, 4→4, ..., 2→15
}

func (c Card) Display() string {
	if c.ID == 52 {
		return "🃏"
	}
	if c.ID == 53 {
		return "👑"
	}
	return suitChars[c.Suit()] + rankChars[c.ID%13]
}

func NewDeck() []Card {
	cards := make([]Card, 54)
	for i := 0; i < 54; i++ {
		cards[i] = Card{ID: i}
	}
	return cards
}

func Shuffle(deck []Card) {
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	rng.Shuffle(len(deck), func(i, j int) {
		deck[i], deck[j] = deck[j], deck[i]
	})
}

func SortCards(cards []Card) {
	sort.Slice(cards, func(i, j int) bool {
		ri, rj := cards[i].Rank(), cards[j].Rank()
		if ri != rj {
			return ri > rj // higher rank first
		}
		return cards[i].Suit() < cards[j].Suit()
	})
}

func Deal(deck []Card) (hand1, hand2, hand3, remaining []Card) {
	return deck[0:17], deck[17:34], deck[34:51], deck[51:54]
}
