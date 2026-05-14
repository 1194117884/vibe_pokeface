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

	if a.Provider != nil {
		prompt := a.buildAgentPrompt(phase)
		userMsg := a.buildUserMessage(phase)

		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		result, err := a.Provider.Complete(ctx, prompt, userMsg)
		cancel()

		if err != nil {
			log.Printf("AI agent LLM error: %v, falling back to rule-based", err)
		} else {
			call, parseErr := ExtractToolCall(result.Content)
			if parseErr != nil {
				log.Printf("AI agent parse error: %v, raw: %s", parseErr, result.Content)
			} else if a.isActionTool(call.Name) {
				a.executeToolCall(result.Content)
				return
			} else {
				log.Printf("AI agent returned non-action tool: %s, falling back", call.Name)
			}
		}
	}

	a.ruleBasedAction(phase)
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

	if phase == "bidding" {
		sb.WriteString("## 当前阶段：叫地主\n\n")
		sb.WriteString("根据手牌强度决定是否叫地主。有炸弹、多张2、有大王时应该叫地主。\n")
		sb.WriteString("手牌较弱时选择不叫。\n\n")
	} else {
		sb.WriteString("## 当前阶段：出牌\n\n")
		sb.WriteString("规则：单张、对子、三张、三带一、三带二、顺子(5张+)、连对(3对+)、飞机、炸弹、火箭。\n")
		sb.WriteString("必须出比上家更大的牌，或选择过牌。牌型相同才能比较大小。\n")
		sb.WriteString("地主目标：尽快出完手牌。农民目标：配合队友阻止地主。\n\n")
	}

	sb.WriteString("## 可用工具\n\n")
	if phase == "bidding" {
		sb.WriteString("- bid_landlord：叫地主\n")
		sb.WriteString("- pass_bid：不叫\n\n")
	} else {
		sb.WriteString("- play_cards：出牌，参数cards=要出的牌ID列表，cards=[]表示过牌\n")
	}

	sb.WriteString("请以JSON格式回复，只包含一个工具调用：\n")
	sb.WriteString(`{"tool": "工具名", "args": {...}}` + "\n")

	return sb.String()
}

func (a *AIAgent) buildUserMessage(phase string) string {
	var sb strings.Builder

	// Show the agent's hand cards clearly
	sb.WriteString(fmt.Sprintf("你的手牌（%d张）：\n", len(a.HandCards)))
	sb.WriteString(formatCards(a.HandCards))
	sb.WriteString("\n\n")

	// Show game state context
	if a.stateJSON != "" {
		var state struct {
			Phase        int `json:"phase"`
			CurrentSeat  int `json:"current_seat"`
			LandlordSeat int `json:"landlord_seat"`
			LastPlay     *struct {
				Seat  int   `json:"seat"`
				Cards []int `json:"cards"`
			} `json:"last_play"`
			Players []struct {
				Seat       int   `json:"seat"`
				IsLandlord bool  `json:"is_landlord"`
				Hand       []int `json:"hand"`
			} `json:"players"`
		}
		if json.Unmarshal([]byte(a.stateJSON), &state) == nil {
			sb.WriteString(fmt.Sprintf("你的座位：%d\n", a.Seat))
			sb.WriteString(fmt.Sprintf("当前轮到座位：%d\n", state.CurrentSeat))

			if state.LandlordSeat >= 0 {
				sb.WriteString(fmt.Sprintf("地主座位：%d\n", state.LandlordSeat))
				if state.LandlordSeat == a.Seat {
					sb.WriteString("你是地主！\n")
				} else {
					sb.WriteString("你是农民，队友也是农民。\n")
				}
			}

			if state.LastPlay != nil && len(state.LastPlay.Cards) > 0 {
				sb.WriteString(fmt.Sprintf("上家(座位%d)出了：%s\n",
					state.LastPlay.Seat, formatCards(state.LastPlay.Cards)))
				sb.WriteString("你需要出更大的牌型，或选择过牌。\n")
			} else {
				sb.WriteString("你可以自由出牌。\n")
			}
		}
	}

	sb.WriteString("\n请选择一个工具执行。")
	return sb.String()
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
	if phase == "bidding" {
		a.Executor.ExecuteAction(a.UserID, "bid_pass", nil)
	} else {
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
		return
	}
	if phase == "bidding" {
		a.Executor.ExecuteAction(a.UserID, a.ruleBasedBid(), nil)
	} else {
		cards := a.ruleBasedPlay(a.isFreePlay())
		if len(cards) == 0 {
			a.Executor.ExecuteAction(a.UserID, "pass", nil)
		} else {
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
	if !isFreePlay {
		return nil
	}
	if len(a.HandCards) == 0 {
		return nil
	}
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
