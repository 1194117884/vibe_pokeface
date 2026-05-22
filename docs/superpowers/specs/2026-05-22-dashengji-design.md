# Dashengji (打升级) Game Mode — Design Spec

**Date:** 2026-05-22
**Status:** Approved

## Overview

Add a new 4-player, 3-deck, team-based climbing card game (打升级/双升) to the Pokeface platform. Hebei Dingzhou regional variant.

## Key Decisions

| Decision | Choice |
|----------|--------|
| Sequencing | Engine-first TDD, then UI, defer AI to v2 |
| Architecture | Mirror doudizhu pattern (Approach A) — separate engine package, separate frontend page |
| Frontend routing | Separate route paths: `/room/[id]/doudizhu` and `/room/[id]/dashengji` |
| Rule scope | Full rules implementation from `dashengji_doc.md` |
| AI | Deferred to v2 |

## Game Rules Summary

See `dashengji_doc.md` for complete rules. Key mechanics:

- **3 decks** (162 cards), 4 players (2v2 teams), 39 cards/player + 6 bottom cards
- **Level progression**: 3 → 4 → ... → A, first team to finish A wins
- **Dynamic trump**: Trump suit and level card change each round
- **Follow constraints**: Must match suit + hand type + main/side category
- **Point-based**: 300 points total (5s=5, 10s=10, Ks=10), scoring brackets determine level changes
- **Hand types**: Singles, pairs, triples (刻子), tractors (拖拉机, ≥3 consecutive pairs same suit)
- **Card ranking**: 副牌 < 主花色牌 < 副2 < 本2 < 副级牌 < 本级牌 < 小王 < 大王

## Server Design

### Package: `server/internal/game/dashengji/`

```
engine.go     — GameEngine interface implementation
state.go      — GameState, GamePhase enum, phase-action whitelist
card.go       — 162-card deck (3×54), deal, shuffle, rank calculation
hand.go       — ParsePlay, CanBeat, tractor detection, follow-suit validation
trump.go      — Trump eligibility (定主/反主), card classification (本主/副主/主牌/副牌)
score.go      — Point tracking, level progression, dealer swap logic
```

### Card Encoding

- IDs 0–161: `face = ID % 54`, `copy = ID / 54`
- Faces 0–51: standard cards (suit×13 + rank), 52: small joker, 53: big joker

### State Machine

```
Init → SetTrump → CounterTrump → TakeBottom → DiscardBottom → Playing → Ended
```

Each phase has a whitelist of allowed actions (same pattern as doudizhu's `phaseActions` map).

### Engine Selection

Add registry map `gameType → Engine factory` in `server/internal/game/`. WebSocket handler uses it instead of hardcoded `&doudizhu.Engine{}`.

### GameRoom Adjustments

- Seat count: configurable (3 for doudizhu, 4 for dashengji)
- Team logic: dealer-team (0,2) vs non-dealer-team (1,3)
- State broadcast carries game-specific fields

## Frontend Design

### Routes

```
app/(main)/room/[id]/
  doudzhu/page.tsx       — Doudizhu game (relocated from room/[id]/page.tsx)
  dashengji/page.tsx     — Dashengji game (new)
  page.tsx               — Redirect based on gameType
```

### Component Split

**Shared** (`components/game/shared/`): Card, WebSocket client, ReadyBar, ChatPanel, themes, audio

**Dashengji-specific** (`components/game/dashengji/`): Table layout (4-player, partners opposite), ActionBar, TrumpSelector, BottomCardsReveal, ScoreBoard

### 4-Player Layout

Team A (dealer): Seats 0+2 (opposite). Team B: Seats 1+3 (opposite).

## Testing Strategy (TDD)

Go tests only (no frontend tests initially):

| Test File | Coverage |
|-----------|----------|
| `card_test.go` | 162-card deck, deal distribution, sort, rank calculation under trump configs |
| `trump_test.go` | 定主/反主 eligibility, 定死/反死, card classification |
| `hand_test.go` | ParsePlay (4 hand types), tractor edge cases, CanBeat with dynamic ranking |
| `engine_test.go` | Phase transitions, follow-suit enforcement, 枪毙/垫牌 logic, action validation |
| `score_test.go` | Point brackets, level progression, dealer swap |

## Out of Scope (v2)

- LLM-powered AI players for Dashengji
- Frontend component tests
- Database schema changes for Dashengji game records
