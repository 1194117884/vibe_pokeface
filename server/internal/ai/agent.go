package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
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
		// already triggered, don't block
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

// makeDecision determines what action the AI should take
func (a *AIAgent) makeDecision() {
	if a.MakeDecisionFunc != nil {
		result := a.MakeDecisionFunc(a, "playing", a.HandCards, a.stateJSON)
		a.executeToolCall(result)
		return
	}

	// Determine phase from state
	phase := a.detectPhase()

	maxIterations := 5
	for i := 0; i < maxIterations; i++ {
		prompt := a.buildAgentPrompt(phase)
		userMsg := a.buildUserMessage(phase)

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		result, err := a.Provider.Complete(ctx, prompt, userMsg)
		cancel()

		if err != nil {
			log.Printf("AI agent LLM error: %v", err)
			a.fallbackAction(phase)
			return
		}

		call, err := ExtractToolCall(result.Content)
		if err != nil {
			log.Printf("AI agent parse error: %v, raw: %s", err, result.Content)
			continue
		}

		if a.isActionTool(call.Name) {
			a.executeToolCall(result.Content)
			return
		}

		// Info tool: execute, feed back result, let LLM decide again
		toolResult := a.executeInfoTool(call.Name)
		resultJSON, _ := json.Marshal(toolResult)

		// In production, this would continue the conversation
		// For now, use the tool result and retry
		_ = resultJSON
	}

	// Fallback if max iterations reached
	a.fallbackAction(phase)
}

func (a *AIAgent) detectPhase() string {
	if a.stateJSON == "" {
		return "bidding"
	}
	var state struct {
		Phase int `json:"phase"`
	}
	if err := json.Unmarshal([]byte(a.stateJSON), &state); err != nil {
		return "bidding"
	}
	if state.Phase == 0 {
		return "bidding"
	}
	return "playing"
}

func (a *AIAgent) buildAgentPrompt(phase string) string {
	var sb strings.Builder
	name := "AI玩家"
	personality := "冷静分析"
	if a.Character != nil {
		if a.Character.Name != "" {
			name = a.Character.Name
		}
		if a.Character.Personality != nil && *a.Character.Personality != "" {
			personality = *a.Character.Personality
		}
	}

	sb.WriteString(fmt.Sprintf("你是斗地主AI玩家「%s」。性格：%s。\n\n", name, personality))
	sb.WriteString("你用繁体中文思考。你通过调用工具来观察牌局并做出决策。\n\n")

	if phase == "bidding" {
		sb.WriteString("## 当前阶段：叫地主\n\n")
		sb.WriteString("牌力强（有炸弹、多张2、有大王）时叫地主，否则不叫。\n\n")
	} else {
		sb.WriteString("## 当前阶段：出牌\n\n")
		sb.WriteString("规则：可出单张、对子、三张、三带一、三带二、顺子(5+)、连对(3+)、飞机、炸弹、火箭。\n")
		sb.WriteString("上家出牌后，你必须出更大的牌或选择过牌。\n")
		sb.WriteString("地主的目标是先出完牌；农民要配合阻止地主。\n\n")
	}

	sb.WriteString("## 可用工具\n\n")
	sb.WriteString("先用信息工具观察局势，再用决策工具行动：\n")
	sb.WriteString("- check_my_hand：查看自己的手牌\n")
	sb.WriteString("- check_game_status：查看当前牌局\n")
	sb.WriteString("- play_cards：出牌（cards=[]表示过牌），可选chat字段附加聊天\n")
	sb.WriteString("- bid_landlord：叫地主\n")
	sb.WriteString("- pass_bid：不叫\n")
	sb.WriteString("- say：在聊天室说话\n\n")
	sb.WriteString("请每次只用一個工具，以JSON格式回复：{\"tool\": \"工具名\", \"args\": {...}}\n")

	return sb.String()
}

func (a *AIAgent) buildUserMessage(phase string) string {
	if a.stateJSON != "" {
		return fmt.Sprintf("当前牌局状态：\n%s\n请使用工具决策。", a.stateJSON)
	}
	cards := ""
	for _, c := range a.HandCards {
		cards += fmt.Sprintf("%d ", c)
	}
	return fmt.Sprintf("你的手牌：%s", cards)
}

func (a *AIAgent) isActionTool(name string) bool {
	switch name {
	case "play_cards", "bid_landlord", "pass_bid", "say":
		return true
	}
	return false
}

func (a *AIAgent) executeToolCall(raw string) {
	call, err := ExtractToolCall(raw)
	if err != nil {
		a.fallbackAction("playing")
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
			if a.Executor != nil {
				a.Executor.ExecuteAction(a.UserID, action, args.Cards)
			}
			if args.Chat != "" && a.Executor != nil {
				a.Executor.SendChat(a.UserID, args.Chat, "text")
			}
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

	case "say":
		if a.Executor != nil {
			a.Executor.SendChat(a.UserID, string(call.Args), "text")
		}
	}
}

func (a *AIAgent) executeInfoTool(name string) *ToolResult {
	switch name {
	case "check_my_hand":
		cards := make([]string, len(a.HandCards))
		for i, id := range a.HandCards {
			cards[i] = fmt.Sprintf("%d", id)
		}
		return &ToolResult{
			Tool:   name,
			Result: fmt.Sprintf("你的手牌(%d张): %s", len(a.HandCards), strings.Join(cards, " ")),
		}
	case "check_game_status":
		if a.stateJSON != "" {
			return &ToolResult{
				Tool:   name,
				Result: fmt.Sprintf("当前牌局：\n%s", a.stateJSON),
			}
		}
		return &ToolResult{
			Tool:   name,
			Result: "暂无牌局信息",
		}
	default:
		return &ToolResult{
			Tool:   name,
			Result: fmt.Sprintf("未知工具: %s", name),
		}
	}
}

func (a *AIAgent) fallbackAction(phase string) {
	if a.Executor == nil {
		return
	}
	if phase == "bidding" {
		a.Executor.ExecuteAction(a.UserID, "bid_pass", nil)
	} else {
		a.Executor.ExecuteAction(a.UserID, "pass", nil)
	}
}
