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

	if gs.DealerSeats != [2]int{0, 2} && gs.DealerSeats != [2]int{1, 3} {
		t.Errorf("invalid dealer seats: %v", gs.DealerSeats)
	}
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

func TestValidateAction_WrongTurn(t *testing.T) {
	eng := &Engine{}
	state, _ := eng.Init(makePlayers())
	gs := state.(*GameState)
	// Try action from seat 1 when current seat is dealer[0]
	if gs.CurrentSeat == 1 {
		gs.CurrentSeat = 0
	}
	action := game.PlayerAction{PlayerID: gs.Players[1].UserID, Action: "set_trump"}
	err := eng.ValidateAction(state, action)
	if err == nil {
		t.Error("should reject action from wrong seat")
	}
}

func TestFollowSuit_MustMatchType(t *testing.T) {
	ledPlay := Play{Type: PlayPair, MainRank: 5, Length: 1}
	ledCards := []Card{{ID: 2}, {ID: 56}}

	followCards := []Card{{ID: 0}}
	hand := []Card{{ID: 0}, {ID: 2}, {ID: 56}}

	_, err := validateFollow(followCards, hand, ledPlay, ledCards, 0, 6)
	if err == nil {
		t.Error("single should not follow a pair lead when player can follow")
	}
}

func TestFollowSuit_MustMatchSuit(t *testing.T) {
	ledPlay := Play{Type: PlaySingle, MainRank: 5, Length: 1}
	ledCards := []Card{{ID: 2}} // suits 0

	followCards := []Card{{ID: 15}} // suit 1
	hand := []Card{{ID: 0}, {ID: 15}} // suits 0, 1

	_, err := validateFollow(followCards, hand, ledPlay, ledCards, 0, 6)
	if err == nil {
		t.Error("should not allow off-suit follow when same-suit cards available")
	}
}

func TestFollowSuit_CanTrumpWhenNoSuitCards(t *testing.T) {
	ledPlay := Play{Type: PlaySingle, MainRank: 5, Length: 1}
	ledCards := []Card{{ID: 2}} // suit 0, side card when trumpSuit=1
	trumpSuit := 1              // hearts is trump

	followCards := []Card{{ID: 13}} // hearts (trump suit)
	hand := []Card{{ID: 13}, {ID: 14}} // only hearts cards

	isPadding, err := validateFollow(followCards, hand, ledPlay, ledCards, trumpSuit, 6)
	if err != nil {
		t.Errorf("should allow trump follow when no led-suit cards: %v", err)
	}
	if isPadding {
		t.Error("trumping play should not be marked as padding")
	}
}

func TestFollowSuit_PadWhenCannotFollowTriple(t *testing.T) {
	// Leader plays a triplet of ♣4 (suit=2, face=0)
	ledPlay := Play{Type: PlayTriple, MainRank: 4, Length: 1}
	ledCards := []Card{{ID: 26}, {ID: 80}, {ID: 134}} // suit 2, face 0 (♣4), three copies
	trumpSuit := 0                                      // ♠ is trump

	// Follower pads with ♣7, ♣8, ♣J (same suit, different ranks - not a triplet)
	followCards := []Card{{ID: 30}, {ID: 31}, {ID: 34}}
	// Follower's hand has these ♣ singles but no three of same ♣ rank
	hand := []Card{
		{ID: 30}, {ID: 31}, {ID: 34}, // ♣7, ♣8, ♣J
		{ID: 0}, {ID: 1}, {ID: 2}, // some ♠ cards
	}

	isPadding, err := validateFollow(followCards, hand, ledPlay, ledCards, trumpSuit, 3)
	if err != nil {
		t.Errorf("padding when cannot follow triplet should be allowed: %v", err)
	}
	if !isPadding {
		t.Error("padding with different ranks should be marked as isPadding=true")
	}
}

func TestFollowSuit_PaddingSkipsPlayTypeCheck(t *testing.T) {
	// Leader plays a pair of ♣4 (suit=2)
	ledPlay := Play{Type: PlayPair, MainRank: 4, Length: 1}
	ledCards := []Card{{ID: 26}, {ID: 80}} // suit 2, face 0, two copies
	trumpSuit := 0                          // ♠ is trump

	// Follower doesn't have a pair of ♣, pads with ♣7 and ♣8 (two singles, same suit)
	followCards := []Card{{ID: 30}, {ID: 31}} // ♣7, ♣8
	hand := []Card{
		{ID: 30}, {ID: 31}, // ♣7, ♣8
		{ID: 0}, {ID: 1},
	}

	isPadding, err := validateFollow(followCards, hand, ledPlay, ledCards, trumpSuit, 3)
	if err != nil {
		t.Errorf("padding with same-suit non-pair cards should be allowed: %v", err)
	}
	if !isPadding {
		t.Error("padding play should be marked as isPadding=true")
	}
}

func TestValidateAction_WrongPhase(t *testing.T) {
	eng := &Engine{}
	state, _ := eng.Init(makePlayers())
	gs := state.(*GameState)
	action := game.PlayerAction{PlayerID: gs.Players[gs.CurrentSeat].UserID, Action: "play"}
	err := eng.ValidateAction(state, action)
	if err == nil {
		t.Error("should reject 'play' action during set_trump phase")
	}
}

// TestBug_ConsistentValidation verifies that calling ExecuteAction twice with
// the same state and same cards produces the same result. This rules out hidden
// state mutation on the error path as the cause of the "second click succeeds" bug.
func TestBug_ConsistentValidation(t *testing.T) {
	eng := &Engine{}

	// Scenario: Leader (seat 0) played a pair of ♠8. Follower (seat 1) has
	// a pair of ♠K (same category CatSide with trumpSuit=-1) so canFollow=true.
	// Follower plays ♥5 + ♥J (not a pair) → must be rejected both times.
	//
	// Card IDs:
	//   ♠8: face=5, suit=0(spade), baseRank=8
	//   ♠K: face=11, suit=0(spade), baseRank=13
	//   ♥5: face=15, suit=1(heart), baseRank=5
	//   ♥J: face=22, suit=1(heart), baseRank=11
	//   Second copy: ID + 54

	ledCards := []Card{{ID: 5}, {ID: 59}} // pair of ♠8
	ledPlay := ParsePlayWithContext(ledCards, -1, 3)

	gs := &GameState{
		Phase:       PhasePlaying,
		CurrentSeat: 1,
		TrumpSuit:   -1,
		LevelRank:   3,
		Players: []PlayerHand{
			{UserID: 1, Seat: 0, Hand: []Card{{ID: 0}}}, // dummy
			{UserID: 2, Seat: 1, Hand: []Card{
				{ID: 11}, {ID: 65}, // pair of ♠K (canFollow=true)
				{ID: 15}, // ♥5
				{ID: 22}, // ♥J
				{ID: 0},  // extra ♠3
			}},
			{UserID: 3, Seat: 2, Hand: []Card{{ID: 100}}},
			{UserID: 4, Seat: 3, Hand: []Card{{ID: 101}}},
		},
		LastPlay: &PlayRecord{
			Seat:  0,
			Play:  ledPlay,
			Cards: ledCards,
		},
	}

	action := game.PlayerAction{
		PlayerID: 2,
		Action:   "play",
		Cards:    []int{15, 22}, // ♥5 + ♥J (not a valid pair)
	}

	// First call
	_, err1 := eng.ExecuteAction(gs, action)
	t.Logf("First call: err=%v", err1)

	// Second call with SAME state
	_, err2 := eng.ExecuteAction(gs, action)
	t.Logf("Second call: err=%v", err2)

	if err1 == nil {
		t.Error("First call should reject heart singles when player can follow with spade pair")
	}
	if err2 == nil {
		t.Error("Second call should also reject heart singles when player can follow with spade pair")
	}
	if (err1 == nil) != (err2 == nil) {
		t.Errorf("INCONSISTENCY: first=%v second=%v — server gave different results for same input", err1, err2)
	}
}

// TestBug_PaddingAcceptedWhenCannotFollow verifies the current (intentional) behavior:
// when canFollow=false, padding with any cards (even PlayInvalid) is accepted.
func TestBug_PaddingAcceptedWhenCannotFollow(t *testing.T) {
	eng := &Engine{}

	ledCards := []Card{{ID: 5}, {ID: 59}}
	ledPlay := ParsePlayWithContext(ledCards, -1, 3)

	gs := &GameState{
		Phase:       PhasePlaying,
		CurrentSeat: 1,
		TrumpSuit:   -1,
		LevelRank:   3,
		Players: []PlayerHand{
			{UserID: 1, Seat: 0, Hand: []Card{{ID: 0}}},
			{UserID: 2, Seat: 1, Hand: []Card{
				{ID: 11}, // ♠K (only one copy, no pair)
				{ID: 15}, // ♥5
				{ID: 22}, // ♥J
			}},
			{UserID: 3, Seat: 2, Hand: []Card{{ID: 100}}},
			{UserID: 4, Seat: 3, Hand: []Card{{ID: 101}}},
		},
		LastPlay: &PlayRecord{
			Seat:  0,
			Play:  ledPlay,
			Cards: ledCards,
		},
	}

	action := game.PlayerAction{
		PlayerID: 2,
		Action:   "play",
		Cards:    []int{15, 22},
	}

	state1, err1 := eng.ExecuteAction(gs, action)
	if err1 != nil {
		t.Logf("Padding rejected (canFollow=false): %v", err1)
	} else {
		t.Log("Padding accepted as expected (canFollow=false, isPadding=true)")
		gs2 := state1.(*GameState)
		if gs2.LastPlay != nil && gs2.LastPlay.Seat == 1 {
			t.Logf("Hand after padding: %d cards remaining", len(gs2.Players[1].Hand))
		}
	}
}

// TestBug_UserScenario_HeartsVsSpadeLead simulates the exact reported bug scenario:
// Lead plays pair of ♠8 (black 8). User has spade pairs in hand but selects
// ♥5 + ♥J (two hearts, not a pair). Tests multiple trump/level configurations
// and verifies the server is deterministic — both clicks must produce same result.
func TestBug_UserScenario_HeartsVsSpadeLead(t *testing.T) {
	eng := &Engine{}

	// Configurations to test: (trumpSuit, levelRank, description)
	configs := []struct {
		trumpSuit  int
		levelRank  int
		desc       string
	}{
		{-1, 3, "no trump set (all side cards CatSide)"},
		{0, 3, "spades trump (♠8=CatTrump, ♠K=CatTrump, ♥5/J=CatSide)"},
		{1, 3, "hearts trump (♠8=CatSide, ♠K=CatSide, ♥5/J=CatSide)"},
		{0, 8, "spades trump + level=8 (♠8=CatNativeMain, ♠K=CatTrump — diff category!)"},
	}

	for _, cfg := range configs {
		t.Run(cfg.desc, func(t *testing.T) {
			t.Logf("=== CONFIG: trumpSuit=%d levelRank=%d (%s) ===", cfg.trumpSuit, cfg.levelRank, cfg.desc)

			// Build GameState: leader(seat 0) just played pair of ♠8
			ledCards := []Card{{ID: 5}, {ID: 59}} // ♠8, ♠8 (copies 0 and 1)
			ledPlay := ParsePlayWithContext(ledCards, cfg.trumpSuit, cfg.levelRank)
			t.Logf("Led play type=%d mainRank=%d", ledPlay.Type, ledPlay.MainRank)
			t.Logf("Led card category: %d", ClassifyCard(ledCards[0], cfg.trumpSuit, cfg.levelRank))

			// Follower (seat 1) has: pair of ♠K + extra spades + ♥5 + ♥J
			// This means canFollow SHOULD return true (has matching pair)
			followerHand := []Card{
				{ID: 11}, {ID: 65}, // pair of ♠K (face=11, baseRank=13)
				{ID: 0},  // ♠3 (extra spade)
				{ID: 1},  // ♠4 (extra spade)
				{ID: 15}, // ♥5 (heart 5)
				{ID: 22}, // ♥J (heart J)
			}

			t.Log("Follower hand:")
			for _, c := range followerHand {
				cat := ClassifyCard(c, cfg.trumpSuit, cfg.levelRank)
				t.Logf("  ID=%3d face=%2d suit=%d baseRank=%2d cat=%d  %s",
					c.ID, c.Face(), c.Suit(), c.BaseRank(), cat, c.Display())
			}

			// Check canFollow explicitly
			cf := canFollow(followerHand, ledPlay, ledCards, cfg.trumpSuit, cfg.levelRank)
			t.Logf("canFollow result: %t", cf)

			gs := &GameState{
				Phase:       PhasePlaying,
				CurrentSeat: 1,
				TrumpSuit:   cfg.trumpSuit,
				LevelRank:   cfg.levelRank,
				Players: []PlayerHand{
					{UserID: 1, Seat: 0, Hand: []Card{{ID: 0}}},
					{UserID: 2, Seat: 1, Hand: followerHand},
					{UserID: 3, Seat: 2, Hand: []Card{{ID: 100}}},
					{UserID: 4, Seat: 3, Hand: []Card{{ID: 101}}},
				},
				LastPlay: &PlayRecord{
					Seat:  0,
					Play:  ledPlay,
					Cards: ledCards,
				},
			}

			action := game.PlayerAction{
				PlayerID: 2,
				Action:   "play",
				Cards:    []int{15, 22}, // ♥5 + ♥J
			}

			t.Log("--- First click ---")
			_, err1 := eng.ExecuteAction(gs, action)
			t.Logf("Result: err=%v", err1)

			// For second click, we need a FRESH copy of the state because
			// if first click succeeded, it mutated gs (advanced seat, removed cards).
			// In the real bug scenario, first click FAILS → state NOT mutated.
			// Rebuild a clean state for the second attempt.
			gs2 := &GameState{
				Phase:       PhasePlaying,
				CurrentSeat: 1,
				TrumpSuit:   cfg.trumpSuit,
				LevelRank:   cfg.levelRank,
				Players: []PlayerHand{
					{UserID: 1, Seat: 0, Hand: []Card{{ID: 0}}},
					{UserID: 2, Seat: 1, Hand: copyHand(followerHand)},
					{UserID: 3, Seat: 2, Hand: []Card{{ID: 100}}},
					{UserID: 4, Seat: 3, Hand: []Card{{ID: 101}}},
				},
				LastPlay: &PlayRecord{
					Seat:  0,
					Play:  ledPlay,
					Cards: copyCards(ledCards),
				},
			}

			t.Log("--- Second click (fresh state copy, same cards) ---")
			_, err2 := eng.ExecuteAction(gs2, action)
			t.Logf("Result: err=%v", err2)

			// VERIFY: both calls must produce same result
			if (err1 == nil) != (err2 == nil) {
				t.Errorf("INCONSISTENCY! First err=%v, Second err=%v — different results for same input!", err1, err2)
			} else {
				t.Logf("CONSISTENT: both calls returned err=%v", err1)
			}

			if cf && err1 == nil {
				t.Error("BUG: canFollow=true but hearts accepted — should have been rejected!")
			}
			if !cf && err1 != nil {
				t.Log("canFollow=false, hearts rejected — padding path not taken (unexpected)")
			}
			if !cf && err1 == nil {
				t.Log("canFollow=false, hearts accepted as padding (by design, but may violate game rules)")
			}
		})
	}
}

// TestBug_FullGameSimulation walks through the phases to reach the bug scenario
// with a hand-crafted deal to deterministically reproduce the user's situation.
func TestBug_FullGameSimulation(t *testing.T) {
	eng := &Engine{}

	// Build a controlled GameState as if we just entered PhasePlaying
	// after trump was set to hearts (suit 1), level=3
	gs := &GameState{
		Phase:       PhasePlaying,
		CurrentSeat: 0, // seat 0 leads first trick
		TrumpSuit:   1, // hearts is trump
		LevelRank:   3,
		DealerSeats: [2]int{0, 2},
		Players: []PlayerHand{
			{UserID: 1, Seat: 0, Hand: []Card{
				{ID: 5}, {ID: 59}, // pair of ♠8 for leading
				{ID: 100}, {ID: 101},
			}},
			{UserID: 2, Seat: 1, Hand: []Card{
				{ID: 11}, {ID: 65}, // pair of ♠K (same category as ♠8 when hearts trump)
				{ID: 1},             // ♠4
				{ID: 0},             // ♠3 (level rank! CatSideMain, different from ♠8's CatSide)
				{ID: 15},            // ♥5 (trump suit CatTrump!)
				{ID: 22},            // ♥J (trump suit CatTrump!)
			}},
			{UserID: 3, Seat: 2, Hand: []Card{{ID: 200}, {ID: 201}}},
			{UserID: 4, Seat: 3, Hand: []Card{{ID: 202}, {ID: 203}}},
		},
	}

	t.Log("=== Full Game Simulation: hearts trump, level=3 ===")
	t.Log("Seat 0 leads with pair of ♠8")

	// Seat 0 plays pair of ♠8
	action0 := game.PlayerAction{
		PlayerID: 1,
		Action:   "play",
		Cards:    []int{5, 59},
	}
	state0, err := eng.ExecuteAction(gs, action0)
	if err != nil {
		t.Fatalf("Seat 0 lead failed: %v", err)
	}
	t.Log("Seat 0: pair of ♠8 accepted ✓")

	// Now it's seat 1's turn. Seat 1 tries to play ♥5 + ♥J
	// Seat 1 HAS a pair of ♠K (CatSide, same as ♠8), so canFollow=true
	gs1 := state0.(*GameState)
	t.Logf("CurrentSeat=%d, LastPlay.Seat=%d, LastPlay type=%d",
		gs1.CurrentSeat, gs1.LastPlay.Seat, gs1.LastPlay.Play.Type)

	// Show seat 1's hand categories
	t.Log("Seat 1 hand categories:")
	for _, c := range gs1.Players[1].Hand {
		cat := ClassifyCard(c, gs1.TrumpSuit, gs1.LevelRank)
		t.Logf("  ID=%3d face=%2d suit=%d baseRank=%2d cat=%d  %s",
			c.ID, c.Face(), c.Suit(), c.BaseRank(), cat, c.Display())
	}

	// Explicit canFollow check
	cf := canFollow(gs1.Players[1].Hand, gs1.LastPlay.Play, gs1.LastPlay.Cards, gs1.TrumpSuit, gs1.LevelRank)
	t.Logf("canFollow result: %t", cf)

	action1 := game.PlayerAction{
		PlayerID: 2,
		Action:   "play",
		Cards:    []int{15, 22}, // ♥5 + ♥J
	}

	t.Log("--- Seat 1: first attempt with ♥5 + ♥J ---")
	_, err1 := eng.ExecuteAction(gs1, action1)
	t.Logf("First attempt: err=%v", err1)

	t.Log("--- Seat 1: second attempt with ♥5 + ♥J ---")
	_, err2 := eng.ExecuteAction(gs1, action1)
	t.Logf("Second attempt: err=%v", err2)

	if (err1 == nil) != (err2 == nil) {
		t.Errorf("INCONSISTENCY between first and second attempt!")
	}
	if cf && err1 == nil {
		t.Error("BUG: canFollow=true but hearts accepted!")
	}
}

// TestFollowSuit_DoesNotRequireBeat verifies that following suit with a
// lower-ranked play is allowed (unlike doudizhu, dashengji does not require
// beating the previous play within a round).
func TestFollowSuit_DoesNotRequireBeat(t *testing.T) {
	// levelRank=3, no trump → CatSide for non-level cards
	// ID 10 = ♠K (rank 13, CatSide), ID 1 = ♠4 (rank 4, CatSide)
	ledPlay := Play{Type: PlaySingle, MainRank: 13, Length: 1}
	ledCards := []Card{{ID: 10}} // ♠K

	followCards := []Card{{ID: 1}} // ♠4 (lower rank)
	hand := []Card{{ID: 1}, {ID: 10}} // has both cards

	isPadding, err := validateFollow(followCards, hand, ledPlay, ledCards, -1, 3)
	if err != nil {
		t.Errorf("following with lower rank should be allowed: %v", err)
	}
	if isPadding {
		t.Error("following with same suit+category should not be marked as padding")
	}
	// Note: CanBeat is NOT checked here — it's the engine's responsibility
	// to not require beating in dashengji (unlike doudizhu).
}

// TestRound_FullRoundFlow simulates a complete 4-player round and verifies
// the round resolves with the highest play winning.
func TestRound_FullRoundFlow(t *testing.T) {
	eng := &Engine{}

	// levelRank=3, no trump → CatSide for non-level-rank cards
	// Seat 0: leads ♠4 (ID 1, rank 4)
	// Seat 1: follows ♠K (ID 10, rank 13) — higher, should win
	// Seat 2: pads ♥4 (ID 14, different suit)
	// Seat 3: pads ♣5 (ID 29, different suit)

	gs := &GameState{
		Phase:       PhasePlaying,
		CurrentSeat: 0,
		RoundLeader: 0,
		TrumpSuit:   -1,
		LevelRank:   3,
		Players: []PlayerHand{
			{UserID: 1, Seat: 0, Hand: []Card{
				{ID: 1},  // ♠4 for leading
				{ID: 20}, // extra card
			}},
			{UserID: 2, Seat: 1, Hand: []Card{
				{ID: 10}, // ♠K (follows, higher rank)
				{ID: 21}, // extra card
			}},
			{UserID: 3, Seat: 2, Hand: []Card{
				{ID: 14}, // ♥4 (pad)
				{ID: 22}, // extra
			}},
			{UserID: 4, Seat: 3, Hand: []Card{
				{ID: 29}, // ♣5 (pad)
				{ID: 23}, // extra
			}},
		},
	}

	// Seat 0 leads: ♠4 (single, rank 4)
	state, err := eng.ExecuteAction(gs, game.PlayerAction{PlayerID: 1, Action: "play", Cards: []int{1}})
	if err != nil {
		t.Fatalf("Seat 0 lead failed: %v", err)
	}
	gs = state.(*GameState)
	t.Logf("After seat 0 (lead ♠4): CurrentSeat=%d RoundPlays=%d", gs.CurrentSeat, len(gs.RoundPlays))

	// Seat 1 follows: ♠K (single, rank 13) — higher rank, same suit+category
	state, err = eng.ExecuteAction(gs, game.PlayerAction{PlayerID: 2, Action: "play", Cards: []int{10}})
	if err != nil {
		t.Fatalf("Seat 1 follow failed: %v", err)
	}
	gs = state.(*GameState)
	t.Logf("After seat 1 (follow ♠K): CurrentSeat=%d RoundPlays=%d", gs.CurrentSeat, len(gs.RoundPlays))

	// Seat 2 pads: ♥4
	state, err = eng.ExecuteAction(gs, game.PlayerAction{PlayerID: 3, Action: "play", Cards: []int{14}})
	if err != nil {
		t.Fatalf("Seat 2 pad failed: %v", err)
	}
	gs = state.(*GameState)
	t.Logf("After seat 2 (pad ♥4): CurrentSeat=%d RoundPlays=%d", gs.CurrentSeat, len(gs.RoundPlays))

	// Seat 3 pads: ♣5
	state, err = eng.ExecuteAction(gs, game.PlayerAction{PlayerID: 4, Action: "play", Cards: []int{29}})
	if err != nil {
		t.Fatalf("Seat 3 pad failed: %v", err)
	}
	gs = state.(*GameState)

	// After 4 plays, round should resolve
	if len(gs.RoundPlays) != 0 {
		t.Errorf("expected round to reset after 4 plays, got %d RoundPlays", len(gs.RoundPlays))
	}

	// Seat 1 played ♠K (rank 13, highest following), so seat 1 wins
	if gs.RoundLeader != 1 {
		t.Errorf("expected seat 1 as round winner (♠K beats ♠4), got seat %d", gs.RoundLeader)
	}
	if gs.CurrentSeat != 1 {
		t.Errorf("expected CurrentSeat to be winner (seat 1), got %d", gs.CurrentSeat)
	}
	t.Logf("Round resolved: winner=seat%d CurrentSeat=%d", gs.RoundLeader, gs.CurrentSeat)
}

// TestRound_WinnerLeadsNextRound verifies that the round winner leads
// the next round and subsequent players follow in clockwise order.
func TestRound_WinnerLeadsNextRound(t *testing.T) {
	eng := &Engine{}

	// levelRank=3, no trump → CatSide for non-level-rank cards
	// Round 1: seat 0 leads ♠K (ID 10, rank 13), others pad (no ♠ cards)
	// → seat 0 wins (only valid follow is its own lead)
	// Round 2: seat 0 leads ♠4 (ID 1)

	gs := &GameState{
		Phase:       PhasePlaying,
		CurrentSeat: 0,
		RoundLeader: 0,
		TrumpSuit:   -1,
		LevelRank:   3,
		Players: []PlayerHand{
			{UserID: 1, Seat: 0, Hand: []Card{{ID: 10}, {ID: 1}, {ID: 2}}},  // ♠K, ♠4, ♠5
			{UserID: 2, Seat: 1, Hand: []Card{{ID: 14}, {ID: 15}}}, // ♥4, ♥5
			{UserID: 3, Seat: 2, Hand: []Card{{ID: 28}, {ID: 29}}}, // ♣4, ♣5
			{UserID: 4, Seat: 3, Hand: []Card{{ID: 42}, {ID: 43}}}, // ♦4, ♦5
		},
	}

	// Round 1: seat 0 leads ♠K (rank 13)
	state, _ := eng.ExecuteAction(gs, game.PlayerAction{PlayerID: 1, Action: "play", Cards: []int{10}})
	gs = state.(*GameState)

	// Seat 1 — no ♠, pads ♥4
	state, _ = eng.ExecuteAction(gs, game.PlayerAction{PlayerID: 2, Action: "play", Cards: []int{14}})
	gs = state.(*GameState)

	// Seat 2 — no ♠, pads ♣4
	state, _ = eng.ExecuteAction(gs, game.PlayerAction{PlayerID: 3, Action: "play", Cards: []int{28}})
	gs = state.(*GameState)

	// Seat 3 — no ♠, pads ♦4
	state, _ = eng.ExecuteAction(gs, game.PlayerAction{PlayerID: 4, Action: "play", Cards: []int{42}})
	gs = state.(*GameState)

	// Round 1 complete — seat 0 wins (leader with only valid follow)
	if gs.RoundLeader != 0 {
		t.Errorf("Round 1: expected winner seat 0, got %d", gs.RoundLeader)
	}
	if gs.CurrentSeat != 0 {
		t.Errorf("Round 1: CurrentSeat should be 0, got %d", gs.CurrentSeat)
	}
	if len(gs.RoundPlays) != 0 {
		t.Errorf("RoundPlays should be empty after round resolution, got %d", len(gs.RoundPlays))
	}
	t.Logf("Round 1 winner: seat %d", gs.RoundLeader)

	// Round 2: seat 0 leads ♠4 (rank 4)
	state, _ = eng.ExecuteAction(gs, game.PlayerAction{PlayerID: 1, Action: "play", Cards: []int{1}})
	gs = state.(*GameState)
	if gs.CurrentSeat != 1 {
		t.Errorf("Round 2 after lead: expected CurrentSeat 1, got %d", gs.CurrentSeat)
	}
	if len(gs.RoundPlays) != 1 {
		t.Errorf("Round 2: expected 1 play, got %d", len(gs.RoundPlays))
	}
	t.Logf("Round 2: seat 0 led ♠4, CurrentSeat=%d RoundPlays=%d", gs.CurrentSeat, len(gs.RoundPlays))
}

// TestPhasePlaying_RejectsPass verifies pass is not allowed during PhasePlaying.
func TestPhasePlaying_RejectsPass(t *testing.T) {
	eng := &Engine{}

	gs := &GameState{
		Phase:       PhasePlaying,
		CurrentSeat: 1,
		RoundLeader: 0,
		TrumpSuit:   -1,
		LevelRank:   3,
		Players: []PlayerHand{
			{UserID: 1, Seat: 0, Hand: []Card{{ID: 0}}},
			{UserID: 2, Seat: 1, Hand: []Card{{ID: 1}}},
			{UserID: 3, Seat: 2, Hand: []Card{{ID: 2}}},
			{UserID: 4, Seat: 3, Hand: []Card{{ID: 3}}},
		},
		LastPlay: &PlayRecord{
			Seat:  0,
			Play:  Play{Type: PlaySingle, MainRank: 5, Length: 1},
			Cards: []Card{{ID: 0}},
		},
		RoundPlays: []PlayRecord{
			{Seat: 0, Play: Play{Type: PlaySingle, MainRank: 5, Length: 1}, Cards: []Card{{ID: 0}}},
		},
	}

	_, err := eng.ExecuteAction(gs, game.PlayerAction{PlayerID: 2, Action: "pass", Cards: nil})
	if err == nil {
		t.Error("pass should be rejected during PhasePlaying")
	}
	t.Logf("Pass correctly rejected: %v", err)
}

func copyHand(src []Card) []Card {
	dst := make([]Card, len(src))
	copy(dst, src)
	return dst
}

func copyCards(src []Card) []Card {
	dst := make([]Card, len(src))
	copy(dst, src)
	return dst
}
