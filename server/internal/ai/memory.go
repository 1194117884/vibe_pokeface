package ai

import (
	"encoding/json"
	"fmt"
	"strings"
)

type memoryCard struct {
	ID int `json:"id"`
}

type memoryPlay struct {
	Type     int `json:"type"`
	MainRank int `json:"main_rank"`
	Length   int `json:"length"`
}

type memoryPlayRecord struct {
	Seat  int          `json:"seat"`
	Play  memoryPlay   `json:"play"`
	Cards []memoryCard `json:"cards"`
}

type memoryBidRecord struct {
	Seat   int  `json:"seat"`
	Called bool `json:"called"`
}

type memoryPlayer struct {
	Seat       int          `json:"seat"`
	IsLandlord bool         `json:"is_landlord"`
	Hand       []memoryCard `json:"hand"`
}

type memoryState struct {
	Phase        int                `json:"phase"`
	RoundNum     int                `json:"round_num"`
	CurrentSeat  int                `json:"current_seat"`
	LandlordSeat int                `json:"landlord_seat"`
	Multiplier   int                `json:"multiplier"`
	DealerSeats  [2]int             `json:"dealer_seats"`
	TeamLevels   [2]int             `json:"team_levels"`
	CurrentLevel int                `json:"current_level"`
	LevelRank    int                `json:"level_rank"`
	TrumpSuit    int                `json:"trump_suit"`
	RoundPoints  int                `json:"round_points"`
	Players      []memoryPlayer     `json:"players"`
	PlayHistory  []memoryPlayRecord `json:"play_history"`
	RoundPlays   []memoryPlayRecord `json:"round_plays"`
	BidHistory   []memoryBidRecord  `json:"bid_history"`
	WinnerSeat   *int               `json:"winner_seat,omitempty"`
}

func (a *AIAgent) buildDecisionMessages(phase string) []ChatMessage {
	a.appendCurrentContextToMemory(phase)

	systemPrompt := a.buildSystemPrompt(phase) + "\n\n" +
		"## 连贯上下文\n" +
		"你会看到从本局开始持续追加的牌局日志和你自己的历史决策。请延续之前形成的全局策略。\n" +
		"当前日志已经包含你的手牌、身份、剩余牌数、上一轮/当前轮新增出牌等信息；除非信息明显缺失，不要重复调用check_my_hand、check_game_status或check_playing_records。\n" +
		"如果当前阶段允许动作且上下文足够，优先直接调用动作工具。"

	messages := []ChatMessage{{Role: "system", Content: systemPrompt}}
	messages = append(messages, a.memoryMessages...)
	return messages
}

func (a *AIAgent) appendCurrentContextToMemory(phase string) {
	state, ok := a.parseMemoryState()
	if !ok {
		content := "【当前行动上下文】\n" + a.buildUserMessage(phase)
		a.appendMemoryMessage(ChatMessage{Role: "user", Content: content})
		return
	}

	if state.RoundNum == 0 {
		state.RoundNum = 1
	}

	gameType := a.gameTypeSnapshot()
	shouldReset := !a.memoryStarted ||
		a.memoryGameType != gameType ||
		state.RoundNum < a.memoryRoundNum ||
		len(state.PlayHistory) < a.memoryPlayLen ||
		len(state.BidHistory) < a.memoryBidLen ||
		(state.RoundNum == a.memoryRoundNum && state.Phase < a.memoryPhase && len(state.PlayHistory) == 0)

	if shouldReset {
		a.resetMemory()
		a.memoryStarted = true
		a.memoryGameType = gameType
	}

	hand := a.handCardsSnapshot()
	key := fmt.Sprintf("%s:%d:%d:%d:%d:%d:%d", gameType, state.RoundNum, state.Phase, state.CurrentSeat, len(state.PlayHistory), len(state.BidHistory), len(hand))
	if key == a.memoryKey {
		return
	}

	content := a.buildMemoryContextMessage(phase, state, shouldReset)
	a.appendMemoryMessage(ChatMessage{Role: "user", Content: content})

	a.memoryRoundNum = state.RoundNum
	a.memoryPhase = state.Phase
	a.memoryPlayLen = len(state.PlayHistory)
	a.memoryBidLen = len(state.BidHistory)
	a.memoryKey = key
}

func (a *AIAgent) resetMemory() {
	a.memoryMessages = nil
	a.memoryStarted = false
	a.memoryRoundNum = 0
	a.memoryPhase = 0
	a.memoryPlayLen = 0
	a.memoryBidLen = 0
	a.memoryKey = ""
}

func (a *AIAgent) parseMemoryState() (memoryState, bool) {
	var state memoryState
	stateJSON := a.stateSnapshot()
	if stateJSON == "" {
		return state, false
	}
	if err := json.Unmarshal([]byte(stateJSON), &state); err != nil {
		return state, false
	}
	return state, true
}

func (a *AIAgent) buildMemoryContextMessage(phase string, state memoryState, isNewGame bool) string {
	var sb strings.Builder
	if isNewGame {
		sb.WriteString("【本局开始】\n")
	} else {
		sb.WriteString("【牌局更新】\n")
	}

	sb.WriteString(fmt.Sprintf("当前阶段：%s；轮到座位%d行动；你的座位：%d。\n", phase, state.CurrentSeat, a.Seat))
	hand := a.handCardsSnapshot()
	sb.WriteString(fmt.Sprintf("你的当前手牌（%d张）：%s\n", len(hand), a.formatCardsWithIDs(hand)))
	a.writeMemoryIdentity(&sb, state)
	a.writeMemoryPlayerCounts(&sb, state)
	a.writeMemoryNewBids(&sb, state)
	a.writeMemoryNewPlays(&sb, state)

	if state.WinnerSeat != nil {
		sb.WriteString(fmt.Sprintf("本局胜者座位：%d。\n", *state.WinnerSeat))
	}

	sb.WriteString("\n当前完整可行动快照：\n")
	sb.WriteString(a.buildUserMessage(phase))
	return sb.String()
}

func (a *AIAgent) writeMemoryIdentity(sb *strings.Builder, state memoryState) {
	if a.gameTypeSnapshot() == "dashengji" {
		level := state.LevelRank
		if level == 0 {
			level = state.CurrentLevel
		}
		sb.WriteString(fmt.Sprintf("庄家队座位：%d,%d；队伍等级：%d/%d；当前级牌：%d；主花色：%s；闲家累计抢分：%d。\n",
			state.DealerSeats[0], state.DealerSeats[1], state.TeamLevels[0], state.TeamLevels[1], level, dashengjiSuitName(state.TrumpSuit), state.RoundPoints))
		return
	}

	if state.LandlordSeat >= 0 {
		role := "农民"
		if state.LandlordSeat == a.Seat {
			role = "地主"
		}
		sb.WriteString(fmt.Sprintf("你的身份：%s；地主座位：%d；当前倍率：%d。\n", role, state.LandlordSeat, state.Multiplier))
	}
}

func (a *AIAgent) writeMemoryPlayerCounts(sb *strings.Builder, state memoryState) {
	if len(state.Players) == 0 {
		return
	}
	sb.WriteString("各座位当前剩余手牌：")
	for i, p := range state.Players {
		if i > 0 {
			sb.WriteString("；")
		}
		role := ""
		if a.gameTypeSnapshot() == "dashengji" {
			if p.Seat == state.DealerSeats[0] || p.Seat == state.DealerSeats[1] {
				role = "庄"
			} else {
				role = "闲"
			}
		} else if p.IsLandlord {
			role = "地主"
		} else {
			role = "农民"
		}
		sb.WriteString(fmt.Sprintf("座位%d(%s)%d张", p.Seat, role, len(p.Hand)))
	}
	sb.WriteString("。\n")
}

func (a *AIAgent) writeMemoryNewBids(sb *strings.Builder, state memoryState) {
	if len(state.BidHistory) <= a.memoryBidLen {
		return
	}
	sb.WriteString("自你上次行动后的叫/抢记录：\n")
	for i := a.memoryBidLen; i < len(state.BidHistory); i++ {
		action := "不叫/不抢"
		if state.BidHistory[i].Called {
			action = "叫/抢"
		}
		sb.WriteString(fmt.Sprintf("  %d. 座位%d：%s\n", i+1, state.BidHistory[i].Seat, action))
	}
}

func (a *AIAgent) writeMemoryNewPlays(sb *strings.Builder, state memoryState) {
	if len(state.PlayHistory) <= a.memoryPlayLen {
		if len(state.RoundPlays) > 0 {
			lead := state.RoundPlays[0]
			sb.WriteString(fmt.Sprintf("当前轮领出：座位%d %s。\n", lead.Seat, a.formatCardsWithIDs(memoryCardIDs(lead.Cards))))
		}
		return
	}

	sb.WriteString("自你上次行动后的新增出牌：\n")
	for i := a.memoryPlayLen; i < len(state.PlayHistory); i++ {
		record := state.PlayHistory[i]
		ids := memoryCardIDs(record.Cards)
		if len(ids) == 0 || record.Play.Type == 0 {
			sb.WriteString(fmt.Sprintf("  %d. 座位%d：过牌\n", i+1, record.Seat))
			continue
		}
		sb.WriteString(fmt.Sprintf("  %d. 座位%d：%s\n", i+1, record.Seat, a.formatCardsWithIDs(ids)))
	}
}

func memoryCardIDs(cards []memoryCard) []int {
	ids := make([]int, len(cards))
	for i, c := range cards {
		ids[i] = c.ID
	}
	return ids
}

func (a *AIAgent) rememberActionDecision(toolName, argsJSON string) {
	if toolName == "" {
		return
	}
	a.appendMemoryMessage(ChatMessage{
		Role:    "assistant",
		Content: fmt.Sprintf("【我的决策】调用%s，参数：%s。后续行动要记住这个意图，并围绕整局策略调整。", toolName, argsJSON),
	})
}

func (a *AIAgent) appendMemoryMessage(msg ChatMessage) {
	a.memoryMessages = append(a.memoryMessages, msg)
}
