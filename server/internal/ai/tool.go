package ai

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
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

// ExtractToolCallFlexible also handles prose such as:
// "【我的决策】调用play_cards，参数：{\"cards\":[1],\"chat\":\"...\"}".
func ExtractToolCallFlexible(text string) (*ToolCall, error) {
	if call, err := ExtractToolCall(text); err == nil {
		return call, nil
	}
	toolName := ""
	toolIndex := -1
	for _, name := range []string{
		"set_trump", "pass_trump", "counter_trump", "pass_counter",
		"take_bottom", "pass_take_bottom", "discard_bottom", "play_cards",
		"bid_landlord", "pass_bid", "reveal_cards", "pass_reveal",
		"choose_double", "choose_no_double",
	} {
		re := regexp.MustCompile(`(^|[^A-Za-z0-9_])` + regexp.QuoteMeta(name) + `([^A-Za-z0-9_]|$)`)
		matches := re.FindAllStringIndex(text, -1)
		if len(matches) == 0 {
			continue
		}
		idx := matches[len(matches)-1][0]
		if idx > toolIndex {
			toolName = name
			toolIndex = idx
		}
	}
	if toolName == "" {
		return nil, fmt.Errorf("no tool name in text")
	}
	args, ok := firstJSONObjectAfter(text, "参数")
	if !ok {
		args, ok = firstJSONObjectAfter(text, toolName)
	}
	if !ok {
		args = "{}"
	}
	var raw json.RawMessage = []byte(args)
	if !json.Valid(raw) {
		return nil, fmt.Errorf("invalid tool args in text")
	}
	return &ToolCall{Name: toolName, Args: raw}, nil
}

func firstJSONObjectAfter(text string, marker string) (string, bool) {
	startAt := 0
	if marker != "" {
		if idx := strings.Index(text, marker); idx >= 0 {
			startAt = idx + len(marker)
		}
	}
	start := strings.Index(text[startAt:], "{")
	if start < 0 {
		return "", false
	}
	start += startAt
	depth := 0
	inString := false
	escaped := false
	for i := start; i < len(text); i++ {
		ch := text[i]
		if inString {
			if escaped {
				escaped = false
				continue
			}
			if ch == '\\' {
				escaped = true
				continue
			}
			if ch == '"' {
				inString = false
			}
			continue
		}
		if ch == '"' {
			inString = true
			continue
		}
		if ch == '{' {
			depth++
		}
		if ch == '}' {
			depth--
			if depth == 0 {
				return text[start : i+1], true
			}
		}
	}
	return "", false
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
			ToolSchema{Type: "function", Function: FuncDef{Name: "set_trump", Description: "定主，传入当前手牌中的合法亮主牌ID：王 + 同色2 + 同花色级牌；两张同花色级牌为定死。", Parameters: cardArgs}},
			ToolSchema{Type: "function", Function: FuncDef{Name: "pass_trump", Description: "不定主", Parameters: emptyArgs}},
		)
	case "counter_trump":
		return append(base,
			ToolSchema{Type: "function", Function: FuncDef{Name: "counter_trump", Description: "反主，传入当前手牌中的合法反主牌ID：王 + 同色2 + 两张同花色级牌；单张级牌不能反主。", Parameters: cardArgs}},
			ToolSchema{Type: "function", Function: FuncDef{Name: "pass_counter", Description: "不反主", Parameters: emptyArgs}},
		)
	case "take_bottom":
		return append(base,
			ToolSchema{Type: "function", Function: FuncDef{Name: "take_bottom", Description: "自己起底", Parameters: emptyArgs}},
			ToolSchema{Type: "function", Function: FuncDef{Name: "pass_take_bottom", Description: "把起底交给队友", Parameters: emptyArgs}},
		)
	case "discard_bottom":
		return append(base,
			ToolSchema{Type: "function", Function: FuncDef{Name: "discard_bottom", Description: "扣底，必须传入当前手牌中的正好6张牌ID；优先扣低价值副牌，避免扣分牌和关键主牌。", Parameters: cardArgs}},
		)
	default:
		return append(base,
			ToolSchema{Type: "function", Function: FuncDef{Name: "play_cards", Description: "打升级出牌，cards必须非空且全部来自当前手牌；跟牌必须先满足领出张数、牌型、花色和主副类别。", Parameters: cardArgs}},
		)
	}
}
