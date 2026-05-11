package ai

import (
    "encoding/json"
    "fmt"
    "strings"
)

const systemPrompt = `你是一个斗地主AI玩家。根据你的手牌和当前牌局状态，选择最佳出牌策略。
你必须严格遵守斗地主规则：
- 可以出单张、对子、三张、三带一、三带二、顺子(5+)、连对(3+)、飞机、炸弹、火箭
- 上一轮出牌后，你必须出比上家更大的牌型，或者选择过(不出)
- 如果你是地主，你的目标是先出完所有牌
- 如果你是农民，你的目标是配合队友阻止地主出完

请以JSON格式回复：{"action": "play"|"pass", "cards": [card_ids]}`

const bidSystemPrompt = `你是一个斗地主AI玩家。根据你的手牌质量，决定是否叫地主。
牌力评估：有2+个2、有炸弹、有大王/小王时应该叫地主。
回复JSON格式：{"action": "bid_call"|"bid_pass"}`

type PromptBuilder struct{}

func NewPromptBuilder() *PromptBuilder {
    return &PromptBuilder{}
}

func (pb *PromptBuilder) BuildPlayDecision(handSize int, myHand []int, lastPlay []int) string {
    var sb strings.Builder
    sb.WriteString(fmt.Sprintf("当前你的手牌(%d张): ", handSize))
    for _, c := range myHand {
        sb.WriteString(fmt.Sprintf("%d ", c))
    }
    sb.WriteString("\n")
    if len(lastPlay) > 0 {
        sb.WriteString(fmt.Sprintf("上家出牌: %v\n", lastPlay))
    } else {
        sb.WriteString("你是本轮第一个出牌\n")
    }
    sb.WriteString("请选择出牌或过牌。")
    return sb.String()
}

func (pb *PromptBuilder) BuildBidDecision(handCards []int, handSize int) string {
    return fmt.Sprintf("你的手牌(%d张): %v\n请选择叫地主或过牌。", handSize, handCards)
}

func (pb *PromptBuilder) BuildChatMessage(characterName string, personality string, handCards []int) string {
    return fmt.Sprintf("你是%s，%s。根据当前牌局情况，说一句简短有趣的话（不超过20字）。", characterName, personality)
}

// stripCodeFences removes markdown code fence markers from LLM output.
// Many LLMs wrap JSON in ```json ... ``` fences.
func stripCodeFences(s string) string {
    s = strings.TrimSpace(s)
    s = strings.TrimPrefix(s, "```json")
    s = strings.TrimPrefix(s, "```")
    s = strings.TrimSuffix(s, "```")
    s = strings.TrimSpace(s)
    return s
}

func ParseAIPlayResponse(jsonStr string) (action string, cards []int, err error) {
    var resp struct {
        Action string `json:"action"`
        Cards  []int  `json:"cards"`
    }
    if err := json.Unmarshal([]byte(stripCodeFences(jsonStr)), &resp); err != nil {
        return "", nil, err
    }
    return resp.Action, resp.Cards, nil
}

func ParseAIBidResponse(jsonStr string) (string, error) {
    var resp struct {
        Action string `json:"action"`
    }
    if err := json.Unmarshal([]byte(stripCodeFences(jsonStr)), &resp); err != nil {
        return "", err
    }
    return resp.Action, nil
}

func ParseAIChatResponse(jsonStr string) (string, error) {
    var resp struct {
        Message string `json:"message"`
    }
    if err := json.Unmarshal([]byte(stripCodeFences(jsonStr)), &resp); err != nil {
        return "", err
    }
    return resp.Message, nil
}
