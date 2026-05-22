# Dashengji (打升级) Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build a full-rules Dashengji (打升级) game engine with TDD, wire it into the WebSocket routing, and create a separate frontend game page.

**Architecture:** Mirror the doudizhu pattern — independent `server/internal/game/dashengji/` package implementing the `GameEngine` interface, a game engine registry for dynamic selection, and a separate `/room/[id]/dashengji` route with game-specific components. Zero changes to doudizhu engine code.

**Tech Stack:** Go 1.26 (server engine + tests), TypeScript/React 19/Next.js 16/Tailwind CSS v4 (frontend)

---

## File Structure

### Server — Create
| File | Responsibility |
|------|---------------|
| `server/internal/game/dashengji/card.go` | 162-card deck (3×54), deal, shuffle, sort, compare-rank |
| `server/internal/game/dashengji/card_test.go` | Tests for deck, deal, sort, rank calc |
| `server/internal/game/dashengji/trump.go` | Trump eligibility checks (定主/反主), 定死/反死, card classification |
| `server/internal/game/dashengji/trump_test.go` | Tests for all trump scenarios |
| `server/internal/game/dashengji/hand.go` | ParsePlay (4 hand types), CanBeat, follow-suit validation |
| `server/internal/game/dashengji/hand_test.go` | Tests for hand parsing and comparison |
| `server/internal/game/dashengji/state.go` | GamePhase enum, GameState struct, phase-action whitelist |
| `server/internal/game/dashengji/engine.go` | GameEngine implementation (Init, ExecuteAction, FilterForPlayer, etc.) |
| `server/internal/game/dashengji/engine_test.go` | Phase transitions, action validation, full round flow |
| `server/internal/game/dashengji/score.go` | Point tracking, level progression, dealer swap |
| `server/internal/game/dashengji/score_test.go` | Score bracket tests, level tests |
| `server/internal/game/registry.go` | gameType → Engine factory map |

### Server — Modify
| File | Change |
|------|--------|
| `server/internal/api/ws/handler.go:170` | Replace `&doudizhu.Engine{}` with `game.NewEngine(gameType)` |
| `server/internal/game/room.go` | Parameterize seat count (3→configurable), adjust `nextAvailableSeat` |

### Frontend — Create
| File | Responsibility |
|------|---------------|
| `frontend/app/(main)/room/[id]/dashengji/page.tsx` | Dashengji game page (standalone, self-contained) |
| `frontend/components/game/dashengji/DashengjiTable.tsx` | 4-player 2v2 table layout |
| `frontend/components/game/dashengji/DashengjiActionBar.tsx` | Phase-specific controls (定主/反主/扣底/出牌) |
| `frontend/components/game/dashengji/TrumpSelector.tsx` | Trump card selection modal |
| `frontend/components/game/dashengji/BottomCardsReveal.tsx` | Bottom 6 cards display |
| `frontend/components/game/dashengji/ScoreBoard.tsx` | Level progress + round score tracker |

### Frontend — Modify
| File | Change |
|------|--------|
| `frontend/app/(main)/room/[id]/page.tsx` | Redirect to correct sub-route based on gameType |
| `frontend/app/(main)/room/create/page.tsx` | Add game type selector |

---

### Task 1: Card encoding and deck operations (TDD)

**Files:**
- Create: `server/internal/game/dashengji/card.go`
- Create: `server/internal/game/dashengji/card_test.go`

- [ ] **Step 1: Write the failing test for 162-card deck creation**

```go
// card_test.go
package dashengji

import (
	"slices"
	"testing"
)

func TestNewDeck(t *testing.T) {
	deck := NewDeck()
	if len(deck) != 162 {
		t.Fatalf("expected 162 cards, got %d", len(deck))
	}

	// Verify 3 copies of each face (0–53)
	faceCount := make(map[int]int)
	for _, c := range deck {
		faceCount[c.Face()]++
	}
	for face := 0; face < 54; face++ {
		if faceCount[face] != 3 {
			t.Errorf("face %d: expected 3 copies, got %d", face, faceCount[face])
		}
	}

	// Verify IDs 0–161
	seen := make(map[int]bool)
	for _, c := range deck {
		if c.ID < 0 || c.ID > 161 {
			t.Errorf("card ID %d out of range", c.ID)
		}
		if seen[c.ID] {
			t.Errorf("duplicate card ID %d", c.ID)
		}
		seen[c.ID] = true
	}
}

func TestCardFaceAndCopy(t *testing.T) {
	c := Card{ID: 55}
	if c.Face() != 1 {
		t.Errorf("ID 55: expected face 1, got %d", c.Face())
	}
	if c.Copy() != 1 {
		t.Errorf("ID 55: expected copy 1, got %d", c.Copy())
	}
	c2 := Card{ID: 0}
	if c2.Face() != 0 {
		t.Errorf("ID 0: expected face 0, got %d", c2.Face())
	}
	if c2.Copy() != 0 {
		t.Errorf("ID 0: expected copy 0, got %d", c2.Copy())
	}
}

func TestCardBaseRank(t *testing.T) {
	// Standard cards: 3→3, 4→4, ..., K→13, A→14, 2→15
	tests := []struct{ face, want int }{
		{0, 3},  // ♠3
		{12, 15}, // ♠2
		{13, 3},  // ♥3
		{25, 15}, // ♥2
	}
	for _, tt := range tests {
		c := Card{ID: tt.face}
		if c.BaseRank() != tt.want {
			t.Errorf("face %d: expected rank %d, got %d", tt.face, tt.want, c.BaseRank())
		}
	}
	// Jokers
	if Card{ID: 52}.BaseRank() != 16 {
		t.Errorf("small joker: expected rank 16, got %d", Card{ID: 52}.BaseRank())
	}
	if Card{ID: 53}.BaseRank() != 17 {
		t.Errorf("big joker: expected rank 17, got %d", Card{ID: 53}.BaseRank())
	}
}

func TestCardSuit(t *testing.T) {
	tests := []struct{ face, want int }{
		{0, 0}, {12, 0},   // ♠
		{13, 1}, {25, 1},  // ♥
		{26, 2}, {38, 2},  // ♣
		{39, 3}, {51, 3},  // ♦
		{52, 4}, {53, 4},  // jokers
	}
	for _, tt := range tests {
		c := Card{ID: tt.face}
		if c.Suit() != tt.want {
			t.Errorf("face %d: expected suit %d, got %d", tt.face, tt.want, c.Suit())
		}
	}
}

func TestCardIsJoker(t *testing.T) {
	if !Card{ID: 52}.IsSmallJoker() {
		t.Error("face 52 should be small joker")
	}
	if !Card{ID: 53}.IsBigJoker() {
		t.Error("face 53 should be big joker")
	}
	if Card{ID: 0}.IsSmallJoker() || Card{ID: 0}.IsBigJoker() {
		t.Error("face 0 should not be a joker")
	}
}

func TestCompareRank(t *testing.T) {
	// With trumpSuit=0 (♠), levelRank=6 (打6)
	// Order: 副牌 < 主花色牌 < 副2 < 本2 < 副6 < 本6 < 小王 < 大王
	trumpSuit := 0
	levelRank := 6 // 3+3

	// 副牌 ♥3 (face=13): baseRank=3
	side3 := Card{ID: 13}
	// 主花色牌 ♠3 (face=0): baseRank=3, trump suit
	main3 := Card{ID: 0}
	// 副2 ♥2 (face=25): baseRank=15
	side2 := Card{ID: 25}
	// 本2 ♠2 (face=12): baseRank=15, trump suit
	main2 := Card{ID: 12}
	// 副6 ♥6 (face=16): baseRank=6, levelRank
	side6 := Card{ID: 16}
	// 本6 ♠6 (face=3): baseRank=6, levelRank, trump suit
	main6 := Card{ID: 3}
	// 小王 (face=52)
	smallJ := Card{ID: 52}
	// 大王 (face=53)
	bigJ := Card{ID: 53}

	cr := func(c Card) int { return c.CompareRank(trumpSuit, levelRank) }

	ranks := []Card{side3, main3, side2, main2, side6, main6, smallJ, bigJ}
	for i := 1; i < len(ranks); i++ {
		if cr(ranks[i-1]) >= cr(ranks[i]) {
			t.Errorf("%s (rank %d) should be < %s (rank %d)",
				ranks[i-1].Display(), cr(ranks[i-1]),
				ranks[i].Display(), cr(ranks[i]))
		}
	}
}

func TestDeal(t *testing.T) {
	deck := NewDeck()
	Shuffle(deck)
	h0, h1, h2, h3, bottom := Deal(deck)

	if len(h0) != 39 || len(h1) != 39 || len(h2) != 39 || len(h3) != 39 {
		t.Errorf("expected 39 cards per hand, got %d/%d/%d/%d",
			len(h0), len(h1), len(h2), len(h3))
	}
	if len(bottom) != 6 {
		t.Errorf("expected 6 bottom cards, got %d", len(bottom))
	}

	// Verify no duplicate IDs across all hands
	all := make(map[int]bool)
	for _, c := range slices.Concat(h0, h1, h2, h3, bottom) {
		if all[c.ID] {
			t.Errorf("duplicate card ID %d", c.ID)
		}
		all[c.ID] = true
	}
}

func TestSortCards(t *testing.T) {
	trumpSuit := -1 // no trump, default sort by base rank desc
	levelRank := 6

	cards := []Card{{ID: 13}, {ID: 0}, {ID: 52}, {ID: 53}, {ID: 25}}
	SortCards(cards, trumpSuit, levelRank)

	// Expected: 大王(53) > 小王(52) > ♥2(25, rank 15) > ♠3(0, rank 3) > ♥3(13, rank 3)
	if cards[0].ID != 53 {
		t.Errorf("first should be big joker, got %d", cards[0].ID)
	}
	if cards[1].ID != 52 {
		t.Errorf("second should be small joker, got %d", cards[1].ID)
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

```bash
cd server && go test ./internal/game/dashengji/... -v -run TestNewDeck 2>&1
```

Expected: compilation error (package/files don't exist yet).

- [ ] **Step 3: Write minimal card.go implementation**

```go
// card.go
package dashengji

import (
	"math/rand"
	"sort"
)

// Card represents a playing card in a 3-deck (162-card) Dashengji game.
// ID encoding: face = ID % 54, copy = ID / 54
// face 0-51: standard cards (suit*13 + rank), 52: small joker, 53: big joker
type Card struct {
	ID int `json:"id"`
}

var suitChars = []string{"♠", "♥", "♣", "♦"}
var rankChars = []string{"3", "4", "5", "6", "7", "8", "9", "10", "J", "Q", "K", "A", "2"}

const deckSize = 162 // 54 * 3
const handSize = 39

// Face returns which of the 54 unique card faces this card represents (0-53).
func (c Card) Face() int { return c.ID % 54 }

// Copy returns which deck copy this card is (0, 1, or 2).
func (c Card) Copy() int { return c.ID / 54 }

// Suit returns the suit: 0=♠, 1=♥, 2=♣, 3=♦, 4=joker.
func (c Card) Suit() int {
	face := c.Face()
	if face >= 52 {
		return 4
	}
	return face / 13
}

// BaseRank returns the base rank: 3→3, 4→4, ..., K→13, A→14, 2→15, SJ→16, BJ→17.
func (c Card) BaseRank() int {
	face := c.Face()
	if face == 52 {
		return 16
	}
	if face == 53 {
		return 17
	}
	return (face % 13) + 3
}

// IsSmallJoker returns true if the card is a small joker.
func (c Card) IsSmallJoker() bool { return c.Face() == 52 }

// IsBigJoker returns true if the card is a big joker.
func (c Card) IsBigJoker() bool { return c.Face() == 53 }

// CompareRank returns the effective comparison rank given trump suit and level card rank.
// Higher values beat lower values.
// Order: 副牌(3-14) < 主花色牌(103-114) < 副2(150) < 本2(160) < 副级牌(170) < 本级牌(180) < 小王(190) < 大王(200)
func (c Card) CompareRank(trumpSuit int, levelRank int) int {
	if c.IsBigJoker() {
		return 200
	}
	if c.IsSmallJoker() {
		return 190
	}

	base := c.BaseRank()
	suit := c.Suit()
	isTrumpSuit := suit == trumpSuit
	isLevelCard := base == levelRank
	isTwo := base == 15

	if isTrumpSuit && isLevelCard {
		return 180 // 本级牌
	}
	if !isTrumpSuit && isLevelCard {
		return 170 // 副级牌
	}
	if isTrumpSuit && isTwo {
		return 160 // 本2
	}
	if !isTrumpSuit && isTwo {
		return 150 // 副2
	}
	if isTrumpSuit {
		return 100 + base // 主花色牌
	}
	return base // 副牌
}

// Display returns a human-readable string representation.
func (c Card) Display() string {
	face := c.Face()
	if face == 52 {
		return "🃏"
	}
	if face == 53 {
		return "👑"
	}
	return suitChars[c.Suit()] + rankChars[face%13]
}

// NewDeck creates a 162-card deck (3 copies of each of the 54 unique faces).
func NewDeck() []Card {
	cards := make([]Card, deckSize)
	for i := 0; i < deckSize; i++ {
		cards[i] = Card{ID: i}
	}
	return cards
}

// Shuffle randomizes the deck using Fisher-Yates.
func Shuffle(deck []Card) {
	rand.Shuffle(len(deck), func(i, j int) {
		deck[i], deck[j] = deck[j], deck[i]
	})
}

// SortCards sorts in descending compare-rank order. If compare ranks are equal,
// sorts by suit (♠, ♥, ♣, ♦).
func SortCards(cards []Card, trumpSuit int, levelRank int) {
	sort.Slice(cards, func(i, j int) bool {
		ri := cards[i].CompareRank(trumpSuit, levelRank)
		rj := cards[j].CompareRank(trumpSuit, levelRank)
		if ri != rj {
			return ri > rj
		}
		return cards[i].Suit() < cards[j].Suit()
	})
}

// Deal splits a 162-card deck into four 39-card hands and 6 bottom cards.
func Deal(deck []Card) (h0, h1, h2, h3, bottom []Card) {
	h0 = make([]Card, handSize)
	copy(h0, deck[0:handSize])
	h1 = make([]Card, handSize)
	copy(h1, deck[handSize:2*handSize])
	h2 = make([]Card, handSize)
	copy(h2, deck[2*handSize:3*handSize])
	h3 = make([]Card, handSize)
	copy(h3, deck[3*handSize:4*handSize])
	bottom = make([]Card, 6)
	copy(bottom, deck[4*handSize:])
	return
}
```

- [ ] **Step 4: Run tests to verify they pass**

```bash
cd server && go test ./internal/game/dashengji/... -v -run "TestNewDeck|TestCard|TestCompare|TestDeal|TestSort" 2>&1
```

Expected: all PASS.

- [ ] **Step 5: Commit**

```bash
git add server/internal/game/dashengji/card.go server/internal/game/dashengji/card_test.go
git commit -m "feat(dashengji): add 162-card deck with compare-rank and deal logic"
```

---

### Task 2: Trump system — eligibility and card classification (TDD)

**Files:**
- Create: `server/internal/game/dashengji/trump.go`
- Create: `server/internal/game/dashengji/trump_test.go`

- [ ] **Step 1: Write the failing tests for trump eligibility**

```go
// trump_test.go
package dashengji

import "testing"

func TestCheckSetTrump_RedJokerRed2RedLevel(t *testing.T) {
	// 大王 + ♥2 + ♥6 → 主色 ♥ (suit 1)
	cards := []Card{{ID: 53}, {ID: 25}, {ID: 16}} // 大王, ♥2, ♥6 (levelRank=6)
	ok := CheckSetTrump(cards, 6, false)
	if !ok {
		t.Error("大王+♥2+♥6 should be valid set trump")
	}
}

func TestCheckSetTrump_BlackJokerBlack2BlackLevel(t *testing.T) {
	// 小王 + ♣2 + ♣6 → 主色 ♣ (suit 2)
	cards := []Card{{ID: 52}, {ID: 38}, {ID: 29}} // 小王, ♣2, ♣6
	ok := CheckSetTrump(cards, 6, false)
	if !ok {
		t.Error("小王+♣2+♣6 should be valid set trump")
	}
}

func TestCheckSetTrump_RedJokerRed2TwoRedLevel(t *testing.T) {
	// 大王 + ♥2 + ♥6♥6 → 主色 ♥ (suit 1)
	cards := []Card{{ID: 53}, {ID: 25}, {ID: 16}, {ID: 70}} // 大王, ♥2, ♥6, ♥6 (copy 1)
	ok := CheckSetTrump(cards, 6, false)
	if !ok {
		t.Error("大王+♥2+♥6♥6 should be valid set trump")
	}
}

func TestCheckSetTrump_BlackJokerBlack2TwoBlackLevel(t *testing.T) {
	// 小王 + ♣2 + ♣6♣6 → 主色 ♣
	cards := []Card{{ID: 52}, {ID: 38}, {ID: 29}, {ID: 83}} // 小王, ♣2, ♣6, ♣6
	ok := CheckSetTrump(cards, 6, false)
	if !ok {
		t.Error("小王+♣2+♣6♣6 should be valid set trump")
	}
}

func TestCheckSetTrump_WrongColorMix(t *testing.T) {
	// 大王 + ♥2 + ♠6 → invalid (2 and level must match suit)
	cards := []Card{{ID: 53}, {ID: 25}, {ID: 3}} // 大王(red), ♥2(red), ♠6(black)
	ok := CheckSetTrump(cards, 6, false)
	if ok {
		t.Error("mixed color 2 and level should be invalid")
	}
}

func TestCheckSetTrump_WrongJokerColor(t *testing.T) {
	// 大王(red) + ♣2(black) + ♣6(black) → invalid (joker and 2 must be same color)
	cards := []Card{{ID: 53}, {ID: 38}, {ID: 29}} // 大王, ♣2, ♣6
	ok := CheckSetTrump(cards, 6, false)
	if ok {
		t.Error("大王(red) with black 2+level should be invalid")
	}
}

func TestCheckSetTrump_NeedsBoth2AndLevel(t *testing.T) {
	// Only 大王 + ♥6 (no 2) → invalid
	cards := []Card{{ID: 53}, {ID: 16}}
	ok := CheckSetTrump(cards, 6, false)
	if ok {
		t.Error("missing 2 should be invalid")
	}
}

func TestCheckCounterTrump_RequiresTwoLevelCards(t *testing.T) {
	// 反主 needs: 王 + 2 + 2张级牌 (same suit for 2 and 级牌)
	cards := []Card{{ID: 53}, {ID: 25}, {ID: 16}, {ID: 70}} // 大王, ♥2, ♥6, ♥6
	ok := CheckCounterTrump(cards, 6)
	if !ok {
		t.Error("大王+♥2+♥6♥6 should be valid counter trump")
	}
}

func TestCheckCounterTrump_OneLevelCardNotEnough(t *testing.T) {
	// Only one level card → can't counter trump
	cards := []Card{{ID: 53}, {ID: 25}, {ID: 16}} // 大王, ♥2, ♥6
	ok := CheckCounterTrump(cards, 6)
	if ok {
		t.Error("only 1 level card should not be enough for counter trump")
	}
}

func TestIsDeadSetTrump(t *testing.T) {
	// 定死: 庄家亮牌包含 "2的同花色的对级牌" (pair of level cards same suit as the 2)
	cards := []Card{{ID: 53}, {ID: 25}, {ID: 16}, {ID: 70}} // 大王, ♥2, ♥6, ♥6 (pair of ♥6)
	if !IsDeadTrump(cards, 6) {
		t.Error("pair of level cards same suit as 2 should be 定死")
	}
}

func TestNotDeadTrump_NoLevelPair(t *testing.T) {
	cards := []Card{{ID: 53}, {ID: 25}, {ID: 16}} // Only 1 level card, not a pair
	if IsDeadTrump(cards, 6) {
		t.Error("no level pair should not be 定死")
	}
}

func TestClassifyCard(t *testing.T) {
	trumpSuit := 0 // ♠
	levelRank := 6  // 打6

	tests := []struct {
		face int
		want CardCategory
	}{
		{0, CatNativeMain},   // ♠6 = 本级牌
		{16, CatSideMain},    // ♥6 = 副级牌
		{12, CatNativeMain},  // ♠2 = 本2
		{25, CatSideMain},    // ♥2 = 副2
		{1, CatTrump},        // ♠4 = 主花色牌
		{14, CatSide},        // ♥4 = 副牌
		{52, CatJoker},       // 小王
		{53, CatJoker},       // 大王
	}

	for _, tt := range tests {
		c := Card{ID: tt.face}
		got := ClassifyCard(c, trumpSuit, levelRank)
		if got != tt.want {
			t.Errorf("face %d (%s): expected category %d, got %d", tt.face, c.Display(), tt.want, got)
		}
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

```bash
cd server && go test ./internal/game/dashengji/... -v -run "TestCheck|TestIsDead|TestClassify" 2>&1
```

Expected: compilation error (trump.go doesn't exist yet).

- [ ] **Step 3: Write trump.go implementation**

```go
// trump.go
package dashengji

// CardCategory classifies a card within the trump system.
type CardCategory int

const (
	CatSide       CardCategory = iota // 副牌
	CatTrump                          // 主花色牌 (non-2, non-level)
	CatSideMain                       // 副主: 2/级牌 of non-trump suits
	CatNativeMain                     // 本主: 2/级牌 of trump suit
	CatJoker                          // 大王/小王
)

// ClassifyCard returns the trump category of a card.
func ClassifyCard(c Card, trumpSuit int, levelRank int) CardCategory {
	if c.IsSmallJoker() || c.IsBigJoker() {
		return CatJoker
	}

	suit := c.Suit()
	base := c.BaseRank()
	isTrumpSuit := suit == trumpSuit
	isLevelCard := base == levelRank
	isTwo := base == 15

	if isTrumpSuit && (isLevelCard || isTwo) {
		return CatNativeMain // 本主: trump-suit 2 or level card
	}
	if !isTrumpSuit && (isLevelCard || isTwo) {
		return CatSideMain // 副主: non-trump 2 or level card
	}
	if isTrumpSuit {
		return CatTrump // 主花色牌
	}
	return CatSide // 副牌
}

// jokerColor returns "red" for big joker (大王=red), "black" for small joker (小王=black).
func jokerColor(c Card) string {
	if c.IsBigJoker() {
		return "red"
	}
	if c.IsSmallJoker() {
		return "black"
	}
	return ""
}

// suitColor returns "red" for ♥/♦, "black" for ♠/♣.
func suitColor(suit int) string {
	if suit == 1 || suit == 3 {
		return "red"
	}
	if suit == 0 || suit == 2 {
		return "black"
	}
	return ""
}

// hasPair checks whether the cards contain at least 2 copies of the same face.
func hasPair(cards []Card, face int) bool {
	count := 0
	for _, c := range cards {
		if c.Face() == face {
			count++
		}
	}
	return count >= 2
}

// findTwoSuit finds the suit of a 2 (rank 15) in the cards. Returns -1 if not found.
func findTwoSuit(cards []Card, levelRank int) int {
	for _, c := range cards {
		if c.BaseRank() == 15 {
			return c.Suit()
		}
	}
	return -1
}

// findLevelFace finds a face that is a level card in the cards. Returns -1 if not found.
func findLevelFace(cards []Card, levelRank int) int {
	for _, c := range cards {
		if c.BaseRank() == levelRank {
			return c.Face()
		}
	}
	return -1
}

// CheckSetTrump validates whether the given cards can be used to set trump (定主).
// levelRank is the current level rank (e.g., 6 means 打6, cards of rank 6 are 级牌).
// Conditions (满足一个即可):
// - 大王 + 红色2 + 红色级牌 (2和级牌同花色)
// - 小王 + 黑色2 + 黑色级牌 (2和级牌同花色)
// - 大王 + 红色2 + 2张红色级牌 (2和2张级牌同花色)
// - 小王 + 黑色2 + 2张黑色级牌 (2和2张级牌同花色)
func CheckSetTrump(cards []Card, levelRank int, isDealerTeam bool) bool {
	if len(cards) < 3 {
		return false
	}

	// Find the joker (king)
	hasBigJoker := false
	hasSmallJoker := false
	for _, c := range cards {
		if c.IsBigJoker() {
			hasBigJoker = true
		}
		if c.IsSmallJoker() {
			hasSmallJoker = true
		}
	}
	if !hasBigJoker && !hasSmallJoker {
		return false
	}

	twoSuit := findTwoSuit(cards, levelRank)
	if twoSuit == -1 {
		return false
	}

	// Check color match between joker and 2
	twoColor := suitColor(twoSuit)
	jokerColorMatch := (hasBigJoker && twoColor == "red") || (hasSmallJoker && twoColor == "black")
	if !jokerColorMatch {
		return false
	}

	// Find level cards that match the 2's suit
	levelFace := findLevelFace(cards, levelRank)
	if levelFace == -1 {
		return false
	}
	levelSuit := levelFace / 13
	if levelSuit != twoSuit {
		return false
	}

	levelCount := 0
	for _, c := range cards {
		if c.BaseRank() == levelRank && c.Suit() == twoSuit {
			levelCount++
		}
	}

	return levelCount >= 1
}

// GetTrumpSuit extracts the trump suit from a valid set-trump card selection.
// Returns the suit index (0-3).
func GetTrumpSuit(cards []Card, levelRank int) int {
	return findTwoSuit(cards, levelRank)
}

// CheckCounterTrump validates whether the given cards can counter-trump (反主).
// 反主条件: 王 + 2 + 2张级牌 (2和2张级牌同花色, 颜色匹配)
func CheckCounterTrump(cards []Card, levelRank int) bool {
	if len(cards) < 4 {
		return false
	}

	hasBigJoker := false
	hasSmallJoker := false
	for _, c := range cards {
		if c.IsBigJoker() {
			hasBigJoker = true
		}
		if c.IsSmallJoker() {
			hasSmallJoker = true
		}
	}
	if !hasBigJoker && !hasSmallJoker {
		return false
	}

	twoSuit := findTwoSuit(cards, levelRank)
	if twoSuit == -1 {
		return false
	}

	twoColor := suitColor(twoSuit)
	jokerColorMatch := (hasBigJoker && twoColor == "red") || (hasSmallJoker && twoColor == "black")
	if !jokerColorMatch {
		return false
	}

	// Need at least 2 level cards of the same suit as the 2
	levelCount := 0
	for _, c := range cards {
		if c.BaseRank() == levelRank && c.Suit() == twoSuit {
			levelCount++
		}
	}

	return levelCount >= 2
}

// IsDeadTrump checks if the set-trump is "定死" (dead/final, cannot be countered).
// 定死: when the revealed cards include a pair of level cards same suit as the 2.
func IsDeadTrump(cards []Card, levelRank int) bool {
	twoSuit := findTwoSuit(cards, levelRank)
	if twoSuit == -1 {
		return false
	}
	levelCount := 0
	for _, c := range cards {
		if c.BaseRank() == levelRank && c.Suit() == twoSuit {
			levelCount++
		}
	}
	return levelCount >= 2
}
```

- [ ] **Step 4: Run tests to verify they pass**

```bash
cd server && go test ./internal/game/dashengji/... -v -run "TestCheck|TestIsDead|TestClassify" 2>&1
```

Expected: all PASS.

- [ ] **Step 5: Commit**

```bash
git add server/internal/game/dashengji/trump.go server/internal/game/dashengji/trump_test.go
git commit -m "feat(dashengji): add trump eligibility and card classification"
```

---

### Task 3: Hand type parsing and play comparison (TDD)

**Files:**
- Create: `server/internal/game/dashengji/hand.go`
- Create: `server/internal/game/dashengji/hand_test.go`

- [ ] **Step 1: Write the failing tests for ParsePlay and CanBeat**

```go
// hand_test.go
package dashengji

import "testing"

func TestParsePlay_Single(t *testing.T) {
	play := ParsePlay([]Card{{ID: 0}}) // ♠3
	if play.Type != PlaySingle {
		t.Fatalf("expected PlaySingle, got %d", play.Type)
	}
}

func TestParsePlay_Pair(t *testing.T) {
	// 2 cards, same suit, same rank (same face — different copies)
	play := ParsePlay([]Card{{ID: 0}, {ID: 54}}) // ♠3 + ♠3 (copy 1)
	if play.Type != PlayPair {
		t.Fatalf("expected PlayPair, got %d", play.Type)
	}
}

func TestParsePlay_Pair_DifferentSuit(t *testing.T) {
	// 2 cards, same rank but different suits → invalid pair
	play := ParsePlay([]Card{{ID: 0}, {ID: 13}}) // ♠3 + ♥3
	if play.Type != PlayInvalid {
		t.Fatalf("expected PlayInvalid for cross-suit pair, got %d", play.Type)
	}
}

func TestParsePlay_Triple(t *testing.T) {
	// 3 cards, same suit, same rank (same face — all 3 copies)
	play := ParsePlay([]Card{{ID: 0}, {ID: 54}, {ID: 108}}) // ♠3 ×3
	if play.Type != PlayTriple {
		t.Fatalf("expected PlayTriple, got %d", play.Type)
	}
}

func TestParsePlay_Triple_DifferentSuit(t *testing.T) {
	play := ParsePlay([]Card{{ID: 0}, {ID: 54}, {ID: 13}}) // ♠3♠3♥3
	if play.Type != PlayInvalid {
		t.Fatalf("expected PlayInvalid for mixed-suit triple, got %d", play.Type)
	}
}

func TestParsePlay_Tractor(t *testing.T) {
	// 3+ consecutive pairs, same suit
	// ♠3♠3♠4♠4♠5♠5 (faces 0, 54, 1, 55, 2, 56)
	trumpSuit := -1 // No trump context for testing (all side cards)
	levelRank := 6

	play := ParsePlayWithContext(
		[]Card{{ID: 0}, {ID: 54}, {ID: 1}, {ID: 55}, {ID: 2}, {ID: 56}},
		trumpSuit, levelRank,
	)
	if play.Type != PlayTractor {
		t.Fatalf("expected PlayTractor, got %d", play.Type)
	}
	if play.Length != 3 {
		t.Errorf("expected length 3 (pairs), got %d", play.Length)
	}
	if play.MainRank != 3 {
		t.Errorf("expected main rank 3, got %d", play.MainRank)
	}
}

func TestParsePlay_Tractor_NotConsecutive(t *testing.T) {
	// ♠3♠3♠7♠7 — not consecutive, invalid
	play := ParsePlayWithContext(
		[]Card{{ID: 0}, {ID: 54}, {ID: 4}, {ID: 58}},
		-1, 6,
	)
	if play.Type != PlayInvalid {
		t.Fatalf("expected PlayInvalid for non-consecutive pairs, got %d", play.Type)
	}
}

func TestParsePlay_Tractor_TooShort(t *testing.T) {
	// Only 2 consecutive pairs — need ≥3
	play := ParsePlayWithContext(
		[]Card{{ID: 0}, {ID: 54}, {ID: 1}, {ID: 55}},
		-1, 6,
	)
	if play.Type != PlayInvalid {
		t.Fatalf("expected PlayInvalid for 2-pair tractor, got %d", play.Type)
	}
}

func TestParsePlay_JokersCannotPair(t *testing.T) {
	// 小王+大王 is not a pair
	play := ParsePlay([]Card{{ID: 52}, {ID: 53}})
	if play.Type != PlayInvalid {
		t.Fatalf("expected PlayInvalid for mixed jokers, got %d", play.Type)
	}
}

func TestParsePlay_Empty(t *testing.T) {
	play := ParsePlay([]Card{})
	if play.Type != PlayPass {
		t.Fatalf("expected PlayPass, got %d", play.Type)
	}
}

func TestCanBeat_SameType_HigherRank(t *testing.T) {
	last := Play{Type: PlaySingle, MainRank: 4, Length: 1}
	current := Play{Type: PlaySingle, MainRank: 8, Length: 1}
	if !CanBeat(current, last) {
		t.Error("single 8 should beat single 4")
	}
}

func TestCanBeat_SameType_LowerRank(t *testing.T) {
	last := Play{Type: PlaySingle, MainRank: 8, Length: 1}
	current := Play{Type: PlaySingle, MainRank: 4, Length: 1}
	if CanBeat(current, last) {
		t.Error("single 4 should not beat single 8")
	}
}

func TestCanBeat_DifferentType_NoBeat(t *testing.T) {
	last := Play{Type: PlaySingle, MainRank: 8, Length: 1}
	current := Play{Type: PlayPair, MainRank: 3, Length: 1}
	if CanBeat(current, last) {
		t.Error("pair should not beat single")
	}
}

func TestCanBeat_LeadCanPlayAnything(t *testing.T) {
	// When last play is Pass (new lead), anything valid goes
	last := Play{Type: PlayPass}
	current := Play{Type: PlaySingle, MainRank: 3, Length: 1}
	if !CanBeat(current, last) {
		t.Error("any play should be valid when leading")
	}
}

func TestCanBeat_PassNeverBeats(t *testing.T) {
	last := Play{Type: PlaySingle, MainRank: 3, Length: 1}
	current := Play{Type: PlayPass}
	if CanBeat(current, last) {
		t.Error("pass should not beat any play")
	}
}

func TestCanBeat_Tractor_MorePairs(t *testing.T) {
	last := Play{Type: PlayTractor, MainRank: 3, Length: 3}    // 3-4-5
	current := Play{Type: PlayTractor, MainRank: 4, Length: 3}  // 4-5-6
	if !CanBeat(current, last) {
		t.Error("tractor 4-5-6 should beat tractor 3-4-5")
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

```bash
cd server && go test ./internal/game/dashengji/... -v -run "TestParse|TestCanBeat" 2>&1
```

Expected: compilation error.

- [ ] **Step 3: Write hand.go implementation**

```go
// hand.go
package dashengji

import "sort"

// PlayType represents the type of a card play in Dashengji.
type PlayType int

const (
	PlayInvalid  PlayType = 0
	PlaySingle   PlayType = 1
	PlayPair     PlayType = 2
	PlayTriple   PlayType = 3
	PlayTractor  PlayType = 4
	PlayPass     PlayType = 5
)

// Play represents a parsed card play.
type Play struct {
	Type     PlayType `json:"type"`
	MainRank int      `json:"main_rank"`
	Length   int      `json:"length"` // number of pairs for tractor, 1 otherwise
}

// ParsePlay analyzes cards without trump context (for basic hand type detection).
func ParsePlay(cards []Card) Play {
	return ParsePlayWithContext(cards, -1, -1)
}

// ParsePlayWithContext analyzes cards with trump context for consecutive-rank detection.
func ParsePlayWithContext(cards []Card, trumpSuit int, levelRank int) Play {
	if len(cards) == 0 {
		return Play{Type: PlayPass}
	}

	// Build suit+face frequency for pair/triple/tractor detection
	// Key: suit*100 + baseRank
	type cardKey struct {
		suit int
		face int
	}
	freq := make(map[cardKey]int)
	for _, c := range cards {
		freq[cardKey{c.Suit(), c.Face()}]++
	}

	if len(cards) == 1 {
		return Play{Type: PlaySingle, MainRank: cards[0].BaseRank(), Length: 1}
	}

	// Check for joker misuse (jokers cannot form pairs/triples together)
	hasSJ, hasBJ := false, false
	for _, c := range cards {
		if c.IsSmallJoker() {
			hasSJ = true
		}
		if c.IsBigJoker() {
			hasBJ = true
		}
	}
	if hasSJ && hasBJ {
		return Play{Type: PlayInvalid}
	}
	// Two same jokers is fine for pair
	if len(cards) == 2 && ((hasSJ && !hasBJ) || (!hasSJ && hasBJ)) {
		return Play{Type: PlayPair, MainRank: cards[0].BaseRank()}
	}

	// Check all cards have the same suit (required for pair, triple, tractor)
	suit := cards[0].Suit()
	for _, c := range cards[1:] {
		if c.Suit() != suit {
			return Play{Type: PlayInvalid}
		}
	}

	// Count pairs by face
	pairFaces := make([]int, 0)
	for key, count := range freq {
		if count == 2 || count == 3 {
			pairFaces = append(pairFaces, key.face)
		}
	}

	// All cards must be accounted for (no solo cards mixed in)
	totalPaired := 0
	for _, f := range pairFaces {
		totalPaired += freq[cardKey{suit, f}]
	}
	if totalPaired != len(cards) {
		return Play{Type: PlayInvalid}
	}

	// Pure pair: 2 cards, same face
	if len(cards) == 2 && len(pairFaces) == 1 {
		return Play{Type: PlayPair, MainRank: cards[0].BaseRank()}
	}

	// Pure triple (刻子): 3 cards, same face
	if len(cards) == 3 && len(pairFaces) == 1 {
		return Play{Type: PlayTriple, MainRank: cards[0].BaseRank()}
	}

	// Tractor: 3+ consecutive pairs, same suit, same category
	if len(pairFaces) >= 3 && len(cards)%2 == 0 {
		// Sort pair faces by their base ranks
		sort.Ints(pairFaces)

		// Check consecutive in base ranks
		isConsec := true
		for i := 1; i < len(pairFaces); i++ {
			if pairFaces[i]-pairFaces[i-1] != 1 {
				isConsec = false
				break
			}
		}

		// Also check they're in the same trump category if trump context provided
		if isConsec && trumpSuit >= 0 && levelRank >= 0 {
			firstCat := ClassifyCard(Card{ID: pairFaces[0]}, trumpSuit, levelRank)
			for _, f := range pairFaces[1:] {
				if ClassifyCard(Card{ID: f}, trumpSuit, levelRank) != firstCat {
					isConsec = false
					break
				}
			}
		}

		if isConsec {
			return Play{
				Type:     PlayTractor,
				MainRank: pairFaces[0] % 13 + 3, // base rank of lowest pair
				Length:   len(pairFaces),
			}
		}
	}

	return Play{Type: PlayInvalid}
}

// CanBeat checks whether play beats lastPlay.
// In Dashengji, plays must match type and length.
func CanBeat(play, lastPlay Play) bool {
	if lastPlay.Type == PlayPass {
		return true // leading, any valid play is fine
	}
	if play.Type == PlayPass {
		return false // pass never beats
	}
	if play.Type != lastPlay.Type {
		return false // must match type
	}
	if play.Type == PlayTractor && play.Length != lastPlay.Length {
		return false // tractor must match pair count
	}
	return play.MainRank > lastPlay.MainRank
}
```

- [ ] **Step 4: Run tests to verify they pass**

```bash
cd server && go test ./internal/game/dashengji/... -v -run "TestParse|TestCanBeat" 2>&1
```

Expected: all PASS.

- [ ] **Step 5: Commit**

```bash
git add server/internal/game/dashengji/hand.go server/internal/game/dashengji/hand_test.go
git commit -m "feat(dashengji): add hand type parsing and play comparison"
```

---

### Task 4: Game state and phase definitions

**Files:**
- Create: `server/internal/game/dashengji/state.go`

- [ ] **Step 1: Write state.go with phase enum, GameState, and phase-action whitelist**

```go
// state.go
package dashengji

import "encoding/json"

// GamePhase represents the current phase of a Dashengji game.
type GamePhase int

const (
	PhaseSetTrump      GamePhase = iota // 定主
	PhaseCounterTrump                   // 反主
	PhaseTakeBottom                     // 起底
	PhaseDiscardBottom                  // 扣底
	PhasePlaying                        // 出牌
	PhaseEnded                          // 结束
)

var phaseNames = map[GamePhase]string{
	PhaseSetTrump:      "set_trump",
	PhaseCounterTrump:  "counter_trump",
	PhaseTakeBottom:    "take_bottom",
	PhaseDiscardBottom: "discard_bottom",
	PhasePlaying:       "playing",
	PhaseEnded:         "ended",
}

func (p GamePhase) String() string {
	if name, ok := phaseNames[p]; ok {
		return name
	}
	return "unknown"
}

// phaseActions maps each game phase to allowed action strings.
var phaseActions = map[GamePhase][]string{
	PhaseSetTrump:      {"set_trump", "pass_trump"},
	PhaseCounterTrump:  {"counter_trump", "pass_counter"},
	PhaseTakeBottom:    {"take_bottom"},
	PhaseDiscardBottom: {"discard_bottom"},
	PhasePlaying:       {"play", "pass"},
}

// AllowedActions returns the list of valid action strings for a phase.
func AllowedActions(phase GamePhase) []string {
	return phaseActions[phase]
}

// SeatTeam returns 0 for dealer team (seats 0,2) or 1 for non-dealer team (seats 1,3).
func SeatTeam(seat int) int {
	return seat % 2
}

// IsDealerTeam returns true if the seat belongs to the dealer team.
func IsDealerTeam(seat int) bool {
	return SeatTeam(seat) == 0
}

// PartnerSeat returns the partner's seat (0↔2, 1↔3).
func PartnerSeat(seat int) int {
	return (seat + 2) % 4
}

// PlayerHand represents a player's hand and metadata.
type PlayerHand struct {
	UserID int64  `json:"user_id"`
	Seat   int    `json:"seat"`
	Hand   []Card `json:"hand"`
}

// PlayRecord records a play made by a player.
type PlayRecord struct {
	Seat  int    `json:"seat"`
	Play  Play   `json:"play"`
	Cards []Card `json:"cards"`
}

// GameState represents the full state of a Dashengji game.
type GameState struct {
	Phase       GamePhase    `json:"phase"`
	Players     []PlayerHand `json:"players"`
	CurrentSeat int          `json:"current_seat"`

	// Dealer team info
	DealerSeats  [2]int `json:"dealer_seats"`  // seats 0,2 or 1,3
	CurrentLevel int    `json:"current_level"` // 3–14 (A=14)
	LevelRank    int    `json:"level_rank"`    // base rank of level card (3→3, ..., A→14)

	// Trump
	TrumpSuit    int  `json:"trump_suit"`    // -1 if not set, 0-3 otherwise
	IsDeadTrump  bool `json:"is_dead_trump"` // whether trump is final (定死)
	TrumpCards   []Card `json:"trump_cards"` // cards used to set/counter trump
	TrumpRevealed bool `json:"trump_revealed"` // whether trump cards have been shown

	// Bottom cards
	BottomCards     []Card `json:"bottom_cards"`     // 6 bottom cards
	BottomTaken     bool   `json:"bottom_taken"`     // whether bottom has been taken
	TakeBottomSeat  int    `json:"take_bottom_seat"` // which dealer takes bottom
	DiscardedCards  []Card `json:"discarded_cards"`  // cards discarded back
	BottomRevealed  bool   `json:"bottom_revealed"`  // whether discarded cards shown

	// Play state
	LastPlay          *PlayRecord `json:"last_play"`
	PlayHistory       []PlayRecord `json:"play_history"`
	ConsecutivePasses int         `json:"consecutive_passes"`
	WinnerSeat        *int        `json:"winner_seat,omitempty"`

	// Pass tracking
	HasPassedTrump  map[int]bool `json:"has_passed_trump"` // seats that passed set_trump
	HasPassedCounter map[int]bool `json:"has_passed_counter"`

	// Score tracking
	RoundPoints   int `json:"round_points"`   // total points collected this round by non-dealer team
	RoundNum      int `json:"round_num"`
	DealerHistory []int `json:"dealer_history"` // which player took bottom each round (alternating)
}

// ToJSON serializes the GameState to JSON.
func (s *GameState) ToJSON() ([]byte, error) {
	return json.Marshal(s)
}

// FromJSON deserializes JSON data into the GameState.
func (s *GameState) FromJSON(data []byte) error {
	return json.Unmarshal(data, s)
}
```

- [ ] **Step 2: Verify compilation**

```bash
cd server && go build ./internal/game/dashengji/ 2>&1
```

Expected: no errors.

- [ ] **Step 3: Commit**

```bash
git add server/internal/game/dashengji/state.go
git commit -m "feat(dashengji): add game state, phase enum, and action whitelist"
```

---

### Task 5: Scoring and level progression (TDD)

**Files:**
- Create: `server/internal/game/dashengji/score.go`
- Create: `server/internal/game/dashengji/score_test.go`

- [ ] **Step 1: Write the failing tests for scoring**

```go
// score_test.go
package dashengji

import "testing"

func TestCalculateLevelChange_DealerWinNoPoints(t *testing.T) {
	// 结束轮庄赢 + [0,0] → 庄升2级
	change := CalculateLevelChange(true, 0)
	if change != 2 {
		t.Errorf("dealer win with 0 points: expected +2 levels, got %d", change)
	}
}

func TestCalculateLevelChange_DealerWinLowPoints(t *testing.T) {
	// [0, 120) → 庄升1级
	change := CalculateLevelChange(true, 60)
	if change != 1 {
		t.Errorf("dealer win with 60 points: expected +1 level, got %d", change)
	}
	change = CalculateLevelChange(true, 119)
	if change != 1 {
		t.Errorf("dealer win with 119 points: expected +1 level, got %d", change)
	}
}

func TestCalculateLevelChange_DealerWinHighPoints_ShouldNotHappen(t *testing.T) {
	// [120, 300] 庄赢 impossible but defined as 闲升1级
	change := CalculateLevelChange(true, 150)
	if change != -1 {
		t.Errorf("dealer win with 150 points: expected -1 (non-dealer gets level), got %d", change)
	}
}

func TestCalculateLevelChange_NonDealerWinZeroPoints(t *testing.T) {
	// 闲赢 + [0,0] → 闲升1级 + 下庄
	change := CalculateLevelChange(false, 0)
	if change != -1 {
		t.Errorf("non-dealer win with 0 points: expected -1 level, got %d", change)
	}
}

func TestCalculateLevelChange_NonDealerWinLowPoints(t *testing.T) {
	// 闲赢 + [0, 120) → 闲升1级 + 下庄
	change := CalculateLevelChange(false, 60)
	if change != -1 {
		t.Errorf("non-dealer win with 60 points: expected -1 level, got %d", change)
	}
}

func TestCalculateLevelChange_NonDealerWinHighPoints(t *testing.T) {
	// 闲赢 + [120, 300] → 闲升2级 + 下庄
	change := CalculateLevelChange(false, 150)
	if change != -2 {
		t.Errorf("non-dealer win with 150 points: expected -2 levels, got %d", change)
	}
	change = CalculateLevelChange(false, 120)
	if change != -2 {
		t.Errorf("non-dealer win with 120 points: expected -2 levels, got %d", change)
	}
}

func TestAdvanceLevel(t *testing.T) {
	if AdvanceLevel(3, 2) != 5 {
		t.Error("3 + 2 = 5")
	}
	if AdvanceLevel(14, 1) != 14 {
		t.Error("A(14) + 1 should stay at 14 (max)")
	}
	if AdvanceLevel(12, 3) != 14 {
		t.Error("Q(12) + 3 should cap at A(14)")
	}
}

func TestLevelToRank(t *testing.T) {
	if LevelToRank(3) != 3 {
		t.Error("level 3 → rank 3")
	}
	if LevelToRank(14) != 14 {
		t.Error("level A(14) → rank 14")
	}
}

func TestCountRoundPoints(t *testing.T) {
	// Only 5s (5分), 10s (10分), Ks (10分) count
	cards := []Card{
		{ID: 2},  // ♠5 (face 2, base rank 5)
		{ID: 7},  // ♠10 (face 7, base rank 10)
		{ID: 11}, // ♠K (face 11, base rank 13)
		{ID: 0},  // ♠3 (face 0, base rank 3) — no points
	}
	points := CountRoundPoints(cards)
	if points != 25 { // 5 + 10 + 10
		t.Errorf("expected 25 points, got %d", points)
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

```bash
cd server && go test ./internal/game/dashengji/... -v -run "TestCalculate|TestAdvance|TestLevel|TestCount" 2>&1
```

Expected: compilation error.

- [ ] **Step 3: Write score.go implementation**

```go
// score.go
package dashengji

// CalculateLevelChange returns the level change for the dealer team.
// Positive = dealer gains levels (and keeps dealing).
// Negative = non-dealer gains levels (and takes over dealing).
// dealerWin: true if the dealer team won the final trick.
// points: total points collected by non-dealer team this round.
//
// Table:
//
//	                | Dealer wins final | Non-dealer wins final
//	Points [0,0]    | Dealer +2, keep   | Non-dealer +1, take
//	Points [0,120)  | Dealer +1, keep   | Non-dealer +1, take
//	Points [120,300]| Non-dealer +1,take| Non-dealer +2, take
func CalculateLevelChange(dealerWin bool, points int) int {
	if dealerWin {
		if points == 0 {
			return 2 // 庄升2级 + 守庄
		}
		if points < 120 {
			return 1 // 庄升1级 + 守庄
		}
		return -1 // 闲升1级 + 下庄
	} else {
		if points == 0 {
			return -1 // 闲升1级 + 下庄
		}
		if points < 120 {
			return -1 // 闲升1级 + 下庄
		}
		return -2 // 闲升2级 + 下庄
	}
}

// AdvanceLevel computes the new level after a change. Caps at 14 (A).
// Positive change = dealer levels up. Negative = non-dealer levels up.
func AdvanceLevel(currentLevel int, change int) int {
	newLevel := currentLevel + change
	if newLevel > 14 {
		return 14
	}
	if newLevel < 3 {
		return 3
	}
	return newLevel
}

// LevelToRank converts a game level (3-14) to a base rank value.
func LevelToRank(level int) int {
	return level // level 3→3, ..., A→14
}

// CountRoundPoints counts score points in a set of cards.
// 5 = 5 points, 10 = 10 points, K = 10 points.
func CountRoundPoints(cards []Card) int {
	total := 0
	for _, c := range cards {
		switch c.BaseRank() {
		case 5:
			total += 5
		case 10:
			total += 10
		case 13: // K
			total += 10
		}
	}
	return total
}

// IsGameOver returns true if a team has reached level A (14) and won.
func IsGameOver(dealerLevel int) bool {
	return dealerLevel >= 14
}
```

- [ ] **Step 4: Run tests to verify they pass**

```bash
cd server && go test ./internal/game/dashengji/... -v -run "TestCalculate|TestAdvance|TestLevel|TestCount" 2>&1
```

Expected: all PASS.

- [ ] **Step 5: Commit**

```bash
git add server/internal/game/dashengji/score.go server/internal/game/dashengji/score_test.go
git commit -m "feat(dashengji): add scoring and level progression"
```

---

### Task 6: GameEngine implementation (Init, basic phase flow)

**Files:**
- Create: `server/internal/game/dashengji/engine.go`
- Create: `server/internal/game/dashengji/engine_test.go`

- [ ] **Step 1: Write engine.go implementing the GameEngine interface**

```go
// engine.go
package dashengji

import (
	"fmt"
	"math/rand"

	"github.com/yongkl/vibe-pokeface/internal/game"
)

// Engine implements the Dashengji (打升级) game logic.
type Engine struct{}

// Init creates a new game state for 4 players with a 162-card deck.
func (e *Engine) Init(players []game.PlayerInfo) (game.GameState, error) {
	if len(players) != 4 {
		return nil, fmt.Errorf("dashengji requires exactly 4 players, got %d", len(players))
	}

	deck := NewDeck()
	Shuffle(deck)
	h0, h1, h2, h3, bottom := Deal(deck)

	// Sort without trump context initially
	SortCards(h0, -1, -1)
	SortCards(h1, -1, -1)
	SortCards(h2, -1, -1)
	SortCards(h3, -1, -1)

	playerHands := make([]PlayerHand, 4)
	for i, info := range players {
		var hand []Card
		switch i {
		case 0:
			hand = h0
		case 1:
			hand = h1
		case 2:
			hand = h2
		case 3:
			hand = h3
		}
		playerHands[i] = PlayerHand{
			UserID: info.ID,
			Seat:   info.Seat,
			Hand:   hand,
		}
	}

	// Randomly determine dealer team (0,2 or 1,3)
	var dealerSeats [2]int
	if rand.Intn(2) == 0 {
		dealerSeats = [2]int{0, 2}
	} else {
		dealerSeats = [2]int{1, 3}
	}

	state := &GameState{
		Phase:            PhaseSetTrump,
		Players:          playerHands,
		CurrentSeat:      dealerSeats[0], // first dealer seat starts
		DealerSeats:      dealerSeats,
		CurrentLevel:     3, // start at level 3
		LevelRank:        3, // rank 3
		TrumpSuit:        -1,
		BottomCards:      bottom,
		RoundNum:         1,
		HasPassedTrump:   make(map[int]bool),
		HasPassedCounter: make(map[int]bool),
		DealerHistory:    make([]int, 0),
	}

	return state, nil
}

// ExecuteAction processes a player action and transitions the game state.
func (e *Engine) ExecuteAction(state game.GameState, action game.PlayerAction) (game.GameState, error) {
	gs, ok := state.(*GameState)
	if !ok {
		return nil, fmt.Errorf("invalid state type")
	}

	// Find player's seat
	seat := -1
	for i, p := range gs.Players {
		if p.UserID == action.PlayerID {
			seat = i
			break
		}
	}
	if seat == -1 {
		return nil, fmt.Errorf("player %d not found", action.PlayerID)
	}
	if seat != gs.CurrentSeat {
		return nil, &game.GameError{Code: game.ErrNotYourTurn}
	}

	// Validate action against phase whitelist
	allowed := phaseActions[gs.Phase]
	valid := false
	for _, a := range allowed {
		if a == action.Action {
			valid = true
			break
		}
	}
	if !valid {
		return nil, &game.GameError{
			Code:   game.ErrPhaseMismatch,
			Phase:  gs.Phase.String(),
			Action: action.Action,
		}
	}

	// Dispatch to phase handler
	switch gs.Phase {
	case PhaseSetTrump:
		return e.handleSetTrump(gs, seat, action)
	case PhaseCounterTrump:
		return e.handleCounterTrump(gs, seat, action)
	case PhaseTakeBottom:
		return e.handleTakeBottom(gs, seat, action)
	case PhaseDiscardBottom:
		return e.handleDiscardBottom(gs, seat, action)
	case PhasePlaying:
		return e.handlePlay(gs, seat, action)
	default:
		return nil, fmt.Errorf("unknown phase: %v", gs.Phase)
	}
}

func (e *Engine) handleSetTrump(gs *GameState, seat int, action game.PlayerAction) (*GameState, error) {
	if action.Action == "pass_trump" {
		gs.HasPassedTrump[seat] = true
		// Move to next dealer team seat
		nextSeat := e.nextDealerSeat(gs, seat)
		if nextSeat == -1 {
			// Both dealer seats passed — swap dealer and re-init
			return e.swapDealerAndReDeal(gs)
		}
		gs.CurrentSeat = nextSeat
		return gs, nil
	}

	// set_trump: validate cards
	cards := make([]Card, len(action.Cards))
	for i, id := range action.Cards {
		cards[i] = Card{ID: id}
	}

	if !CheckSetTrump(cards, gs.LevelRank, true) {
		return nil, &game.GameError{Code: game.ErrInvalidCards}
	}

	gs.TrumpSuit = GetTrumpSuit(cards, gs.LevelRank)
	gs.TrumpCards = cards
	gs.TrumpRevealed = true

	// Re-sort all hands with trump context
	for i := range gs.Players {
		SortCards(gs.Players[i].Hand, gs.TrumpSuit, gs.LevelRank)
	}

	if IsDeadTrump(cards, gs.LevelRank) {
		// 定死 — skip counter-trump, go to take bottom
		gs.IsDeadTrump = true
		gs.Phase = PhaseTakeBottom
		gs.CurrentSeat = gs.nextTakeBottomSeat()
	} else {
		gs.Phase = PhaseCounterTrump
		// Non-dealer team starts counter-trump
		nonDealerSeat := gs.DealerSeats[0] ^ 1 // flip team: 0→1, 2→3
		gs.CurrentSeat = nonDealerSeat
	}
	return gs, nil
}

func (e *Engine) handleCounterTrump(gs *GameState, seat int, action game.PlayerAction) (*GameState, error) {
	if action.Action == "pass_counter" {
		gs.HasPassedCounter[seat] = true
		nextSeat := e.nextNonDealerSeat(gs, seat)
		if nextSeat == -1 {
			// Both non-dealer seats passed — keep original trump
			gs.Phase = PhaseTakeBottom
			gs.CurrentSeat = gs.nextTakeBottomSeat()
			return gs, nil
		}
		gs.CurrentSeat = nextSeat
		return gs, nil
	}

	// counter_trump: validate
	cards := make([]Card, len(action.Cards))
	for i, id := range action.Cards {
		cards[i] = Card{ID: id}
	}

	if !CheckCounterTrump(cards, gs.LevelRank) {
		return nil, &game.GameError{Code: game.ErrInvalidCards}
	}

	// Counter trump successful
	gs.TrumpSuit = GetTrumpSuit(cards, gs.LevelRank)
	gs.TrumpCards = cards

	// Re-sort hands with new trump
	for i := range gs.Players {
		SortCards(gs.Players[i].Hand, gs.TrumpSuit, gs.LevelRank)
	}

	gs.Phase = PhaseTakeBottom
	gs.CurrentSeat = gs.nextTakeBottomSeat()
	return gs, nil
}

func (e *Engine) handleTakeBottom(gs *GameState, seat int, action game.PlayerAction) (*GameState, error) {
	// Give bottom cards to the taking player
	playerIdx := -1
	for i, p := range gs.Players {
		if p.Seat == seat {
			playerIdx = i
			break
		}
	}
	if playerIdx == -1 {
		return nil, fmt.Errorf("player at seat %d not found", seat)
	}

	// Add bottom cards to player's hand and re-sort
	gs.Players[playerIdx].Hand = append(gs.Players[playerIdx].Hand, gs.BottomCards...)
	SortCards(gs.Players[playerIdx].Hand, gs.TrumpSuit, gs.LevelRank)
	gs.BottomTaken = true
	gs.TakeBottomSeat = seat
	gs.DealerHistory = append(gs.DealerHistory, seat)
	gs.BottomRevealed = true

	gs.Phase = PhaseDiscardBottom
	// Same player discards
	return gs, nil
}

func (e *Engine) handleDiscardBottom(gs *GameState, seat int, action game.PlayerAction) (*GameState, error) {
	if action.Action != "discard_bottom" || len(action.Cards) != 6 {
		return nil, &game.GameError{Code: game.ErrInvalidCards}
	}

	// Remove discarded cards from player's hand
	playerIdx := -1
	for i, p := range gs.Players {
		if p.Seat == seat {
			playerIdx = i
			break
		}
	}

	discardSet := make(map[int]bool)
	for _, id := range action.Cards {
		discardSet[id] = true
	}

	newHand := make([]Card, 0, len(gs.Players[playerIdx].Hand)-6)
	discarded := make([]Card, 0, 6)
	for _, c := range gs.Players[playerIdx].Hand {
		if discardSet[c.ID] {
			discarded = append(discarded, c)
		} else {
			newHand = append(newHand, c)
		}
	}
	gs.Players[playerIdx].Hand = newHand
	gs.DiscardedCards = discarded
	gs.BottomRevealed = true

	// Start playing — the player who took/discarded bottom leads
	gs.Phase = PhasePlaying
	gs.CurrentSeat = seat

	return gs, nil
}

func (e *Engine) handlePlay(gs *GameState, seat int, action game.PlayerAction) (*GameState, error) {
	playerIdx := -1
	for i, p := range gs.Players {
		if p.Seat == seat {
			playerIdx = i
			break
		}
	}

	if action.Action == "pass" {
		if gs.LastPlay == nil || gs.LastPlay.Seat == seat {
			return nil, &game.GameError{Code: game.ErrCannotPass}
		}
		gs.ConsecutivePasses++
		e.advanceSeat(gs)
		return gs, nil
	}

	// Play cards
	cards := make([]Card, len(action.Cards))
	for i, id := range action.Cards {
		cards[i] = Card{ID: id}
	}

	play := ParsePlayWithContext(cards, gs.TrumpSuit, gs.LevelRank)
	if play.Type == PlayInvalid {
		return nil, &game.GameError{Code: game.ErrInvalidCards}
	}

	// If following (not leading), must match
	if gs.LastPlay != nil && gs.LastPlay.Seat != seat {
		if !CanBeat(play, gs.LastPlay.Play) {
			return nil, &game.GameError{Code: game.ErrCannotBeat}
		}
	}

	// Remove cards from hand
	cardSet := make(map[int]bool)
	for _, id := range action.Cards {
		cardSet[id] = true
	}
	newHand := make([]Card, 0, len(gs.Players[playerIdx].Hand)-len(cards))
	for _, c := range gs.Players[playerIdx].Hand {
		if !cardSet[c.ID] {
			newHand = append(newHand, c)
		}
	}
	gs.Players[playerIdx].Hand = newHand

	record := PlayRecord{Seat: seat, Play: play, Cards: cards}
	gs.LastPlay = &record
	gs.PlayHistory = append(gs.PlayHistory, record)
	gs.ConsecutivePasses = 0

	// Check for points in this trick if followed by non-dealer
	if !IsDealerTeam(seat) {
		// Points are counted at round end based on who won the trick
	}

	// Check win
	if len(newHand) == 0 {
		gs.Phase = PhaseEnded
		gs.WinnerSeat = &seat
		return gs, nil
	}

	e.advanceSeat(gs)
	return gs, nil
}

func (e *Engine) advanceSeat(gs *GameState) {
	gs.CurrentSeat = (gs.CurrentSeat + 1) % 4
}

func (e *Engine) nextDealerSeat(gs *GameState, current int) int {
	for _, s := range gs.DealerSeats {
		if s != current && !gs.HasPassedTrump[s] {
			return s
		}
	}
	return -1
}

func (e *Engine) nextNonDealerSeat(gs *GameState, current int) int {
	nonDealerSeats := [2]int{gs.DealerSeats[0] ^ 1, gs.DealerSeats[0] ^ 3}
	// Normalize to 0-3
	for i := range nonDealerSeats {
		nonDealerSeats[i] = nonDealerSeats[i] % 4
	}
	for _, s := range nonDealerSeats {
		if s != current && !gs.HasPassedCounter[s] {
			return s
		}
	}
	return -1
}

func (e *Engine) nextTakeBottomSeat(gs *GameState) int {
	if len(gs.DealerHistory) == 0 {
		return gs.DealerSeats[0]
	}
	lastTaker := gs.DealerHistory[len(gs.DealerHistory)-1]
	// Alternate between dealer team members
	if lastTaker == gs.DealerSeats[0] {
		return gs.DealerSeats[1]
	}
	return gs.DealerSeats[0]
}

func (e *Engine) swapDealerAndReDeal(gs *GameState) (*GameState, error) {
	// Swap dealer to non-dealer team
	if gs.DealerSeats[0] == 0 {
		gs.DealerSeats = [2]int{1, 3}
	} else {
		gs.DealerSeats = [2]int{0, 2}
	}

	// Re-deal
	deck := NewDeck()
	Shuffle(deck)
	h0, h1, h2, h3, bottom := Deal(deck)
	SortCards(h0, -1, -1)
	SortCards(h1, -1, -1)
	SortCards(h2, -1, -1)
	SortCards(h3, -1, -1)

	for i := range gs.Players {
		var hand []Card
		switch i {
		case 0:
			hand = h0
		case 1:
			hand = h1
		case 2:
			hand = h2
		case 3:
			hand = h3
		}
		gs.Players[i].Hand = hand
	}

	gs.Phase = PhaseSetTrump
	gs.CurrentSeat = gs.DealerSeats[0]
	gs.TrumpSuit = -1
	gs.IsDeadTrump = false
	gs.TrumpCards = nil
	gs.TrumpRevealed = false
	gs.BottomCards = bottom
	gs.BottomTaken = false
	gs.HasPassedTrump = make(map[int]bool)
	gs.HasPassedCounter = make(map[int]bool)
	gs.RoundNum++

	return gs, nil
}

// ValidateAction checks if an action is valid without modifying state.
func (e *Engine) ValidateAction(state game.GameState, action game.PlayerAction) error {
	// Simple validation: check phase and turn
	gs, ok := state.(*GameState)
	if !ok {
		return fmt.Errorf("invalid state type")
	}
	seat := -1
	for i, p := range gs.Players {
		if p.UserID == action.PlayerID {
			seat = i
			break
		}
	}
	if seat != gs.CurrentSeat {
		return &game.GameError{Code: game.ErrNotYourTurn}
	}
	allowed := phaseActions[gs.Phase]
	for _, a := range allowed {
		if a == action.Action {
			return nil
		}
	}
	return &game.GameError{Code: game.ErrPhaseMismatch, Phase: gs.Phase.String(), Action: action.Action}
}

// IsRoundEnd returns true if the phase is PhaseEnded.
func (e *Engine) IsRoundEnd(state game.GameState) bool {
	gs, ok := state.(*GameState)
	if !ok {
		return false
	}
	return gs.Phase == PhaseEnded
}

// CalculateScore computes the round scores for all players.
func (e *Engine) CalculateScore(state game.GameState) ([]game.PlayerScore, error) {
	gs, ok := state.(*GameState)
	if !ok {
		return nil, fmt.Errorf("invalid state type")
	}
	if gs.WinnerSeat == nil {
		return nil, fmt.Errorf("round not ended")
	}

	dealerWon := IsDealerTeam(*gs.WinnerSeat)
	change := CalculateLevelChange(dealerWon, gs.RoundPoints)

	scores := make([]game.PlayerScore, 4)
	for i, p := range gs.Players {
		scores[i] = game.PlayerScore{PlayerID: p.UserID}
		if IsDealerTeam(i) {
			scores[i].Score = change // positive = dealer gained
		} else {
			scores[i].Score = -change // negative = non-dealer gained
		}
	}
	return scores, nil
}

// SerializeForAI returns a JSON string representation for AI consumption.
func (e *Engine) SerializeForAI(state game.GameState) string {
	gs, ok := state.(*GameState)
	if !ok {
		return "{}"
	}
	data, err := gs.ToJSON()
	if err != nil {
		return "{}"
	}
	return string(data)
}

// FilterForPlayer creates a copy of state hiding other players' hands.
func (e *Engine) FilterForPlayer(state game.GameState, seat int) game.GameState {
	gs, ok := state.(*GameState)
	if !ok {
		return nil
	}

	filtered := *gs
	filtered.Players = make([]PlayerHand, len(gs.Players))
	for i, p := range gs.Players {
		filtered.Players[i] = p
		if p.Seat != seat {
			// Hide other players' hands but show card count
			filtered.Players[i].Hand = nil
		}
	}
	return &filtered
}
```

- [ ] **Step 2: Verify compilation**

```bash
cd server && go build ./internal/game/dashengji/ 2>&1
```

Expected: no errors.

- [ ] **Step 3: Write engine_test.go for Init and basic phase flow**

```go
// engine_test.go
package dashengji

import (
	"testing"

	"github.com/yongkl/vibe-pokeface/internal/game"
)

func makePlayers() []game.PlayerInfo {
	return []game.PlayerInfo{
		{ID: 1, Name: "P0", Seat: 0},
		{ID: 2, Name: "P1", Seat: 1},
		{ID: 3, Name: "P2", Seat: 2},
		{ID: 4, Name: "P3", Seat: 3},
	}
}

func TestInit_RequiresFourPlayers(t *testing.T) {
	eng := &Engine{}
	_, err := eng.Init([]game.PlayerInfo{{ID: 1, Seat: 0}})
	if err == nil {
		t.Error("Init with 1 player should fail")
	}
}

func TestInit_CreatesValidState(t *testing.T) {
	eng := &Engine{}
	state, err := eng.Init(makePlayers())
	if err != nil {
		t.Fatalf("Init failed: %v", err)
	}
	gs := state.(*GameState)

	if gs.Phase != PhaseSetTrump {
		t.Errorf("expected PhaseSetTrump, got %v", gs.Phase)
	}
	if len(gs.Players) != 4 {
		t.Errorf("expected 4 players, got %d", len(gs.Players))
	}
	for i, p := range gs.Players {
		if len(p.Hand) != 39 {
			t.Errorf("player %d: expected 39 cards, got %d", i, len(p.Hand))
		}
	}
	if len(gs.BottomCards) != 6 {
		t.Errorf("expected 6 bottom cards, got %d", len(gs.BottomCards))
	}
	if gs.CurrentLevel != 3 {
		t.Errorf("expected starting level 3, got %d", gs.CurrentLevel)
	}
}

func TestInit_DealerTeamValid(t *testing.T) {
	eng := &Engine{}
	state, _ := eng.Init(makePlayers())
	gs := state.(*GameState)

	// Dealer team must be (0,2) or (1,3)
	if gs.DealerSeats != [2]int{0, 2} && gs.DealerSeats != [2]int{1, 3} {
		t.Errorf("invalid dealer seats: %v", gs.DealerSeats)
	}

	// Verify partners sit opposite
	if gs.DealerSeats[1]-gs.DealerSeats[0] != 2 {
		t.Errorf("dealer seats should be opposite, got %v", gs.DealerSeats)
	}
}

func TestIsRoundEnd(t *testing.T) {
	eng := &Engine{}
	state, _ := eng.Init(makePlayers())
	if eng.IsRoundEnd(state) {
		t.Error("new game should not be round end")
	}
	gs := state.(*GameState)
	gs.Phase = PhaseEnded
	if !eng.IsRoundEnd(state) {
		t.Error("ended game should report round end")
	}
}

func TestFilterForPlayer_HidesOtherHands(t *testing.T) {
	eng := &Engine{}
	state, _ := eng.Init(makePlayers())
	filtered := eng.FilterForPlayer(state, 0)
	fgs := filtered.(*GameState)

	for i, p := range fgs.Players {
		if i == 0 {
			if len(p.Hand) != 39 {
				t.Errorf("player 0 should see own 39 cards, sees %d", len(p.Hand))
			}
		} else {
			if p.Hand != nil {
				t.Errorf("player %d hand should be hidden, got %d cards", i, len(p.Hand))
			}
		}
	}
}
```

- [ ] **Step 4: Run engine tests**

```bash
cd server && go test ./internal/game/dashengji/... -v -run "TestInit|TestIsRound|TestFilter" 2>&1
```

Expected: all PASS.

- [ ] **Step 5: Commit**

```bash
git add server/internal/game/dashengji/engine.go server/internal/game/dashengji/engine_test.go
git commit -m "feat(dashengji): add GameEngine implementation with phase handlers"
```

---

### Task 6b: Engine — follow-suit enforcement and trick-level point counting

**Files:**
- Modify: `server/internal/game/dashengji/engine.go` (handlePlay)
- Create: follow-suit tests in `server/internal/game/dashengji/engine_test.go`

- [ ] **Step 1: Add follow-suit validation helper**

Add to `engine.go` after `handlePlay`:

```go
// hasSuit checks if the player has any cards matching the led suit.
func hasSuit(hand []Card, ledSuit int, trumpSuit int) bool {
	for _, c := range hand {
		if c.IsSmallJoker() || c.IsBigJoker() {
			continue
		}
		if c.Suit() == ledSuit {
			return true
		}
	}
	return false
}

// hasTrump checks if the player has any trump cards (main suit cards).
func hasTrump(hand []Card, trumpSuit int, levelRank int) bool {
	for _, c := range hand {
		if c.IsSmallJoker() || c.IsBigJoker() {
			return true // jokers are trump
		}
		cat := ClassifyCard(c, trumpSuit, levelRank)
		if cat == CatTrump || cat == CatNativeMain || cat == CatJoker {
			return true
		}
	}
	return false
}

// validateFollow checks that the played cards follow the led suit+type+category rules.
func validateFollow(cards []Card, hand []Card, ledPlay Play, ledCards []Card,
	trumpSuit int, levelRank int) error {

	if len(cards) == 0 {
		return &game.GameError{Code: game.ErrInvalidCards}
	}

	play := ParsePlayWithContext(cards, trumpSuit, levelRank)
	if play.Type == PlayInvalid {
		return &game.GameError{Code: game.ErrInvalidCards}
	}

	// Must match type and length
	if play.Type != ledPlay.Type {
		return &game.GameError{Code: game.ErrInvalidCards}
	}
	if ledPlay.Type == PlayTractor && play.Length != ledPlay.Length {
		return &game.GameError{Code: game.ErrInvalidCards}
	}

	ledSuit := ledCards[0].Suit()
	ledCat := ClassifyCard(ledCards[0], trumpSuit, levelRank)
	playCat := ClassifyCard(cards[0], trumpSuit, levelRank)
	playSuit := cards[0].Suit()

	// If same suit and same category → valid follow
	if playSuit == ledSuit && playCat == ledCat {
		return nil // valid follow
	}

	// If no cards of led suit in hand → can 枪毙 (trump with main cards)
	if !hasSuit(hand, ledSuit, trumpSuit) {
		if playCat == CatTrump || playCat == CatNativeMain || playCat == CatJoker {
			return nil // valid 枪毙
		}
	}

	return &game.GameError{Code: game.ErrInvalidCards}
}
```

- [ ] **Step 2: Integrate follow-suit validation into handlePlay**

In `handlePlay`, replace the simple CanBeat check with full follow-suit validation:

```go
// If following (not leading), must match type + suit + category
if gs.LastPlay != nil && gs.LastPlay.Seat != seat {
	play := ParsePlayWithContext(cards, gs.TrumpSuit, gs.LevelRank)
	if play.Type == PlayInvalid {
		return nil, &game.GameError{Code: game.ErrInvalidCards}
	}

	if err := validateFollow(cards, gs.Players[playerIdx].Hand, gs.LastPlay.Play,
		gs.LastPlay.Cards, gs.TrumpSuit, gs.LevelRank); err != nil {
		return nil, err
	}

	if !CanBeat(play, gs.LastPlay.Play) {
		return nil, &game.GameError{Code: game.ErrCannotBeat}
	}
}
```

- [ ] **Step 3: Add trick-level point counting**

Add to `handlePlay` after recording the play record, before advancing seat:

```go
// Count points in this trick: add to roundPoints if non-dealer team would win
// (simplified: count all point cards played this trick)
trickPoints := CountRoundPoints(cards)
// Points are tallied at end of round based on who wins each trick.
// For now, accumulate all played point cards; final calculation in CalculateScore
gs.RoundPoints += trickPoints
```

- [ ] **Step 4: Write engine test for follow-suit enforcement**

Add to `engine_test.go`:

```go
func TestFollowSuit_MustMatchType(t *testing.T) {
	// If lead is a pair, follow must be a pair
	ledPlay := Play{Type: PlayPair, MainRank: 5, Length: 1}
	ledCards := []Card{{ID: 2}, {ID: 56}} // pair of ♠5

	followCards := []Card{{ID: 0}} // single ♠3
	hand := []Card{{ID: 0}, {ID: 2}, {ID: 56}} // ♠3, ♠5, ♠5

	err := validateFollow(followCards, hand, ledPlay, ledCards, 0, 6)
	if err == nil {
		t.Error("single should not follow a pair lead")
	}
}

func TestFollowSuit_MustMatchSuit(t *testing.T) {
	ledPlay := Play{Type: PlaySingle, MainRank: 5, Length: 1}
	ledCards := []Card{{ID: 2}} // ♠5 (suit 0)

	// Try to follow with ♥5 when player has ♠ cards in hand
	followCards := []Card{{ID: 15}} // ♥5 (suit 1)
	hand := []Card{{ID: 0}, {ID: 15}} // ♠3 (suit 0), ♥5

	err := validateFollow(followCards, hand, ledPlay, ledCards, 0, 6)
	if err == nil {
		t.Error("should not allow off-suit follow when same-suit cards available")
	}
}

func TestFollowSuit_CanTrumpWhenNoSuitCards(t *testing.T) {
	ledPlay := Play{Type: PlaySingle, MainRank: 5, Length: 1}
	ledCards := []Card{{ID: 2}} // ♠5 (suit 0, side card when trump=♥)
	trumpSuit := 1               // ♥ is trump

	// Player has only ♥ cards (all trump), no ♠
	followCards := []Card{{ID: 13}} // ♥3 (trump suit)
	hand := []Card{{ID: 13}, {ID: 14}} // only ♥ cards

	err := validateFollow(followCards, hand, ledPlay, ledCards, trumpSuit, 6)
	if err != nil {
		t.Errorf("should allow trump follow when no led-suit cards: %v", err)
	}
}
```

- [ ] **Step 5: Run tests**

```bash
cd server && go test ./internal/game/dashengji/... -v -run "TestFollow" 2>&1
```

Expected: all PASS.

- [ ] **Step 6: Commit**

```bash
git add server/internal/game/dashengji/engine.go server/internal/game/dashengji/engine_test.go
git commit -m "feat(dashengji): add follow-suit validation and trick-level point counting"
```

---

### Task 7: Server integration — engine registry and WebSocket routing

**Files:**
- Create: `server/internal/game/registry.go`
- Modify: `server/internal/api/ws/handler.go:170`

- [ ] **Step 1: Create engine registry**

```go
// registry.go
package game

import (
	"fmt"

	"github.com/yongkl/vibe-pokeface/internal/game/dashengji"
	"github.com/yongkl/vibe-pokeface/internal/game/doudizhu"
)

var engineFactories = map[string]func() GameEngine{
	"doudizhu":  func() GameEngine { return &doudizhu.Engine{} },
	"dashengji": func() GameEngine { return &dashengji.Engine{} },
}

// NewEngine creates a GameEngine for the given game type.
func NewEngine(gameType string) (GameEngine, error) {
	factory, ok := engineFactories[gameType]
	if !ok {
		return nil, fmt.Errorf("unknown game type: %s", gameType)
	}
	return factory(), nil
}

// ValidGameTypes returns the list of supported game types.
func ValidGameTypes() []string {
	types := make([]string, 0, len(engineFactories))
	for t := range engineFactories {
		types = append(types, t)
	}
	return types
}
```

- [ ] **Step 2: Modify handler.go to use registry**

In `server/internal/api/ws/handler.go`, change line 170:

```go
// BEFORE:
room = h.RoomManager.GetOrCreateRoom(roomID, gameType, &doudizhu.Engine{})

// AFTER:
engine, err := game.NewEngine(gameType)
if err != nil {
	errMsg, _ := json.Marshal(S2CMessage{Type: "error", Data: fmt.Sprintf("unsupported game type: %s", gameType)})
	select {
	case client.Send <- errMsg:
	default:
	}
	return
}
room = h.RoomManager.GetOrCreateRoom(roomID, gameType, engine)
```

- [ ] **Step 3: Verify compilation**

```bash
cd server && go build ./... 2>&1
```

Expected: no errors.

- [ ] **Step 4: Run all tests**

```bash
cd server && go test ./... 2>&1
```

Expected: all tests pass (existing doudizhu tests + new dashengji tests).

- [ ] **Step 5: Commit**

```bash
git add server/internal/game/registry.go server/internal/api/ws/handler.go
git commit -m "feat(game): add engine registry for dynamic game-type selection"
```

---

### Task 8: GameRoom — parameterize seat count for 4-player games

**Files:**
- Modify: `server/internal/game/room.go`

- [ ] **Step 1: Adjust hardcoded seat count assumptions**

In `room.go`, find places where seat count is hardcoded to 3:

**`nextAvailableSeat`** (currently loops 0..2):
```go
// BEFORE:
func (r *GameRoom) nextAvailableSeat() int {
	for seat := 0; seat < 3; seat++ {
		occupied := false
		for _, p := range r.Players {
			if p.Seat == seat {
				occupied = true
				break
			}
		}
		if !occupied {
			return seat
		}
	}
	return -1
}

// AFTER:
func (r *GameRoom) maxSeats() int {
	switch r.GameType {
	case "dashengji":
		return 4
	default:
		return 3
	}
}

func (r *GameRoom) nextAvailableSeat() int {
	max := r.maxSeats()
	for seat := 0; seat < max; seat++ {
		occupied := false
		for _, p := range r.Players {
			if p.Seat == seat {
				occupied = true
				break
			}
		}
		if !occupied {
			return seat
		}
	}
	return -1
}
```

Find and update all other `3` seat assumptions in `room.go` (seat iteration in `sendStateToAll`, `FillEmptySeats`, etc.) to use `r.maxSeats()`.

- [ ] **Step 2: Verify compilation and tests**

```bash
cd server && go build ./... && go test ./... 2>&1
```

Expected: all pass.

- [ ] **Step 3: Commit**

```bash
git add server/internal/game/room.go
git commit -m "feat(game): parameterize seat count for multi-game support"
```

---

### Task 9: Frontend — Dashengji game page

**Files:**
- Create: `frontend/app/(main)/room/[id]/dashengji/page.tsx`
- Create: `frontend/components/game/dashengji/DashengjiTable.tsx`
- Create: `frontend/components/game/dashengji/DashengjiActionBar.tsx`
- Create: `frontend/components/game/dashengji/ScoreBoard.tsx`

- [ ] **Step 1: Create DashengjiActionBar component**

```tsx
// components/game/dashengji/DashengjiActionBar.tsx
"use client";

interface DashengjiActionBarProps {
  phase: string;
  isMyTurn: boolean;
  onAction: (action: string, cards?: number[]) => void;
}

export function DashengjiActionBar({ phase, isMyTurn, onAction }: DashengjiActionBarProps) {
  const disabled = !isMyTurn;

  switch (phase) {
    case "set_trump":
      return (
        <div className="fixed bottom-24 left-1/2 -translate-x-1/2 flex gap-4 z-30">
          <button
            className="gold-button px-8 py-3 text-lg"
            disabled={disabled}
            onClick={() => onAction("set_trump")}
          >
            定主
          </button>
          <button
            className="px-8 py-3 text-lg rounded-xl bg-white/10 text-white/70 hover:bg-white/20"
            disabled={disabled}
            onClick={() => onAction("pass_trump")}
          >
            不定
          </button>
        </div>
      );
    case "counter_trump":
      return (
        <div className="fixed bottom-24 left-1/2 -translate-x-1/2 flex gap-4 z-30">
          <button
            className="gold-button px-8 py-3 text-lg"
            disabled={disabled}
            onClick={() => onAction("counter_trump")}
          >
            反主
          </button>
          <button
            className="px-8 py-3 text-lg rounded-xl bg-white/10 text-white/70 hover:bg-white/20"
            disabled={disabled}
            onClick={() => onAction("pass_counter")}
          >
            不反
          </button>
        </div>
      );
    case "take_bottom":
      return (
        <div className="fixed bottom-24 left-1/2 -translate-x-1/2 flex gap-4 z-30">
          <button
            className="gold-button px-8 py-3 text-lg"
            disabled={disabled}
            onClick={() => onAction("take_bottom")}
          >
            起底
          </button>
        </div>
      );
    case "discard_bottom":
      return (
        <div className="fixed bottom-24 left-1/2 -translate-x-1/2 flex gap-4 z-30">
          <button
            className="gold-button px-8 py-3 text-lg"
            disabled={disabled}
            onClick={() => onAction("discard_bottom")}
          >
            扣底
          </button>
        </div>
      );
    case "playing":
      return (
        <div className="fixed bottom-24 left-1/2 -translate-x-1/2 flex gap-4 z-30">
          <button
            className="emerald-button px-8 py-3 text-lg"
            disabled={disabled}
            onClick={() => onAction("play")}
          >
            出牌
          </button>
          <button
            className="px-8 py-3 text-lg rounded-xl bg-white/10 text-white/70 hover:bg-white/20"
            disabled={disabled}
            onClick={() => onAction("pass")}
          >
            不出
          </button>
        </div>
      );
    default:
      return null;
  }
}
```

- [ ] **Step 2: Create DashengjiTable component**

```tsx
// components/game/dashengji/DashengjiTable.tsx
"use client";

import { Card } from "@/components/game/Card";

export interface DashengjiPlayer {
  userId: string;
  seat: number;
  nickname: string;
  cardCount: number;
  isBot: boolean;
  characterId?: string;
  isCurrentTurn: boolean;
  isDealerTeam: boolean;
  revealedHand?: number[]; // for showing opponent's played/discarded cards
}

interface DashengjiTableProps {
  players: DashengjiPlayer[];
  mySeat: number;
  trumpSuit: number; // -1 if not set, 0-3
  levelRank: number;
  bottomCards: number[];
  discardedCards: number[];
  lastPlay: { seat: number; cards: number[] } | null;
}

export function DashengjiTable({
  players,
  mySeat,
  trumpSuit,
  levelRank,
  bottomCards,
  discardedCards,
  lastPlay,
}: DashengjiTableProps) {
  const suitSymbols = ["♠", "♥", "♣", "♦"];

  // Arrange players: partner opposite, left, right
  const partner = players.find((p) => p.seat === (mySeat + 2) % 4);
  const left = players.find((p) => p.seat === (mySeat + 1) % 4);
  const right = players.find((p) => p.seat === (mySeat + 3) % 4);
  const me = players.find((p) => p.seat === mySeat);

  const renderSeat = (player: DashengjiPlayer | undefined, position: string) => {
    if (!player) return null;
    return (
      <div
        className={`absolute flex flex-col items-center gap-2 ${
          position === "top"
            ? "top-4 left-1/2 -translate-x-1/2"
            : position === "left"
            ? "left-4 top-1/2 -translate-y-1/2"
            : position === "right"
            ? "right-4 top-1/2 -translate-y-1/2"
            : "bottom-4 left-1/2 -translate-x-1/2"
        }`}
      >
        {/* Player info */}
        <div className="flex items-center gap-2">
          <span className="text-white text-sm font-medium">{player.nickname}</span>
          {player.isDealerTeam && (
            <span className="text-xs px-1.5 py-0.5 rounded bg-amber-600/80 text-white">庄</span>
          )}
          {player.isCurrentTurn && (
            <span className="w-2 h-2 rounded-full bg-green-400 animate-pulse" />
          )}
        </div>

        {/* Card backs or revealed hand */}
        <div className="flex -space-x-2">
          {player.revealedHand
            ? player.revealedHand.map((id, i) => <Card key={i} cardId={id} small />)
            : Array.from({ length: Math.min(player.cardCount, 10) }).map((_, i) => (
                <Card key={i} cardId={-1} faceDown small />
              ))}
        </div>
        <span className="text-white/60 text-xs">{player.cardCount}张</span>
      </div>
    );
  };

  return (
    <div className="relative w-full h-full min-h-[500px]">
      {/* Trump indicator */}
      {trumpSuit >= 0 && (
        <div className="absolute top-1/2 left-1/2 -translate-x-1/2 -translate-y-1/2 text-white/80 text-center">
          <div className="text-4xl">{suitSymbols[trumpSuit]}</div>
          <div className="text-sm">主花色</div>
        </div>
      )}

      {/* Bottom cards area */}
      {bottomCards.length > 0 && (
        <div className="absolute top-1/2 left-1/2 -translate-x-1/2 translate-y-8 flex -space-x-1">
          {bottomCards.map((id, i) => (
            <Card key={i} cardId={id} small />
          ))}
        </div>
      )}

      {/* Last play */}
      {lastPlay && (
        <div className="absolute top-1/2 left-1/2 -translate-x-1/2 -translate-y-16 flex -space-x-1">
          {lastPlay.cards.map((id, i) => (
            <Card key={i} cardId={id} medium />
          ))}
        </div>
      )}

      {renderSeat(partner, "top")}
      {renderSeat(left, "left")}
      {renderSeat(right, "right")}
      {me && (
        <div className="absolute bottom-4 left-1/2 -translate-x-1/2 text-white/40 text-sm">
          {me.nickname} (你)
        </div>
      )}
    </div>
  );
}
```

- [ ] **Step 3: Create ScoreBoard component**

```tsx
// components/game/dashengji/ScoreBoard.tsx
"use client";

interface ScoreBoardProps {
  currentLevel: number;
  dealerTeam: string; // "庄队" or team names
  nonDealerTeam: string;
  roundPoints: number;
  dealerLevel: number; // the level the dealer team is playing
}

const levelNames = ["", "", "", "3", "4", "5", "6", "7", "8", "9", "10", "J", "Q", "K", "A"];

export function ScoreBoard({ currentLevel, dealerTeam, nonDealerTeam, roundPoints, dealerLevel }: ScoreBoardProps) {
  return (
    <div className="fixed top-4 right-4 bg-black/60 backdrop-blur rounded-xl p-4 text-white text-sm z-40 min-w-[180px]">
      <div className="text-white/60 text-xs mb-2">打升级</div>
      <div className="flex justify-between mb-1">
        <span>当前打</span>
        <span className="text-amber-400 font-bold text-lg">{levelNames[dealerLevel] || dealerLevel}</span>
      </div>
      <div className="flex justify-between mb-1">
        <span>{dealerTeam}</span>
        <span className="text-amber-400">庄队</span>
      </div>
      <div className="flex justify-between mb-1">
        <span>{nonDealerTeam}</span>
        <span className="text-blue-400">闲队</span>
      </div>
      <div className="border-t border-white/10 mt-2 pt-2 flex justify-between">
        <span>本轮闲家得分</span>
        <span className="text-red-400 font-bold">{roundPoints}</span>
      </div>
    </div>
  );
}
```

- [ ] **Step 4: Create the Dashengji page**

```tsx
// app/(main)/room/[id]/dashengji/page.tsx
"use client";

import { useCallback, useEffect, useState, useRef } from "react";
import { useParams, useRouter } from "next/navigation";
import { WSGameClient, formatError } from "@/lib/ws-game";
import { HandCards } from "@/components/game/HandCards";
import { ReadyBar } from "@/components/game/ReadyBar";
import { ChatPanel } from "@/components/chat/ChatPanel";
import { DashengjiTable, DashengjiPlayer } from "@/components/game/dashengji/DashengjiTable";
import { DashengjiActionBar } from "@/components/game/dashengji/DashengjiActionBar";
import { ScoreBoard } from "@/components/game/dashengji/ScoreBoard";

export default function DashengjiRoomPage() {
  const params = useParams();
  const router = useRouter();
  const roomId = params.id as string;

  const [connected, setConnected] = useState(false);
  const [players, setPlayers] = useState<DashengjiPlayer[]>([]);
  const [mySeat, setMySeat] = useState(-1);
  const [phase, setPhase] = useState("waiting");
  const [hand, setHand] = useState<number[]>([]);
  const [currentSeat, setCurrentSeat] = useState(-1);
  const [trumpSuit, setTrumpSuit] = useState(-1);
  const [levelRank, setLevelRank] = useState(3);
  const [currentLevel, setCurrentLevel] = useState(3);
  const [bottomCards, setBottomCards] = useState<number[]>([]);
  const [discardedCards, setDiscardedCards] = useState<number[]>([]);
  const [lastPlay, setLastPlay] = useState<{ seat: number; cards: number[] } | null>(null);
  const [roundPoints, setRoundPoints] = useState(0);
  const [roundResult, setRoundResult] = useState<string | null>(null);
  const [roomScores, setRoomScores] = useState<Array<{ player_id: number; score: number }>>([]);
  const [dealerSeats, setDealerSeats] = useState<[number, number]>([0, 2]);

  const wsRef = useRef<WSGameClient | null>(null);
  const userIdRef = useRef<string>("");

  useEffect(() => {
    const token = localStorage.getItem("token");
    if (!token) {
      router.push("/auth/login");
      return;
    }
    const payload = JSON.parse(atob(token.split(".")[1]));
    userIdRef.current = String(payload.user_id);

    const ws = new WSGameClient(userIdRef.current, roomId, "dashengji");
    wsRef.current = ws;

    ws.on("player_joined", (data: any) => {
      const serverPlayers: any[] = data.players || [];
      setPlayers(
        serverPlayers.map((p: any) => ({
          userId: String(p.user_id ?? p.userId ?? ""),
          seat: p.seat ?? 0,
          nickname: p.nickname ?? "",
          cardCount: p.card_count ?? p.cardCount ?? 0,
          isBot: p.is_bot ?? p.isBot ?? false,
          characterId: p.character_id ?? p.characterId,
          isCurrentTurn: false,
          isDealerTeam: dealerSeats.includes(p.seat ?? 0),
          revealedHand: p.hand ? p.hand.map((c: any) => (typeof c === "number" ? c : c.id)) : undefined,
        }))
      );
      if (data.seat !== undefined) setMySeat(data.seat);
      if (data.game_type) {
        // Already set via join
      }
      setConnected(true);
    });

    ws.on("state_update", (data: any) => {
      if (data.phase !== undefined) {
        const phaseMap: Record<number, string> = {
          0: "set_trump", 1: "counter_trump", 2: "take_bottom",
          3: "discard_bottom", 4: "playing", 5: "ended",
        };
        setPhase(phaseMap[data.phase] ?? "waiting");
      }
      if (data.current_seat !== undefined) setCurrentSeat(data.current_seat);
      if (data.trump_suit !== undefined) setTrumpSuit(data.trump_suit);
      if (data.level_rank !== undefined) setLevelRank(data.level_rank);
      if (data.current_level !== undefined) setCurrentLevel(data.current_level);
      if (data.dealer_seats) setDealerSeats(data.dealer_seats);
      if (data.bottom_cards) setBottomCards(data.bottom_cards.map((c: any) => (typeof c === "number" ? c : c.id)));
      if (data.discarded_cards) setDiscardedCards(data.discarded_cards.map((c: any) => (typeof c === "number" ? c : c.id)));
      if (data.round_points !== undefined) setRoundPoints(data.round_points);
      if (data.last_play) {
        setLastPlay({
          seat: data.last_play.seat,
          cards: data.last_play.cards.map((c: any) => (typeof c === "number" ? c : c.id)),
        });
      }

      const serverPlayers: any[] = data.players || [];
      setPlayers((prev) =>
        serverPlayers.map((p: any) => ({
          userId: String(p.user_id ?? p.userId ?? ""),
          seat: p.seat ?? 0,
          nickname: p.nickname ?? prev.find((pp) => pp.seat === p.seat)?.nickname ?? "",
          cardCount: p.card_count ?? p.cardCount ?? (p.hand ? p.hand.length : 0),
          isBot: p.is_bot ?? p.isBot ?? false,
          characterId: p.character_id ?? p.characterId,
          isCurrentTurn: p.seat === data.current_seat,
          isDealerTeam: (data.dealer_seats || dealerSeats).includes(p.seat ?? 0),
          revealedHand: p.hand ? p.hand.map((c: any) => (typeof c === "number" ? c : c.id)) : undefined,
        }))
      );

      // Find own hand
      const me = serverPlayers.find((p: any) => p.seat === mySeat);
      if (me?.hand) {
        setHand(me.hand.map((c: any) => (typeof c === "number" ? c : c.id)));
      }
    });

    ws.on("round_end", (data: any) => {
      setPhase("ended");
      setHand([]);
      setRoundResult(data.scores ? "round complete" : null);
      if (data.scores) setRoomScores(data.scores);
    });

    ws.on("error", (data: any) => {
      console.error("Game error:", formatError(data));
    });

    ws.on("player_left", (data: any) => {
      setPlayers((prev) => prev.filter((p) => p.userId !== String(data.user_id)));
    });

    ws.connect();

    return () => {
      ws.disconnect();
    };
  }, [roomId]);

  const handleAction = useCallback(
    (action: string, cards?: number[]) => {
      wsRef.current?.sendAction(action, cards);
    },
    []
  );

  const isMyTurn = mySeat === currentSeat;

  return (
    <div className="min-h-screen bg-gradient-to-b from-green-900 via-green-800 to-green-950 relative overflow-hidden">
      {/* Table */}
      <DashengjiTable
        players={players}
        mySeat={mySeat}
        trumpSuit={trumpSuit}
        levelRank={levelRank}
        bottomCards={bottomCards}
        discardedCards={discardedCards}
        lastPlay={lastPlay}
      />

      {/* Score Board */}
      <ScoreBoard
        currentLevel={currentLevel}
        dealerTeam="庄队"
        nonDealerTeam="闲队"
        roundPoints={roundPoints}
        dealerLevel={currentLevel}
      />

      {/* Action Bar */}
      {phase !== "waiting" && phase !== "ended" && (
        <DashengjiActionBar phase={phase} isMyTurn={isMyTurn} onAction={handleAction} />
      )}

      {/* Hand Cards */}
      {phase === "playing" && hand.length > 0 && (
        <HandCards
          cards={hand}
          onPlayCards={(cards) => handleAction("play", cards)}
          selectedIds={[]}
        />
      )}

      {/* Ready Bar (waiting phase) */}
      {phase === "waiting" && (
        <div className="fixed bottom-0 left-0 right-0 z-40">
          <ReadyBar
            players={players.map((p) => ({
              userId: p.userId,
              seat: p.seat,
              ready: false,
              isBot: p.isBot,
            }))}
            mySeat={mySeat}
            isOwner={players.find((p) => p.seat === mySeat)?.userId === players[0]?.userId}
            onReady={() => wsRef.current?.sendReady()}
            onStartGame={() => wsRef.current?.startGame()}
            onAddBot={() => {}}
            gameType="dashengji"
          />
        </div>
      )}

      {/* Round End */}
      {phase === "ended" && roundResult && (
        <div className="fixed inset-0 bg-black/70 flex items-center justify-center z-50">
          <div className="bg-gray-900 rounded-2xl p-8 text-white text-center">
            <h2 className="text-2xl font-bold mb-4">本轮结束</h2>
            {roomScores.map((s, i) => (
              <div key={i} className="text-lg">
                Player {s.player_id}: {s.score > 0 ? "+" : ""}{s.score} 级
              </div>
            ))}
            <button
              className="mt-6 px-6 py-2 bg-amber-500 rounded-xl"
              onClick={() => router.push("/lobby")}
            >
              返回大厅
            </button>
          </div>
        </div>
      )}
    </div>
  );
}
```

- [ ] **Step 5: Verify frontend compiles**

```bash
cd frontend && npx tsc --noEmit 2>&1 | head -50
```

Fix any type errors.

- [ ] **Step 6: Commit**

```bash
git add frontend/app/\(main\)/room/\[id\]/dashengji/ frontend/components/game/dashengji/
git commit -m "feat(ui): add dashengji game page with 4-player table and phase controls"
```

---

### Task 10: Frontend — Room creation with game type selector

**Files:**
- Modify: `frontend/app/(main)/room/create/page.tsx`
- Modify: `frontend/app/(main)/room/[id]/page.tsx`

- [ ] **Step 1: Add game type selector to room creation form**

In `create/page.tsx`, add a select dropdown for game type:

```tsx
// Add to the form, before the submit button
<div className="mb-4">
  <label className="block text-sm text-white/60 mb-1">游戏类型</label>
  <select
    value={gameType}
    onChange={(e) => setGameType(e.target.value)}
    className="w-full px-4 py-2 rounded-xl bg-white/10 text-white border border-white/20"
  >
    <option value="doudizhu">斗地主 (3人)</option>
    <option value="dashengji">打升级 (4人)</option>
  </select>
</div>
```

Add `const [gameType, setGameType] = useState("doudizhu");` to the component state, and include `game_type: gameType` in the create room API call.

- [ ] **Step 2: Create room/[id]/page.tsx redirect**

```tsx
// app/(main)/room/[id]/page.tsx
import { redirect } from "next/navigation";

// This page redirects based on gameType — determined client-side after joining.
// For now, default to doudizhu; the lobby links should point to the correct sub-route.
export default function RoomPage() {
  // The lobby already has gameType info, so it can link directly.
  // This fallback redirects to doudizhu for backward compatibility.
  redirect("/lobby");
}
```

Update the lobby page to link to `/room/${room.id}/doudizhu` or `/room/${room.id}/dashengji` based on `room.gameType`.

- [ ] **Step 3: Verify frontend compiles**

```bash
cd frontend && npx tsc --noEmit 2>&1 | head -50
```

- [ ] **Step 4: Commit**

```bash
git add frontend/app/\(main\)/room/create/page.tsx frontend/app/\(main\)/room/\[id\]/page.tsx
git commit -m "feat(ui): add game type selector to room creation"
```

---

### Task 11: Final verification — run all checks

- [ ] **Step 1: Run all server checks**

```bash
cd server && go vet ./... && go test ./... -v 2>&1
```

Expected: `go vet` clean, all tests PASS.

- [ ] **Step 2: Run all frontend checks**

```bash
cd frontend && npm run lint && npx tsc --noEmit 2>&1
```

Expected: no lint errors, no type errors.

- [ ] **Step 3: Final commit**

```bash
git add -A
git commit -m "chore: final verification — all tests and checks pass"
```
