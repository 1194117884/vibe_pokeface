# AI Agent Game Flow — Test Plan

## Overview

Verifies the AI agent (`server/internal/ai/agent.go`) correctly handles all 5 Dou Di Zhu game phases: calling (叫地主), snatching (抢地主), revealing (明牌), doubling (加倍), and playing (出牌).

## Test Scenarios

### 1. Phase Detection

| Test | Input Phase | Expected Output |
|------|-------------|-----------------|
| detectPhase with phase=0 | `stateJSON` with `{"phase":0}` | `"calling"` |
| detectPhase with phase=1 | `stateJSON` with `{"phase":1}` | `"snatching"` |
| detectPhase with phase=2 | `stateJSON` with `{"phase":2}` | `"revealing"` |
| detectPhase with phase=3 | `stateJSON` with `{"phase":3}` | `"doubling"` |
| detectPhase with phase=4 | `stateJSON` with `{"phase":4}` | `"playing"` |
| detectPhase empty state | `stateJSON: ""` | `"calling"` |

### 2. Prompt Construction Per Phase

| Phase | Expected Prompt Content |
|-------|------------------------|
| calling | "当前阶段：叫地主" + bid_landlord/pass_bid tools |
| snatching | "当前阶段：抢地主" + bid_landlord/pass_bid tools |
| revealing | "当前阶段：明牌" + reveal_cards/pass_reveal tools |
| doubling | "当前阶段：加倍" + choose_double/choose_no_double tools |
| playing | "当前阶段：出牌" + play_cards tool |

### 3. User Message Context

For a landlord AI with state containing bid history and opponent info:
- Shows "地主牌（底牌）" with formatted cards
- Shows "叫地主/抢地主记录" with seat-level actions
- Shows "各玩家剩余手牌" with per-seat card counts
- Shows "当前倍率" multiplier
- Shows last play with type name (e.g., "单张", "对子")

### 4. Tool Execution

| LLM Tool | Engine Action | Phase |
|----------|--------------|-------|
| `bid_landlord` | `bid_call` | calling/snatching |
| `pass_bid` | `bid_pass` | calling/snatching |
| `reveal_cards` | `reveal_all` | revealing |
| `pass_reveal` | `pass` | revealing |
| `choose_double` | `double` | doubling |
| `choose_no_double` | `no_double` | doubling |
| `play_cards` | `play`/`pass` | playing |

### 5. Fallback Behavior

| Phase | fallbackAction | ruleBasedAction |
|-------|---------------|-----------------|
| calling | `bid_pass` | `ruleBasedBid()` |
| snatching | `bid_pass` | `ruleBasedBid()` |
| revealing | `pass` | `pass` |
| doubling | `no_double` | `no_double` |
| playing | `pass` | `ruleBasedPlay()` |

### 6. Rule-Based Play

- Free play: leads with lowest single card
- Not free play with Single: tries higher single
- Not free play with Pair: tries higher pair
- Not free play with other: passes

## Verification Steps

### Server-side

```bash
cd server
go vet ./internal/ai/...
go test ./internal/ai/... -v -count=1
```

### End-to-End (browser)

1. Navigate to http://localhost:3000
2. Login or register
3. Create a doudizhu room
4. Add 2 AI bots
5. Click "准备" then "开始游戏"
6. Observe phase transitions: 叫地主 → 抢地主 → 明牌 → 加倍 → 出牌
7. Verify AI bots take turns without game hanging
8. Verify 底牌 displayed to all players
9. Verify player card counts shown in UI

### Key Observables

- Game does NOT hang during any phase transition
- AI bots respond within 1-2 seconds (rule-based) or 5-15 seconds (LLM)
- 底牌 visible after landlord finalized
- Score multiplier updates after snatching
- Play history visible during playing phase

## Current Status

All 47 existing tests pass. Browser E2E test confirmed: calling → snatching → revealing → doubling → playing transition works correctly with the new phase detection.
