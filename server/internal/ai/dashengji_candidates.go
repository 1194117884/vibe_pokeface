package ai

import (
	"fmt"
	"log"
	"sort"
	"strings"
)

type dashengjiCandidate struct {
	Tool   string
	Cards  []int
	Label  string
	Reason string
}

type dashengjiTrickView struct {
	WinnerSeat int
	Points     int
	OK         bool
}

func (a *AIAgent) dashengjiCandidatesForPhase(phase string) []dashengjiCandidate {
	state, _ := a.parseDashengjiState()
	switch phase {
	case "set_trump":
		return a.dashengjiTrumpCandidates("set_trump", state.LevelRank, false)
	case "counter_trump":
		candidates := a.dashengjiTrumpCandidates("counter_trump", state.LevelRank, true)
		if len(candidates) == 0 {
			return []dashengjiCandidate{{Tool: "pass_counter", Label: "不反主", Reason: "当前手牌没有合法反主组合"}}
		}
		return candidates
	case "take_bottom":
		return []dashengjiCandidate{{Tool: "take_bottom", Label: "自己起底", Reason: "默认由当前庄家队行动者起底"}}
	case "discard_bottom":
		hand := a.handCardsSnapshot()
		return []dashengjiCandidate{{
			Tool:   "discard_bottom",
			Cards:  chooseDashengjiDiscard(hand, state.TrumpSuit, state.LevelRank),
			Label:  "低价值扣底",
			Reason: "优先扣低价值副牌，保留主牌、分牌和对子结构",
		}}
	case "playing":
		return a.dashengjiPlayCandidates(state)
	default:
		return nil
	}
}

func (a *AIAgent) buildDashengjiCandidateHints(phase string) string {
	candidates := a.dashengjiCandidatesForPhase(phase)
	if len(candidates) == 0 {
		return "合法动作建议不可用；仍必须只使用当前手牌中的ID并遵守阶段规则。"
	}
	var sb strings.Builder
	sb.WriteString("## 合法动作建议\n")
	for i, c := range candidates {
		sb.WriteString(fmt.Sprintf("%d. tool=%s", i+1, c.Tool))
		if len(c.Cards) > 0 {
			sb.WriteString(fmt.Sprintf(" cards=%v %s", c.Cards, formatDashengjiCardsWithIDs(c.Cards)))
		} else {
			sb.WriteString(" cards=[]")
		}
		if c.Label != "" {
			sb.WriteString("；" + c.Label)
		}
		if c.Reason != "" {
			sb.WriteString("；" + c.Reason)
		}
		sb.WriteString("\n")
	}
	sb.WriteString("这些是推荐动作，不是唯一选择。你可以自行组合其他合法cards，但必须全部来自当前手牌，并严格满足阶段规则、跟牌张数、花色/主副类别和牌型约束。\n")
	return sb.String()
}

func (a *AIAgent) dashengjiCandidateByToolCards(tool string, cards []int) (dashengjiCandidate, bool) {
	for _, c := range a.dashengjiCandidatesForPhase(a.detectPhase()) {
		if c.Tool == tool && sameIntSetOrdered(c.Cards, cards) {
			return c, true
		}
	}
	return dashengjiCandidate{}, false
}

func (a *AIAgent) dashengjiFallbackCandidate(phase string) (dashengjiCandidate, bool) {
	candidates := a.dashengjiCandidatesForPhase(phase)
	if len(candidates) == 0 {
		return dashengjiCandidate{}, false
	}
	return candidates[0], true
}

func (a *AIAgent) isLegalDashengjiAction(tool string, cards []int) bool {
	if a.stateSnapshot() == "" {
		return true
	}
	state, ok := a.parseDashengjiState()
	if !ok {
		return false
	}
	phase := a.detectPhase()
	if !dashengjiCardsInHand(a.handCardsSnapshot(), cards) {
		return false
	}

	switch phase {
	case "set_trump":
		return tool == "set_trump" && dashengjiTrumpSelectionLegal(cards, state.LevelRank, false) ||
			tool == "pass_trump" && len(cards) == 0
	case "counter_trump":
		return tool == "counter_trump" && dashengjiTrumpSelectionLegal(cards, state.LevelRank, true) ||
			tool == "pass_counter" && len(cards) == 0
	case "take_bottom":
		return (tool == "take_bottom" || tool == "pass_take_bottom") && len(cards) == 0
	case "discard_bottom":
		return tool == "discard_bottom" && len(cards) == 6
	case "playing":
		return tool == "play_cards" && len(cards) > 0 && dashengjiPlayLegal(cards, state)
	default:
		return false
	}
}

func dashengjiCardsInHand(hand []int, cards []int) bool {
	available := make(map[int]int, len(hand))
	for _, id := range hand {
		available[id]++
	}
	for _, id := range cards {
		if available[id] == 0 {
			return false
		}
		available[id]--
	}
	return true
}

func dashengjiTrumpSelectionLegal(cards []int, levelRank int, counter bool) bool {
	if len(cards) < 3 || counter && len(cards) < 4 {
		return false
	}
	hasBigJoker := false
	hasSmallJoker := false
	twoSuit := -1
	levelCount := 0
	for _, id := range cards {
		face := id % 54
		if face == 53 {
			hasBigJoker = true
		}
		if face == 52 {
			hasSmallJoker = true
		}
		if face < 52 && face%13+3 == 15 {
			twoSuit = face / 13
		}
	}
	if twoSuit < 0 {
		return false
	}
	red := twoSuit == 1 || twoSuit == 3
	if red && !hasBigJoker || !red && !hasSmallJoker {
		return false
	}
	for _, id := range cards {
		face := id % 54
		if face < 52 && face/13 == twoSuit && face%13+3 == levelRank {
			levelCount++
		}
	}
	if counter {
		return levelCount >= 2
	}
	return levelCount >= 1
}

func dashengjiPlayLegal(cards []int, state dashengjiState) bool {
	if len(state.RoundPlays) == 0 {
		play := dashengjiParsePlay(cards, state.TrumpSuit, state.LevelRank)
		return play.Type != 0
	}
	led := state.RoundPlays[0]
	ledCards := cardIDs(led.Cards)
	if len(cards) != len(ledCards) {
		return false
	}
	if len(ledCards) == 0 {
		return dashengjiParsePlay(cards, state.TrumpSuit, state.LevelRank).Type != 0
	}

	ledSuit := dashengjiSuit(ledCards[0])
	ledCat := dashengjiCategory(ledCards[0]%54, state.TrumpSuit, state.LevelRank)
	availableMatching := 0
	playedMatching := 0
	currentHand := stateCurrentHandIDs(state)
	if currentHand == nil {
		return false
	}
	for _, id := range currentHand {
		if dashengjiFollowGroupMatch(id, ledSuit, ledCat, state.TrumpSuit, state.LevelRank) {
			availableMatching++
		}
	}
	for _, id := range cards {
		if dashengjiFollowGroupMatch(id, ledSuit, ledCat, state.TrumpSuit, state.LevelRank) {
			playedMatching++
		}
	}
	requiredMatching := len(ledCards)
	if availableMatching < requiredMatching {
		requiredMatching = availableMatching
	}
	if playedMatching < requiredMatching {
		return false
	}

	canFollow := dashengjiCanFollow(currentHand, led.Play, ledCards, state.TrumpSuit, state.LevelRank)
	if !canFollow {
		return true
	}

	play := dashengjiParsePlay(cards, state.TrumpSuit, state.LevelRank)
	if play.Type == 0 || play.Type != led.Play.Type {
		return false
	}
	if led.Play.Type == 4 && play.Length != led.Play.Length {
		return false
	}
	return dashengjiFollowGroupMatch(cards[0], ledSuit, ledCat, state.TrumpSuit, state.LevelRank)
}

func stateCurrentHandIDs(state dashengjiState) []int {
	for _, p := range state.Players {
		if p.Seat == state.CurrentSeat {
			return cardIDs(p.Hand)
		}
	}
	return nil
}

func dashengjiCanFollow(hand []int, ledPlay dashengjiPlay, ledCards []int, trumpSuit int, levelRank int) bool {
	if len(ledCards) == 0 {
		return false
	}
	matching, _ := splitDashengjiMatching(hand, ledCards[0], trumpSuit, levelRank)
	switch ledPlay.Type {
	case 1:
		return len(matching) >= 1
	case 2:
		for _, group := range firstDashengjiGroups(matching, 2, trumpSuit, levelRank) {
			if len(group) == 2 {
				return true
			}
		}
	case 3:
		for _, group := range firstDashengjiGroups(matching, 3, trumpSuit, levelRank) {
			if len(group) == 3 {
				return true
			}
		}
	case 4:
		return len(firstDashengjiTractor(matching, ledPlay.Length, trumpSuit, levelRank)) == ledPlay.Length*2
	}
	return false
}

func dashengjiParsePlay(cards []int, trumpSuit int, levelRank int) dashengjiPlay {
	if len(cards) == 0 {
		return dashengjiPlay{}
	}
	if len(cards) == 1 {
		return dashengjiPlay{Type: 1, MainRank: dashengjiCompareRank(cards[0], trumpSuit, levelRank), Length: 1}
	}
	hasSmall := false
	hasBig := false
	for _, id := range cards {
		face := id % 54
		if face == 52 {
			hasSmall = true
		}
		if face == 53 {
			hasBig = true
		}
	}
	if hasSmall || hasBig {
		if hasSmall && hasBig {
			return dashengjiPlay{}
		}
		for _, id := range cards {
			face := id % 54
			if hasSmall && face != 52 || hasBig && face != 53 {
				return dashengjiPlay{}
			}
		}
		if len(cards) == 2 {
			return dashengjiPlay{Type: 2, MainRank: dashengjiCompareRank(cards[0], trumpSuit, levelRank), Length: 1}
		}
		if len(cards) == 3 {
			return dashengjiPlay{Type: 3, MainRank: dashengjiCompareRank(cards[0], trumpSuit, levelRank), Length: 1}
		}
		return dashengjiPlay{}
	}

	suit := dashengjiSuit(cards[0])
	byFace := make(map[int][]int)
	for _, id := range cards {
		if dashengjiSuit(id) != suit {
			return dashengjiPlay{}
		}
		face := id % 54
		byFace[face] = append(byFace[face], id)
	}
	if len(byFace) == 1 {
		switch len(cards) {
		case 2:
			return dashengjiPlay{Type: 2, MainRank: dashengjiCompareRank(cards[0], trumpSuit, levelRank), Length: 1}
		case 3:
			return dashengjiPlay{Type: 3, MainRank: dashengjiCompareRank(cards[0], trumpSuit, levelRank), Length: 1}
		}
	}
	if len(cards)%2 != 0 || len(cards) < 6 {
		return dashengjiPlay{}
	}
	faces := make([]int, 0, len(byFace))
	for face, ids := range byFace {
		if len(ids) < 2 {
			return dashengjiPlay{}
		}
		faces = append(faces, face)
	}
	sort.Slice(faces, func(i, j int) bool { return faces[i]%13 < faces[j]%13 })
	if !dashengjiConsecutiveSameCategory(faces, trumpSuit, levelRank) {
		return dashengjiPlay{}
	}
	return dashengjiPlay{Type: 4, MainRank: dashengjiBaseRank(faces[0]), Length: len(faces)}
}

func sameIntSetOrdered(a []int, b []int) bool {
	if len(a) != len(b) {
		return false
	}
	aa := append([]int(nil), a...)
	bb := append([]int(nil), b...)
	sort.Ints(aa)
	sort.Ints(bb)
	for i := range aa {
		if aa[i] != bb[i] {
			return false
		}
	}
	return true
}

func (a *AIAgent) dashengjiTrumpCandidates(tool string, levelRank int, counter bool) []dashengjiCandidate {
	byFace := make(map[int][]int)
	for _, id := range a.handCardsSnapshot() {
		byFace[id%54] = append(byFace[id%54], id)
	}
	for _, ids := range byFace {
		sort.Ints(ids)
	}
	candidates := make([]dashengjiCandidate, 0)
	for suit := 0; suit < 4; suit++ {
		red := suit == 1 || suit == 3
		jokerFace := 52
		if red {
			jokerFace = 53
		}
		jokers := byFace[jokerFace]
		twos := byFace[suit*13+12]
		levels := byFace[suit*13+(levelRank-3)]
		needLevels := 1
		label := "合法定主"
		if counter {
			needLevels = 2
			label = "合法反主"
		}
		if len(jokers) == 0 || len(twos) == 0 || len(levels) < needLevels {
			continue
		}
		cards := []int{jokers[0], twos[0]}
		cards = append(cards, levels[:needLevels]...)
		candidates = append(candidates, dashengjiCandidate{
			Tool:   tool,
			Cards:  cards,
			Label:  label,
			Reason: fmt.Sprintf("%s王色 + %s2 + %d张同花色级牌", dashengjiColorName(red), dashengjiSuitName(suit), needLevels),
		})
	}
	if len(candidates) == 0 && !counter {
		return []dashengjiCandidate{{Tool: "pass_trump", Label: "不定主", Reason: "当前手牌没有合法定主组合"}}
	}
	return candidates
}

func (a *AIAgent) dashengjiPlayCandidates(state dashengjiState) []dashengjiCandidate {
	hand := a.handCardsSnapshot()
	sortDashengjiLowest(hand, state.TrumpSuit, state.LevelRank)
	if len(hand) == 0 {
		return nil
	}
	if len(state.RoundPlays) == 0 {
		return dashengjiLeadCandidates(hand, state)
	}
	return dashengjiFollowCandidates(hand, state)
}

func dashengjiLeadCandidates(hand []int, state dashengjiState) []dashengjiCandidate {
	candidates := make([]dashengjiCandidate, 0, 8)
	candidates = append(candidates, dashengjiCandidate{
		Tool:   "play_cards",
		Cards:  []int{hand[0]},
		Label:  "保守单张领出",
		Reason: "先出低价值牌探路",
	})
	for _, group := range firstDashengjiGroups(hand, 2, state.TrumpSuit, state.LevelRank) {
		reason := "利用对子结构施压"
		if risk := dashengjiLeadRisk(group, state); risk != "" {
			reason = risk
		}
		candidates = append(candidates, dashengjiCandidate{Tool: "play_cards", Cards: group, Label: "对子领出", Reason: reason})
		if len(candidates) >= 4 {
			break
		}
	}
	for _, group := range firstDashengjiGroups(hand, 3, state.TrumpSuit, state.LevelRank) {
		candidates = append(candidates, dashengjiCandidate{Tool: "play_cards", Cards: group, Label: "刻子领出", Reason: "利用刻子结构减少手牌"})
		if len(candidates) >= 6 {
			break
		}
	}
	candidates = uniqueDashengjiCandidates(candidates)
	sortDashengjiLeadCandidates(candidates, state)
	return candidates
}

func dashengjiFollowCandidates(hand []int, state dashengjiState) []dashengjiCandidate {
	led := state.RoundPlays[0]
	ledCards := cardIDs(led.Cards)
	if len(ledCards) == 0 {
		return dashengjiLeadCandidates(hand, state)
	}

	matching, others := splitDashengjiMatching(hand, ledCards[0], state.TrumpSuit, state.LevelRank)
	need := len(ledCards)
	view := dashengjiCurrentTrickView(state)
	if len(matching) >= need {
		if plays := dashengjiSameTypeCandidates(matching, led.Play, state.TrumpSuit, state.LevelRank); len(plays) > 0 {
			candidates := make([]dashengjiCandidate, 0, len(plays))
			for _, play := range plays {
				candidates = append(candidates, dashengjiCandidate{
					Tool:   "play_cards",
					Cards:  play,
					Label:  "严格跟牌",
					Reason: dashengjiFollowReason(play, state, view),
				})
			}
			candidates = uniqueDashengjiCandidates(candidates)
			sortDashengjiFollowCandidates(candidates, state, view)
			return candidates
		}
	}

	// Cannot follow exact type: still satisfy the engine's required matching count.
	candidates := make([]dashengjiCandidate, 0, 4)
	if len(matching) == 0 {
		for _, play := range dashengjiTrumpingCandidates(others, led.Play, state.TrumpSuit, state.LevelRank) {
			if len(play) != need {
				continue
			}
			candidates = append(candidates, dashengjiCandidate{
				Tool:   "play_cards",
				Cards:  play,
				Label:  "主牌枪毙",
				Reason: dashengjiFollowReason(play, state, view),
			})
			if len(candidates) >= 3 {
				break
			}
		}
	}

	cards := make([]int, 0, need)
	takeMatching := len(matching)
	if takeMatching > need {
		takeMatching = need
	}
	cards = append(cards, matching[:takeMatching]...)
	for _, id := range others {
		if len(cards) >= need {
			break
		}
		cards = append(cards, id)
	}
	if len(cards) == need {
		candidates = append(candidates, dashengjiCandidate{
			Tool:   "play_cards",
			Cards:  cards,
			Label:  "合法垫牌",
			Reason: dashengjiFollowReason(cards, state, view),
		})
		candidates = uniqueDashengjiCandidates(candidates)
		sortDashengjiFollowCandidates(candidates, state, view)
		return candidates
	}
	return nil
}

func dashengjiSameTypeCandidates(cards []int, led dashengjiPlay, trumpSuit int, levelRank int) [][]int {
	switch led.Type {
	case 1:
		if len(cards) == 0 {
			return nil
		}
		plays := make([][]int, 0, len(cards))
		for _, id := range cards {
			plays = append(plays, []int{id})
		}
		return plays
	case 2:
		return firstDashengjiGroups(cards, 2, trumpSuit, levelRank)
	case 3:
		return firstDashengjiGroups(cards, 3, trumpSuit, levelRank)
	case 4:
		if tractor := firstDashengjiTractor(cards, led.Length, trumpSuit, levelRank); len(tractor) == led.Length*2 {
			return [][]int{tractor}
		}
		return nil
	default:
		return nil
	}
}

func dashengjiTrumpingCandidates(cards []int, led dashengjiPlay, trumpSuit int, levelRank int) [][]int {
	main := make([]int, 0, len(cards))
	for _, id := range cards {
		if dashengjiIsMainCategory(dashengjiCategory(id%54, trumpSuit, levelRank)) {
			main = append(main, id)
		}
	}
	return dashengjiSameTypeCandidates(main, led, trumpSuit, levelRank)
}

func sortDashengjiLeadCandidates(candidates []dashengjiCandidate, state dashengjiState) {
	sort.SliceStable(candidates, func(i, j int) bool {
		si := dashengjiLeadScore(candidates[i], state)
		sj := dashengjiLeadScore(candidates[j], state)
		if si != sj {
			return si > sj
		}
		return dashengjiCardsTotalRank(candidates[i].Cards, state.TrumpSuit, state.LevelRank) < dashengjiCardsTotalRank(candidates[j].Cards, state.TrumpSuit, state.LevelRank)
	})
}

func dashengjiLeadScore(candidate dashengjiCandidate, state dashengjiState) int {
	score := 100 - dashengjiCardsTotalRank(candidate.Cards, state.TrumpSuit, state.LevelRank)/10
	points := dashengjiCardsPoints(candidate.Cards)
	score -= points * 8
	if strings.Contains(candidate.Reason, "对A未出") {
		score -= 90
	}
	play := dashengjiParsePlay(candidate.Cards, state.TrumpSuit, state.LevelRank)
	if play.Type == 2 || play.Type == 3 {
		score += 20
	}
	if dashengjiIsMainCategory(dashengjiCategory(candidate.Cards[0]%54, state.TrumpSuit, state.LevelRank)) {
		score -= 20
	}
	return score
}

func sortDashengjiFollowCandidates(candidates []dashengjiCandidate, state dashengjiState, view dashengjiTrickView) {
	sort.SliceStable(candidates, func(i, j int) bool {
		si := dashengjiFollowScore(candidates[i], state, view)
		sj := dashengjiFollowScore(candidates[j], state, view)
		if si != sj {
			return si > sj
		}
		return dashengjiCardsTotalRank(candidates[i].Cards, state.TrumpSuit, state.LevelRank) < dashengjiCardsTotalRank(candidates[j].Cards, state.TrumpSuit, state.LevelRank)
	})
}

func dashengjiFollowScore(candidate dashengjiCandidate, state dashengjiState, view dashengjiTrickView) int {
	score := 100
	cardPoints := dashengjiCardsPoints(candidate.Cards)
	canWin := dashengjiCandidateWins(candidate.Cards, state)
	if view.OK && dashengjiSameTeam(state.CurrentSeat, view.WinnerSeat) {
		score += cardPoints * 30
		if canWin && view.WinnerSeat != state.CurrentSeat {
			score -= 60
		}
	} else {
		if canWin {
			score += (view.Points + cardPoints) * 40
			if !dashengjiIsDealer(state.CurrentSeat, state) && state.RoundPoints < 120 {
				score += 60
			}
		} else {
			score -= cardPoints * 30
		}
	}
	score -= dashengjiCardsTotalRank(candidate.Cards, state.TrumpSuit, state.LevelRank) / 15
	if strings.Contains(candidate.Label, "主牌枪毙") && (view.Points+cardPoints) == 0 {
		score -= 80
	}
	return score
}

func dashengjiFollowReason(cards []int, state dashengjiState, view dashengjiTrickView) string {
	cardPoints := dashengjiCardsPoints(cards)
	totalPoints := view.Points + cardPoints
	canWin := dashengjiCandidateWins(cards, state)
	if view.OK && dashengjiSameTeam(state.CurrentSeat, view.WinnerSeat) {
		if cardPoints > 0 {
			return fmt.Sprintf("队友当前最大，送%d分给队友收", cardPoints)
		}
		return "队友当前最大，垫低价值牌保留控制力"
	}
	if canWin && totalPoints > 0 {
		return fmt.Sprintf("能抢回本轮%d分，优先抢分", totalPoints)
	}
	if canWin {
		return "可拿回牌权，但本轮无分，避免过度消耗"
	}
	if cardPoints > 0 {
		return "对手当前最大且抢不回，避免继续送分"
	}
	return "对手当前最大，垫低价值牌避分"
}

func dashengjiLeadRisk(cards []int, state dashengjiState) string {
	if len(cards) < 2 {
		return ""
	}
	face := cards[0] % 54
	base := dashengjiBaseRank(face)
	if base != 13 || dashengjiIsMainCategory(dashengjiCategory(face, state.TrumpSuit, state.LevelRank)) {
		return ""
	}
	suit := dashengjiSuit(cards[0])
	aceFace := suit*13 + (14 - 3)
	if dashengjiVisibleFaceCount(state, aceFace) < 2 {
		return fmt.Sprintf("%s对K有风险：对A未出，不能贸然强攻", dashengjiSuitName(suit))
	}
	return ""
}

func dashengjiVisibleFaceCount(state dashengjiState, face int) int {
	count := 0
	for _, record := range state.PlayHistory {
		for _, card := range record.Cards {
			if card.ID%54 == face {
				count++
			}
		}
	}
	for _, record := range state.RoundPlays {
		for _, card := range record.Cards {
			if card.ID%54 == face {
				count++
			}
		}
	}
	return count
}

func dashengjiCurrentTrickView(state dashengjiState) dashengjiTrickView {
	view := dashengjiTrickView{WinnerSeat: -1}
	if len(state.RoundPlays) == 0 {
		return view
	}
	winner, ok := dashengjiWinnerForPlays(state.RoundPlays, state.TrumpSuit, state.LevelRank)
	if !ok {
		return view
	}
	view.WinnerSeat = winner
	view.Points = dashengjiRoundPlayPoints(state.RoundPlays)
	view.OK = true
	return view
}

func dashengjiCandidateWins(cards []int, state dashengjiState) bool {
	if len(state.RoundPlays) == 0 {
		return true
	}
	play := dashengjiParsePlay(cards, state.TrumpSuit, state.LevelRank)
	if play.Type == 0 {
		return false
	}
	plays := append([]dashengjiPlayRecord(nil), state.RoundPlays...)
	plays = append(plays, dashengjiPlayRecord{
		Seat:  state.CurrentSeat,
		Play:  play,
		Cards: dashengjiCardsFromIDs(cards),
	})
	winner, ok := dashengjiWinnerForPlays(plays, state.TrumpSuit, state.LevelRank)
	return ok && winner == state.CurrentSeat
}

func dashengjiWinnerForPlays(plays []dashengjiPlayRecord, trumpSuit int, levelRank int) (int, bool) {
	if len(plays) == 0 || len(plays[0].Cards) == 0 {
		return -1, false
	}
	winner := plays[0].Seat
	bestPlay := plays[0].Play
	bestIsTrump := false
	bestIsPadding := true
	ledID := plays[0].Cards[0].ID
	ledSuit := dashengjiSuit(ledID)
	ledCat := dashengjiCategory(ledID%54, trumpSuit, levelRank)
	ledIsMain := dashengjiIsMainCategory(ledCat)

	for _, r := range plays {
		if len(r.Cards) == 0 {
			continue
		}
		id := r.Cards[0].ID
		playCat := dashengjiCategory(id%54, trumpSuit, levelRank)
		isFollowing := dashengjiFollowGroupMatch(id, ledSuit, ledCat, trumpSuit, levelRank)
		isTrump := !isFollowing && !ledIsMain && dashengjiIsMainCategory(playCat)
		isPadding := !isFollowing && !isTrump
		if isPadding {
			continue
		}
		if bestIsPadding {
			winner = r.Seat
			bestPlay = r.Play
			bestIsTrump = isTrump
			bestIsPadding = false
			continue
		}
		if isTrump && !bestIsTrump {
			winner = r.Seat
			bestPlay = r.Play
			bestIsTrump = true
			continue
		}
		if !isTrump && bestIsTrump {
			continue
		}
		if r.Play.MainRank > bestPlay.MainRank {
			winner = r.Seat
			bestPlay = r.Play
			bestIsTrump = isTrump
		}
	}
	return winner, true
}

func dashengjiCardsFromIDs(ids []int) []dashengjiCard {
	cards := make([]dashengjiCard, len(ids))
	for i, id := range ids {
		cards[i] = dashengjiCard{ID: id}
	}
	return cards
}

func dashengjiRoundPlayPoints(plays []dashengjiPlayRecord) int {
	points := 0
	for _, record := range plays {
		points += dashengjiCardsPoints(cardIDs(record.Cards))
	}
	return points
}

func dashengjiCardsPoints(cards []int) int {
	points := 0
	for _, id := range cards {
		switch dashengjiBaseRank(id % 54) {
		case 5:
			points += 5
		case 10, 13:
			points += 10
		}
	}
	return points
}

func dashengjiCardsTotalRank(cards []int, trumpSuit int, levelRank int) int {
	total := 0
	for _, id := range cards {
		total += dashengjiCompareRank(id, trumpSuit, levelRank)
	}
	return total
}

func dashengjiSameTeam(a int, b int) bool {
	return a >= 0 && b >= 0 && a%2 == b%2
}

func dashengjiIsDealer(seat int, state dashengjiState) bool {
	return seat == state.DealerSeats[0] || seat == state.DealerSeats[1]
}

func firstDashengjiGroups(cards []int, n int, trumpSuit int, levelRank int) [][]int {
	byFace := make(map[int][]int)
	for _, id := range cards {
		face := id % 54
		byFace[face] = append(byFace[face], id)
	}
	faces := make([]int, 0, len(byFace))
	for face := range byFace {
		faces = append(faces, face)
	}
	sort.Slice(faces, func(i, j int) bool {
		ri := dashengjiCompareRank(faces[i], trumpSuit, levelRank)
		rj := dashengjiCompareRank(faces[j], trumpSuit, levelRank)
		if ri != rj {
			return ri < rj
		}
		return faces[i] < faces[j]
	})
	groups := make([][]int, 0)
	for _, face := range faces {
		ids := byFace[face]
		sort.Ints(ids)
		if len(ids) >= n {
			groups = append(groups, append([]int(nil), ids[:n]...))
		}
	}
	return groups
}

func uniqueDashengjiCandidates(candidates []dashengjiCandidate) []dashengjiCandidate {
	seen := make(map[string]bool)
	result := make([]dashengjiCandidate, 0, len(candidates))
	for _, c := range candidates {
		key := fmt.Sprintf("%s:%v", c.Tool, c.Cards)
		if seen[key] {
			continue
		}
		seen[key] = true
		result = append(result, c)
	}
	return result
}

func dashengjiColorName(red bool) string {
	if red {
		return "红色"
	}
	return "黑色"
}

func logDashengjiCandidateRejected(userID string, tool string, cards []int, fallback dashengjiCandidate) {
	log.Printf("[AI:%s] CANDIDATE-REJECT tool=%s cards=%v fallback_tool=%s fallback_cards=%v reason=not_in_legal_candidates",
		userID, tool, cards, fallback.Tool, fallback.Cards)
}
