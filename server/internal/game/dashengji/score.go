package dashengji

// CalculateLevelChange returns level change for the dealer team.
// Positive = dealer gains levels. Negative = non-dealer gains levels.
func CalculateLevelChange(dealerWin bool, points int) int {
	if dealerWin {
		if points == 0 {
			return 2
		}
		if points < 120 {
			return 1
		}
		return 0
	} else {
		if points < 120 {
			return 0
		}
		return -1
	}
}

// AdvanceLevel computes new level after a change. Caps at 14 (A). Floors at 3.
func AdvanceLevel(currentLevel int, change int) int {
	newLevel := currentLevel + change
	if newLevel > 14 {
		return 14
	}
	if newLevel < 3 {
		return 3
	}
	return newLevel
}

// LevelToRank converts a game level (3-14) to a base rank value.
func LevelToRank(level int) int {
	return level
}

// CountRoundPoints counts score points in cards (5=5, 10=10, K=10).
func CountRoundPoints(cards []Card) int {
	total := 0
	for _, c := range cards {
		switch c.BaseRank() {
		case 5:
			total += 5
		case 10:
			total += 10
		case 13: // K
			total += 10
		}
	}
	return total
}

// IsGameOver returns true if level A (14) is reached.
func IsGameOver(dealerLevel int) bool {
	return dealerLevel >= 14
}
