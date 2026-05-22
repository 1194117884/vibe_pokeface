package dashengji

// CardCategory classifies a card within the trump system.
type CardCategory int

const (
	CatSide       CardCategory = iota // 副牌
	CatTrump                          // 主花色牌 (non-2, non-level)
	CatSideMain                       // 副主: 2/级牌 of non-trump suits
	CatNativeMain                     // 本主: 2/级牌 of trump suit
	CatJoker                          // 大王/小王
)

// ClassifyCard returns the trump category of a card.
func ClassifyCard(c Card, trumpSuit int, levelRank int) CardCategory {
	if c.IsSmallJoker() || c.IsBigJoker() {
		return CatJoker
	}

	suit := c.Suit()
	base := c.BaseRank()
	isTrumpSuit := suit == trumpSuit
	isLevelCard := base == levelRank
	isTwo := base == 15

	if isTrumpSuit && (isLevelCard || isTwo) {
		return CatNativeMain
	}
	if !isTrumpSuit && (isLevelCard || isTwo) {
		return CatSideMain
	}
	if isTrumpSuit {
		return CatTrump
	}
	return CatSide
}

// suitColor returns "red" for ♥/♦, "black" for ♠/♣.
func suitColor(suit int) string {
	if suit == 1 || suit == 3 {
		return "red"
	}
	if suit == 0 || suit == 2 {
		return "black"
	}
	return ""
}

// findTwoSuit finds the suit of a 2 (rank 15) in the cards. Returns -1 if not found.
func findTwoSuit(cards []Card, levelRank int) int {
	for _, c := range cards {
		if c.BaseRank() == 15 {
			return c.Suit()
		}
	}
	return -1
}

// CheckSetTrump validates whether the given cards can be used to set trump (定主).
// Conditions (满足一个即可):
// - 大王 + 红色2 + 红色级牌 (2和级牌同花色)
// - 小王 + 黑色2 + 黑色级牌 (2和级牌同花色)
// - 大王 + 红色2 + 2张红色级牌 (2和2张级牌同花色)
// - 小王 + 黑色2 + 2张黑色级牌 (2和2张级牌同花色)
func CheckSetTrump(cards []Card, levelRank int, isDealerTeam bool) bool {
	if len(cards) < 3 {
		return false
	}

	hasBigJoker := false
	hasSmallJoker := false
	for _, c := range cards {
		if c.IsBigJoker() {
			hasBigJoker = true
		}
		if c.IsSmallJoker() {
			hasSmallJoker = true
		}
	}
	if !hasBigJoker && !hasSmallJoker {
		return false
	}

	twoSuit := findTwoSuit(cards, levelRank)
	if twoSuit == -1 {
		return false
	}

	twoColor := suitColor(twoSuit)
	jokerColorMatch := (hasBigJoker && twoColor == "red") || (hasSmallJoker && twoColor == "black")
	if !jokerColorMatch {
		return false
	}

	// Count level cards that match the 2's suit
	levelCount := 0
	for _, c := range cards {
		if c.BaseRank() == levelRank && c.Suit() == twoSuit {
			levelCount++
		}
	}

	return levelCount >= 1
}

// GetTrumpSuit extracts the trump suit from a valid set-trump card selection.
func GetTrumpSuit(cards []Card, levelRank int) int {
	return findTwoSuit(cards, levelRank)
}

// CheckCounterTrump validates whether the given cards can counter-trump (反主).
// 反主条件: 王 + 2 + 2张级牌 (2和2张级牌同花色, 颜色匹配)
func CheckCounterTrump(cards []Card, levelRank int) bool {
	if len(cards) < 4 {
		return false
	}

	hasBigJoker := false
	hasSmallJoker := false
	for _, c := range cards {
		if c.IsBigJoker() {
			hasBigJoker = true
		}
		if c.IsSmallJoker() {
			hasSmallJoker = true
		}
	}
	if !hasBigJoker && !hasSmallJoker {
		return false
	}

	twoSuit := findTwoSuit(cards, levelRank)
	if twoSuit == -1 {
		return false
	}

	twoColor := suitColor(twoSuit)
	jokerColorMatch := (hasBigJoker && twoColor == "red") || (hasSmallJoker && twoColor == "black")
	if !jokerColorMatch {
		return false
	}

	levelCount := 0
	for _, c := range cards {
		if c.BaseRank() == levelRank && c.Suit() == twoSuit {
			levelCount++
		}
	}

	return levelCount >= 2
}

// IsDeadTrump checks if the set-trump is 定死 (dead/final, cannot be countered).
func IsDeadTrump(cards []Card, levelRank int) bool {
	twoSuit := findTwoSuit(cards, levelRank)
	if twoSuit == -1 {
		return false
	}
	levelCount := 0
	for _, c := range cards {
		if c.BaseRank() == levelRank && c.Suit() == twoSuit {
			levelCount++
		}
	}
	return levelCount >= 2
}
