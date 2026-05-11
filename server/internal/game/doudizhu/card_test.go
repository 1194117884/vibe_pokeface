package doudizhu

import "testing"

func TestNewDeck_Has54Cards(t *testing.T) {
	deck := NewDeck()
	if len(deck) != 54 {
		t.Errorf("deck length = %d, want 54", len(deck))
	}
}

func TestNewDeck_HasCorrectCards(t *testing.T) {
	deck := NewDeck()
	seen := make(map[int]bool)
	for _, c := range deck {
		if seen[c.ID] {
			t.Errorf("duplicate card: %d", c.ID)
		}
		seen[c.ID] = true
	}
	// 54 unique cards = 52 standard + 2 jokers
	if len(seen) != 54 {
		t.Errorf("unique cards = %d, want 54", len(seen))
	}
}

func TestShuffle_ChangesOrder(t *testing.T) {
	d1 := NewDeck()
	d2 := NewDeck()
	Shuffle(d2)
	same := true
	for i := range d1 {
		if d1[i].ID != d2[i].ID {
			same = false
			break
		}
	}
	if same {
		t.Error("shuffle did not change card order")
	}
}

func TestDeal_17_17_20(t *testing.T) {
	deck := NewDeck()
	Shuffle(deck)
	h1, h2, h3, remaining := Deal(deck)
	if len(h1) != 17 {
		t.Errorf("hand 1 = %d cards, want 17", len(h1))
	}
	if len(h2) != 17 {
		t.Errorf("hand 2 = %d cards, want 17", len(h2))
	}
	if len(h3) != 17 {
		t.Errorf("hand 3 = %d cards, want 17", len(h3))
	}
	if len(remaining) != 3 {
		t.Errorf("remaining = %d cards, want 3", len(remaining))
	}
}

func TestCard_Display(t *testing.T) {
	tests := []struct {
		id   int
		want string
	}{
		{0, "♠3"}, {12, "♠2"}, {13, "♥3"}, {25, "♥2"},
		{26, "♣3"}, {38, "♣2"}, {39, "♦3"}, {51, "♦2"},
		{52, "🃏"}, {53, "👑"},
	}
	for _, tt := range tests {
		c := Card{ID: tt.id}
		if got := c.Display(); got != tt.want {
			t.Errorf("Card{%d}.Display() = %q, want %q", tt.id, got, tt.want)
		}
	}
}

func TestSortCards(t *testing.T) {
	// 3 cards: 2♠(12), 3♠(0), A♠(11) - should sort by rank desc, then suit
	cards := []Card{{ID: 12}, {ID: 0}, {ID: 11}}
	SortCards(cards)
	if cards[0].ID != 12 {
		t.Errorf("first card should be 2♠ (highest rank)")
	}
}

func TestCardRank(t *testing.T) {
	// rank: 3=3, 4=4, ..., K=13, A=14, 2=15, small joker=16, big joker=17
	tests := []struct {
		id   int
		rank int
	}{
		{0, 3}, {12, 15}, {13, 3}, {25, 15},
		{52, 16}, {53, 17},
	}
	for _, tt := range tests {
		c := Card{ID: tt.id}
		if r := c.Rank(); r != tt.rank {
			t.Errorf("Card{%d}.Rank() = %d, want %d", tt.id, r, tt.rank)
		}
	}
}
