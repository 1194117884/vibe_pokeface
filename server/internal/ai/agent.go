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
}

// NewAIAgent creates a new AI agent
func NewAIAgent(userID string, seat int, character *model.AICharacter, provider LLMProvider, executor ActionExecutor) *AIAgent {
	return &AIAgent{
		UserID:      userID,
		Seat:        seat,
		Character:   character,
		Provider:    provider,
		Executor:    executor,
		triggerChan: make(chan struct{}, 1),
		stopChan:    make(chan struct{}),
	}
}

// SetAIStore sets the AIStore for LLM call logging.
func (a *AIAgent) SetAIStore(store *model.AIStore) {
	a.aiStore = store
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

// Trigger tells the agent it's its turn to act
func (a *AIAgent) Trigger() {
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

	messages := []ChatMessage{
		{Role: "system", Content: a.buildSystemPrompt(phase)},
		{Role: "user", Content: a.buildUserMessage(phase)},
	}

	tools := GetToolSchemas()

	for turn := 0; turn < 10; turn++ {
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		result, err := a.Provider.CompleteWithTools(ctx, messages, tools)
		cancel()

		// Log LLM call to database
		if a.aiStore != nil {
			llmCall := &model.LLMCallLog{
				Provider: a.Provider.ProviderName(),
				Model:    a.Provider.ModelName(),
				CallType: "play_decision",
				Success:  err == nil,
			}
			if err == nil {
				llmCall.PromptTokens = result.PromptTokens
				llmCall.CompletionTokens = result.CompletionTokens
				llmCall.DurationMs = int(result.DurationMs)
			} else {
				errMsg := err.Error()
				llmCall.ErrorMessage = &errMsg
			}
			if logErr := a.aiStore.LogLLMCall(context.Background(), llmCall); logErr != nil {
				log.Printf("[AI:%s] failed to log LLM call: %v", a.UserID, logErr)
			}
		}

		if err != nil {
			log.Printf("[AI:%s] tool loop error at turn %d: %v, falling back", a.UserID, turn, err)
			break
		}

		if len(result.ToolCalls) == 0 {
			// Try parsing text content as JSON tool call (fallback for models without function calling)
			if result.Content != "" {
				if call, parseErr := ExtractToolCall(result.Content); parseErr == nil && a.isActionTool(call.Name) {
					log.Printf("[AI:%s] extracted tool call from text content: %s", a.UserID, call.Name)
					a.executeToolCall(result.Content)
					return
				}
			}
			log.Printf("[AI:%s] no tool calls at turn %d, falling back", a.UserID, turn)
			break
		}

		// Build assistant message with tool calls
		assistantMsg := ChatMessage{
			Role:      "assistant",
			Content:   result.Content,
			ToolCalls: result.ToolCalls,
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
			} else {
				// Action tool (or say) — convert to JSON and execute via existing path
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

func (a *AIAgent) isInfoTool(name string) bool {
	return name == "check_my_hand" || name == "check_game_status"
}

func (a *AIAgent) executeInfoTool(name string) string {
	switch name {
	case "check_my_hand":
		return fmt.Sprintf("你的手牌（%d张）：%s", len(a.HandCards), formatCardsWithIDs(a.HandCards))
	case "check_game_status":
		if a.stateJSON == "" {
			return "游戏状态不可用"
		}
		var state struct {
			Phase        int `json:"phase"`
			CurrentSeat  int `json:"current_seat"`
			LandlordSeat int `json:"landlord_seat"`
			Multiplier   int `json:"multiplier"`
			LastPlay     *struct {
				Seat  int   `json:"seat"`
				Cards []int `json:"cards"`
				Play  *struct {
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
			sb.WriteString(fmt.Sprintf("上家(座位%d)出了：%s\n", state.LastPlay.Seat, formatCardsWithIDs(state.LastPlay.Cards)))
		} else {
			sb.WriteString("当前没有上家出牌（你是首家，可以自由出牌）\n")
		}
		return sb.String()
	}
	return "unknown tool"
}

func toolCallToJSON(tc AssistantToolCall) string {
	data, _ := json.Marshal(map[string]interface{}{
		"tool": tc.Function.Name,
		"args": json.RawMessage(tc.Function.Arguments),
	})
	return string(data)
}

func (a *AIAgent) buildSystemPrompt(phase string) string {
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

	sb.WriteString(fmt.Sprintf("你是斗地主AI玩家「%s」。\n", name))
	sb.WriteString(fmt.Sprintf("性格：%s。\n\n", personality))
	sb.WriteString("规则：单张、对子、三张、三带一、三带二、顺子(5张+)、连对(3对+)、飞机、炸弹、火箭。\n")
	sb.WriteString("必须出比上家更大的牌型，或选择过牌。牌型相同才能比较大小。\n")
	sb.WriteString("地主目标：尽快出完手牌。农民目标：配合队友阻止地主。\n")
	sb.WriteString("出牌时使用卡牌ID（整数），不要使用文字描述。\n")
	sb.WriteString("先使用check_my_hand和check_game_status了解情况，再做出牌决定。\n")

	return sb.String()
}

func (a *AIAgent) detectPhase() string {
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

func (a *AIAgent) buildAgentPrompt(phase string) string {
	var sb strings.Builder
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

	sb.WriteString(fmt.Sprintf("你是斗地主AI玩家「%s」。\n", name))
	sb.WriteString(fmt.Sprintf("性格：%s。出牌风格：%s。\n\n", personality, playStyle))

	switch phase {
	case "calling":
		sb.WriteString("## 当前阶段：叫地主\n\n")
		sb.WriteString("根据手牌强度决定是否叫地主。有炸弹、多张2、有大王时应该叫地主。\n")
		sb.WriteString("手牌较弱时选择不叫。\n\n")
	case "snatching":
		sb.WriteString("## 当前阶段：抢地主\n\n")
		sb.WriteString("有人已经叫地主了。根据手牌强度决定是否抢地主。\n")
		sb.WriteString("抢地主会使倍数翻倍。手牌很强时抢，否则不抢。\n\n")
	case "revealing":
		sb.WriteString("## 当前阶段：明牌\n\n")
		sb.WriteString("选择是否明牌。明牌会让所有人看到你的手牌，但倍数翻倍。\n")
		sb.WriteString("手牌非常好（有炸弹、火箭）时可以考虑明牌，否则选择不显示。\n\n")
	case "doubling":
		sb.WriteString("## 当前阶段：加倍\n\n")
		sb.WriteString("选择是否加倍。加倍会让你的得分翻倍（赢多输多）。\n")
		sb.WriteString("手牌很好或你是地主时可以考虑加倍，否则选择不加倍。\n\n")
	default:
		sb.WriteString("## 当前阶段：出牌\n\n")
		sb.WriteString("规则：单张、对子、三张、三带一、三带二、顺子(5张+)、连对(3对+)、飞机、炸弹、火箭。\n")
		sb.WriteString("必须出比上家更大的牌，或选择过牌。牌型相同才能比较大小。\n")
		sb.WriteString("地主目标：尽快出完手牌。农民目标：配合队友阻止地主。\n")
			sb.WriteString("重要：出牌时必须使用卡牌的整数ID（如27），不要使用文字描述（如♣4）。手牌列表中的格式为ID(显示)。\n\n")
	}

	sb.WriteString("## 可用工具\n\n")
	sb.WriteString("所有工具都支持可选的chat参数，用于在操作时发送聊天消息（不超过30字）。\n")
	sb.WriteString("- say：在聊天室说话，参数message=要说的话\n\n")
	switch phase {
	case "calling", "snatching":
		sb.WriteString("- bid_landlord：叫地主/抢地主，可选参数chat=叫地主时说的话\n")
		sb.WriteString("- pass_bid：不叫/不抢，可选参数chat=不叫时说的话\n\n")
	case "revealing":
		sb.WriteString("- reveal_cards：明牌（亮出手牌）\n")
		sb.WriteString("- pass_reveal：不明牌\n\n")
	case "doubling":
		sb.WriteString("- choose_double：加倍\n")
		sb.WriteString("- choose_no_double：不加倍\n\n")
	default:
		sb.WriteString("- play_cards：出牌，参数cards=要出的牌ID列表（cards=[]表示过牌），可选参数chat=出牌时说的话\n")
	}

	sb.WriteString("请以JSON格式回复，只包含一个工具调用：\n")
	sb.WriteString(`{"tool": "工具名", "args": {...}}` + "\n")

	return sb.String()
}

func (a *AIAgent) buildUserMessage(phase string) string {
	var sb strings.Builder

	// Show the agent's hand cards clearly
	sb.WriteString(fmt.Sprintf("你的手牌（%d张）：\n", len(a.HandCards)))
	sb.WriteString(formatCardsWithIDs(a.HandCards))
	sb.WriteString("\n（每张牌格式：ID(显示)，使用ID来出牌）\n\n")

	// Show game state context
	if a.stateJSON != "" {
		var state struct {
			Phase         int `json:"phase"`
			CurrentSeat   int `json:"current_seat"`
			LandlordSeat  int `json:"landlord_seat"`
			Multiplier    int `json:"multiplier"`
			LandlordCards []int `json:"landlord_cards"`
			LastPlay      *struct {
				Seat  int   `json:"seat"`
				Cards []int `json:"cards"`
				Play  *struct {
					Type     int `json:"type"`
					MainRank int `json:"main_rank"`
					Length   int `json:"length"`
				} `json:"play"`
			} `json:"last_play"`
			Players []struct {
				Seat       int   `json:"seat"`
				IsLandlord bool  `json:"is_landlord"`
				Hand       []int `json:"hand"`
			} `json:"players"`
			BidHistory []struct {
				Seat   int  `json:"seat"`
				Called bool `json:"called"`
			} `json:"bid_history"`
		}
		if json.Unmarshal([]byte(a.stateJSON), &state) == nil {
			sb.WriteString(fmt.Sprintf("你的座位：%d\n", a.Seat))
			sb.WriteString(fmt.Sprintf("当前轮到座位：%d\n", state.CurrentSeat))
			sb.WriteString(fmt.Sprintf("当前倍率：%d\n", state.Multiplier))

			// Landlord info and 底牌
			if state.LandlordSeat >= 0 {
				sb.WriteString(fmt.Sprintf("地主座位：%d\n", state.LandlordSeat))
				if state.LandlordSeat == a.Seat {
					sb.WriteString("你是地主！\n")
					if len(state.LandlordCards) > 0 {
						sb.WriteString(fmt.Sprintf("地主牌（底牌）：%s\n", formatCardsWithIDs(state.LandlordCards)))
					}
				} else {
					sb.WriteString("你是农民，队友也是农民。\n")
					if len(state.LandlordCards) > 0 {
						sb.WriteString(fmt.Sprintf("地主牌（底牌）：%s\n", formatCardsWithIDs(state.LandlordCards)))
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
				sb.WriteString(fmt.Sprintf("\n上家(座位%d)出了：%s\n",
					state.LastPlay.Seat, formatCardsWithIDs(state.LastPlay.Cards)))
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
				sb.WriteString("\n你可以自由出牌（你是本轮首家）。\n")
			}
		}
	}

	sb.WriteString("\n请选择一个工具执行。")
	return sb.String()
}

func (a *AIAgent) isActionTool(name string) bool {
	switch name {
	case "play_cards", "bid_landlord", "pass_bid", "say",
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

	case "say":
		if a.Executor != nil {
			var args SayArgs
			if json.Unmarshal(call.Args, &args) == nil && args.Message != "" {
				a.Executor.SendChat(a.UserID, args.Message, "text")
			}
		}
	}
}

func (a *AIAgent) fallbackAction(phase string) {
	if a.Executor == nil {
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

// formatCardsWithIDs formats cards as "ID(display)" for LLM consumption,
// e.g. "27(♣4) 12(♠2)" so the LLM knows to use integer IDs.
func formatCardsWithIDs(cards []int) string {
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

// formatCards converts card IDs to human-readable format like "♠A ♥K ♦7"
func formatCards(cards []int) string {
	suits := []string{"♠", "♥", "♣", "♦"}
	ranks := []string{"3", "4", "5", "6", "7", "8", "9", "10", "J", "Q", "K", "A", "2"}
	parts := make([]string, len(cards))
	for i, id := range cards {
		if id == 52 {
			parts[i] = "小王"
		} else if id == 53 {
			parts[i] = "大王"
		} else {
			parts[i] = suits[id/13] + ranks[id%13]
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
