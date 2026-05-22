package dashengji

import (
	"math/rand"
	"sort"
)

// Card represents a playing card in a 3-deck (162-card) Dashengji game.
// ID encoding: face = ID % 54, copy = ID / 54
// face 0-51: standard cards (suit*13 + rank), 52: small joker, 53: big joker
type Card struct {
	ID int `json:"id"`
}

var suitChars = []string{"♠", "♥", "♣", "♦"}
var rankChars = []string{"3", "4", "5", "6", "7", "8", "9", "10", "J", "Q", "K", "A", "2"}

const deckSize = 162
const handSize = 39

func (c Card) Face() int { return c.ID % 54 }

func (c Card) Copy() int { return c.ID / 54 }

func (c Card) Suit() int {
	face := c.Face()
	if face >= 52 {
		return 4
	}
	return face / 13
}

func (c Card) BaseRank() int {
	face := c.Face()
	if face == 52 {
		return 16
	}
	if face == 53 {
		return 17
	}
	return (face % 13) + 3
}

func (c Card) IsSmallJoker() bool { return c.Face() == 52 }
func (c Card) IsBigJoker() bool   { return c.Face() == 53 }

// CompareRank returns the effective comparison rank given trump suit and level card rank.
// Order: 副牌(3-14) < 主花色牌(103-114) < 副2(150) < 本2(160) < 副级牌(170) < 本级牌(180) < 小王(190) < 大王(200)
func (c Card) CompareRank(trumpSuit int, levelRank int) int {
	if c.IsBigJoker() {
		return 200
	}
	if c.IsSmallJoker() {
		return 190
	}

	base := c.BaseRank()
	suit := c.Suit()
	isTrumpSuit := suit == trumpSuit
	isLevelCard := base == levelRank
	isTwo := base == 15

	if isTrumpSuit && isLevelCard {
		return 180
	}
	if !isTrumpSuit && isLevelCard {
		return 170
	}
	if isTrumpSuit && isTwo {
		return 160
	}
	if !isTrumpSuit && isTwo {
		return 150
	}
	if isTrumpSuit {
		return 100 + base
	}
	return base
}

func (c Card) Display() string {
	face := c.Face()
	if face == 52 {
		return "🃏"
	}
	if face == 53 {
		return "👑"
	}
	return suitChars[c.Suit()] + rankChars[face%13]
}

func NewDeck() []Card {
	cards := make([]Card, deckSize)
	for i := 0; i < deckSize; i++ {
		cards[i] = Card{ID: i}
	}
	return cards
}

func Shuffle(deck []Card) {
	rand.Shuffle(len(deck), func(i, j int) {
		deck[i], deck[j] = deck[j], deck[i]
	})
}

func SortCards(cards []Card, trumpSuit int, levelRank int) {
	sort.Slice(cards, func(i, j int) bool {
		ri := cards[i].CompareRank(trumpSuit, levelRank)
		rj := cards[j].CompareRank(trumpSuit, levelRank)
		if ri != rj {
			return ri > rj
		}
		return cards[i].Suit() < cards[j].Suit()
	})
}

func Deal(deck []Card) (h0, h1, h2, h3, bottom []Card) {
	h0 = make([]Card, handSize)
	copy(h0, deck[0:handSize])
	h1 = make([]Card, handSize)
	copy(h1, deck[handSize:2*handSize])
	h2 = make([]Card, handSize)
	copy(h2, deck[2*handSize:3*handSize])
	h3 = make([]Card, handSize)
	copy(h3, deck[3*handSize:4*handSize])
	bottom = make([]Card, 6)
	copy(bottom, deck[4*handSize:])
	return
}
