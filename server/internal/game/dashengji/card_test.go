package dashengji

import (
	"slices"
	"testing"
)

func TestNewDeck(t *testing.T) {
	deck := NewDeck()
	if len(deck) != 162 {
		t.Fatalf("expected 162 cards, got %d", len(deck))
	}
	faceCount := make(map[int]int)
	for _, c := range deck {
		faceCount[c.Face()]++
	}
	for face := 0; face < 54; face++ {
		if faceCount[face] != 3 {
			t.Errorf("face %d: expected 3 copies, got %d", face, faceCount[face])
		}
	}
	seen := make(map[int]bool)
	for _, c := range deck {
		if c.ID < 0 || c.ID > 161 {
			t.Errorf("card ID %d out of range", c.ID)
		}
		if seen[c.ID] {
			t.Errorf("duplicate card ID %d", c.ID)
		}
		seen[c.ID] = true
	}
}

func TestCardFaceAndCopy(t *testing.T) {
	c := Card{ID: 55}
	if c.Face() != 1 {
		t.Errorf("ID 55: expected face 1, got %d", c.Face())
	}
	if c.Copy() != 1 {
		t.Errorf("ID 55: expected copy 1, got %d", c.Copy())
	}
	c2 := Card{ID: 0}
	if c2.Face() != 0 {
		t.Errorf("ID 0: expected face 0, got %d", c2.Face())
	}
	if c2.Copy() != 0 {
		t.Errorf("ID 0: expected copy 0, got %d", c2.Copy())
	}
}

func TestCardBaseRank(t *testing.T) {
	tests := []struct{ face, want int }{
		{0, 3},   // spade3
		{12, 15}, // spade2
		{13, 3},  // heart3
		{25, 15}, // heart2
	}
	for _, tt := range tests {
		c := Card{ID: tt.face}
		if c.BaseRank() != tt.want {
			t.Errorf("face %d: expected rank %d, got %d", tt.face, tt.want, c.BaseRank())
		}
	}
	if (Card{ID: 52}).BaseRank() != 16 {
		t.Errorf("small joker: expected rank 16, got %d", (Card{ID: 52}).BaseRank())
	}
	if (Card{ID: 53}).BaseRank() != 17 {
		t.Errorf("big joker: expected rank 17, got %d", (Card{ID: 53}).BaseRank())
	}
}

func TestCardSuit(t *testing.T) {
	tests := []struct{ face, want int }{
		{0, 0}, {12, 0},   // spade
		{13, 1}, {25, 1},  // heart
		{26, 2}, {38, 2},  // club
		{39, 3}, {51, 3},  // diamond
		{52, 4}, {53, 4},  // jokers
	}
	for _, tt := range tests {
		c := Card{ID: tt.face}
		if c.Suit() != tt.want {
			t.Errorf("face %d: expected suit %d, got %d", tt.face, tt.want, c.Suit())
		}
	}
}

func TestCardIsJoker(t *testing.T) {
	if !(Card{ID: 52}).IsSmallJoker() {
		t.Error("face 52 should be small joker")
	}
	if !(Card{ID: 53}).IsBigJoker() {
		t.Error("face 53 should be big joker")
	}
	if (Card{ID: 0}).IsSmallJoker() || (Card{ID: 0}).IsBigJoker() {
		t.Error("face 0 should not be a joker")
	}
}

func TestCompareRank(t *testing.T) {
	trumpSuit := 0 // spade
	levelRank := 6 // rank value for 6

	side3 := Card{ID: 13}  // heart3
	main3 := Card{ID: 0}   // spade3
	side2 := Card{ID: 25}  // heart2
	main2 := Card{ID: 12}  // spade2
	side6 := Card{ID: 16}  // heart6 (level card, non-trump)
	main6 := Card{ID: 3}   // spade6 (level card, trump)
	smallJ := Card{ID: 52}
	bigJ := Card{ID: 53}

	cr := func(c Card) int { return c.CompareRank(trumpSuit, levelRank) }

	ranks := []Card{side3, main3, side2, main2, side6, main6, smallJ, bigJ}
	for i := 1; i < len(ranks); i++ {
		if cr(ranks[i-1]) >= cr(ranks[i]) {
			t.Errorf("%s (rank %d) should be < %s (rank %d)",
				ranks[i-1].Display(), cr(ranks[i-1]),
				ranks[i].Display(), cr(ranks[i]))
		}
	}
}

func TestDeal(t *testing.T) {
	deck := NewDeck()
	Shuffle(deck)
	h0, h1, h2, h3, bottom := Deal(deck)

	if len(h0) != 39 || len(h1) != 39 || len(h2) != 39 || len(h3) != 39 {
		t.Errorf("expected 39 cards per hand, got %d/%d/%d/%d",
			len(h0), len(h1), len(h2), len(h3))
	}
	if len(bottom) != 6 {
		t.Errorf("expected 6 bottom cards, got %d", len(bottom))
	}

	all := make(map[int]bool)
	for _, c := range slices.Concat(h0, h1, h2, h3, bottom) {
		if all[c.ID] {
			t.Errorf("duplicate card ID %d", c.ID)
		}
		all[c.ID] = true
	}
}

func TestSortCards(t *testing.T) {
	trumpSuit := -1
	levelRank := 6

	cards := []Card{{ID: 13}, {ID: 0}, {ID: 52}, {ID: 53}, {ID: 25}}
	SortCards(cards, trumpSuit, levelRank)

	if cards[0].ID != 53 {
		t.Errorf("first should be big joker, got %d", cards[0].ID)
	}
	if cards[1].ID != 52 {
		t.Errorf("second should be small joker, got %d", cards[1].ID)
	}
}
