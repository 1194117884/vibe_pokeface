package ai

import (
	"encoding/json"
	"fmt"
)

// ToolCall represents an LLM decision to call a tool
type ToolCall struct {
	Name string          `json:"tool"`
	Args json.RawMessage `json:"args"`
}

// ToolSchema defines an LLM-callable tool (OpenAI function-calling format)
type ToolSchema struct {
	Type     string  `json:"type"`
	Function FuncDef `json:"function"`
}

type FuncDef struct {
	Name        string      `json:"name"`
	Description string      `json:"description"`
	Parameters  ParamSchema `json:"parameters"`
}

type ParamSchema struct {
	Type       string                    `json:"type"`
	Properties map[string]ParamProperty `json:"properties,omitempty"`
	Required   []string                  `json:"required,omitempty"`
}

type ParamProperty struct {
	Type        string        `json:"type"`
	Description string        `json:"description"`
	Enum        []string      `json:"enum,omitempty"`
	Items       *ItemsSchema  `json:"items,omitempty"`
}

// ItemsSchema is used for array-type properties in tool schemas
type ItemsSchema struct {
	Type string `json:"type"`
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

// GetToolSchemas returns tool definitions appropriate for the given game phase.
// Bidding phases get bidding tools; playing phase gets play_cards.
// Info tools (check_my_hand, check_game_status) and say are always available.
func GetToolSchemas(phase string) []ToolSchema {
	infoTools := []ToolSchema{
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
	}

	sayTool := ToolSchema{
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
	}

	biddingTools := []ToolSchema{
		{
			Type: "function",
			Function: FuncDef{
				Name:        "bid_landlord",
				Description: "叫地主/抢地主",
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
				Description: "不叫地主/不抢地主",
				Parameters: ParamSchema{
					Type: "object",
					Properties: map[string]ParamProperty{
						"chat": {
							Type:        "string",
							Description: "不叫时说的话（可选，不超过30字）",
						},
					},
				},
			},
		},
	}

	revealTools := []ToolSchema{
		{
			Type: "function",
			Function: FuncDef{
				Name:        "reveal_cards",
				Description: "明牌（亮出自己的手牌，倍数翻倍）",
				Parameters: ParamSchema{
					Type:       "object",
					Properties: map[string]ParamProperty{},
				},
			},
		},
		{
			Type: "function",
			Function: FuncDef{
				Name:        "pass_reveal",
				Description: "不明牌",
				Parameters: ParamSchema{
					Type:       "object",
					Properties: map[string]ParamProperty{},
				},
			},
		},
	}

	doubleTools := []ToolSchema{
		{
			Type: "function",
			Function: FuncDef{
				Name:        "choose_double",
				Description: "加倍（得分翻倍，赢多输也多）",
				Parameters: ParamSchema{
					Type:       "object",
					Properties: map[string]ParamProperty{},
				},
			},
		},
		{
			Type: "function",
			Function: FuncDef{
				Name:        "choose_no_double",
				Description: "不加倍",
				Parameters: ParamSchema{
					Type:       "object",
					Properties: map[string]ParamProperty{},
				},
			},
		},
	}

	playTool := ToolSchema{
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
						Items:       &ItemsSchema{Type: "integer"},
					},
					"chat": {
						Type:        "string",
						Description: "出牌时说的话（可选，不超过30字）",
					},
				},
				Required: []string{"cards"},
			},
		},
	}

	base := append(infoTools, sayTool)

	switch phase {
	case "calling", "snatching":
		return append(base, biddingTools...)
	case "revealing":
		return append(base, revealTools...)
	case "doubling":
		return append(base, doubleTools...)
	default:
		return append(base, playTool)
	}
}
