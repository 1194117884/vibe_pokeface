package dashengji

import "sort"

// PlayType represents the type of a card play in Dashengji.
type PlayType int

const (
	PlayInvalid PlayType = 0
	PlaySingle  PlayType = 1
	PlayPair    PlayType = 2
	PlayTriple  PlayType = 3
	PlayTractor PlayType = 4
	PlayPass    PlayType = 5
)

// Play represents a parsed card play.
type Play struct {
	Type     PlayType `json:"type"`
	MainRank int      `json:"main_rank"`
	Length   int      `json:"length"`
}

// ParsePlay analyzes cards without trump context.
func ParsePlay(cards []Card) Play {
	return ParsePlayWithContext(cards, -1, -1)
}

// ParsePlayWithContext analyzes cards with trump context for tractor same-category detection.
func ParsePlayWithContext(cards []Card, trumpSuit int, levelRank int) Play {
	if len(cards) == 0 {
		return Play{Type: PlayPass}
	}

	// Single card (including any single joker) is a valid single
	if len(cards) == 1 {
		return Play{Type: PlaySingle, MainRank: playRank(cards[0], trumpSuit, levelRank), Length: 1}
	}

	// Mixed jokers (small + big) is invalid
	hasSJ, hasBJ := false, false
	for _, c := range cards {
		if c.IsSmallJoker() {
			hasSJ = true
		}
		if c.IsBigJoker() {
			hasBJ = true
		}
	}
	if hasSJ && hasBJ {
		return Play{Type: PlayInvalid}
	}
	if hasSJ || hasBJ {
		return parseJokerPlay(cards, hasSJ, hasBJ, trumpSuit, levelRank)
	}

	// Two same jokers = pair
	// handled by parseJokerPlay above.

	// All cards must have the same suit
	suit := cards[0].Suit()
	for _, c := range cards[1:] {
		if c.Suit() != suit {
			return Play{Type: PlayInvalid}
		}
	}

	// Build face frequency within the suit
	type cardKey struct {
		suit int
		face int
	}
	freq := make(map[cardKey]int)
	for _, c := range cards {
		freq[cardKey{c.Suit(), c.Face()}]++
	}

	// Collect unique faces that appear 2+ times
	pairFaces := make([]int, 0)
	for key, count := range freq {
		if count >= 2 {
			pairFaces = append(pairFaces, key.face)
		}
	}

	// All cards must be in complete groups of 2+
	totalPaired := 0
	for _, f := range pairFaces {
		totalPaired += freq[cardKey{suit, f}]
	}
	if totalPaired != len(cards) {
		return Play{Type: PlayInvalid}
	}

	// Pure pair: 2 cards, same face
	if len(cards) == 2 && len(pairFaces) == 1 {
		return Play{Type: PlayPair, MainRank: playRank(cards[0], trumpSuit, levelRank), Length: 1}
	}

	// Pure triple: 3 cards, same face
	if len(cards) == 3 && len(pairFaces) == 1 {
		return Play{Type: PlayTriple, MainRank: playRank(cards[0], trumpSuit, levelRank), Length: 1}
	}

	// Tractor: 3+ consecutive pairs, same suit, same category
	if len(pairFaces) >= 3 && len(cards)%2 == 0 {
		// Convert faces to base rank indices for consecutive check
		faceRanks := make([]int, len(pairFaces))
		for i, f := range pairFaces {
			faceRanks[i] = f % 13 // 0-based rank index (0=3, 1=4, ..., 12=2)
		}
		sort.Ints(faceRanks)

		isConsec := true
		for i := 1; i < len(faceRanks); i++ {
			if faceRanks[i]-faceRanks[i-1] != 1 {
				isConsec = false
				break
			}
		}

		// Check same trump category if trump context provided
		if isConsec && trumpSuit >= 0 {
			firstCat := ClassifyCard(Card{ID: pairFaces[0]}, trumpSuit, levelRank)
			for _, f := range pairFaces[1:] {
				if ClassifyCard(Card{ID: f}, trumpSuit, levelRank) != firstCat {
					isConsec = false
					break
				}
			}
		}

		if isConsec {
			return Play{
				Type:     PlayTractor,
				MainRank: faceRanks[0] + 3, // base rank of lowest pair
				Length:   len(pairFaces),
			}
		}
	}

	return Play{Type: PlayInvalid}
}

func playRank(card Card, trumpSuit int, levelRank int) int {
	if trumpSuit >= 0 || levelRank >= 0 {
		return card.CompareRank(trumpSuit, levelRank)
	}
	return card.BaseRank()
}

func groupPlayRank(cards []Card, trumpSuit int, levelRank int) int {
	rank := playRank(cards[0], trumpSuit, levelRank)
	for _, c := range cards[1:] {
		if r := playRank(c, trumpSuit, levelRank); r > rank {
			rank = r
		}
	}
	return rank
}

func parseJokerPlay(cards []Card, hasSJ bool, hasBJ bool, trumpSuit int, levelRank int) Play {
	if hasSJ && hasBJ {
		return Play{Type: PlayInvalid}
	}
	for _, c := range cards {
		if (hasSJ && !c.IsSmallJoker()) || (hasBJ && !c.IsBigJoker()) {
			return Play{Type: PlayInvalid}
		}
	}
	return parseMatchedGroup(cards, groupPlayRank(cards, trumpSuit, levelRank))
}

func parseMatchedGroup(cards []Card, rank int) Play {
	switch len(cards) {
	case 2:
		return Play{Type: PlayPair, MainRank: rank, Length: 1}
	case 3:
		return Play{Type: PlayTriple, MainRank: rank, Length: 1}
	default:
		return Play{Type: PlayInvalid}
	}
}

// CanBeat checks whether play beats lastPlay.
func CanBeat(play, lastPlay Play) bool {
	if lastPlay.Type == PlayPass {
		return true
	}
	if play.Type == PlayPass {
		return false
	}
	if play.Type != lastPlay.Type {
		return false
	}
	if play.Type == PlayTractor && play.Length != lastPlay.Length {
		return false
	}
	return play.MainRank > lastPlay.MainRank
}
