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
	stateJSON := a.stateSnapshot()
	if stateJSON == "" {
		return "set_trump"
	}
	var state struct {
		Phase int `json:"phase"`
	}
	if err := json.Unmarshal([]byte(stateJSON), &state); err != nil {
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
	sb.WriteString(fmt.Sprintf("你在进行一场四人打升级/双升比赛，你是「%s」。\n", name))
	sb.WriteString(fmt.Sprintf("性格：%s。出牌风格：%s。\n", personality, playStyle))
	sb.WriteString("必须调用当前阶段允许的工具；所有牌都必须使用整数ID，并且只能使用当前手牌里的整数ID。\n\n")
	sb.WriteString("## 全局规则\n")
	sb.WriteString("- 这是河北定州四人打升级：3副牌，4人2v2，对座为队友（0-2、1-3）。\n")
	sb.WriteString("- 庄家队目标是跑分并守庄/升级；闲家队目标是抢分，达到下庄或升级。\n")
	sb.WriteString("- 分牌：5=5分，10/K=10分；闲家赢一轮才累计该轮分，庄家赢一轮则这些分被跑掉。\n")
	sb.WriteString("- 主牌：大王、小王、所有2、所有级牌、主花色牌。副牌是非主牌。\n")
	sb.WriteString("- 大小顺序：副牌 < 主花色牌 < 副2 < 本2 < 副级牌 < 本级牌 < 小王 < 大王。\n")
	sb.WriteString("- 支持牌型：单张、对子、刻子、拖拉机；领出后，本轮所有人必须出相同张数。\n\n")
	sb.WriteString("## 出牌决策顺序\n")
	sb.WriteString("- 先判主副：所有级牌和所有2都是主牌，不能再当原花色副牌跟出。\n")
	sb.WriteString("- 再判跟牌：先满足领出花色/主副类别、牌型和张数，不能为了抢分破坏硬约束。\n")
	sb.WriteString("- 再判本轮谁最大：队友最大就送分或垫低牌；对手最大才考虑抢回或避分。\n")
	sb.WriteString("- 再判本轮分数：有分且能确定抢回时积极抢；抢不回时避免扔5、10、K。\n")
	sb.WriteString("- 领出对子/高分牌前先看关键大牌是否已出；同花色对A未出时不要贸然领出对K。\n\n")
	sb.WriteString("## 分阶段规则\n")
	sb.WriteString("- 定主阶段：庄家队行动。只有王 + 同色2 + 同花色级牌才可set_trump；没有合法组合就pass_trump。带两张同花色级牌属于定死，通常更强。\n")
	sb.WriteString("- 反主阶段：闲家队行动。只有王 + 同色2 + 两张同花色级牌才可counter_trump；不能合法反主就pass_counter，不要用单张级牌反主。\n")
	sb.WriteString("- 起底阶段：轮到庄家队指定玩家时，通常take_bottom；只有明确要让队友起底时才pass_take_bottom。\n")
	sb.WriteString("- 扣底阶段：discard_bottom必须正好6张，优先扣低价值副牌；保留主牌、分牌、对子/拖拉机结构和控牌。\n")
	sb.WriteString("- 出牌阶段：必须play_cards，不能过牌，cards不能为空，并附带chat字段说一句不超过30字的台词。\n\n")
	sb.WriteString("## 跟牌硬约束\n")
	sb.WriteString("- 跟牌时必须先看领出的张数、牌型、花色和主副类别。\n")
	sb.WriteString("- 如果手里有可跟的同花色/同主副类别牌，必须先跟同花色/同主副类别，不能垫其他花色，也不能随便出主。\n")
	sb.WriteString("- 领出单张跟单张，领出对子跟对子，领出刻子跟刻子，领出拖拉机跟同长度拖拉机；无法保持牌型时仍要优先用同类牌补足张数。\n")
	sb.WriteString("- 只有没有领出花色/类别可跟时，才可以垫牌；只有没有副牌可跟且要争夺本轮时，才用主牌枪毙。\n")
	sb.WriteString("- 如果不确定，选择当前手牌中最保守、最可能合法的动作，不要编造不存在的牌ID。\n\n")
	sb.WriteString("## 配合与赢面\n")
	sb.WriteString("- 队友当前最大时优先送分或垫低价值牌，帮助队友收分/跑分。\n")
	sb.WriteString("- 对手当前最大时少送分；除非能确定抢回本轮，否则避免扔5、10、K。\n")
	sb.WriteString("- 闲家落后时更积极抢分和争领出；庄家队领先时稳守主牌和关键牌权。\n")
	sb.WriteString("- 领出时优先选择能减少手牌负担、保护分牌和保留控制力的牌型。\n")
	sb.WriteString(fmt.Sprintf("当前阶段：%s。\n", phase))
	return sb.String()
}

func (a *AIAgent) buildDashengjiUserMessage(phase string) string {
	status := a.buildDashengjiStatus()
	records := a.buildDashengjiPlayingRecords()
	candidates := a.buildDashengjiCandidateHints(phase)
	handCards := a.handCardsSnapshot()
	hand := formatDashengjiCardsWithIDs(handCards)
	if hand == "" {
		hand = "（空）"
	}
	return fmt.Sprintf("## 当前阶段：%s\n你的当前手牌（%d张）：%s\n只能从上面的手牌ID中选择；不要使用已经出过、扣掉或不存在的ID。\n%s\n%s\n%s\n请根据当前阶段调用一个动作工具。", phase, len(handCards), hand, status, records, candidates)
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
		sb.WriteString(fmt.Sprintf("本轮需要跟出张数：%d\n", len(led.Cards)))
		sb.WriteString(fmt.Sprintf("本轮领出牌型：%s\n", dashengjiPlayTypeName(led.Play.Type)))
		sb.WriteString(a.buildDashengjiTrickSummary(state))
	}
	if summary := buildDashengjiKeyCardSummary(state); summary != "" {
		sb.WriteString(summary)
	}
	return sb.String()
}

func (a *AIAgent) buildDashengjiTrickSummary(state dashengjiState) string {
	view := dashengjiCurrentTrickView(state)
	if !view.OK {
		return ""
	}
	owner := "对手"
	if dashengjiSameTeam(a.Seat, view.WinnerSeat) {
		owner = "队友"
	}
	advice := "避分"
	if owner == "队友" {
		advice = "送分/垫低牌"
	} else if dashengjiCanCurrentSeatWinAnyCandidate(state) {
		advice = "能抢则抢分"
	}
	return fmt.Sprintf("本轮局势摘要：当前最大座位%d（%s），桌面已有%d分，建议：%s。\n", view.WinnerSeat, owner, view.Points, advice)
}

func buildDashengjiKeyCardSummary(state dashengjiState) string {
	if state.TrumpSuit < 0 {
		return ""
	}
	var risks []string
	for suit := 0; suit < 4; suit++ {
		if suit == state.TrumpSuit {
			continue
		}
		aceFace := suit*13 + (14 - 3)
		if dashengjiVisibleFaceCount(state, aceFace) < 2 {
			risks = append(risks, fmt.Sprintf("%s对A未完全打出", dashengjiSuitName(suit)))
		}
	}
	if len(risks) == 0 {
		return ""
	}
	return "关键牌摘要：" + strings.Join(risks, "；") + "，领出同花色对K要谨慎。\n"
}

func dashengjiCanCurrentSeatWinAnyCandidate(state dashengjiState) bool {
	hand := stateCurrentHandIDs(state)
	if len(hand) == 0 || len(state.RoundPlays) == 0 {
		return false
	}
	candidates := dashengjiFollowCandidates(append([]int(nil), hand...), state)
	for _, candidate := range candidates {
		if dashengjiCandidateWins(candidate.Cards, state) {
			return true
		}
	}
	return false
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
		a.reportActionStatus(call.Name)
		var args PlayCardsArgs
		if err := json.Unmarshal(call.Args, &args); err != nil {
			a.fallbackDashengjiAction(a.detectPhase())
			return
		}
		if !a.executeDashengjiCandidateOrFallback(call.Name, args.Cards, args.Chat) {
			return
		}
		if args.Chat != "" {
			a.Executor.SendChat(a.UserID, args.Chat, "text")
		}
		a.Executor.ExecuteAction(a.UserID, call.Name, args.Cards)
	case "play_cards":
		a.reportActionStatus(call.Name)
		var args PlayCardsArgs
		if err := json.Unmarshal(call.Args, &args); err != nil || len(args.Cards) == 0 {
			a.fallbackDashengjiAction(a.detectPhase())
			return
		}
		if !a.executeDashengjiCandidateOrFallback(call.Name, args.Cards, args.Chat) {
			return
		}
		if args.Chat != "" {
			a.Executor.SendChat(a.UserID, args.Chat, "text")
		}
		a.Executor.ExecuteAction(a.UserID, "play", args.Cards)
	case "pass_trump", "pass_counter", "take_bottom", "pass_take_bottom":
		a.reportActionStatus(call.Name)
		a.Executor.ExecuteAction(a.UserID, call.Name, nil)
	default:
		a.fallbackDashengjiAction(a.detectPhase())
	}
}

func (a *AIAgent) executeDashengjiCandidateOrFallback(tool string, cards []int, chat string) bool {
	if a.stateSnapshot() == "" {
		return true
	}
	if a.isLegalDashengjiAction(tool, cards) {
		return true
	}
	phase := a.detectPhase()
	fallback, ok := a.dashengjiFallbackCandidate(phase)
	if !ok {
		a.fallbackDashengjiAction(phase)
		return false
	}
	logDashengjiCandidateRejected(a.UserID, tool, cards, fallback)
	a.executeDashengjiCandidate(fallback, "")
	return false
}

func (a *AIAgent) executeDashengjiCandidate(candidate dashengjiCandidate, chat string) {
	if chat != "" {
		a.Executor.SendChat(a.UserID, chat, "text")
	}
	switch candidate.Tool {
	case "play_cards":
		a.Executor.ExecuteAction(a.UserID, "play", candidate.Cards)
	case "set_trump", "counter_trump", "discard_bottom":
		a.Executor.ExecuteAction(a.UserID, candidate.Tool, candidate.Cards)
	case "pass_trump", "pass_counter", "take_bottom", "pass_take_bottom":
		a.Executor.ExecuteAction(a.UserID, candidate.Tool, nil)
	default:
		a.fallbackDashengjiAction(a.detectPhase())
	}
}

func (a *AIAgent) fallbackDashengjiAction(phase string) {
	if a.Executor == nil {
		return
	}
	state, _ := a.parseDashengjiState()
	hand := a.handCardsSnapshot()
	switch phase {
	case "set_trump":
		if cards := findDashengjiTrumpCards(hand, state.LevelRank, false); len(cards) > 0 {
			a.Executor.ExecuteAction(a.UserID, "set_trump", cards)
			return
		}
		a.Executor.ExecuteAction(a.UserID, "pass_trump", nil)
	case "counter_trump":
		if cards := findDashengjiTrumpCards(hand, state.LevelRank, true); len(cards) > 0 {
			a.Executor.ExecuteAction(a.UserID, "counter_trump", cards)
			return
		}
		a.Executor.ExecuteAction(a.UserID, "pass_counter", nil)
	case "take_bottom":
		a.Executor.ExecuteAction(a.UserID, "take_bottom", nil)
	case "discard_bottom":
		a.Executor.ExecuteAction(a.UserID, "discard_bottom", chooseDashengjiDiscard(hand, state.TrumpSuit, state.LevelRank))
	default:
		a.Executor.ExecuteAction(a.UserID, "play", chooseDashengjiPlay(hand, state))
	}
}

func (a *AIAgent) parseDashengjiState() (dashengjiState, bool) {
	var state dashengjiState
	state.TrumpSuit = -1
	state.LevelRank = 3
	stateJSON := a.stateSnapshot()
	if stateJSON == "" {
		return state, false
	}
	if err := json.Unmarshal([]byte(stateJSON), &state); err != nil {
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

func dashengjiPlayTypeName(playType int) string {
	switch playType {
	case 1:
		return "单张"
	case 2:
		return "对子"
	case 3:
		return "刻子"
	case 4:
		return "拖拉机"
	default:
		return "未知"
	}
}
