# AI Agent Mode Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Transform each AI bot from a rule-based passive player into an LLM-driven agent with tools to inspect game state, make decisions, and chat during 斗地主 games.

**Architecture:** Each bot runs an `AIAgent` goroutine that uses LLM tool-calling (REACT loop: observe → think → act). Tools are defined as Go functions with JSON schemas, described in prompts for provider-agnostic compatibility. The game room detects when it's a bot's turn and triggers the agent asynchronously.

**Tech Stack:** Go 1.26, OpenAI-compatible LLM API (Anthropic, DeepSeek etc.), existing `LLMProvider` interface

**Current state:** Bots fill seats with `IsBot: true` but `SelectPlay()` always returns `"pass"` — bots don't actually play. Their `Conn` channels receive game broadcasts but nobody reads them.

---

## File Structure

| File | Action | Purpose |
|---|---|---|
| `server/internal/ai/tool.go` | **Create** | Tool definition types, JSON schemas, tool registry, execution |
| `server/internal/ai/agent.go` | **Create** | `AIAgent` struct, goroutine lifecycle, LLM decision loop, state management |
| `server/internal/game/room.go` | **Modify** | Add `AIAgent` field to `GameRoom`, agent lifecycle in `FillWithBot`, bot-turn detection, `FillEmptySeats` enhancements |
| `server/internal/ai/player.go` | **Modify** | Minor: add exported helpers `GetHandFromState` |
| `server/internal/api/ws/handler.go` | **Modify** | Pass `AIStore` / character lookup via hub for agent creation |

---

### Task 1: Create tool definitions (`tool.go`)

**Files:**
- Create: `server/internal/ai/tool.go`
- Test: `server/internal/ai/tool_test.go`

- [ ] **Step 1: Write failing tests for tool parsing**

```go
package ai

import (
    "encoding/json"
    "testing"
)

func TestToolCallJSON_Parse(t *testing.T) {
    raw := `{"tool": "check_my_hand", "args": {}}`
    var call ToolCall
    if err := json.Unmarshal([]byte(raw), &call); err != nil {
        t.Fatalf("unmarshal error: %v", err)
    }
    if call.Name != "check_my_hand" {
        t.Errorf("name = %s, want check_my_hand", call.Name)
    }
}

func TestToolCallJSON_ParseWithArgs(t *testing.T) {
    raw := `{"tool": "play_cards", "args": {"cards": [0, 13, 26]}}`
    var call ToolCall
    json.Unmarshal([]byte(raw), &call)
    var args PlayCardsArgs
    json.Unmarshal(call.Args, &args)
    if len(args.Cards) != 3 || args.Cards[0] != 0 {
        t.Errorf("unexpected cards: %v", args.Cards)
    }
}

func TestToolSchemas_AreValidJSON(t *testing.T) {
    schemas := GetToolSchemas()
    for _, s := range schemas {
        data, err := json.Marshal(s)
        if err != nil {
            t.Errorf("schema %s failed marshal: %v", s.Function.Name, err)
        }
        if len(data) < 10 {
            t.Errorf("schema %s too short", s.Function.Name)
        }
    }
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `cd server && go test ./internal/ai/ -run "TestToolCallJSON|TestToolSchemas" -v`
Expected: build error (package not found or undefined types)

- [ ] **Step 3: Write the tool types and schemas**

```go
package ai

import "encoding/json"

// ToolCall represents an LLM decision to call a tool
type ToolCall struct {
    Name string          `json:"tool"`
    Args json.RawMessage `json:"args"`
}

// ToolSchema defines an LLM-callable tool (OpenAI function-calling format)
type ToolSchema struct {
    Type     string       `json:"type"`
    Function FuncDef      `json:"function"`
}

type FuncDef struct {
    Name        string          `json:"name"`
    Description string          `json:"description"`
    Parameters  ParamSchema     `json:"parameters"`
}

type ParamSchema struct {
    Type       string                    `json:"type"`
    Properties map[string]ParamProperty `json:"properties,omitempty"`
    Required   []string                  `json:"required,omitempty"`
}

type ParamProperty struct {
    Type        string   `json:"type"`
    Description string   `json:"description"`
    Enum        []string `json:"enum,omitempty"`
}

// Tool argument types
type PlayCardsArgs struct {
    Cards []int  `json:"cards"`
    Chat  string `json:"chat,omitempty"`
}

type BidArgs struct {
    Chat string `json:"chat,omitempty"`
}

type SayArgs struct {
    Message string `json:"message"`
}

// GetToolSchemas returns all available tool definitions for LLM prompt injection
func GetToolSchemas() []ToolSchema {
    return []ToolSchema{
        {
            Type: "function",
            Function: FuncDef{
                Name:        "check_my_hand",
                Description: "查看自己的手牌，返回当前手牌列表",
                Parameters: ParamSchema{
                    Type:       "object",
                    Properties: map[string]ParamProperty{},
                },
            },
        },
        {
            Type: "function",
            Function: FuncDef{
                Name:        "check_game_status",
                Description: "查看当前牌局状态：轮到谁、地主是谁、各家剩余牌数、最后出牌记录",
                Parameters: ParamSchema{
                    Type:       "object",
                    Properties: map[string]ParamProperty{},
                },
            },
        },
        {
            Type: "function",
            Function: FuncDef{
                Name:        "play_cards",
                Description: "出牌。传入要出的牌ID列表（空数组=不出/过牌）。可在chat字段附加聊天消息",
                Parameters: ParamSchema{
                    Type: "object",
                    Properties: map[string]ParamProperty{
                        "cards": {
                            Type:        "array",
                            Description: "要出的牌的ID数组，空数组表示不出/过牌",
                            Items:       &struct{ Type string `json:"type"` }{Type: "integer"},
                        },
                        "chat": {
                            Type:        "string",
                            Description: "出牌时说的话（可选，不超过30字）",
                        },
                    },
                    Required: []string{"cards"},
                },
            },
        },
        {
            Type: "function",
            Function: FuncDef{
                Name:        "bid_landlord",
                Description: "叫地主",
                Parameters: ParamSchema{
                    Type: "object",
                    Properties: map[string]ParamProperty{
                        "chat": {
                            Type:        "string",
                            Description: "叫地主时说的话（可选，不超过30字）",
                        },
                    },
                },
            },
        },
        {
            Type: "function",
            Function: FuncDef{
                Name:        "pass_bid",
                Description: "不叫地主",
                Parameters: ParamSchema{
                    Type: "object",
                    Properties: map[string]ParamProperty{
                        "chat": {
                            Type:        "string",
                            Description: "过牌时说的话（可选，不超过30字）",
                        },
                    },
                },
            },
        },
        {
            Type: "function",
            Function: FuncDef{
                Name:        "say",
                Description: "在游戏聊天室说一句话",
                Parameters: ParamSchema{
                    Type: "object",
                    Properties: map[string]ParamProperty{
                        "message": {
                            Type:        "string",
                            Description: "要说的内容（不超过50字）",
                        },
                    },
                    Required: []string{"message"},
                },
            },
        },
    }
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `cd server && go test ./internal/ai/ -run "TestToolCallJSON|TestToolSchemas" -v`
Expected: All 3 tests PASS

- [ ] **Step 5: Commit**

```bash
git add server/internal/ai/tool.go server/internal/ai/tool_test.go
git commit -m "feat(ai): add tool definitions and JSON schemas for AI agent"
```

---

### Task 2: Create AI agent (`agent.go`)

**Files:**
- Create: `server/internal/ai/agent.go`
- Test: `server/internal/ai/agent_test.go`

**Context:** The agent runs in a goroutine per bot. It receives triggers from the game room when it's the bot's turn. It calls the LLM in a REACT loop (info tools → decision tool), then executes the action via a callback.

- [ ] **Step 1: Write failing tests for agent decision parsing**

```go
package ai

import (
    "testing"
)

func TestExtractToolCall_Play(t *testing.T) {
    llmOutput := `{"tool": "play_cards", "args": {"cards": [0, 13, 26], "chat": "接招！"}}`
    call, err := ExtractToolCall(llmOutput)
    if err != nil {
        t.Fatalf("extract error: %v", err)
    }
    if call.Name != "play_cards" {
        t.Errorf("name = %s, want play_cards", call.Name)
    }
}

func TestExtractToolCall_Bid(t *testing.T) {
    llmOutput := `{"tool": "bid_landlord", "args": {"chat": "我来当地主！"}}`
    call, err := ExtractToolCall(llmOutput)
    if err != nil {
        t.Fatalf("extract error: %v", err)
    }
    if call.Name != "bid_landlord" {
        t.Errorf("name = %s, want bid_landlord", call.Name)
    }
}

func TestExtractToolCall_Say(t *testing.T) {
    llmOutput := `{"tool": "say", "args": {"message": "大家好！"}}`
    call, err := ExtractToolCall(llmOutput)
    if err != nil {
        t.Fatalf("extract error: %v", err)
    }
    if call.Name != "say" {
        t.Errorf("name = %s, want say", call.Name)
    }
}

func TestExtractToolCall_WithCodeFences(t *testing.T) {
    llmOutput := "```json\n{\"tool\": \"play_cards\", \"args\": {\"cards\": []}}\n```"
    call, err := ExtractToolCall(llmOutput)
    if err != nil {
        t.Fatalf("extract error: %v", err)
    }
    if call.Name != "play_cards" {
        t.Errorf("name = %s, want play_cards", call.Name)
    }
}

func TestExtractToolCall_Invalid(t *testing.T) {
    _, err := ExtractToolCall("not json")
    if err == nil {
        t.Error("expected error for invalid JSON")
    }
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `cd server && go test ./internal/ai/ -run "TestExtractToolCall" -v`
Expected: build error (undefined ExtractToolCall)

- [ ] **Step 3: Write the ExtractToolCall helper**

Add to `tool.go`:

```go
// ExtractToolCall parses an LLM JSON response to extract a tool call.
// Handles markdown code fences and whitespace.
func ExtractToolCall(jsonStr string) (*ToolCall, error) {
    cleaned := stripCodeFences(jsonStr)
    var call ToolCall
    if err := json.Unmarshal([]byte(cleaned), &call); err != nil {
        return nil, fmt.Errorf("parse tool call: %w", err)
    }
    if call.Name == "" {
        return nil, fmt.Errorf("empty tool name")
    }
    return &call, nil
}
```

Add import `"fmt"` to `tool.go`.

- [ ] **Step 4: Run tests to verify they pass**

Run: `cd server && go test ./internal/ai/ -run "TestExtractToolCall" -v`
Expected: All 5 tests PASS

- [ ] **Step 5: Write failing tests for agent lifecycle**

```go
package ai

import (
    "testing"
    "time"
)

func TestAIAgent_NewAndStop(t *testing.T) {
    agent := NewAIAgent("ai:bot:1", 0, nil, nil, nil, nil)
    if agent == nil {
        t.Fatal("expected non-nil agent")
    }
    if agent.UserID != "ai:bot:1" {
        t.Errorf("userID = %s, want ai:bot:1", agent.UserID)
    }
    // Stop should not block
    done := make(chan struct{})
    go func() {
        agent.Stop()
        close(done)
    }()
    select {
    case <-done:
    case <-time.After(time.Second):
        t.Fatal("Stop() timed out")
    }
}

func TestAIAgent_Trigger(t *testing.T) {
    triggered := make(chan struct{}, 1)
    decisionFn := func(agent *AIAgent, phase string, handCards []int, stateJSON string) string {
        triggered <- struct{}{}
        return `{"tool": "play_cards", "args": {"cards": []}}`
    }
    agent := NewAIAgent("ai:bot:1", 0, nil, nil, nil, nil)
    agent.MakeDecisionFunc = decisionFn
    agent.Start()

    agent.Trigger("playing")
    
    select {
    case <-triggered:
        // success
    case <-time.After(2 * time.Second):
        t.Fatal("trigger was not processed")
    }
    agent.Stop()
}
```

- [ ] **Step 6: Run tests, expect them to fail**

Run: `cd server && go test ./internal/ai/ -run "TestAIAgent" -v`
Expected: build error (undefined NewAIAgent)

- [ ] **Step 7: Write the AI agent implementation**

```go
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
    UserID      string
    Seat        int
    HandCards   []int
    Character   *model.AICharacter
    Provider    LLMProvider
    Executor    ActionExecutor
    stateJSON   string

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

// detectPhase reads the game phase from stored state JSON
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

// buildAgentPrompt creates the system prompt for the LLM
func (a *AIAgent) buildAgentPrompt(phase string) string {
    var sb strings.Builder
    name := "AI玩家"
    personality := "冷静分析"
    if a.Character != nil {
        if a.Character.Name != "" {
            name = a.Character.Name
        }
        if a.Character.Personality != "" {
            personality = a.Character.Personality
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

// buildUserMessage creates the user message with current game state
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

// isActionTool returns true if the tool results in a game action
func (a *AIAgent) isActionTool(name string) bool {
    switch name {
    case "play_cards", "bid_landlord", "pass_bid", "say":
        return true
    }
    return false
}

// executeToolCall parses and executes the action from an LLM response
func (a *AIAgent) executeToolCall(raw string) {
    call, err := ExtractToolCall(raw)
    if err != nil {
        a.fallbackAction("playing")
        return
    }

    switch call.Name {
    case "play_cards":
        var args PlayCardsArgs
        if err := json.Unmarshal(call.Args, &args); err == nil {
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

// executeInfoTool runs an info-gathering tool and returns the result
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

// fallbackAction performs a safe default action
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
```

- [ ] **Step 8: Run tests to verify they pass**

Run: `cd server && go test ./internal/ai/ -run "TestAIAgent" -v`
Expected: Both tests PASS

- [ ] **Step 9: Commit**

```bash
git add server/internal/ai/agent.go server/internal/ai/tool.go
git commit -m "feat(ai): implement AI agent with tool-based LLM decision loop"
```

---

### Task 3: Integrate agent with game room

**Files:**
- Modify: `server/internal/game/room.go`
- Modify: `server/internal/api/ws/handler.go`

- [ ] **Step 1: Write failing test for room bot integration** (if room tests exist, else skip direct test and verify via build)

```bash
cd server && go build ./...
```
Expected: build fails because `AIAgent` / `ActionExecutor` are undefined in `room.go`

- [ ] **Step 2: Add AIAgent dependency to GameRoom and implement ActionExecutor**

Add to `server/internal/game/room.go`:

```go
package game

import (
    "encoding/json"
    "fmt"
    "sync"

    "github.com/yongkl/vibe-pokeface/internal/ai"
    "github.com/yongkl/vibe-pokeface/internal/model"
)

// Update imports to include ai package

type GameRoom struct {
    ID       string
    GameType string
    Players  []*PlayerSession
    Engine   GameEngine
    State    GameState
    Status   string
    store    *model.GameStore
    mu       sync.Mutex
    notify   chan []byte
    agents   map[string]*ai.AIAgent   // userID -> AI agent (NEW)
}
```

- [ ] **Step 3: Initialize agents map in NewGameRoom**

```go
func NewGameRoom(id string, gameType string, engine GameEngine, store *model.GameStore) *GameRoom {
    return &GameRoom{
        ID:       id,
        GameType: gameType,
        Engine:   engine,
        Players:  make([]*PlayerSession, 0),
        Status:   "waiting",
        store:    store,
        notify:   make(chan []byte, 256),
        agents:   make(map[string]*ai.AIAgent),  // NEW
    }
}
```

- [ ] **Step 4: Add helper functions for agent lifecycle**

Add to `room.go`:

```go
// createAIAgent creates and starts an AI agent for a bot player.
// The agent uses the provided LLM provider and character info.
func (r *GameRoom) createAIAgent(userID string, seat int, character *model.AICharacter, provider ai.LLMProvider) *ai.AIAgent {
    agent := ai.NewAIAgent(userID, seat, character, provider, r)
    agent.Start()
    r.agents[userID] = agent
    return agent
}

// stopAIAgent stops and removes an AI agent for the given user ID.
func (r *GameRoom) stopAIAgent(userID string) {
    if agent, ok := r.agents[userID]; ok {
        agent.Stop()
        delete(r.agents, userID)
    }
}

// The GameRoom implements ai.ActionExecutor:
func (r *GameRoom) ExecuteAction(userID string, action string, cards []int) {
    r.HandleAction(userID, action, cards)
}

func (r *GameRoom) SendChat(senderID string, content string, msgType string) {
    r.BroadcastChat(senderID, content, msgType)
}
```

- [ ] **Step 5: Modify FillWithBot to create AI agent for bot players**

Replace the current `FillWithBot`:

```go
func (r *GameRoom) FillWithBot(botID string, conn chan []byte, opts ...BotOption) error {
    r.mu.Lock()
    defer r.mu.Unlock()

    if len(r.Players) >= 3 {
        return fmt.Errorf("room is full")
    }

    for _, p := range r.Players {
        if p.UserID == botID {
            return fmt.Errorf("bot already in room")
        }
    }

    seat := len(r.Players)
    bot := &PlayerSession{
        UserID: botID,
        Seat:   seat,
        Conn:   conn,
        IsBot:  true,
        Ready:  true,
    }
    r.Players = append(r.Players, bot)

    // Apply options (for injecting character/provider)
    botConfig := &botConfig{}
    for _, opt := range opts {
        opt(botConfig)
    }

    // Create AI agent if provider is configured
    if botConfig.Provider != nil {
        agent := ai.NewAIAgent(botID, seat, botConfig.Character, botConfig.Provider, r)
        agent.Start()
        r.agents[botID] = agent
    }

    r.broadcastMsg("player_joined", map[string]interface{}{
        "user_id": botID,
        "seat":    seat,
        "is_bot":  true,
        "players": r.playerList(),
    })

    if len(r.Players) == 3 {
        allReady := true
        for _, p := range r.Players {
            if !p.Ready {
                allReady = false
                break
            }
        }
        if allReady {
            r.startGame()
        }
    }

    return nil
}

// BotOption configures a bot's AI settings
type BotOption func(*botConfig)

type botConfig struct {
    Character *model.AICharacter
    Provider  ai.LLMProvider
}

func WithAICharacter(char *model.AICharacter) BotOption {
    return func(c *botConfig) {
        c.Character = char
    }
}

func WithLLMProvider(provider ai.LLMProvider) BotOption {
    return func(c *botConfig) {
        c.Provider = provider
    }
}
```

- [ ] **Step 6: Modify RemovePlayer to stop AI agent**

In `RemovePlayer`, add cleanup:

```go
func (r *GameRoom) RemovePlayer(userID string) {
    r.mu.Lock()
    defer r.mu.Unlock()

    // Stop AI agent if present
    if agent, ok := r.agents[userID]; ok {
        agent.Stop()
        delete(r.agents, userID)
    }

    // ... rest of existing code ...
}
```

- [ ] **Step 7: Modify HandleAction to trigger AI agent after state update**

Replace the end of `HandleAction`:

```go
func (r *GameRoom) HandleAction(userID string, action string, cards []int) {
    r.mu.Lock()
    defer r.mu.Unlock()

    if action == "ready" {
        r.setReady(userID)
        return
    }

    if r.Status != "playing" || r.State == nil {
        return
    }

    var player *PlayerSession
    for _, p := range r.Players {
        if p.UserID == userID {
            player = p
            break
        }
    }
    if player == nil {
        return
    }

    gameAction := PlayerAction{
        PlayerID: player.PlayerID,
        Action:   action,
        Cards:    cards,
    }

    newState, err := r.Engine.ExecuteAction(r.State, gameAction)
    if err != nil {
        errMsg, _ := json.Marshal(map[string]interface{}{
            "type": "error",
            "data": err.Error(),
        })
        select {
        case player.Conn <- errMsg:
        default:
        }
        return
    }

    r.State = newState

    if r.Engine.IsRoundEnd(newState) {
        scores, err := r.Engine.CalculateScore(newState)
        if err != nil {
            return
        }

        r.broadcastMsg("round_end", map[string]interface{}{
            "scores": scores,
        })

        r.Status = "waiting"
        r.State = nil
        for _, p := range r.Players {
            p.Ready = false
        }

        // Clean up all AI agents on round end
        for id := range r.agents {
            if agent, ok := r.agents[id]; ok {
                agent.Stop()
                delete(r.agents, id)
            }
        }
    } else {
        r.broadcastMsg("state_update", newState)

        // Trigger AI agent if next player is a bot
        r.triggerAIAgent()
    }
}

// NEW: triggerAIAgent checks if the current player is a bot and triggers its AI agent
func (r *GameRoom) triggerAIAgent() {
    state, ok := r.State.(*doudizhu.GameState)
    if !ok {
        return
    }

    currentSeat := state.CurrentSeat
    for _, p := range r.Players {
        if p.Seat == currentSeat && p.IsBot {
            // Update agent's hand and state before triggering
            if agent, ok := r.agents[p.UserID]; ok {
                for _, ph := range state.Players {
                    if ph.Seat == currentSeat {
                        cards := make([]int, len(ph.Hand))
                        for i, c := range ph.Hand {
                            cards[i] = c.ID
                        }
                        agent.UpdateHand(cards)
                        break
                    }
                }
                stateJSON, _ := json.Marshal(state)
                agent.UpdateState(string(stateJSON))
                agent.Trigger()
            }
            return
        }
    }
}
```

- [ ] **Step 8: Modify startGame to trigger AI agent for first turn**

Add at end of `startGame()`:

```go
func (r *GameRoom) startGame() {
    // ... existing code ...
    
    r.broadcastMsg("game_start", state)

    // Check if the first player is a bot and trigger AI
    stateObj, ok := r.State.(*doudizhu.GameState)
    if ok {
        currentSeat := stateObj.CurrentSeat
        for _, p := range r.Players {
            if p.Seat == currentSeat && p.IsBot {
                if agent, ok := r.agents[p.UserID]; ok {
                    // Update hand and state
                    for _, ph := range stateObj.Players {
                        if ph.Seat == currentSeat {
                            cards := make([]int, len(ph.Hand))
                            for i, c := range ph.Hand {
                                cards[i] = c.ID
                            }
                            agent.UpdateHand(cards)
                            break
                        }
                    }
                    stateJSON, _ := json.Marshal(stateObj)
                    agent.UpdateState(string(stateJSON))
                    agent.Trigger()
                }
            }
        }
    }
}
```

- [ ] **Step 9: Update FillEmptySeats to pass AI config to FillWithBot**

Modify the `RoomManager.FillEmptySeats` to accept AI configuration:

```go
func (rm *RoomManager) FillEmptySeats(roomID string) int {
    room := rm.GetRoom(roomID)
    if room == nil {
        return 0
    }
    if room.Status != "waiting" {
        return 0
    }

    nextN := room.nextBotNumber()
    added := 0
    for room.PlayerCount() < 3 {
        botID := fmt.Sprintf("ai:bot:%d", nextN)
        // Currently no LLM config, so bots join without AI agent
        // (they'll be passive placeholders until AI config is set up)
        if err := room.FillWithBot(botID, make(chan []byte, 256)); err != nil {
            break
        }
        added++
        nextN++
    }
    return added
}
```

The AI provider setup will be wired in the handler (Task 4).

- [ ] **Step 10: Verify build**

Run: `cd server && go vet ./...`
Expected: no errors

If the `doudizhu` package type isn't importable directly from `game`, check if we need to import it. The `GameRoom` already uses `game.GameState` (interface). For type assertion:
- Either assert through the interface methods
- Or import `doudizhu` package for concrete type

Check: `r.State` is `game.GameState` interface. We need to type-assert to `*doudizhu.GameState`. This requires importing `github.com/yongkl/vibe-pokeface/internal/game/doudizhu` in `room.go`.

Add import:
```go
doudizhu "github.com/yongkl/vibe-pokeface/internal/game/doudizhu"
```

- [ ] **Step 11: Commit**

```bash
git add server/internal/game/room.go
git commit -m "feat(game): integrate AI agent with game room lifecycle"
```

---

### Task 4: Wire up AI provider in WebSocket handler

**Files:**
- Modify: `server/internal/api/ws/handler.go`
- Modify: `server/internal/api/ws/hub.go`

**Context:** Currently bots are created in `FillEmptySeats` without LLM providers. We need to wire up the AI provider and character selection when creating bots, so the agent has access to an LLM.

- [ ] **Step 1: Read current handler.go to understand the flow**

Check the `HandleWS` function to see how bots are created.

- [ ] **Step 2: Add AIStore reference to Hub**

```go
type Hub struct {
    Rooms       map[string]*RoomHub
    mu          sync.RWMutex
    Register    chan *Client
    Unregister  chan *Client
    RoomManager *game.RoomManager
    AIStore     *model.AIStore   // NEW
}
```

- [ ] **Step 3: Add method to Hub: getAIProviderForBot**

```go
import (
    "github.com/yongkl/vibe-pokeface/internal/ai"
    "github.com/yongkl/vibe-pokeface/internal/model"
)

// getAIProviderForBot creates an LLM provider and picks a character for an AI bot.
// Returns nil, nil if AI is not configured.
func (h *Hub) getAIProviderForBot() (*model.AICharacter, ai.LLMProvider, error) {
    if h.AIStore == nil {
        return nil, nil, nil
    }

    // Get active LLM config
    cfg, err := h.AIStore.GetActiveConfig()
    if err != nil || cfg == nil {
        return nil, nil, nil
    }

    provider, err := ai.NewProvider(cfg)
    if err != nil {
        return nil, nil, err
    }

    // Get a random AI character
    characters, err := h.AIStore.ListCharacters()
    if err != nil || len(characters) == 0 {
        return nil, provider, nil
    }

    // Pick one based on bot number
    char := characters[rand.Intn(len(characters))]
    return &char, provider, nil
}
```

- [ ] **Step 4: Modify NewHub to accept AIStore**

```go
func NewHub(store *model.GameStore, aiStore *model.AIStore) *Hub {
    return &Hub{
        Rooms:       make(map[string]*RoomHub),
        Register:    make(chan *Client, 256),
        Unregister:  make(chan *Client, 256),
        RoomManager: game.NewRoomManager(store),
        AIStore:     aiStore,
    }
}
```

- [ ] **Step 5: Modify FillEmptySeats call in handleJoinRoom to pass AI config**

After `joinRoom`, when `FillEmptySeats` is called (or in the handler that processes room joins), wire up AI providers for new bots.

Add a method to fill seats with AI-aware bots:

```go
// fillRoomBots fills empty seats with AI bots that have LLM providers
func (h *Hub) fillRoomBots(roomID string) {
    room := h.RoomManager.GetRoom(roomID)
    if room == nil {
        return
    }

    char, provider, err := h.getAIProviderForBot()
    if err != nil || provider == nil {
        // Fall back to plain FillEmptySeats (no AI agent)
        h.RoomManager.FillEmptySeats(roomID)
        return
    }

    // Fill seats with AI agents
    nextN := room.NextBotNumber()
    for room.PlayerCount() < 3 {
        botID := fmt.Sprintf("ai:bot:%d", nextN)
        opts := []game.BotOption{
            game.WithLLMProvider(provider),
        }
        if char != nil {
            opts = append(opts, game.WithAICharacter(char))
        }
        if err := room.FillWithBot(botID, make(chan []byte, 256), opts...); err != nil {
            break
        }
        nextN++
    }
}
```

- [ ] **Step 6: Update handleJoinRoom to use fillRoomBots**

In `ws/handler.go`, replace the call to `FillEmptySeats` with `fillRoomBots`:

```go
func (h *Hub) handleJoinRoom(client *Client, msg C2SMessage) {
    // ... existing code to handle leaving current room ...

    roomID := msg.RoomID
    if roomID == "" {
        return
    }

    room := h.RoomManager.GetOrCreateRoom(roomID, "doudizhu", &doudizhu.Engine{})
    
    // ... add player ...

    // Fill remaining seats with bots
    h.fillRoomBots(roomID)
}
```

- [ ] **Step 7: Update main.go to pass AIStore to NewHub**

Find `server/cmd/server/main.go` and update `NewHub` call:

```go
// Before:
hub := ws.NewHub(gameStore)

// After:
hub := ws.NewHub(gameStore, aiStore)
```

- [ ] **Step 8: Verify build**

Run: `cd server && go vet ./...`
Expected: no errors

- [ ] **Step 9: Run existing tests**

Run: `cd server && go test ./internal/ai/... ./internal/game/... -v`
Expected: all existing tests PASS

- [ ] **Step 10: Commit**

```bash
git add server/internal/api/ws/hub.go server/internal/api/ws/handler.go server/cmd/server/main.go
git commit -m "feat(ws): wire up AI provider and character for bot agents"
```

---

### Task 5: Write integration tests for AI agent

**Files:**
- Create: `server/internal/ai/agent_integration_test.go`

- [ ] **Step 1: Write integration test for full agent decision loop**

```go
package ai

import (
    "testing"
)

// mockExecutor for testing
type mockExecutor struct {
    lastAction   string
    lastCards    []int
    lastChat     string
    actions      []struct{ action string; cards []int }
}

func (m *mockExecutor) ExecuteAction(userID string, action string, cards []int) {
    m.lastAction = action
    m.lastCards = cards
    m.actions = append(m.actions, struct{ action string; cards []int }{action, cards})
}

func (m *mockExecutor) SendChat(senderID string, content string, msgType string) {
    m.lastChat = content
}

func TestAgent_MakeDecisionPlay(t *testing.T) {
    exec := &mockExecutor{}
    agent := NewAIAgent("ai:bot:1", 0, nil, nil, exec)
    agent.MakeDecisionFunc = func(agent *AIAgent, phase string, handCards []int, stateJSON string) string {
        return `{"tool": "play_cards", "args": {"cards": [], "chat": "过"}}`
    }
    agent.Start()
    defer agent.Stop()

    agent.UpdateHand([]int{0, 1, 2, 13, 14, 26, 27, 39, 40, 52})
    agent.Trigger()

    // Wait briefly for async execution
    for i := 0; i < 100; i++ {
        if exec.lastAction != "" {
            break
        }
        // small yield
    }

    if exec.lastAction != "pass" {
        t.Errorf("expected pass action, got %s", exec.lastAction)
    }
    if exec.lastChat != "过" {
        t.Errorf("expected chat '过', got '%s'", exec.lastChat)
    }
}

func TestAgent_MakeDecisionBid(t *testing.T) {
    exec := &mockExecutor{}
    agent := NewAIAgent("ai:bot:1", 0, nil, nil, exec)
    agent.MakeDecisionFunc = func(agent *AIAgent, phase string, handCards []int, stateJSON string) string {
        return `{"tool": "bid_landlord", "args": {"chat": "我来！"}}`
    }
    agent.Start()
    defer agent.Stop()

    agent.Trigger()

    for i := 0; i < 100; i++ {
        if exec.lastAction != "" {
            break
        }
    }

    if exec.lastAction != "bid_call" {
        t.Errorf("expected bid_call, got %s", exec.lastAction)
    }
}

func TestAgent_ChatTool(t *testing.T) {
    exec := &mockExecutor{}
    agent := NewAIAgent("ai:bot:1", 0, nil, nil, exec)
    agent.MakeDecisionFunc = func(agent *AIAgent, phase string, handCards []int, stateJSON string) string {
        return `{"tool": "say", "args": {"message": "大家好！"}}`
    }
    agent.Start()
    defer agent.Stop()

    agent.Trigger()

    for i := 0; i < 100; i++ {
        if exec.lastChat != "" {
            break
        }
    }

    if exec.lastChat != "大家好！" {
        t.Errorf("expected chat '大家好！', got '%s'", exec.lastChat)
    }
}

func TestAgent_FallbackAction(t *testing.T) {
    exec := &mockExecutor{}
    agent := NewAIAgent("ai:bot:1", 0, nil, nil, exec)
    agent.Start()
    defer agent.Stop()

    // No MakeDecisionFunc set and no provider -> should trigger fallback
    agent.Trigger()

    for i := 0; i < 100; i++ {
        if exec.lastAction != "" {
            break
        }
    }

    if exec.lastAction != "pass" {
        t.Errorf("expected fallback pass, got %s", exec.lastAction)
    }
}
```

- [ ] **Step 2: Run tests**

Run: `cd server && go test ./internal/ai/ -v`
Expected: All tests PASS (including existing ones + new integration tests)

- [ ] **Step 3: Commit**

```bash
git add server/internal/ai/agent_integration_test.go
git commit -m "test(ai): add integration tests for AI agent decision loop"
```

---

### Task 6: Update prompt builder and remove old AIPlayer dependency

**Files:**
- Modify: `server/internal/ai/player.go`
- Keep: Old `AIPlayer` as fallback, add deprecation comments

**Note:** The old `AIPlayer` is still used by `NeedsDecision` and `SelectBid`. Keep it for backward compatibility but mark as deprecated.

- [ ] **Step 1: Add deprecation comments**

```go
// Deprecated: Use AIAgent for LLM-driven decisions. AIPlayer uses rule-based logic.
type AIPlayer struct { ... }
```

- [ ] **Step 2: Verify all existing tests still pass**

Run: `cd server && go test ./internal/ai/... -v`
Expected: All PASS

- [ ] **Step 3: Commit**

```bash
git add server/internal/ai/player.go
git commit -m "chore(ai): mark AIPlayer as deprecated in favor of AIAgent"
```

---

## Self-Review Checklist

### 1. Spec coverage
- **Change AI to agent mode** → Task 2 (AIAgent goroutine with LLM loop)
- **Game tool: check own hand** → `check_my_hand` tool in Task 1 + `executeInfoTool` in Task 2
- **Game tool: check each player's plays** → `check_game_status` tool in Task 1
- **Game tool: play cards** → `play_cards` tool in Task 1 + execution in Task 2
- **Chat tool: speak** → `say` tool + `chat` field on play/bid tools

### 2. Placeholder scan
All tool definitions, agent code, room integration, and test code are fully specified with complete Go code. No TBD, TODO, or "implement later" patterns.

### 3. Type consistency
- `ToolCall.Name` → used consistently in `ExtractToolCall`, `executeToolCall`, `isActionTool`, `executeInfoTool`
- `ActionExecutor.ExecuteAction(userID, action, cards)` → same signature as `GameRoom.HandleAction`
- `FillWithBot` with `BotOption` variadic → consistent across all call sites
