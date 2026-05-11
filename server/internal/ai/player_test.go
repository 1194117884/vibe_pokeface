package ai

import (
	"testing"
)

func TestAIPlayer_NeedsBidDecision(t *testing.T) {
	p := &AIPlayer{CharacterID: 1, HandCards: []int{0, 1, 2, 13, 14, 26, 27, 39, 40, 52}}
	if !p.NeedsDecision("bidding") {
		t.Error("expected bidding decision needed")
	}
	if p.NeedsDecision("playing") {
		t.Error("expected no playing decision yet")
	}
}

func TestAIPlayer_SelectBid(t *testing.T) {
	p := &AIPlayer{
		CharacterID: 1,
		HandCards:   []int{12, 25, 38, 51, 52, 53}, // lots of 2s + jokers (rocket)
		PlayStyle:   "aggressive",
	}
	bid, err := p.SelectBid()
	if err != nil {
		t.Fatalf("SelectBid error: %v", err)
	}
	if bid != "bid_call" {
		t.Errorf("expected bid_call for strong hand, got %s", bid)
	}
}

func TestAIPlayer_HasBomb(t *testing.T) {
	cards := []int{0, 13, 26, 39} // four 3s = bomb
	if !hasBomb(cards) {
		t.Error("expected bomb detected")
	}
	cards2 := []int{0, 1, 2, 3} // not a bomb
	if hasBomb(cards2) {
		t.Error("expected no bomb")
	}
}

func TestAIPlayer_HasBomb_NoJokerCollision(t *testing.T) {
	// Three 3s + small joker should NOT be a bomb (rank collision edge case)
	cards := []int{0, 13, 26, 52} // 3,3,3, small joker
	if hasBomb(cards) {
		t.Error("three 3s + small joker should not be a bomb")
	}
}

func TestAIPlayer_CountHighCards(t *testing.T) {
	cards := []int{12, 25, 52, 53} // 2, 2, small joker, big joker
	count := countHighCards(cards)
	if count < 3 {
		t.Errorf("expected >= 3 high cards, got %d", count)
	}
}
