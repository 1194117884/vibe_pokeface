package doudizhu

import "sort"

// PlayType represents the type of a card play in Dou Di Zhu.
type PlayType int

const (
	PlayInvalid       PlayType = 0
	PlaySingle        PlayType = 1
	PlayPair          PlayType = 2
	PlayTriple        PlayType = 3
	PlayTriplePlus1   PlayType = 4
	PlayTriplePlus2   PlayType = 5
	PlayStraight      PlayType = 6
	PlayPairStraight  PlayType = 7
	PlayAirplane      PlayType = 8
	PlayAirplaneWings PlayType = 9
	PlayFourPlus2     PlayType = 10
	PlayBomb          PlayType = 11
	PlayRocket        PlayType = 12
	PlayPass          PlayType = 13
)

// Play represents a parsed card play with its type, main rank, and length.
type Play struct {
	Type     PlayType `json:"type"`
	MainRank int      `json:"main_rank"`
	Length   int      `json:"length"`
}

// ParsePlay analyzes a set of cards and determines the play type.
// Returns PlayInvalid if the cards don't form a valid play.
// Empty cards are treated as PlayPass.
func ParsePlay(cards []Card) Play {
	if len(cards) == 0 {
		return Play{Type: PlayPass}
	}

	// Build rank frequency map and sorted rank list.
	ranks := make([]int, len(cards))
	freq := make(map[int]int)
	for i, c := range cards {
		r := c.Rank()
		ranks[i] = r
		freq[r]++
	}
	sort.Ints(ranks)

	// Collect unique ranks sorted.
	uniqueRanks := make([]int, 0, len(freq))
	for r := range freq {
		uniqueRanks = append(uniqueRanks, r)
	}
	sort.Ints(uniqueRanks)

	// Count how many ranks have each frequency.
	countByFreq := make(map[int]int)
	for _, f := range freq {
		countByFreq[f]++
	}

	// Check for 2 and jokers (straights and some combos can't include them).
	has2 := false
	hasJoker := false
	for _, r := range uniqueRanks {
		if r == 15 {
			has2 = true
		}
		if r >= 16 {
			hasJoker = true
		}
	}

	// Rocket: both jokers.
	if len(cards) == 2 && freq[16] == 1 && freq[17] == 1 {
		return Play{Type: PlayRocket}
	}

	// Single.
	if len(cards) == 1 {
		return Play{Type: PlaySingle, MainRank: ranks[0], Length: 1}
	}

	// Pair.
	if len(cards) == 2 && countByFreq[2] == 1 {
		return Play{Type: PlayPair, MainRank: uniqueRanks[0]}
	}

	// Bomb: four of a kind.
	if len(cards) == 4 && countByFreq[4] == 1 {
		return Play{Type: PlayBomb, MainRank: uniqueRanks[0]}
	}

	// Triple.
	if len(cards) == 3 && countByFreq[3] == 1 {
		return Play{Type: PlayTriple, MainRank: uniqueRanks[0]}
	}

	// Triple + 1: three of a kind with a single kicker.
	if len(cards) == 4 && countByFreq[3] == 1 && countByFreq[1] == 1 {
		return Play{Type: PlayTriplePlus1, MainRank: findRankWithFreq(freq, 3)}
	}

	// Triple + 2: three of a kind with a pair kicker.
	if len(cards) == 5 && countByFreq[3] == 1 && countByFreq[2] == 1 {
		return Play{Type: PlayTriplePlus2, MainRank: findRankWithFreq(freq, 3)}
	}

	// Four + 2: four of a kind + two singles, or four + two pairs.
	if len(cards) == 6 && countByFreq[4] == 1 && countByFreq[1] == 2 {
		return Play{Type: PlayFourPlus2, MainRank: findRankWithFreq(freq, 4)}
	}
	if len(cards) == 8 && countByFreq[4] == 1 && countByFreq[2] == 2 {
		return Play{Type: PlayFourPlus2, MainRank: findRankWithFreq(freq, 4)}
	}

	// Straight: 5+ consecutive singles, no 2 or jokers.
	if countByFreq[1] == len(cards) && len(cards) >= 5 && !has2 && !hasJoker {
		if isConsecutive(uniqueRanks) {
			return Play{Type: PlayStraight, MainRank: uniqueRanks[0], Length: len(cards)}
		}
	}

	// Pair straight: 3+ consecutive pairs, no 2 or jokers.
	if countByFreq[2] == len(cards)/2 && len(cards) >= 6 && len(cards)%2 == 0 && !has2 && !hasJoker {
		if isConsecutive(uniqueRanks) {
			return Play{Type: PlayPairStraight, MainRank: uniqueRanks[0], Length: len(cards) / 2}
		}
	}

	// Airplane: 2+ consecutive triples with matching wings.
	if len(cards) >= 6 {
		tripleRanks := findRanksWithFreq(freq, 3)
		sort.Ints(tripleRanks)
		if len(tripleRanks) >= 2 && isConsecutive(tripleRanks) {
			has2orJoker := false
			for _, r := range tripleRanks {
				if r == 15 || r >= 16 {
					has2orJoker = true
				}
			}
			if !has2orJoker {
				wings := len(cards) - len(tripleRanks)*3
				// Wings can be singles or pairs.
				if wings == len(tripleRanks) || wings == len(tripleRanks)*2 {
					return Play{Type: PlayAirplane, MainRank: tripleRanks[0], Length: len(tripleRanks)}
				}
			}
		}
	}

	return Play{Type: PlayInvalid}
}

// CanBeat checks whether the play can beat the lastPlay.
// Returns true if play beats lastPlay according to Dou Di Zhu rules:
// - PlayPass beats only PlayPass (or when there is no last play).
// - Rocket beats everything.
// - Bomb beats everything except Rocket.
// - Same type and length: higher main rank wins.
// - Different types: cannot beat (except bombs and rockets).
func CanBeat(play, lastPlay Play) bool {
	if lastPlay.Type == PlayPass {
		return true
	}
	if play.Type == PlayPass {
		return false
	}
	if play.Type == PlayRocket {
		return true
	}
	if lastPlay.Type == PlayRocket {
		return false
	}
	if play.Type == PlayBomb && lastPlay.Type != PlayBomb {
		return true
	}
	if play.Type == lastPlay.Type && play.Length == lastPlay.Length {
		return play.MainRank > lastPlay.MainRank
	}
	return false
}

// findRankWithFreq returns the first rank with the given frequency.
func findRankWithFreq(freq map[int]int, target int) int {
	for r, f := range freq {
		if f == target {
			return r
		}
	}
	return 0
}

// findRanksWithFreq returns all ranks with the given frequency.
func findRanksWithFreq(freq map[int]int, target int) []int {
	var ranks []int
	for r, f := range freq {
		if f == target {
			ranks = append(ranks, r)
		}
	}
	return ranks
}

// isConsecutive checks if the sorted ranks form consecutive integers.
func isConsecutive(ranks []int) bool {
	if len(ranks) < 2 {
		return true
	}
	for i := 1; i < len(ranks); i++ {
		if ranks[i]-ranks[i-1] != 1 {
			return false
		}
	}
	return true
}
