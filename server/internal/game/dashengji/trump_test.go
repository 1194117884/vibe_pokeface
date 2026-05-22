package dashengji

import "testing"

func TestCheckSetTrump_RedJokerRed2RedLevel(t *testing.T) {
	// 大王 + ♥2 + ♥6 → 主色 ♥ (suit 1)
	cards := []Card{{ID: 53}, {ID: 25}, {ID: 16}}
	ok := CheckSetTrump(cards, 6, true)
	if !ok {
		t.Error("大王+♥2+♥6 should be valid set trump")
	}
}

func TestCheckSetTrump_BlackJokerBlack2BlackLevel(t *testing.T) {
	// 小王 + ♣2 + ♣6 → 主色 ♣ (suit 2)
	cards := []Card{{ID: 52}, {ID: 38}, {ID: 29}}
	ok := CheckSetTrump(cards, 6, true)
	if !ok {
		t.Error("小王+♣2+♣6 should be valid set trump")
	}
}

func TestCheckSetTrump_RedJokerRed2TwoRedLevel(t *testing.T) {
	// 大王 + ♥2 + ♥6♥6 → 主色 ♥ (suit 1)
	cards := []Card{{ID: 53}, {ID: 25}, {ID: 16}, {ID: 70}}
	ok := CheckSetTrump(cards, 6, true)
	if !ok {
		t.Error("大王+♥2+♥6♥6 should be valid set trump")
	}
}

func TestCheckSetTrump_BlackJokerBlack2TwoBlackLevel(t *testing.T) {
	// 小王 + ♣2 + ♣6♣6 → 主色 ♣
	cards := []Card{{ID: 52}, {ID: 38}, {ID: 29}, {ID: 83}}
	ok := CheckSetTrump(cards, 6, true)
	if !ok {
		t.Error("小王+♣2+♣6♣6 should be valid set trump")
	}
}

func TestCheckSetTrump_WrongColorMix(t *testing.T) {
	// 大王 + ♥2 + ♠6 → invalid (2 and level must match suit)
	cards := []Card{{ID: 53}, {ID: 25}, {ID: 3}}
	ok := CheckSetTrump(cards, 6, true)
	if ok {
		t.Error("mixed color 2 and level should be invalid")
	}
}

func TestCheckSetTrump_WrongJokerColor(t *testing.T) {
	// 大王(red) + ♣2(black) + ♣6(black) → invalid
	cards := []Card{{ID: 53}, {ID: 38}, {ID: 29}}
	ok := CheckSetTrump(cards, 6, true)
	if ok {
		t.Error("大王(red) with black 2+level should be invalid")
	}
}

func TestCheckSetTrump_NeedsBoth2AndLevel(t *testing.T) {
	// Only 大王 + ♥6 (no 2) → invalid
	cards := []Card{{ID: 53}, {ID: 16}}
	ok := CheckSetTrump(cards, 6, true)
	if ok {
		t.Error("missing 2 should be invalid")
	}
}

func TestCheckCounterTrump_RequiresTwoLevelCards(t *testing.T) {
	cards := []Card{{ID: 53}, {ID: 25}, {ID: 16}, {ID: 70}}
	ok := CheckCounterTrump(cards, 6)
	if !ok {
		t.Error("大王+♥2+♥6♥6 should be valid counter trump")
	}
}

func TestCheckCounterTrump_OneLevelCardNotEnough(t *testing.T) {
	cards := []Card{{ID: 53}, {ID: 25}, {ID: 16}}
	ok := CheckCounterTrump(cards, 6)
	if ok {
		t.Error("only 1 level card should not be enough for counter trump")
	}
}

func TestIsDeadSetTrump(t *testing.T) {
	cards := []Card{{ID: 53}, {ID: 25}, {ID: 16}, {ID: 70}}
	if !IsDeadTrump(cards, 6) {
		t.Error("pair of level cards same suit as 2 should be 定死")
	}
}

func TestNotDeadTrump_NoLevelPair(t *testing.T) {
	cards := []Card{{ID: 53}, {ID: 25}, {ID: 16}}
	if IsDeadTrump(cards, 6) {
		t.Error("no level pair should not be 定死")
	}
}

func TestClassifyCard(t *testing.T) {
	trumpSuit := 0 // ♠
	levelRank := 6 // 打6

	tests := []struct {
		face int
		want CardCategory
	}{
		{3, CatNativeMain},   // ♠6 = 本级牌
		{16, CatSideMain},    // ♥6 = 副级牌
		{12, CatNativeMain},  // ♠2 = 本2
		{25, CatSideMain},    // ♥2 = 副2
		{1, CatTrump},        // ♠4 = 主花色牌
		{14, CatSide},        // ♥4 = 副牌
		{52, CatJoker},       // 小王
		{53, CatJoker},       // 大王
	}

	for _, tt := range tests {
		c := Card{ID: tt.face}
		got := ClassifyCard(c, trumpSuit, levelRank)
		if got != tt.want {
			t.Errorf("face %d (%s): expected category %d, got %d", tt.face, c.Display(), tt.want, got)
		}
	}
}

func TestGetTrumpSuit(t *testing.T) {
	cards := []Card{{ID: 53}, {ID: 25}, {ID: 16}} // 大王, ♥2, ♥6
	suit := GetTrumpSuit(cards, 6)
	if suit != 1 {
		t.Errorf("expected trump suit 1 (♥), got %d", suit)
	}
}
