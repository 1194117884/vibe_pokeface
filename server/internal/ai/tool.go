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
	Type       string                   `json:"type"`
	Properties map[string]ParamProperty `json:"properties,omitempty"`
	Required   []string                 `json:"required,omitempty"`
}

type ParamProperty struct {
	Type        string       `json:"type"`
	Description string       `json:"description"`
	Enum        []string     `json:"enum,omitempty"`
	Items       *ItemsSchema `json:"items,omitempty"`
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
// Info tools (check_my_hand, check_game_status) are always available.
func GetToolSchemas(phase string) []ToolSchema {
	return GetToolSchemasForGame("doudizhu", phase)
}

// GetToolSchemasForGame returns tool definitions for the requested game and phase.
func GetToolSchemasForGame(gameType string, phase string) []ToolSchema {
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
		{
			Type: "function",
			Function: FuncDef{
				Name:        "check_playing_records",
				Description: "查看本轮出牌记录：谁出了什么牌、谁过了牌、出牌的顺序",
				Parameters: ParamSchema{
					Type:       "object",
					Properties: map[string]ParamProperty{},
				},
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

	if gameType == "dashengji" {
		return getDashengjiToolSchemas(infoTools, phase)
	}

	base := infoTools
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

func getDashengjiToolSchemas(base []ToolSchema, phase string) []ToolSchema {
	cardArgs := ParamSchema{
		Type: "object",
		Properties: map[string]ParamProperty{
			"cards": {
				Type:        "array",
				Description: "要使用的牌ID数组",
				Items:       &ItemsSchema{Type: "integer"},
			},
			"chat": {
				Type:        "string",
				Description: "可选聊天内容，不超过30字",
			},
		},
		Required: []string{"cards"},
	}
	emptyArgs := ParamSchema{Type: "object", Properties: map[string]ParamProperty{}}

	switch phase {
	case "set_trump":
		return append(base,
			ToolSchema{Type: "function", Function: FuncDef{Name: "set_trump", Description: "定主，传入符合规则的亮主牌ID", Parameters: cardArgs}},
			ToolSchema{Type: "function", Function: FuncDef{Name: "pass_trump", Description: "不定主", Parameters: emptyArgs}},
		)
	case "counter_trump":
		return append(base,
			ToolSchema{Type: "function", Function: FuncDef{Name: "counter_trump", Description: "反主，传入符合规则的反主牌ID", Parameters: cardArgs}},
			ToolSchema{Type: "function", Function: FuncDef{Name: "pass_counter", Description: "不反主", Parameters: emptyArgs}},
		)
	case "take_bottom":
		return append(base,
			ToolSchema{Type: "function", Function: FuncDef{Name: "take_bottom", Description: "自己起底", Parameters: emptyArgs}},
			ToolSchema{Type: "function", Function: FuncDef{Name: "pass_take_bottom", Description: "把起底交给队友", Parameters: emptyArgs}},
		)
	case "discard_bottom":
		return append(base,
			ToolSchema{Type: "function", Function: FuncDef{Name: "discard_bottom", Description: "扣底，必须传入6张牌ID", Parameters: cardArgs}},
		)
	default:
		return append(base,
			ToolSchema{Type: "function", Function: FuncDef{Name: "play_cards", Description: "大升级出牌，传入要出的牌ID数组", Parameters: cardArgs}},
		)
	}
}
