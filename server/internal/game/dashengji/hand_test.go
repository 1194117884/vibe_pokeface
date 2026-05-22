package dashengji

import "testing"

func TestParsePlay_Single(t *testing.T) {
	play := ParsePlay([]Card{{ID: 0}})
	if play.Type != PlaySingle {
		t.Fatalf("expected PlaySingle, got %d", play.Type)
	}
}

func TestParsePlay_Pair(t *testing.T) {
	play := ParsePlay([]Card{{ID: 0}, {ID: 54}}) // ♠3 + ♠3 (copy 1)
	if play.Type != PlayPair {
		t.Fatalf("expected PlayPair, got %d", play.Type)
	}
}

func TestParsePlay_Pair_DifferentSuit(t *testing.T) {
	play := ParsePlay([]Card{{ID: 0}, {ID: 13}}) // ♠3 + ♥3
	if play.Type != PlayInvalid {
		t.Fatalf("expected PlayInvalid for cross-suit pair, got %d", play.Type)
	}
}

func TestParsePlay_Triple(t *testing.T) {
	play := ParsePlay([]Card{{ID: 0}, {ID: 54}, {ID: 108}}) // ♠3 x3
	if play.Type != PlayTriple {
		t.Fatalf("expected PlayTriple, got %d", play.Type)
	}
}

func TestParsePlay_Triple_DifferentSuit(t *testing.T) {
	play := ParsePlay([]Card{{ID: 0}, {ID: 54}, {ID: 13}}) // ♠3♠3♥3
	if play.Type != PlayInvalid {
		t.Fatalf("expected PlayInvalid for mixed-suit triple, got %d", play.Type)
	}
}

func TestParsePlay_Tractor(t *testing.T) {
	// 3+ consecutive pairs, same suit
	// ♠3♠3♠4♠4♠5♠5 (faces 0,54,1,55,2,56 -- all spade, ranks 3,4,5)
	play := ParsePlayWithContext(
		[]Card{{ID: 0}, {ID: 54}, {ID: 1}, {ID: 55}, {ID: 2}, {ID: 56}},
		-1, 6,
	)
	if play.Type != PlayTractor {
		t.Fatalf("expected PlayTractor, got %d", play.Type)
	}
	if play.Length != 3 {
		t.Errorf("expected length 3 (pairs), got %d", play.Length)
	}
}

func TestParsePlay_Tractor_NotConsecutive(t *testing.T) {
	play := ParsePlayWithContext(
		[]Card{{ID: 0}, {ID: 54}, {ID: 4}, {ID: 58}}, // ♠3♠3♠7♠7
		-1, 6,
	)
	if play.Type != PlayInvalid {
		t.Fatalf("expected PlayInvalid for non-consecutive pairs, got %d", play.Type)
	}
}

func TestParsePlay_Tractor_TooShort(t *testing.T) {
	play := ParsePlayWithContext(
		[]Card{{ID: 0}, {ID: 54}, {ID: 1}, {ID: 55}}, // Only 2 consecutive pairs
		-1, 6,
	)
	if play.Type != PlayInvalid {
		t.Fatalf("expected PlayInvalid for 2-pair tractor, got %d", play.Type)
	}
}

func TestParsePlay_JokersCannotPair(t *testing.T) {
	play := ParsePlay([]Card{{ID: 52}, {ID: 53}}) // small joker + big joker
	if play.Type != PlayInvalid {
		t.Fatalf("expected PlayInvalid for mixed jokers, got %d", play.Type)
	}
}

func TestParsePlay_Empty(t *testing.T) {
	play := ParsePlay([]Card{})
	if play.Type != PlayPass {
		t.Fatalf("expected PlayPass, got %d", play.Type)
	}
}

func TestCanBeat_SameType_HigherRank(t *testing.T) {
	last := Play{Type: PlaySingle, MainRank: 4, Length: 1}
	current := Play{Type: PlaySingle, MainRank: 8, Length: 1}
	if !CanBeat(current, last) {
		t.Error("single 8 should beat single 4")
	}
}

func TestCanBeat_SameType_LowerRank(t *testing.T) {
	last := Play{Type: PlaySingle, MainRank: 8, Length: 1}
	current := Play{Type: PlaySingle, MainRank: 4, Length: 1}
	if CanBeat(current, last) {
		t.Error("single 4 should not beat single 8")
	}
}

func TestCanBeat_DifferentType_NoBeat(t *testing.T) {
	last := Play{Type: PlaySingle, MainRank: 8, Length: 1}
	current := Play{Type: PlayPair, MainRank: 3, Length: 1}
	if CanBeat(current, last) {
		t.Error("pair should not beat single")
	}
}

func TestCanBeat_LeadCanPlayAnything(t *testing.T) {
	last := Play{Type: PlayPass}
	current := Play{Type: PlaySingle, MainRank: 3, Length: 1}
	if !CanBeat(current, last) {
		t.Error("any play should be valid when leading")
	}
}

func TestCanBeat_PassNeverBeats(t *testing.T) {
	last := Play{Type: PlaySingle, MainRank: 3, Length: 1}
	current := Play{Type: PlayPass}
	if CanBeat(current, last) {
		t.Error("pass should not beat any play")
	}
}

func TestCanBeat_Tractor_MorePairs(t *testing.T) {
	last := Play{Type: PlayTractor, MainRank: 3, Length: 3}
	current := Play{Type: PlayTractor, MainRank: 4, Length: 3}
	if !CanBeat(current, last) {
		t.Error("tractor 4-5-6 should beat tractor 3-4-5")
	}
}
