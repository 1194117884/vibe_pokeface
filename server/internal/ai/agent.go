package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"strings"
	"time"

	"github.com/yongkl/vibe-pokeface/internal/model"
)

// ToolResult is the result of executing a tool
type ToolResult struct {
	Tool   string `json:"tool"`
	Result string `json:"result"`
}

// ActionExecutor defines how the agent executes game actions
type ActionExecutor interface {
	ExecuteAction(userID string, action string, cards []int)
	SendChat(senderID string, content string, msgType string)
}

// AIAgent runs as a goroutine per bot, making LLM-driven decisions
type AIAgent struct {
	UserID    string
	Seat      int
	GameType  string
	HandCards []int
	Character *model.AICharacter
	Provider  LLMProvider
	Executor  ActionExecutor
	stateJSON string

	triggerChan chan struct{}
	stopChan    chan struct{}
	stopped     bool

	// Override for testing
	MakeDecisionFunc func(agent *AIAgent, phase string, handCards []int, stateJSON string) string

	aiStore *model.AIStore // for LLM call logging
	roomID  string         // room context for audit logging

	lastToolExecID int64 // most recent action-tool execution ID for game_actions link

	lastError  *GameError // last action error for retry context
	retryCount int        // number of retries after errors

	memoryMessages []ChatMessage // persistent per-game strategy/context log
	memoryStarted  bool
	memoryGameType string
	memoryRoundNum int
	memoryPhase    int
	memoryPlayLen  int
	memoryBidLen   int
	memoryKey      string
}

// GameError mirrors game.GameError so the ai package can accept errors
// without importing the game package (avoiding a potential cycle).
type GameError struct {
	Code   string `json:"code"`
	Phase  string `json:"phase,omitempty"`
	Action string `json:"action,omitempty"`
}

func (e *GameError) Error() string {
	return e.Code + ":" + e.Phase + ":" + e.Action
}

// ReportError stores an action error and re-triggers the decision loop so
// the LLM can retry with context. After 2 retries the agent gives up and
// falls back to a rule-based default.
func (a *AIAgent) ReportError(err *GameError) {
	if a.retryCount >= 2 {
		log.Printf("[AI:%s] retry limit exceeded, falling back to rule-based action", a.UserID)
		go a.ruleBasedAction(a.detectPhase())
		return
	}
	a.lastError = err
	a.retryCount++
	a.Trigger()
}

// ResetError clears the error state (called on successful action or new turn).
func (a *AIAgent) ResetError() {
	a.lastError = nil
	a.retryCount = 0
}

// NewAIAgent creates a new AI agent
func NewAIAgent(userID string, seat int, character *model.AICharacter, provider LLMProvider, executor ActionExecutor) *AIAgent {
	return &AIAgent{
		UserID:      userID,
		Seat:        seat,
		GameType:    "doudizhu",
		Character:   character,
		Provider:    provider,
		Executor:    executor,
		triggerChan: make(chan struct{}, 1),
		stopChan:    make(chan struct{}),
	}
}

// SetGameType sets the game ruleset used for prompts, tools, and fallback play.
func (a *AIAgent) SetGameType(gameType string) {
	if gameType == "" {
		gameType = "doudizhu"
	}
	a.GameType = gameType
}

// SetAIStore sets the AIStore for LLM call logging.
func (a *AIAgent) SetAIStore(store *model.AIStore) {
	a.aiStore = store
}

// SetRoomID sets the room ID for audit logging.
func (a *AIAgent) SetRoomID(roomID string) {
	a.roomID = roomID
}

// LastToolExecID returns the most recent action-tool execution ID, or 0 if none.
func (a *AIAgent) LastToolExecID() int64 {
	return a.lastToolExecID
}

// Start launches the agent's goroutine
func (a *AIAgent) Start() {
	go a.run()
}

// Stop signals the agent to shut down
func (a *AIAgent) Stop() {
	if !a.stopped {
		a.stopped = true
		close(a.stopChan)
	}
}

// Trigger tells the agent it's its turn to act.
// If this is a fresh turn (no pending error), reset the retry counter.
func (a *AIAgent) Trigger() {
	if a.lastError == nil {
		a.retryCount = 0
	}
	select {
	case a.triggerChan <- struct{}{}:
	default:
	}
}

// UpdateHand synchronizes the agent's known hand from a state update
func (a *AIAgent) UpdateHand(cards []int) {
	a.HandCards = make([]int, len(cards))
	copy(a.HandCards, cards)
}

// UpdateState stores the latest game state JSON for tool use
func (a *AIAgent) UpdateState(stateJSON string) {
	a.stateJSON = stateJSON
}

// run is the main agent loop
func (a *AIAgent) run() {
	for {
		select {
		case <-a.triggerChan:
			a.makeDecision()
		case <-a.stopChan:
			return
		}
	}
}

// makeDecision determines what action the AI should take.
// Tries LLM first (single call with all state injected), falls back to rule-based.
func (a *AIAgent) makeDecision() {
	select {
	case <-a.stopChan:
		return
	case <-time.After(800 * time.Millisecond):
	}

	if a.MakeDecisionFunc != nil {
		result := a.MakeDecisionFunc(a, "playing", a.HandCards, a.stateJSON)
		a.executeToolCall(result)
		return
	}

	phase := a.detectPhase()
	log.Printf("[AI:%s seat=%d] phase=%s hand=%d cards", a.UserID, a.Seat, phase, len(a.HandCards))

	if a.Provider != nil {
		log.Printf("[AI:%s] calling LLM with tools (phase=%s)", a.UserID, phase)
		a.makeDecisionWithTools()
		return
	}

	a.ruleBasedAction(phase)
}

// makeDecisionWithTools implements a multi-turn tool-calling loop.
// The agent can call info tools (check_my_hand, check_game_status)
// over up to 10 turns before committing to an action tool.
func (a *AIAgent) makeDecisionWithTools() {
	phase := a.detectPhase()
	messages := a.buildDecisionMessages(phase)

	// Inject error context from previous failed action so the LLM can correct itself.
	if a.lastError != nil {
		retryMsg := fmt.Sprintf("你上一次操作失败了，错误码：%s", a.lastError.Code)
		if a.lastError.Phase != "" {
			retryMsg += fmt.Sprintf("，当前阶段是：%s", a.lastError.Phase)
		}
		if a.lastError.Action != "" {
			retryMsg += fmt.Sprintf("，你尝试的动作是：%s", a.lastError.Action)
		}
		retryMsg += "。请根据当前阶段选择正确的操作工具。"
		messages = append(messages, ChatMessage{Role: "user", Content: retryMsg})
		a.lastError = nil // consumed
	}

	tools := GetToolSchemasForGame(a.GameType, phase)

	for turn := 0; turn < 10; turn++ {
		// Capture request payload BEFORE the LLM call
		requestJSON := captureRequestJSON(messages, tools)

		ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
		result, err := a.Provider.CompleteWithTools(ctx, messages, tools)
		cancel()

		// Capture response payload AFTER the LLM call
		responseJSON := captureResponseJSON(result, err)

		// Log LLM call to database with full payloads
		var callLogID int64
		if a.aiStore != nil {
			llmCall := &model.LLMCallLog{
				Provider:     a.Provider.ProviderName(),
				Model:        a.Provider.ModelName(),
				CallType:     "play_decision",
				Success:      err == nil,
				RoomID:       a.roomID,
				UserID:       a.UserID,
				Seat:         a.Seat,
				Phase:        phase,
				TurnNumber:   turn,
				RequestJSON:  jsonPtr(requestJSON),
				ResponseJSON: jsonPtr(responseJSON),
			}
			if err == nil {
				llmCall.PromptTokens = result.PromptTokens
				llmCall.CompletionTokens = result.CompletionTokens
				llmCall.DurationMs = int(result.DurationMs)
			} else {
				errMsg := err.Error()
				llmCall.ErrorMessage = &errMsg
			}
			var logErr error
			callLogID, logErr = a.aiStore.LogLLMCall(context.Background(), llmCall)
			if logErr != nil {
				log.Printf("[AI:%s] failed to log LLM call: %v", a.UserID, logErr)
			}
		}

		if err != nil {
			log.Printf("[AI:%s] tool loop error at turn %d: %v, falling back", a.UserID, turn, err)
			break
		}

		if len(result.ToolCalls) == 0 {
			if result.Content != "" {
				if call, parseErr := ExtractToolCall(result.Content); parseErr == nil && a.isActionTool(call.Name) {
					log.Printf("[AI:%s] extracted tool call from text content: %s", a.UserID, call.Name)
					a.logToolExecution(callLogID, call.Name, "action", string(call.Args), "")
					a.rememberActionDecision(call.Name, string(call.Args))
					a.executeToolCall(result.Content)
					return
				}
			}
			log.Printf("[AI:%s] no tool calls at turn %d, falling back", a.UserID, turn)
			break
		}

		assistantMsg := ChatMessage{
			Role:             "assistant",
			Content:          result.Content,
			ReasoningContent: result.ReasoningContent,
			ToolCalls:        result.ToolCalls,
		}
		messages = append(messages, assistantMsg)

		hasAction := false
		for _, tc := range result.ToolCalls {
			if a.isActionTool(tc.Function.Name) {
				hasAction = true
			}

			if a.isInfoTool(tc.Function.Name) {
				infoResult := a.executeInfoTool(tc.Function.Name)
				messages = append(messages, ChatMessage{
					Role:       "tool",
					ToolCallID: tc.ID,
					Name:       tc.Function.Name,
					Content:    infoResult,
				})
				a.logToolExecution(callLogID, tc.Function.Name, "info", tc.Function.Arguments, infoResult)
			} else {
				a.logToolExecution(callLogID, tc.Function.Name, "action", tc.Function.Arguments, "")
				a.rememberActionDecision(tc.Function.Name, tc.Function.Arguments)
				raw := toolCallToJSON(tc)
				a.executeToolCall(raw)
			}
		}

		if hasAction {
			return
		}
	}

	log.Printf("[AI:%s] multi-turn loop exhausted or failed, using rule-based fallback", a.UserID)
	a.ruleBasedAction(phase)
}

// captureRequestJSON serializes the full LLM request (messages + tools) for audit logging.
func captureRequestJSON(messages []ChatMessage, tools []ToolSchema) string {
	data, err := json.Marshal(map[string]interface{}{
		"messages": messages,
		"tools":    tools,
	})
	if err != nil {
		return ""
	}
	return string(data)
}

// captureResponseJSON serializes the LLM response for audit logging.
func captureResponseJSON(result *LLMResultWithTools, err error) string {
	if err != nil || result == nil {
		return ""
	}
	data, err := json.Marshal(result)
	if err != nil {
		return ""
	}
	return string(data)
}

func jsonPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func (a *AIAgent) logToolExecution(callLogID int64, toolName, toolType, argsJSON, resultJSON string) {
	if a.aiStore == nil || callLogID == 0 {
		return
	}
	exec := &model.AiToolExecution{
		CallLogID:  callLogID,
		ToolName:   toolName,
		ToolType:   toolType,
		ArgsJSON:   jsonPtr(argsJSON),
		ResultJSON: jsonPtr(resultJSON),
	}
	id, err := a.aiStore.LogToolExecution(context.Background(), exec)
	if err != nil {
		log.Printf("[AI:%s] failed to log tool execution: %v", a.UserID, err)
		return
	}
	if toolType == "action" {
		a.lastToolExecID = id
	}
}

func (a *AIAgent) isInfoTool(name string) bool {
	return name == "check_my_hand" || name == "check_game_status" || name == "check_playing_records"
}

func (a *AIAgent) executeInfoTool(name string) string {
	switch name {
	case "check_my_hand":
		return fmt.Sprintf("你的手牌（%d张）：%s", len(a.HandCards), a.formatCardsWithIDs(a.HandCards))
	case "check_game_status":
		if a.GameType == "dashengji" {
			return a.buildDashengjiStatus()
		}
		if a.stateJSON == "" {
			return "游戏状态不可用"
		}
		var state struct {
			Phase        int `json:"phase"`
			CurrentSeat  int `json:"current_seat"`
			LandlordSeat int `json:"landlord_seat"`
			Multiplier   int `json:"multiplier"`
			LastPlay     *struct {
				Seat  int `json:"seat"`
				Cards []struct {
					ID int `json:"id"`
				} `json:"cards"`
				Play *struct {
					Type     int `json:"type"`
					MainRank int `json:"main_rank"`
				} `json:"play"`
			} `json:"last_play"`
			Players []struct {
				Seat       int  `json:"seat"`
				IsLandlord bool `json:"is_landlord"`
				Hand       []struct {
					ID int `json:"id"`
				} `json:"hand"`
			} `json:"players"`
		}
		if err := json.Unmarshal([]byte(a.stateJSON), &state); err != nil {
			return "无法解析游戏状态"
		}
		var sb strings.Builder
		sb.WriteString(fmt.Sprintf("当前轮到座位：%d（你的座位：%d）\n", state.CurrentSeat, a.Seat))
		sb.WriteString(fmt.Sprintf("地主座位：%d\n", state.LandlordSeat))
		sb.WriteString(fmt.Sprintf("当前倍率：%d\n", state.Multiplier))
		sb.WriteString("各玩家剩余手牌：\n")
		for _, p := range state.Players {
			role := "农民"
			if p.IsLandlord {
				role = "地主"
			}
			sb.WriteString(fmt.Sprintf("  座位%d（%s）：%d张\n", p.Seat, role, len(p.Hand)))
		}
		if state.LastPlay != nil && len(state.LastPlay.Cards) > 0 {
			ids := make([]int, len(state.LastPlay.Cards))
			for i, c := range state.LastPlay.Cards {
				ids[i] = c.ID
			}
			sb.WriteString(fmt.Sprintf("上家(座位%d)出了：%s\n", state.LastPlay.Seat, a.formatCardsWithIDs(ids)))
		} else {
			sb.WriteString("轮到你了，可以自由出牌\n")
		}
		return sb.String()
	case "check_playing_records":
		return a.buildPlayingRecords()
	}
	return "unknown tool"
}

func (a *AIAgent) buildPlayingRecords() string {
	if a.stateJSON == "" {
		return "游戏状态不可用"
	}
	if a.GameType == "dashengji" {
		return a.buildDashengjiPlayingRecords()
	}
	var state struct {
		PlayHistory []struct {
			Seat int `json:"seat"`
			Play *struct {
				Type     int `json:"type"`
				MainRank int `json:"main_rank"`
				Length   int `json:"length"`
			} `json:"play"`
			Cards []struct {
				ID int `json:"id"`
			} `json:"cards"`
		} `json:"play_history"`
		LastPlay *struct {
			Seat int `json:"seat"`
			Play *struct {
				Type     int `json:"type"`
				MainRank int `json:"main_rank"`
				Length   int `json:"length"`
			} `json:"play"`
			Cards []struct {
				ID int `json:"id"`
			} `json:"cards"`
		} `json:"last_play"`
	}
	if err := json.Unmarshal([]byte(a.stateJSON), &state); err != nil {
		return "无法解析游戏状态"
	}

	var sb strings.Builder
	sb.WriteString("本轮出牌记录：\n")
	playTypeNames := []string{"无效", "单张", "对子", "三张", "三带一", "三带二", "顺子", "连对", "飞机", "飞机带翅膀", "四带二", "炸弹", "火箭"}

	seq := 0
	for _, p := range state.PlayHistory {
		seq++
		if p.Play != nil && p.Play.Type > 0 {
			// Actual card play
			cardIDs := make([]int, len(p.Cards))
			for i, c := range p.Cards {
				cardIDs[i] = c.ID
			}
			ptName := "未知"
			if p.Play.Type < len(playTypeNames) {
				ptName = playTypeNames[p.Play.Type]
			}
			sb.WriteString(fmt.Sprintf("  %d. 座位%d 出了%s %s（主牌等级%d）\n",
				seq, p.Seat, ptName, a.formatCardsWithIDs(cardIDs), p.Play.MainRank))
		} else {
			// Pass
			sb.WriteString(fmt.Sprintf("  %d. 座位%d 过牌\n", seq, p.Seat))
		}
	}

	// Also show current last play for quick reference
	if state.LastPlay != nil && len(state.LastPlay.Cards) > 0 {
		cardIDs := make([]int, len(state.LastPlay.Cards))
		for i, c := range state.LastPlay.Cards {
			cardIDs[i] = c.ID
		}
		sb.WriteString(fmt.Sprintf("\n当前上家(座位%d)出了：%s", state.LastPlay.Seat, a.formatCardsWithIDs(cardIDs)))
		if state.LastPlay.Play != nil {
			ptName := "未知"
			if state.LastPlay.Play.Type < len(playTypeNames) {
				ptName = playTypeNames[state.LastPlay.Play.Type]
			}
			sb.WriteString(fmt.Sprintf("（%s，主牌等级%d）", ptName, state.LastPlay.Play.MainRank))
		}
		sb.WriteString("\n")
	} else {
		sb.WriteString("\n当前无上家出牌，你可以自由出牌\n")
	}

	return sb.String()
}

func toolCallToJSON(tc AssistantToolCall) string {
	data, _ := json.Marshal(map[string]interface{}{
		"tool": tc.Function.Name,
		"args": json.RawMessage(tc.Function.Arguments),
	})
	return string(data)
}

func (a *AIAgent) buildSystemPrompt(phase string) string {
	if a.GameType == "dashengji" {
		return a.buildDashengjiSystemPrompt(phase)
	}

	var sb strings.Builder
	name := "AI玩家"
	personality := "冷静分析，稳健出牌"

	if a.Character != nil {
		if a.Character.Name != "" {
			name = a.Character.Name
		}
		if a.Character.Personality != nil && *a.Character.Personality != "" {
			personality = *a.Character.Personality
		}
	}

	sb.WriteString(fmt.Sprintf("你在进行一场斗地主比赛，你是「%s」。\n", name))
	playStyle := "balanced"
	if a.Character != nil {
		playStyle = a.Character.PlayStyle
	}
	sb.WriteString(fmt.Sprintf("性格：%s。出牌风格：%s。\n\n", personality, playStyle))
	sb.WriteString("斗地主分为五阶段：叫地主 -> 抢地主 -> 明牌 -> 加倍 -> 出牌\n")

	sb.WriteString("##叫地主:\n")
	sb.WriteString("游戏开始，三人轮流叫地主。叫分最高者成为地主，获得3张底牌。\n")
	sb.WriteString("如果你叫了地主，后续可被他人抢地主；如果你不叫，轮到下家决定。\n\n")
	sb.WriteString("【重要】你必须调用bid_landlord（叫地主）或pass_bid（不叫）。\n")
	sb.WriteString("根据手牌强度决定：有炸弹、多张2、有大王时应该叫地主，手牌较弱时选择不叫。\n")

	sb.WriteString("##抢地主:\n")
	sb.WriteString("有人已叫地主。现在其他玩家可以抢地主，每抢一次倍数翻倍，无人抢则由叫分者成为地主。\n\n")
	sb.WriteString("【重要】你必须调用bid_landlord（抢地主）或pass_bid（不抢）。\n")
	sb.WriteString("抢地主会使倍数翻倍，手牌很强时才抢，否则不抢。\n")

	sb.WriteString("##明牌:\n")
	sb.WriteString("地主决定是否亮出底牌。明牌后所有玩家可见底牌，倍数翻倍。地主获得底牌后手牌增至20张。\n\n")
	sb.WriteString("【重要】你必须调用reveal_cards（明牌）或pass_reveal（不明牌）。\n")
	sb.WriteString("手牌非常好（有炸弹、火箭）时可以考虑明牌，否则选择不显示。\n")

	sb.WriteString("##加倍:\n")
	sb.WriteString("各玩家轮流决定是否加倍。加倍后该玩家的输赢分数翻倍（赢多输多），农民和地主分别独立翻倍。\n\n")
	sb.WriteString("【重要】你必须调用choose_double（加倍）或choose_no_double（不加倍）。\n")
	sb.WriteString("手牌很好或你是地主时可以考虑加倍，否则选择不加倍。\n")

	sb.WriteString("##出牌:\n")
	sb.WriteString("地主先出牌，之后按逆时针轮流。轮到时必须出比上家更大的同牌型，或无牌可出时选择过牌。\n")
	sb.WriteString("最先出完手牌者获胜。地主赢则农民输，任一农民先出完则地主输。\n\n")
	sb.WriteString("规则：单张、对子、三张、三带一、三带二、顺子(5张+)、连对(3对+)、飞机、炸弹、火箭。\n")
	sb.WriteString("必须出比上家更大的牌型，或选择过牌。牌型相同才能比较大小。\n")
	sb.WriteString("地主目标：尽快出完手牌。农民目标：配合队友阻止地主。\n")
	sb.WriteString("【重要】你必须调用play_cards来出牌或过牌。\n")
	sb.WriteString("- 出牌：play_cards cards=[id1,id2,...] chat=\"要说的话\"\n")
	sb.WriteString("- 过牌：play_cards cards=[]\n")
	sb.WriteString("chat参数可选，用于在出牌时附带简短聊天（不超过30字）。\n")
	sb.WriteString("出牌时必须使用卡牌的整数ID（如27），不要使用文字描述（如♣4）。\n")

	sb.WriteString("##牌:\n")
	sb.WriteString("牌的ID从0到53，分别对应：\n")
	sb.WriteString("- 0-51：普通牌，按花色和点数排序（0=♣3, 1=♦3, 3=♥3, 4=♠3, ..., 48=♣2, 49=♦2, 50=♥2, 51=♠2）\n")
	sb.WriteString("- 52：小王\n")
	sb.WriteString("- 53：大王\n")

	return sb.String()
}

func (a *AIAgent) detectPhase() string {
	if a.GameType == "dashengji" {
		return a.detectDashengjiPhase()
	}
	if a.stateJSON == "" {
		return "calling"
	}
	var state struct {
		Phase int `json:"phase"`
	}
	if err := json.Unmarshal([]byte(a.stateJSON), &state); err != nil {
		return "calling"
	}
	switch state.Phase {
	case 0:
		return "calling"
	case 1:
		return "snatching"
	case 2:
		return "revealing"
	case 3:
		return "doubling"
	default:
		return "playing"
	}
}

func (a *AIAgent) buildUserMessage(phase string) string {
	if a.GameType == "dashengji" {
		return a.buildDashengjiUserMessage(phase)
	}

	var sb strings.Builder

	switch phase {
	case "calling":
		sb.WriteString("## 当前阶段：叫地主\n")
	case "snatching":
		sb.WriteString("## 当前阶段：抢地主\n")
	case "revealing":
		sb.WriteString("## 当前阶段：明牌\n")
	case "doubling":
		sb.WriteString("## 当前阶段：加倍\n")
	default:
		sb.WriteString("## 当前阶段：出牌\n")
	}

	// Show game state context
	if a.stateJSON != "" {
		var state struct {
			Phase         int `json:"phase"`
			CurrentSeat   int `json:"current_seat"`
			LandlordSeat  int `json:"landlord_seat"`
			Multiplier    int `json:"multiplier"`
			LandlordCards []struct {
				ID int `json:"id"`
			} `json:"landlord_cards"`
			LastPlay *struct {
				Seat  int `json:"seat"`
				Cards []struct {
					ID int `json:"id"`
				} `json:"cards"`
				Play *struct {
					Type     int `json:"type"`
					MainRank int `json:"main_rank"`
					Length   int `json:"length"`
				} `json:"play"`
			} `json:"last_play"`
			Players []struct {
				Seat       int  `json:"seat"`
				IsLandlord bool `json:"is_landlord"`
				Hand       []struct {
					ID int `json:"id"`
				} `json:"hand"`
			} `json:"players"`
			BidHistory []struct {
				Seat   int  `json:"seat"`
				Called bool `json:"called"`
			} `json:"bid_history"`
		}
		if json.Unmarshal([]byte(a.stateJSON), &state) == nil {
			// Extract card IDs from struct format for display
			landlordCardIDs := make([]int, len(state.LandlordCards))
			for i, c := range state.LandlordCards {
				landlordCardIDs[i] = c.ID
			}

			sb.WriteString(fmt.Sprintf("你的座位：%d\n", a.Seat))
			sb.WriteString(fmt.Sprintf("当前轮到座位：%d\n", state.CurrentSeat))
			sb.WriteString(fmt.Sprintf("当前倍率：%d\n", state.Multiplier))

			// Landlord info and 底牌
			if state.LandlordSeat >= 0 {
				sb.WriteString(fmt.Sprintf("地主座位：%d\n", state.LandlordSeat))
				if state.LandlordSeat == a.Seat {
					sb.WriteString("你是地主！\n")
					if len(landlordCardIDs) > 0 {
						sb.WriteString(fmt.Sprintf("地主牌（底牌）：%s\n", a.formatCardsWithIDs(landlordCardIDs)))
					}
				} else {
					sb.WriteString("你是农民，队友也是农民。\n")
					if len(landlordCardIDs) > 0 {
						sb.WriteString(fmt.Sprintf("地主牌（底牌）：%s\n", a.formatCardsWithIDs(landlordCardIDs)))
					}
				}
			}

			// Bid history
			if len(state.BidHistory) > 0 {
				sb.WriteString("\n叫地主/抢地主记录：\n")
				for _, bid := range state.BidHistory {
					action := "不叫/不抢"
					if bid.Called {
						action = "叫地主/抢地主"
					}
					sb.WriteString(fmt.Sprintf("  座位%d：%s\n", bid.Seat, action))
				}
			}

			// Opponent card counts
			sb.WriteString("\n各玩家剩余手牌：\n")
			for _, p := range state.Players {
				role := "农民"
				if p.IsLandlord {
					role = "地主"
				}
				sb.WriteString(fmt.Sprintf("  座位%d（%s）：%d张手牌\n", p.Seat, role, len(p.Hand)))
			}

			// Last play
			if state.LastPlay != nil && len(state.LastPlay.Cards) > 0 {
				lastPlayIDs := make([]int, len(state.LastPlay.Cards))
				for i, c := range state.LastPlay.Cards {
					lastPlayIDs[i] = c.ID
				}
				sb.WriteString(fmt.Sprintf("\n上家(座位%d)出了：%s\n",
					state.LastPlay.Seat, a.formatCardsWithIDs(lastPlayIDs)))
				if state.LastPlay.Play != nil {
					playTypeNames := []string{"无效", "单张", "对子", "三张", "三带一", "三带二", "顺子", "连对", "飞机", "飞机带翅膀", "四带二", "炸弹", "火箭"}
					ptName := "未知"
					if state.LastPlay.Play.Type > 0 && state.LastPlay.Play.Type < len(playTypeNames) {
						ptName = playTypeNames[state.LastPlay.Play.Type]
					}
					sb.WriteString(fmt.Sprintf("牌型：%s，主牌等级：%d\n", ptName, state.LastPlay.Play.MainRank))
				}
				sb.WriteString("你需要出更大的牌型，或选择过牌。\n")
			} else {
				sb.WriteString("\n轮到你了，你可以自由出牌。\n")
			}
		}
	}

	sb.WriteString("\n请为了自己阵营的胜利做出决策。")
	return sb.String()
}

func (a *AIAgent) isActionTool(name string) bool {
	if a.GameType == "dashengji" {
		switch name {
		case "set_trump", "pass_trump", "counter_trump", "pass_counter",
			"take_bottom", "pass_take_bottom", "discard_bottom", "play_cards":
			return true
		}
		return false
	}
	switch name {
	case "play_cards", "bid_landlord", "pass_bid",
		"reveal_cards", "pass_reveal", "choose_double", "choose_no_double":
		return true
	}
	return false
}

func (a *AIAgent) executeToolCall(raw string) {
	call, err := ExtractToolCall(raw)
	if err != nil {
		log.Printf("[AI:%s] executeToolCall: ExtractToolCall failed: %v, falling back", a.UserID, err)
		a.fallbackAction(a.detectPhase())
		return
	}

	log.Printf("[AI:%s] executeToolCall: tool=%s", a.UserID, call.Name)
	if a.GameType == "dashengji" {
		a.executeDashengjiToolCall(call)
		return
	}

	switch call.Name {
	case "play_cards":
		var args PlayCardsArgs
		if json.Unmarshal(call.Args, &args) == nil {
			action := "play"
			if len(args.Cards) == 0 {
				action = "pass"
			}
			log.Printf("[AI:%s] play_cards: action=%s cards=%v", a.UserID, action, args.Cards)
			if a.Executor != nil {
				a.Executor.ExecuteAction(a.UserID, action, args.Cards)
			}
			if args.Chat != "" && a.Executor != nil {
				a.Executor.SendChat(a.UserID, args.Chat, "text")
			}
		} else {
			log.Printf("[AI:%s] play_cards: unmarshal args FAILED, args=%s, falling back", a.UserID, string(call.Args))
			a.fallbackAction(a.detectPhase())
		}

	case "bid_landlord":
		if a.Executor != nil {
			a.Executor.ExecuteAction(a.UserID, "bid_call", nil)
		}
		var args BidArgs
		if json.Unmarshal(call.Args, &args) == nil && args.Chat != "" && a.Executor != nil {
			a.Executor.SendChat(a.UserID, args.Chat, "text")
		}

	case "pass_bid":
		if a.Executor != nil {
			a.Executor.ExecuteAction(a.UserID, "bid_pass", nil)
		}
		var args BidArgs
		if json.Unmarshal(call.Args, &args) == nil && args.Chat != "" && a.Executor != nil {
			a.Executor.SendChat(a.UserID, args.Chat, "text")
		}

	case "reveal_cards":
		if a.Executor != nil {
			a.Executor.ExecuteAction(a.UserID, "reveal_all", nil)
		}

	case "pass_reveal":
		if a.Executor != nil {
			a.Executor.ExecuteAction(a.UserID, "pass", nil)
		}

	case "choose_double":
		if a.Executor != nil {
			a.Executor.ExecuteAction(a.UserID, "double", nil)
		}

	case "choose_no_double":
		if a.Executor != nil {
			a.Executor.ExecuteAction(a.UserID, "no_double", nil)
		}

	}
}

func (a *AIAgent) fallbackAction(phase string) {
	if a.Executor == nil {
		return
	}
	if a.GameType == "dashengji" {
		a.fallbackDashengjiAction(phase)
		return
	}
	switch phase {
	case "calling", "snatching":
		a.Executor.ExecuteAction(a.UserID, "bid_pass", nil)
	case "revealing":
		a.Executor.ExecuteAction(a.UserID, "pass", nil)
	case "doubling":
		a.Executor.ExecuteAction(a.UserID, "no_double", nil)
	default:
		a.Executor.ExecuteAction(a.UserID, "pass", nil)
	}
}

// isFreePlay checks whether the AI is first to act (no last play to follow).
func (a *AIAgent) isFreePlay() bool {
	if a.stateJSON == "" {
		return true
	}
	var state struct {
		LastPlay *struct {
			Seat  int   `json:"seat"`
			Cards []int `json:"cards"`
		} `json:"last_play"`
	}
	if err := json.Unmarshal([]byte(a.stateJSON), &state); err != nil {
		return true
	}
	if state.LastPlay != nil && len(state.LastPlay.Cards) > 0 {
		return false
	}
	return true
}

// ruleBasedAction uses fast rule-based logic when LLM is unavailable.
func (a *AIAgent) ruleBasedAction(phase string) {
	if a.Executor == nil {
		log.Printf("[AI:%s] ruleBasedAction: no executor, cannot act", a.UserID)
		return
	}
	if a.GameType == "dashengji" {
		a.fallbackDashengjiAction(phase)
		return
	}
	switch phase {
	case "calling", "snatching":
		bid := a.ruleBasedBid()
		log.Printf("[AI:%s] ruleBasedAction: phase=%s bid=%s", a.UserID, phase, bid)
		a.Executor.ExecuteAction(a.UserID, bid, nil)
	case "revealing":
		log.Printf("[AI:%s] ruleBasedAction: phase=%s action=pass", a.UserID, phase)
		a.Executor.ExecuteAction(a.UserID, "pass", nil)
	case "doubling":
		log.Printf("[AI:%s] ruleBasedAction: phase=%s action=no_double", a.UserID, phase)
		a.Executor.ExecuteAction(a.UserID, "no_double", nil)
	default:
		isFree := a.isFreePlay()
		cards := a.ruleBasedPlay(isFree)
		if len(cards) == 0 {
			log.Printf("[AI:%s] ruleBasedAction: phase=%s free=%v action=pass", a.UserID, phase, isFree)
			a.Executor.ExecuteAction(a.UserID, "pass", nil)
		} else {
			log.Printf("[AI:%s] ruleBasedAction: phase=%s free=%v action=play cards=%v", a.UserID, phase, isFree, cards)
			a.Executor.ExecuteAction(a.UserID, "play", cards)
		}
	}
}

func (a *AIAgent) ruleBasedBid() string {
	if hasBomb(a.HandCards) || countHighCards(a.HandCards) >= 3 {
		return "bid_call"
	}
	style := "balanced"
	if a.Character != nil {
		style = a.Character.PlayStyle
	}
	threshold := 0.3
	if style == "aggressive" {
		threshold = 0.6
	} else if style == "conservative" {
		threshold = 0.15
	}
	if rand.Float64() < threshold {
		return "bid_call"
	}
	return "bid_pass"
}

func (a *AIAgent) ruleBasedPlay(isFreePlay bool) []int {
	if len(a.HandCards) == 0 {
		return nil
	}
	// Free play: lead with lowest single
	if isFreePlay {
		return a.lowestSingle()
	}
	// Must beat last play: try single or pair
	return a.tryBeatLastPlay()
}

func (a *AIAgent) lowestSingle() []int {
	rankFreq := make(map[int]int)
	for _, id := range a.HandCards {
		if id >= 52 {
			continue
		}
		rankFreq[id%13]++
	}
	for rank := 0; rank < 13; rank++ {
		if rankFreq[rank] == 1 {
			for _, id := range a.HandCards {
				if id < 52 && id%13 == rank {
					return []int{id}
				}
			}
		}
	}
	if len(a.HandCards) > 0 {
		return []int{a.HandCards[0]}
	}
	return nil
}

func (a *AIAgent) tryBeatLastPlay() []int {
	// Parse last play info from state
	var state struct {
		LastPlay *struct {
			Seat  int   `json:"seat"`
			Cards []int `json:"cards"`
			Play  *struct {
				Type     int `json:"type"`
				MainRank int `json:"main_rank"`
			} `json:"play"`
		} `json:"last_play"`
	}
	if a.stateJSON == "" {
		return nil
	}
	if err := json.Unmarshal([]byte(a.stateJSON), &state); err != nil {
		return nil
	}
	if state.LastPlay == nil || state.LastPlay.Play == nil {
		return nil
	}

	lp := state.LastPlay.Play
	switch lp.Type {
	case 1: // Single — try higher single
		return a.higherSingle(lp.MainRank)
	case 2: // Pair — try higher pair
		return a.higherPair(lp.MainRank)
	}
	return nil
}

func (a *AIAgent) higherSingle(minRank int) []int {
	for _, id := range a.HandCards {
		if id >= 52 {
			continue
		}
		rank := id % 13
		if rank > minRank || (id >= 52 && minRank < 16) {
			return []int{id}
		}
	}
	// Try jokers
	for _, id := range a.HandCards {
		if id == 52 || id == 53 {
			return []int{id}
		}
	}
	return nil
}

func (a *AIAgent) higherPair(minRank int) []int {
	rankFreq := make(map[int][]int)
	for _, id := range a.HandCards {
		if id >= 52 {
			continue
		}
		rank := id % 13
		rankFreq[rank] = append(rankFreq[rank], id)
	}
	for rank := minRank + 1; rank < 13; rank++ {
		if len(rankFreq[rank]) >= 2 {
			return []int{rankFreq[rank][0], rankFreq[rank][1]}
		}
	}
	return nil
}

// formatCardsWithIDs formats Dou Dizhu cards as "ID(display)" for LLM consumption,
// e.g. "27(♣4) 12(♠2)" so the LLM knows to use integer IDs.
func (a *AIAgent) formatCardsWithIDs(cards []int) string {
	if a.GameType == "dashengji" {
		return formatDashengjiCardsWithIDs(cards)
	}
	suits := []string{"♠", "♥", "♣", "♦"}
	ranks := []string{"3", "4", "5", "6", "7", "8", "9", "10", "J", "Q", "K", "A", "2"}
	parts := make([]string, len(cards))
	for i, id := range cards {
		if id == 52 {
			parts[i] = fmt.Sprintf("%d(小王)", id)
		} else if id == 53 {
			parts[i] = fmt.Sprintf("%d(大王)", id)
		} else {
			parts[i] = fmt.Sprintf("%d(%s%s)", id, suits[id/13], ranks[id%13])
		}
	}
	return strings.Join(parts, " ")
}

// truncateForLog truncates a string to maxLen for logging.
func truncateForLog(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}
