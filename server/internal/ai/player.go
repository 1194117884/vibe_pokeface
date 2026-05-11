package ai

import (
	"math/rand"
)

type AIPlayer struct {
	CharacterID   int
	HandCards     []int
	PlayStyle     string // "aggressive", "balanced", "conservative"
	provider      LLMProvider
	promptBuilder *PromptBuilder
}

func NewAIPlayer(charID int, handCards []int, playStyle string, provider LLMProvider) *AIPlayer {
	return &AIPlayer{
		CharacterID:   charID,
		HandCards:     handCards,
		PlayStyle:     playStyle,
		provider:      provider,
		promptBuilder: NewPromptBuilder(),
	}
}

func (p *AIPlayer) NeedsDecision(phase string) bool {
	switch phase {
	case "bidding":
		// Always needs a bid decision before the round starts.
		return true
	case "playing":
		// Playing decisions require game state (e.g. turn tracking) not yet available.
		return false
	default:
		return false
	}
}

func (p *AIPlayer) SelectBid() (string, error) {
	// Simple rule-based bidding as fallback
	if hasBomb(p.HandCards) || countHighCards(p.HandCards) >= 3 {
		return "bid_call", nil
	}
	// Random chance based on play style
	threshold := 0.3
	if p.PlayStyle == "aggressive" {
		threshold = 0.6
	} else if p.PlayStyle == "conservative" {
		threshold = 0.15
	}
	if rand.Float64() < threshold {
		return "bid_call", nil
	}
	return "bid_pass", nil
}

func (p *AIPlayer) SelectPlay(lastPlay []int) (string, []int, error) {
	// Will integrate with LLM provider when available
	// For now, return pass
	return "pass", []int{}, nil
}

func hasBomb(cards []int) bool {
	rankFreq := make(map[int]int)
	for _, id := range cards {
		if id >= 52 {
			continue // handled separately as rocket
		}
		rank := id % 13
		rankFreq[rank]++
	}
	for _, freq := range rankFreq {
		if freq >= 4 {
			return true
		}
	}
	// Check jokers (both = rocket)
	hasSmall := false
	hasBig := false
	for _, id := range cards {
		if id == 52 {
			hasSmall = true
		}
		if id == 53 {
			hasBig = true
		}
	}
	return hasSmall && hasBig
}

func countHighCards(cards []int) int {
	count := 0
	for _, id := range cards {
		rank := id % 13
		if rank >= 10 || id >= 52 { // rank >= J (10=J, 11=Q, 12=K) or jokers
			count++
		}
	}
	return count
}
