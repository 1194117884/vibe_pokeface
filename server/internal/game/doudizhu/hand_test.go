package doudizhu

import (
	"testing"
)

// helpers to create cards by ID quickly
func ids(ids ...int) []Card {
	cards := make([]Card, len(ids))
	for i, id := range ids {
		cards[i] = Card{ID: id}
	}
	return cards
}

func TestParsePlay_Single(t *testing.T) {
	p := ParsePlay(ids(0)) // ♠3
	if p.Type != PlaySingle {
		t.Errorf("type = %d, want %d", p.Type, PlaySingle)
	}
	if p.MainRank != 3 {
		t.Errorf("mainRank = %d, want %d", p.MainRank, 3)
	}
	if p.Length != 1 {
		t.Errorf("length = %d, want %d", p.Length, 1)
	}
}

func TestParsePlay_Pair(t *testing.T) {
	p := ParsePlay(ids(0, 13)) // ♠3+♥3
	if p.Type != PlayPair {
		t.Errorf("type = %d, want %d", p.Type, PlayPair)
	}
	if p.MainRank != 3 {
		t.Errorf("mainRank = %d, want %d", p.MainRank, 3)
	}
}

func TestParsePlay_Triple(t *testing.T) {
	p := ParsePlay(ids(0, 13, 26)) // all 3s
	if p.Type != PlayTriple {
		t.Errorf("type = %d, want %d", p.Type, PlayTriple)
	}
}

func TestParsePlay_Bomb(t *testing.T) {
	p := ParsePlay(ids(0, 13, 26, 39)) // 4x 3s
	if p.Type != PlayBomb {
		t.Errorf("type = %d, want %d", p.Type, PlayBomb)
	}
}

func TestParsePlay_Rocket(t *testing.T) {
	p := ParsePlay(ids(52, 53)) // both jokers
	if p.Type != PlayRocket {
		t.Errorf("type = %d, want %d", p.Type, PlayRocket)
	}
}

func TestParsePlay_Straight(t *testing.T) {
	// 3,4,5,6,7 (one of each suit)
	p := ParsePlay(ids(0, 14, 28, 42, 4)) // 3,4,5,6,7
	if p.Type != PlayStraight {
		t.Errorf("type = %d, want %d", p.Type, PlayStraight)
	}
	if p.Length != 5 {
		t.Errorf("length = %d, want %d", p.Length, 5)
	}
}

func TestParsePlay_PairStraight(t *testing.T) {
	// 33,44,55 (3 consecutive pairs)
	p := ParsePlay(ids(0, 13, 1, 14, 2, 15)) // 3,3,4,4,5,5
	if p.Type != PlayPairStraight {
		t.Errorf("type = %d, want %d", p.Type, PlayPairStraight)
	}
	if p.Length != 3 {
		t.Errorf("length = %d, want %d", p.Length, 3)
	}
}

func TestParsePlay_TriplePlusOne(t *testing.T) {
	// 333 + 4
	p := ParsePlay(ids(0, 13, 26, 1)) // 3,3,3,4
	if p.Type != PlayTriplePlus1 {
		t.Errorf("type = %d, want %d", p.Type, PlayTriplePlus1)
	}
}

func TestParsePlay_TriplePlusTwo(t *testing.T) {
	// 333 + 44
	p := ParsePlay(ids(0, 13, 26, 1, 14)) // 3,3,3,4,4
	if p.Type != PlayTriplePlus2 {
		t.Errorf("type = %d, want %d", p.Type, PlayTriplePlus2)
	}
}

func TestParsePlay_InvalidCards(t *testing.T) {
	tests := []struct {
		name  string
		cards []Card
	}{
		{"non-consecutive singles", ids(0, 1, 2, 4)},                          // 3,4,5,7 (missing 6)
		{"non-consecutive triples", ids(0, 13, 26, 2, 28, 40)},               // 333555 not consecutive
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := ParsePlay(tt.cards)
			if p.Type != PlayInvalid {
				t.Errorf("expected invalid for %s", tt.name)
			}
		})
	}
}

func TestComparePlay_BombBeatsSingle(t *testing.T) {
	single := ParsePlay(ids(0))
	bomb := ParsePlay(ids(0, 13, 26, 39)) // bomb 3s
	if !CanBeat(bomb, single) {
		t.Error("bomb should beat single")
	}
}

func TestComparePlay_RocketBeatsBomb(t *testing.T) {
	bomb := ParsePlay(ids(0, 13, 26, 39))
	rocket := ParsePlay(ids(52, 53))
	if !CanBeat(rocket, bomb) {
		t.Error("rocket should beat bomb")
	}
}

func TestComparePlay_HigherRank(t *testing.T) {
	p3 := ParsePlay(ids(0)) // 3
	p4 := ParsePlay(ids(1)) // 4
	if !CanBeat(p4, p3) {
		t.Error("4 should beat 3")
	}
	if CanBeat(p3, p4) {
		t.Error("3 should not beat 4")
	}
}

func TestComparePlay_DifferentTypes(t *testing.T) {
	single := ParsePlay(ids(0))
	pair := ParsePlay(ids(0, 13))
	if CanBeat(pair, single) {
		t.Error("pair should not beat single (different types)")
	}
}

func TestParsePlay_Pass(t *testing.T) {
	p := ParsePlay([]Card{})
	if p.Type != PlayPass {
		t.Errorf("empty play should be Pass, got type %d", p.Type)
	}
}

func TestParsePlay_Airplane(t *testing.T) {
	// 333,444 + 5,6 (2 consecutive triples + 2 singles)
	p := ParsePlay(ids(0, 13, 26, 1, 14, 27, 2, 28)) // 3,3,3,4,4,4,5,6
	if p.Type != PlayAirplane {
		t.Errorf("type = %d, want %d", p.Type, PlayAirplane)
	}
}
