package ai

import (
	"strings"
	"testing"
)

func TestDashengjiAgent_DetectPhaseAndTools(t *testing.T) {
	agent := NewAIAgent("ai:bot:1", 0, nil, nil, nil)
	agent.SetGameType("dashengji")
	agent.stateJSON = `{"phase":3}`

	if got := agent.detectPhase(); got != "discard_bottom" {
		t.Fatalf("phase = %s, want discard_bottom", got)
	}

	tools := GetToolSchemasForGame("dashengji", "discard_bottom")
	names := make(map[string]bool)
	for _, tool := range tools {
		names[tool.Function.Name] = true
	}
	if !names["discard_bottom"] {
		t.Fatal("discard_bottom tool is missing")
	}
	if names["bid_landlord"] {
		t.Fatal("doudizhu bid tool should not be exposed for dashengji")
	}
}

func TestDashengjiAgent_FallbackSetTrump(t *testing.T) {
	exec := &mockExecutor{}
	agent := NewAIAgent("ai:bot:1", 0, nil, nil, exec)
	agent.SetGameType("dashengji")
	agent.HandCards = []int{53, 25, 16, 1}
	agent.stateJSON = `{"phase":0,"level_rank":6}`

	agent.fallbackDashengjiAction("set_trump")

	if exec.lastAction != "set_trump" {
		t.Fatalf("action = %s, want set_trump", exec.lastAction)
	}
	if len(exec.lastCards) != 3 {
		t.Fatalf("cards = %v, want 3 trump cards", exec.lastCards)
	}
}

func TestDashengjiAgent_FallbackDiscardBottomChoosesSixCards(t *testing.T) {
	exec := &mockExecutor{}
	agent := NewAIAgent("ai:bot:1", 0, nil, nil, exec)
	agent.SetGameType("dashengji")
	agent.HandCards = []int{0, 1, 2, 3, 4, 5, 9, 10, 13, 16}
	agent.stateJSON = `{"phase":3,"trump_suit":1,"level_rank":6}`

	agent.fallbackDashengjiAction("discard_bottom")

	if exec.lastAction != "discard_bottom" {
		t.Fatalf("action = %s, want discard_bottom", exec.lastAction)
	}
	if len(exec.lastCards) != 6 {
		t.Fatalf("cards = %v, want 6 cards", exec.lastCards)
	}
	handSet := make(map[int]bool)
	for _, id := range agent.HandCards {
		handSet[id] = true
	}
	for _, id := range exec.lastCards {
		if !handSet[id] {
			t.Fatalf("discarded card %d is not in hand %v", id, agent.HandCards)
		}
	}
}

func TestDashengjiAgent_FallbackPlayFollowsLedSuitSingle(t *testing.T) {
	exec := &mockExecutor{}
	agent := NewAIAgent("ai:bot:1", 1, nil, nil, exec)
	agent.SetGameType("dashengji")
	agent.HandCards = []int{0, 14, 28}
	agent.stateJSON = `{
		"phase":4,
		"current_seat":1,
		"trump_suit":1,
		"level_rank":6,
		"round_plays":[{"seat":0,"play":{"type":1,"main_rank":13,"length":1},"cards":[{"id":10}]}]
	}`

	agent.fallbackDashengjiAction("playing")

	if exec.lastAction != "play" {
		t.Fatalf("action = %s, want play", exec.lastAction)
	}
	if len(exec.lastCards) != 1 || exec.lastCards[0] != 0 {
		t.Fatalf("cards = %v, want spade follow card [0]", exec.lastCards)
	}
}

func TestDashengjiAgent_PlayCardsSendsChat(t *testing.T) {
	exec := &mockExecutor{}
	agent := NewAIAgent("ai:bot:1", 1, nil, nil, exec)
	agent.SetGameType("dashengji")

	agent.executeToolCall(`{"tool":"play_cards","args":{"cards":[0],"chat":"这手先探路"}}`)

	if exec.lastAction != "play" {
		t.Fatalf("action = %s, want play", exec.lastAction)
	}
	if exec.lastChat != "这手先探路" {
		t.Fatalf("chat = %q, want 这手先探路", exec.lastChat)
	}
}

func TestDashengjiAgent_SystemPromptIncludesRulesAndStrategy(t *testing.T) {
	agent := NewAIAgent("ai:bot:1", 0, nil, nil, nil)
	agent.SetGameType("dashengji")

	got := agent.buildDashengjiSystemPrompt("playing")
	for _, want := range []string{
		"3副牌",
		"0-2、1-3",
		"主牌：大王、小王、所有2、所有级牌、主花色牌",
		"定主阶段",
		"反主阶段",
		"扣底阶段",
		"出牌阶段",
		"不能过牌",
		"必须先跟同花色/同主副类别",
		"队友当前最大时优先送分",
		"只能使用当前手牌里的整数ID",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("system prompt missing %q:\n%s", want, got)
		}
	}
}

func TestDashengjiAgent_UserMessageIncludesCurrentHandIDs(t *testing.T) {
	agent := NewAIAgent("ai:bot:1", 1, nil, nil, nil)
	agent.SetGameType("dashengji")
	agent.HandCards = []int{0, 14, 53}
	agent.stateJSON = `{
		"phase":4,
		"current_seat":1,
		"dealer_seats":[1,3],
		"level_rank":6,
		"trump_suit":1,
		"round_plays":[{"seat":0,"play":{"type":1,"main_rank":13,"length":1},"cards":[{"id":10}]}],
		"players":[{"seat":0,"hand":[{"id":10}]},{"seat":1,"hand":[{"id":0},{"id":14},{"id":53}]},{"seat":2,"hand":[]},{"seat":3,"hand":[]}]
	}`

	got := agent.buildDashengjiUserMessage("playing")
	for _, want := range []string{
		"你的当前手牌（3张）：0(♠3#1) 14(♥4#1) 53(大王#1)",
		"只能从上面的手牌ID中选择",
		"本轮领出座位0：10(♠K#1)",
		"本轮需要跟出张数：1",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("user message missing %q:\n%s", want, got)
		}
	}
}

func TestDashengjiAgent_RetryPromptWarnsAgainstRepeatedIllegalCards(t *testing.T) {
	exec := &mockExecutor{}
	provider := &mockToolProvider{
		responses: []*LLMResultWithTools{
			{ToolCalls: []AssistantToolCall{{ID: "call_1", Type: "function"}}},
		},
	}
	provider.responses[0].ToolCalls[0].Function.Name = "play_cards"
	provider.responses[0].ToolCalls[0].Function.Arguments = `{"cards":[0]}`

	agent := NewAIAgent("ai:bot:1", 0, nil, provider, exec)
	agent.SetGameType("dashengji")
	agent.HandCards = []int{1, 2, 3}
	agent.stateJSON = `{"phase":4,"current_seat":0,"players":[{"seat":0,"hand":[{"id":1},{"id":2},{"id":3}]}]}`
	agent.lastError = &GameError{Code: "INVALID_CARDS", Phase: "playing", Action: "play"}

	agent.makeDecisionWithTools()

	if len(provider.requests) != 1 {
		t.Fatalf("expected one provider request, got %d", len(provider.requests))
	}
	var combined strings.Builder
	for _, msg := range provider.requests[0] {
		combined.WriteString(msg.Content)
		combined.WriteString("\n")
	}
	got := combined.String()
	for _, want := range []string{"不要重复上一次非法", "只能从当前手牌ID中重新选择", "INVALID_CARDS"} {
		if !strings.Contains(got, want) {
			t.Fatalf("retry prompt missing %q:\n%s", want, got)
		}
	}
}

func TestDashengjiAgent_CandidatesExcludeLevelCardsWhenFollowingSideSuit(t *testing.T) {
	agent := NewAIAgent("ai:bot:2", 1, nil, nil, nil)
	agent.SetGameType("dashengji")
	agent.HandCards = []int{13, 121, 22, 75, 73, 17}
	agent.stateJSON = `{
		"phase":4,
		"current_seat":1,
		"level_rank":3,
		"trump_suit":2,
		"round_plays":[{"seat":3,"play":{"type":2,"main_rank":8,"length":1},"cards":[{"id":72},{"id":126}]}],
		"players":[{"seat":1,"hand":[{"id":13},{"id":121},{"id":22},{"id":75},{"id":73},{"id":17}]}]
	}`

	candidates := agent.dashengjiCandidatesForPhase("playing")
	if len(candidates) != 1 {
		t.Fatalf("candidates = %#v, want exactly one padding candidate", candidates)
	}
	if got := candidates[0].Cards; len(got) != 2 || got[0] != 17 || got[1] != 73 {
		t.Fatalf("candidate cards = %v, want [17 73]", got)
	}
	if strings.Contains(agent.buildDashengjiCandidateHints("playing"), "13(♥3#1) 121(♥3#3)") {
		t.Fatalf("candidate hints should not offer level-card pair as side-suit follow:\n%s", agent.buildDashengjiCandidateHints("playing"))
	}
}

func TestDashengjiAgent_PrevalidationFallsBackWhenLLMChoosesNonCandidate(t *testing.T) {
	exec := &mockExecutor{}
	agent := NewAIAgent("ai:bot:2", 1, nil, nil, exec)
	agent.SetGameType("dashengji")
	agent.HandCards = []int{13, 121, 22, 75, 73, 17}
	agent.stateJSON = `{
		"phase":4,
		"current_seat":1,
		"level_rank":3,
		"trump_suit":2,
		"round_plays":[{"seat":3,"play":{"type":2,"main_rank":8,"length":1},"cards":[{"id":72},{"id":126}]}],
		"players":[{"seat":1,"hand":[{"id":13},{"id":121},{"id":22},{"id":75},{"id":73},{"id":17}]}]
	}`

	agent.executeToolCall(`{"tool":"play_cards","args":{"cards":[13,121],"chat":"跟红桃"}}`)

	if exec.lastAction != "play" {
		t.Fatalf("action = %s, want play", exec.lastAction)
	}
	if len(exec.lastCards) != 2 || exec.lastCards[0] != 17 || exec.lastCards[1] != 73 {
		t.Fatalf("cards = %v, want fallback [17 73]", exec.lastCards)
	}
	if exec.lastChat != "" {
		t.Fatalf("chat = %q, want no original LLM chat for fallback action", exec.lastChat)
	}
}

func TestDashengjiAgent_PrevalidationRejectsMissingCurrentPlayer(t *testing.T) {
	agent := NewAIAgent("ai:bot:2", 1, nil, nil, nil)
	agent.SetGameType("dashengji")
	agent.HandCards = []int{13, 121}
	agent.stateJSON = `{
		"phase":4,
		"current_seat":1,
		"level_rank":3,
		"trump_suit":2,
		"round_plays":[{"seat":3,"play":{"type":2,"main_rank":8,"length":1},"cards":[{"id":72},{"id":126}]}],
		"players":[{"seat":0,"hand":[]},{"seat":2,"hand":[]},{"seat":3,"hand":[]}]
	}`

	if agent.isLegalDashengjiAction("play_cards", []int{13, 121}) {
		t.Fatal("missing current player should not validate cards against the proposed play itself")
	}
}

func TestDashengjiAgent_PrevalidationAllowsFlexibleLegalLead(t *testing.T) {
	exec := &mockExecutor{}
	agent := NewAIAgent("ai:bot:1", 0, nil, nil, exec)
	agent.SetGameType("dashengji")
	agent.HandCards = []int{0, 1, 2}
	agent.stateJSON = `{
		"phase":4,
		"current_seat":0,
		"level_rank":6,
		"trump_suit":2,
		"players":[{"seat":0,"hand":[{"id":0},{"id":1},{"id":2}]}]
	}`

	agent.executeToolCall(`{"tool":"play_cards","args":{"cards":[2],"chat":"换个节奏"}}`)

	if exec.lastAction != "play" {
		t.Fatalf("action = %s, want play", exec.lastAction)
	}
	if len(exec.lastCards) != 1 || exec.lastCards[0] != 2 {
		t.Fatalf("cards = %v, want original legal lead [2]", exec.lastCards)
	}
}

func TestDashengjiAgent_PrevalidationAllowsFlexiblePaddingWhenCannotFollowPair(t *testing.T) {
	exec := &mockExecutor{}
	agent := NewAIAgent("ai:bot:2", 1, nil, nil, exec)
	agent.SetGameType("dashengji")
	agent.HandCards = []int{22, 75, 73, 17}
	agent.stateJSON = `{
		"phase":4,
		"current_seat":1,
		"level_rank":3,
		"trump_suit":2,
		"round_plays":[{"seat":3,"play":{"type":2,"main_rank":8,"length":1},"cards":[{"id":72},{"id":126}]}],
		"players":[{"seat":1,"hand":[{"id":22},{"id":75},{"id":73},{"id":17}]}]
	}`

	agent.executeToolCall(`{"tool":"play_cards","args":{"cards":[22,75],"chat":"贴两张"}}`)

	if exec.lastAction != "play" {
		t.Fatalf("action = %s, want play", exec.lastAction)
	}
	if len(exec.lastCards) != 2 || exec.lastCards[0] != 22 || exec.lastCards[1] != 75 {
		t.Fatalf("cards = %v, want original legal padding [22 75]", exec.lastCards)
	}
}

func TestDashengjiAgent_PrevalidationAllowsPaddingWithoutBreakingPairForTriple(t *testing.T) {
	exec := &mockExecutor{}
	agent := NewAIAgent("ai:bot:2", 1, nil, nil, exec)
	agent.SetGameType("dashengji")
	agent.HandCards = []int{19, 73, 22, 75, 17}
	agent.stateJSON = `{
		"phase":4,
		"current_seat":1,
		"level_rank":3,
		"trump_suit":2,
		"round_plays":[{"seat":3,"play":{"type":3,"main_rank":8,"length":1},"cards":[{"id":72},{"id":126},{"id":18}]}],
		"players":[{"seat":1,"hand":[{"id":19},{"id":73},{"id":22},{"id":75},{"id":17}]}]
	}`

	agent.executeToolCall(`{"tool":"play_cards","args":{"cards":[22,75,17],"chat":"不拆对子"}}`)

	if exec.lastAction != "play" {
		t.Fatalf("action = %s, want play", exec.lastAction)
	}
	if len(exec.lastCards) != 3 || exec.lastCards[0] != 22 || exec.lastCards[1] != 75 || exec.lastCards[2] != 17 {
		t.Fatalf("cards = %v, want original legal padding [22 75 17]", exec.lastCards)
	}
}

func TestDashengjiAgent_FollowCandidatesFeedPointsToWinningTeammate(t *testing.T) {
	agent := NewAIAgent("ai:bot:1", 1, nil, nil, nil)
	agent.SetGameType("dashengji")
	agent.HandCards = []int{27, 28, 33}
	agent.stateJSON = `{
		"phase":4,
		"current_seat":1,
		"dealer_seats":[0,2],
		"level_rank":6,
		"trump_suit":1,
		"round_plays":[
			{"seat":3,"play":{"type":1,"main_rank":14,"length":1},"cards":[{"id":37}]},
			{"seat":0,"play":{"type":1,"main_rank":4,"length":1},"cards":[{"id":27}]}
		],
		"players":[{"seat":1,"hand":[{"id":27},{"id":28},{"id":33}]}]
	}`

	candidates := agent.dashengjiCandidatesForPhase("playing")
	if len(candidates) == 0 {
		t.Fatal("expected candidates")
	}
	if got := candidates[0].Cards; len(got) != 1 || got[0] != 33 {
		t.Fatalf("first candidate = %v, want point card [33] for teammate", got)
	}
	if !strings.Contains(candidates[0].Reason, "队友当前最大") {
		t.Fatalf("reason = %q, want teammate-winning explanation", candidates[0].Reason)
	}
}

func TestDashengjiAgent_FollowCandidatesAvoidPointsWhenOpponentWinning(t *testing.T) {
	agent := NewAIAgent("ai:bot:1", 1, nil, nil, nil)
	agent.SetGameType("dashengji")
	agent.HandCards = []int{27, 33}
	agent.stateJSON = `{
		"phase":4,
		"current_seat":1,
		"dealer_seats":[0,2],
		"level_rank":6,
		"trump_suit":1,
		"round_plays":[{"seat":0,"play":{"type":1,"main_rank":14,"length":1},"cards":[{"id":37}]}],
		"players":[{"seat":1,"hand":[{"id":27},{"id":33}]}]
	}`

	candidates := agent.dashengjiCandidatesForPhase("playing")
	if len(candidates) == 0 {
		t.Fatal("expected candidates")
	}
	if got := candidates[0].Cards; len(got) != 1 || got[0] != 27 {
		t.Fatalf("first candidate = %v, want low non-point [27]", got)
	}
	if !strings.Contains(candidates[0].Reason, "对手当前最大") {
		t.Fatalf("reason = %q, want opponent-winning explanation", candidates[0].Reason)
	}
}

func TestDashengjiAgent_FollowCandidatesTrumpPairToCapturePoints(t *testing.T) {
	agent := NewAIAgent("ai:bot:1", 1, nil, nil, nil)
	agent.SetGameType("dashengji")
	agent.HandCards = []int{72, 74, 128, 21, 75}
	agent.stateJSON = `{
		"phase":4,
		"current_seat":1,
		"dealer_seats":[0,2],
		"round_points":80,
		"level_rank":3,
		"trump_suit":1,
		"round_plays":[
			{"seat":3,"play":{"type":2,"main_rank":10,"length":1},"cards":[{"id":61},{"id":115}]},
			{"seat":0,"play":{"type":2,"main_rank":105,"length":1},"cards":[{"id":15},{"id":123}]}
		],
		"players":[{"seat":1,"hand":[{"id":72},{"id":74},{"id":128},{"id":21},{"id":75}]}]
	}`

	candidates := agent.dashengjiCandidatesForPhase("playing")
	if len(candidates) == 0 {
		t.Fatal("expected candidates")
	}
	if got := candidates[0].Cards; len(got) != 2 || got[0] != 74 || got[1] != 128 {
		t.Fatalf("first candidate = %v, want trumping heart 10 pair [74 128]", got)
	}
	if !strings.Contains(candidates[0].Reason, "抢回本轮") {
		t.Fatalf("reason = %q, want capture-points explanation", candidates[0].Reason)
	}
}

func TestDashengjiAgent_LeadCandidatesDowngradePairKWhenPairAUnseen(t *testing.T) {
	agent := NewAIAgent("ai:bot:1", 1, nil, nil, nil)
	agent.SetGameType("dashengji")
	agent.HandCards = []int{27, 36, 90}
	agent.stateJSON = `{
		"phase":4,
		"current_seat":1,
		"dealer_seats":[0,2],
		"level_rank":6,
		"trump_suit":1,
		"players":[{"seat":1,"hand":[{"id":27},{"id":36},{"id":90}]}],
		"play_history":[]
	}`

	hints := agent.buildDashengjiCandidateHints("playing")
	if !strings.Contains(hints, "对A未出") {
		t.Fatalf("candidate hints should mention unseen pair A risk:\n%s", hints)
	}
	candidates := agent.dashengjiCandidatesForPhase("playing")
	if len(candidates) == 0 {
		t.Fatal("expected candidates")
	}
	if len(candidates[0].Cards) == 2 {
		t.Fatalf("risky pair K should not be first candidate: %#v", candidates)
	}
}

func TestDashengjiAgent_UserMessageIncludesTrickAndKeyCardSummaries(t *testing.T) {
	agent := NewAIAgent("ai:bot:1", 1, nil, nil, nil)
	agent.SetGameType("dashengji")
	agent.HandCards = []int{27, 33}
	agent.stateJSON = `{
		"phase":4,
		"current_seat":1,
		"dealer_seats":[0,2],
		"level_rank":6,
		"trump_suit":1,
		"round_plays":[{"seat":0,"play":{"type":1,"main_rank":14,"length":1},"cards":[{"id":37}]}],
		"players":[{"seat":1,"hand":[{"id":27},{"id":33}]}]
	}`

	got := agent.buildDashengjiUserMessage("playing")
	for _, want := range []string{"本轮局势摘要", "桌面已有", "关键牌摘要", "对A未完全打出"} {
		if !strings.Contains(got, want) {
			t.Fatalf("user message missing %q:\n%s", want, got)
		}
	}
}

func TestDashengjiAgent_SystemPromptIncludesDecisionOrder(t *testing.T) {
	agent := NewAIAgent("ai:bot:1", 0, nil, nil, nil)
	agent.SetGameType("dashengji")

	got := agent.buildDashengjiSystemPrompt("playing")
	for _, want := range []string{"出牌决策顺序", "所有级牌和所有2都是主牌", "对A未出时不要贸然领出对K"} {
		if !strings.Contains(got, want) {
			t.Fatalf("system prompt missing %q:\n%s", want, got)
		}
	}
}
