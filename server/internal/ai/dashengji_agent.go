package ai

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
)

type dashengjiCard struct {
	ID int `json:"id"`
}

type dashengjiPlay struct {
	Type     int `json:"type"`
	MainRank int `json:"main_rank"`
	Length   int `json:"length"`
}

type dashengjiPlayRecord struct {
	Seat  int             `json:"seat"`
	Play  dashengjiPlay   `json:"play"`
	Cards []dashengjiCard `json:"cards"`
}

type dashengjiState struct {
	Phase        int                   `json:"phase"`
	CurrentSeat  int                   `json:"current_seat"`
	DealerSeats  [2]int                `json:"dealer_seats"`
	TeamLevels   [2]int                `json:"team_levels"`
	CurrentLevel int                   `json:"current_level"`
	LevelRank    int                   `json:"level_rank"`
	TrumpSuit    int                   `json:"trump_suit"`
	RoundPoints  int                   `json:"round_points"`
	Players      []dashengjiPlayerHand `json:"players"`
	BottomCards  []dashengjiCard       `json:"bottom_cards"`
	RoundPlays   []dashengjiPlayRecord `json:"round_plays"`
	PlayHistory  []dashengjiPlayRecord `json:"play_history"`
}

type dashengjiPlayerHand struct {
	UserID int64           `json:"user_id"`
	Seat   int             `json:"seat"`
	Hand   []dashengjiCard `json:"hand"`
}

func (a *AIAgent) detectDashengjiPhase() string {
	if a.stateJSON == "" {
		return "set_trump"
	}
	var state struct {
		Phase int `json:"phase"`
	}
	if err := json.Unmarshal([]byte(a.stateJSON), &state); err != nil {
		return "set_trump"
	}
	switch state.Phase {
	case 0:
		return "set_trump"
	case 1:
		return "counter_trump"
	case 2:
		return "take_bottom"
	case 3:
		return "discard_bottom"
	case 4:
		return "playing"
	default:
		return "playing"
	}
}

func (a *AIAgent) buildDashengjiSystemPrompt(phase string) string {
	name := "AI玩家"
	personality := "冷静分析，稳健出牌"
	playStyle := "balanced"
	if a.Character != nil {
		if a.Character.Name != "" {
			name = a.Character.Name
		}
		if a.Character.Personality != nil && *a.Character.Personality != "" {
			personality = *a.Character.Personality
		}
		playStyle = a.Character.PlayStyle
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("你在进行一场四人大升级/双升比赛，你是「%s」。\n", name))
	sb.WriteString(fmt.Sprintf("性格：%s。出牌风格：%s。\n", personality, playStyle))
	sb.WriteString("必须调用当前阶段允许的工具，所有牌都必须使用整数ID。\n")
	sb.WriteString("阶段：定主、反主、起底、扣底、出牌。出牌阶段必须出牌，不能过牌。\n")
	sb.WriteString("出牌阶段调用play_cards时，必须附带chat字段，说一句符合你性格的简短台词（不超过30字）。\n")
	sb.WriteString("跟牌必须尽量跟领出花色/主牌组和牌型；无法跟牌型时按规则垫牌或用主牌枪毙。\n")
	sb.WriteString("如果不确定，选择保守合法动作。\n")
	sb.WriteString(fmt.Sprintf("当前阶段：%s。\n", phase))
	return sb.String()
}

func (a *AIAgent) buildDashengjiUserMessage(phase string) string {
	status := a.buildDashengjiStatus()
	records := a.buildDashengjiPlayingRecords()
	return fmt.Sprintf("## 当前阶段：%s\n%s\n%s\n请根据当前阶段调用一个动作工具。", phase, status, records)
}

func (a *AIAgent) buildDashengjiStatus() string {
	state, ok := a.parseDashengjiState()
	if !ok {
		return "游戏状态不可用"
	}
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("你的座位：%d\n", a.Seat))
	sb.WriteString(fmt.Sprintf("当前轮到座位：%d\n", state.CurrentSeat))
	sb.WriteString(fmt.Sprintf("庄家队座位：%d,%d\n", state.DealerSeats[0], state.DealerSeats[1]))
	sb.WriteString(fmt.Sprintf("当前级牌：%d，主花色：%s\n", state.LevelRank, dashengjiSuitName(state.TrumpSuit)))
	sb.WriteString(fmt.Sprintf("闲家累计抢分：%d\n", state.RoundPoints))
	sb.WriteString("各玩家剩余手牌：\n")
	for _, p := range state.Players {
		role := "闲家"
		if p.Seat == state.DealerSeats[0] || p.Seat == state.DealerSeats[1] {
			role = "庄家队"
		}
		sb.WriteString(fmt.Sprintf("  座位%d（%s）：%d张\n", p.Seat, role, len(p.Hand)))
	}
	if len(state.RoundPlays) > 0 {
		led := state.RoundPlays[0]
		sb.WriteString(fmt.Sprintf("本轮领出座位%d：%s\n", led.Seat, formatDashengjiCardsWithIDs(cardIDs(led.Cards))))
	}
	return sb.String()
}

func (a *AIAgent) buildDashengjiPlayingRecords() string {
	state, ok := a.parseDashengjiState()
	if !ok {
		return "出牌记录不可用"
	}
	if len(state.PlayHistory) == 0 {
		return "暂无出牌记录。"
	}
	var sb strings.Builder
	sb.WriteString("出牌记录：\n")
	start := len(state.PlayHistory) - 8
	if start < 0 {
		start = 0
	}
	for i := start; i < len(state.PlayHistory); i++ {
		record := state.PlayHistory[i]
		sb.WriteString(fmt.Sprintf("  %d. 座位%d：%s\n", i+1, record.Seat, formatDashengjiCardsWithIDs(cardIDs(record.Cards))))
	}
	return sb.String()
}

func (a *AIAgent) executeDashengjiToolCall(call *ToolCall) {
	if a.Executor == nil {
		return
	}
	switch call.Name {
	case "set_trump", "counter_trump", "discard_bottom":
		var args PlayCardsArgs
		if err := json.Unmarshal(call.Args, &args); err != nil {
			a.fallbackDashengjiAction(a.detectPhase())
			return
		}
		a.Executor.ExecuteAction(a.UserID, call.Name, args.Cards)
		if args.Chat != "" {
			a.Executor.SendChat(a.UserID, args.Chat, "text")
		}
	case "play_cards":
		var args PlayCardsArgs
		if err := json.Unmarshal(call.Args, &args); err != nil || len(args.Cards) == 0 {
			a.fallbackDashengjiAction(a.detectPhase())
			return
		}
		a.Executor.ExecuteAction(a.UserID, "play", args.Cards)
		if args.Chat != "" {
			a.Executor.SendChat(a.UserID, args.Chat, "text")
		}
	case "pass_trump", "pass_counter", "take_bottom", "pass_take_bottom":
		a.Executor.ExecuteAction(a.UserID, call.Name, nil)
	default:
		a.fallbackDashengjiAction(a.detectPhase())
	}
}

func (a *AIAgent) fallbackDashengjiAction(phase string) {
	if a.Executor == nil {
		return
	}
	state, _ := a.parseDashengjiState()
	switch phase {
	case "set_trump":
		if cards := findDashengjiTrumpCards(a.HandCards, state.LevelRank, false); len(cards) > 0 {
			a.Executor.ExecuteAction(a.UserID, "set_trump", cards)
			return
		}
		a.Executor.ExecuteAction(a.UserID, "pass_trump", nil)
	case "counter_trump":
		if cards := findDashengjiTrumpCards(a.HandCards, state.LevelRank, true); len(cards) > 0 {
			a.Executor.ExecuteAction(a.UserID, "counter_trump", cards)
			return
		}
		a.Executor.ExecuteAction(a.UserID, "pass_counter", nil)
	case "take_bottom":
		a.Executor.ExecuteAction(a.UserID, "take_bottom", nil)
	case "discard_bottom":
		a.Executor.ExecuteAction(a.UserID, "discard_bottom", chooseDashengjiDiscard(a.HandCards, state.TrumpSuit, state.LevelRank))
	default:
		a.Executor.ExecuteAction(a.UserID, "play", chooseDashengjiPlay(a.HandCards, state))
	}
}

func (a *AIAgent) parseDashengjiState() (dashengjiState, bool) {
	var state dashengjiState
	state.TrumpSuit = -1
	state.LevelRank = 3
	if a.stateJSON == "" {
		return state, false
	}
	if err := json.Unmarshal([]byte(a.stateJSON), &state); err != nil {
		return state, false
	}
	if state.LevelRank == 0 {
		state.LevelRank = state.CurrentLevel
	}
	if state.LevelRank == 0 {
		state.LevelRank = 3
	}
	return state, true
}

func findDashengjiTrumpCards(hand []int, levelRank int, counter bool) []int {
	byFace := make(map[int][]int)
	for _, id := range hand {
		byFace[id%54] = append(byFace[id%54], id)
	}
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
		if counter {
			needLevels = 2
		}
		if len(jokers) > 0 && len(twos) > 0 && len(levels) >= needLevels {
			cards := []int{jokers[0], twos[0]}
			cards = append(cards, levels[:needLevels]...)
			return cards
		}
	}
	return nil
}

func chooseDashengjiDiscard(hand []int, trumpSuit int, levelRank int) []int {
	cards := append([]int(nil), hand...)
	sort.Slice(cards, func(i, j int) bool {
		vi := dashengjiDiscardValue(cards[i], trumpSuit, levelRank)
		vj := dashengjiDiscardValue(cards[j], trumpSuit, levelRank)
		if vi != vj {
			return vi < vj
		}
		return cards[i] < cards[j]
	})
	if len(cards) > 6 {
		cards = cards[:6]
	}
	return cards
}

func chooseDashengjiPlay(hand []int, state dashengjiState) []int {
	if len(hand) == 0 {
		return nil
	}
	cards := append([]int(nil), hand...)
	sortDashengjiLowest(cards, state.TrumpSuit, state.LevelRank)
	if len(state.RoundPlays) == 0 {
		return []int{cards[0]}
	}

	led := state.RoundPlays[0]
	ledCards := cardIDs(led.Cards)
	if len(ledCards) == 0 {
		return []int{cards[0]}
	}

	matching, others := splitDashengjiMatching(cards, ledCards[0], state.TrumpSuit, state.LevelRank)
	need := len(ledCards)
	if len(matching) >= need {
		if play := chooseDashengjiSameType(matching, led.Play, state.TrumpSuit, state.LevelRank); len(play) == need {
			return play
		}
	}

	result := make([]int, 0, need)
	takeMatching := len(matching)
	if takeMatching > need {
		takeMatching = need
	}
	result = append(result, matching[:takeMatching]...)
	for _, id := range others {
		if len(result) >= need {
			break
		}
		result = append(result, id)
	}
	return result
}

func chooseDashengjiSameType(cards []int, led dashengjiPlay, trumpSuit int, levelRank int) []int {
	switch led.Type {
	case 1:
		return []int{cards[0]}
	case 2:
		return firstDashengjiNOfFace(cards, 2)
	case 3:
		return firstDashengjiNOfFace(cards, 3)
	case 4:
		return firstDashengjiTractor(cards, led.Length, trumpSuit, levelRank)
	default:
		return nil
	}
}

func firstDashengjiNOfFace(cards []int, n int) []int {
	byFace := make(map[int][]int)
	for _, id := range cards {
		face := id % 54
		byFace[face] = append(byFace[face], id)
		if len(byFace[face]) == n {
			return append([]int(nil), byFace[face]...)
		}
	}
	return nil
}

func firstDashengjiTractor(cards []int, length int, trumpSuit int, levelRank int) []int {
	if length < 3 {
		return nil
	}
	byFace := make(map[int][]int)
	faces := make([]int, 0)
	for _, id := range cards {
		face := id % 54
		byFace[face] = append(byFace[face], id)
		if len(byFace[face]) == 2 {
			faces = append(faces, face)
		}
	}
	sort.Slice(faces, func(i, j int) bool {
		return dashengjiBaseRank(faces[i]) < dashengjiBaseRank(faces[j])
	})
	for i := 0; i+length <= len(faces); i++ {
		run := faces[i : i+length]
		if !dashengjiConsecutiveSameCategory(run, trumpSuit, levelRank) {
			continue
		}
		result := make([]int, 0, length*2)
		for _, face := range run {
			result = append(result, byFace[face][:2]...)
		}
		return result
	}
	return nil
}

func dashengjiConsecutiveSameCategory(faces []int, trumpSuit int, levelRank int) bool {
	if len(faces) == 0 {
		return false
	}
	cat := dashengjiCategory(faces[0], trumpSuit, levelRank)
	for i := 1; i < len(faces); i++ {
		if dashengjiCategory(faces[i], trumpSuit, levelRank) != cat {
			return false
		}
		if faces[i]%13-faces[i-1]%13 != 1 {
			return false
		}
	}
	return true
}

func splitDashengjiMatching(hand []int, ledID int, trumpSuit int, levelRank int) ([]int, []int) {
	ledSuit := dashengjiSuit(ledID)
	ledCat := dashengjiCategory(ledID%54, trumpSuit, levelRank)
	matching := make([]int, 0)
	others := make([]int, 0)
	for _, id := range hand {
		if dashengjiFollowGroupMatch(id, ledSuit, ledCat, trumpSuit, levelRank) {
			matching = append(matching, id)
		} else {
			others = append(others, id)
		}
	}
	return matching, others
}

func dashengjiFollowGroupMatch(id int, ledSuit int, ledCat int, trumpSuit int, levelRank int) bool {
	cat := dashengjiCategory(id%54, trumpSuit, levelRank)
	if dashengjiIsMainCategory(ledCat) {
		return dashengjiIsMainCategory(cat)
	}
	return dashengjiSuit(id) == ledSuit && cat == ledCat
}

func dashengjiIsMainCategory(cat int) bool {
	return cat == 1 || cat == 2 || cat == 3 || cat == 4
}

func sortDashengjiLowest(cards []int, trumpSuit int, levelRank int) {
	sort.Slice(cards, func(i, j int) bool {
		ri := dashengjiCompareRank(cards[i], trumpSuit, levelRank)
		rj := dashengjiCompareRank(cards[j], trumpSuit, levelRank)
		if ri != rj {
			return ri < rj
		}
		return cards[i] < cards[j]
	})
}

func dashengjiDiscardValue(id int, trumpSuit int, levelRank int) int {
	value := dashengjiCompareRank(id, trumpSuit, levelRank)
	base := dashengjiBaseRank(id % 54)
	if base == 5 || base == 10 || base == 13 {
		value += 80
	}
	if dashengjiIsMainCategory(dashengjiCategory(id%54, trumpSuit, levelRank)) {
		value += 120
	}
	return value
}

func dashengjiCompareRank(id int, trumpSuit int, levelRank int) int {
	face := id % 54
	if face == 53 {
		return 200
	}
	if face == 52 {
		return 190
	}
	base := dashengjiBaseRank(face)
	suit := dashengjiSuit(id)
	isTrumpSuit := suit == trumpSuit
	isLevelCard := base == levelRank
	isTwo := base == 15
	if isTrumpSuit && isLevelCard {
		return 180
	}
	if !isTrumpSuit && isLevelCard {
		return 170
	}
	if isTrumpSuit && isTwo {
		return 160
	}
	if !isTrumpSuit && isTwo {
		return 150
	}
	if isTrumpSuit {
		return 100 + base
	}
	return base
}

func dashengjiCategory(face int, trumpSuit int, levelRank int) int {
	if face == 52 || face == 53 {
		return 4
	}
	suit := face / 13
	base := dashengjiBaseRank(face)
	isTrumpSuit := suit == trumpSuit
	isLevelCard := base == levelRank
	isTwo := base == 15
	if isTrumpSuit && (isLevelCard || isTwo) {
		return 3
	}
	if !isTrumpSuit && (isLevelCard || isTwo) {
		return 2
	}
	if isTrumpSuit {
		return 1
	}
	return 0
}

func dashengjiSuit(id int) int {
	face := id % 54
	if face >= 52 {
		return 4
	}
	return face / 13
}

func dashengjiBaseRank(face int) int {
	if face == 52 {
		return 16
	}
	if face == 53 {
		return 17
	}
	return face%13 + 3
}

func formatDashengjiCardsWithIDs(cards []int) string {
	suits := []string{"♠", "♥", "♣", "♦"}
	ranks := []string{"3", "4", "5", "6", "7", "8", "9", "10", "J", "Q", "K", "A", "2"}
	parts := make([]string, len(cards))
	for i, id := range cards {
		face := id % 54
		copyNo := id/54 + 1
		switch face {
		case 52:
			parts[i] = fmt.Sprintf("%d(小王#%d)", id, copyNo)
		case 53:
			parts[i] = fmt.Sprintf("%d(大王#%d)", id, copyNo)
		default:
			parts[i] = fmt.Sprintf("%d(%s%s#%d)", id, suits[face/13], ranks[face%13], copyNo)
		}
	}
	return strings.Join(parts, " ")
}

func cardIDs(cards []dashengjiCard) []int {
	ids := make([]int, len(cards))
	for i, c := range cards {
		ids[i] = c.ID
	}
	return ids
}

func dashengjiSuitName(suit int) string {
	switch suit {
	case 0:
		return "♠"
	case 1:
		return "♥"
	case 2:
		return "♣"
	case 3:
		return "♦"
	default:
		return "未定"
	}
}
