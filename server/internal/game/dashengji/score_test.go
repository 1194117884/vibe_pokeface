package dashengji

import "testing"

func TestCalculateLevelChange_DealerWinNoPoints(t *testing.T) {
	change := CalculateLevelChange(true, 0)
	if change != 2 {
		t.Errorf("dealer win with 0 points: expected +2 levels, got %d", change)
	}
}

func TestCalculateLevelChange_DealerWinLowPoints(t *testing.T) {
	change := CalculateLevelChange(true, 60)
	if change != 1 {
		t.Errorf("dealer win with 60 points: expected +1 level, got %d", change)
	}
	change = CalculateLevelChange(true, 119)
	if change != 1 {
		t.Errorf("dealer win with 119 points: expected +1 level, got %d", change)
	}
}

func TestCalculateLevelChange_DealerWinHighPoints(t *testing.T) {
	change := CalculateLevelChange(true, 150)
	if change != 0 {
		t.Errorf("dealer win with 150 points: expected no level change and dealer rotation, got %d", change)
	}
}

func TestCalculateLevelChange_NonDealerWinZeroPoints(t *testing.T) {
	change := CalculateLevelChange(false, 0)
	if change != 0 {
		t.Errorf("non-dealer win with 0 points: expected no level change, got %d", change)
	}
}

func TestCalculateLevelChange_NonDealerWinLowPoints(t *testing.T) {
	change := CalculateLevelChange(false, 60)
	if change != 0 {
		t.Errorf("non-dealer win with 60 points: expected no level change, got %d", change)
	}
}

func TestCalculateLevelChange_NonDealerWinHighPoints(t *testing.T) {
	change := CalculateLevelChange(false, 150)
	if change != -1 {
		t.Errorf("non-dealer win with 150 points: expected -1 level, got %d", change)
	}
	change = CalculateLevelChange(false, 120)
	if change != -1 {
		t.Errorf("non-dealer win with 120 points: expected -1 level, got %d", change)
	}
}

func TestAdvanceLevel(t *testing.T) {
	if AdvanceLevel(3, 2) != 5 {
		t.Error("3 + 2 = 5")
	}
	if AdvanceLevel(14, 1) != 14 {
		t.Error("A(14) + 1 should stay at 14 (max)")
	}
	if AdvanceLevel(12, 3) != 14 {
		t.Error("Q(12) + 3 should cap at A(14)")
	}
}

func TestLevelToRank(t *testing.T) {
	if LevelToRank(3) != 3 {
		t.Error("level 3 -> rank 3")
	}
	if LevelToRank(14) != 14 {
		t.Error("level A(14) -> rank 14")
	}
}

func TestCountRoundPoints(t *testing.T) {
	cards := []Card{
		{ID: 2},  // S5 (face 2, base rank 5)
		{ID: 7},  // S10 (face 7, base rank 10)
		{ID: 10}, // SK (face 10, base rank 13)
		{ID: 0},  // S3 - no points
	}
	points := CountRoundPoints(cards)
	if points != 25 {
		t.Errorf("expected 25 points, got %d", points)
	}
}
