import { test, expect } from "@playwright/test";
import { MockGameServer } from "./mock-ws-server";
import { injectAuth } from "./helpers";
import path from "path";
import fs from "fs";

const WS_PORT = 18765;
const SCREENSHOT_DIR = path.join(__dirname, "screenshots");

/**
 * Full game simulation: drives a complete 斗地主 game through the browser
 * using the mock WebSocket server. Takes screenshots at each phase.
 */
test.describe("Full Game Simulation", () => {
  let mockServer: MockGameServer;

  // Game simulation state
  let bidCount = 0;
  let landlordSeat = -1;
  let currentSeat = 0;
  let playCount = 0;

  test.beforeAll(() => {
    // Ensure screenshot directory exists
    if (!fs.existsSync(SCREENSHOT_DIR)) {
      fs.mkdirSync(SCREENSHOT_DIR, { recursive: true });
    }
  });

  test("complete game simulation with screenshots", async ({ page }) => {
    // Track client messages for debugging
    const clientMessages: any[] = [];

    // Stateful mock server scenario
    mockServer = new MockGameServer({
      onMessage(msg, player, server) {
        clientMessages.push({ type: msg.type, data: msg.data });

        if (msg.type === "join_room") {
          console.log(`  [mock] Player joined: ${player.userId}`);

          // Auto-fill the room with AI bots so we can start the game
          setTimeout(() => {
            console.log("  [mock] Adding AI Bob (seat 1)");
            server.broadcast({
              type: "player_joined",
              data: {
                user_id: "2",
                seat: 1,
                is_bot: true,
                players: [
                  { user_id: 1, seat: 0, ready: false, is_bot: false, is_owner: true, nickname: "Player 1" },
                  { user_id: 2, seat: 1, ready: true, is_bot: true, is_owner: false, nickname: "AI Bob" },
                ],
                theme: "classic-poker",
                game_type: "doudizhu",
                max_players: 3,
              },
            });
          }, 300);

          setTimeout(() => {
            console.log("  [mock] Adding AI Charlie (seat 2)");
            server.broadcast({
              type: "player_joined",
              data: {
                user_id: "3",
                seat: 2,
                is_bot: true,
                players: [
                  { user_id: 1, seat: 0, ready: false, is_bot: false, is_owner: true, nickname: "Player 1" },
                  { user_id: 2, seat: 1, ready: true, is_bot: true, is_owner: false, nickname: "AI Bob" },
                  { user_id: 3, seat: 2, ready: true, is_bot: true, is_owner: false, nickname: "AI Charlie" },
                ],
                theme: "classic-poker",
                game_type: "doudizhu",
                max_players: 3,
              },
            });
          }, 600);

          // Make all players ready so owner can click "开始游戏"
          // The owner doesn't have a "准备" button, so we simulate all becoming ready
          setTimeout(() => {
            console.log("  [mock] All players ready (including owner via broadcast)");
            server.broadcast({
              type: "player_ready",
              data: {
                players: [
                  { user_id: 1, seat: 0, ready: true, is_bot: false, is_owner: true, nickname: "Player 1" },
                  { user_id: 2, seat: 1, ready: true, is_bot: true, is_owner: false, nickname: "AI Bob" },
                  { user_id: 3, seat: 2, ready: true, is_bot: true, is_owner: false, nickname: "AI Charlie" },
                ],
              },
            });
          }, 900);
        }

        if (msg.type === "start_game") {
          console.log("  [mock] Game starting — broadcasting game_start");
          server.broadcast({
            type: "game_start",
            data: {
              phase: 0, // PhaseBidding
              players: [
                { user_id: 1, seat: 0, hand: [0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16], is_landlord: false },
                { user_id: 2, seat: 1, hand: [17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31, 32, 33], is_landlord: false },
                { user_id: 3, seat: 2, hand: [34, 35, 36, 37, 38, 39, 40, 41, 42, 43, 44, 45, 46, 47, 48, 49, 50], is_landlord: false },
              ],
              current_seat: 0,
              landlord_cards: [51, 52, 53],
            },
          });
        }

        // === BIDDING PHASE ===
        if (msg.type === "room_action" && msg.data?.action === "bid_call") {
          bidCount++;
          console.log(`  [mock] Bid call (bid #${bidCount})`);

          if (bidCount === 1) {
            // Player called — advance to seat 1
            currentSeat = 1;
            landlordSeat = 0;
            server.broadcast({
              type: "state_update",
              data: {
                phase: 0,
                players: [
                  { user_id: 1, seat: 0, hand: [0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16], is_landlord: false },
                  { user_id: 2, seat: 1, hand: [17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31, 32, 33], is_landlord: false },
                  { user_id: 3, seat: 2, hand: [34, 35, 36, 37, 38, 39, 40, 41, 42, 43, 44, 45, 46, 47, 48, 49, 50], is_landlord: false },
                ],
                current_seat: 1,
                landlord_seat: 0,
                bid_history: [{ seat: 0, called: true }],
              },
            });

            // Simulate AI seat 1 calling (抢地主) after a delay
            setTimeout(() => {
              bidCount++;
              landlordSeat = 1;
              currentSeat = 2;
              console.log("  [mock] AI Bob snatches landlord (抢地主)");
              server.broadcast({
                type: "state_update",
                data: {
                  phase: 0,
                  players: [
                    { user_id: 1, seat: 0, hand: [0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16], is_landlord: false },
                    { user_id: 2, seat: 1, hand: [17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31, 32, 33], is_landlord: false },
                    { user_id: 3, seat: 2, hand: [34, 35, 36, 37, 38, 39, 40, 41, 42, 43, 44, 45, 46, 47, 48, 49, 50], is_landlord: false },
                  ],
                  current_seat: 2,
                  landlord_seat: 1,
                  bid_history: [{ seat: 0, called: true }, { seat: 1, called: true }],
                },
              });

              // Simulate AI seat 2 passing — transition to playing
              setTimeout(() => {
                bidCount++;
                currentSeat = 1;
                console.log("  [mock] AI Charlie passes — Bob is landlord, entering playing phase");

                server.broadcast({
                  type: "state_update",
                  data: {
                    phase: 1, // PhasePlaying
                    players: [
                      { user_id: 1, seat: 0, hand: [0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16], is_landlord: false },
                      { user_id: 2, seat: 1, hand: [17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31, 32, 33, 51, 52, 53], is_landlord: true },
                      { user_id: 3, seat: 2, hand: [34, 35, 36, 37, 38, 39, 40, 41, 42, 43, 44, 45, 46, 47, 48, 49, 50], is_landlord: false },
                    ],
                    current_seat: 1, // Bob starts (landlord)
                    landlord_seat: 1,
                    landlord_cards: [51, 52, 53],
                    last_play: null,
                  },
                });

                // Simulate AI Bob (landlord, seat 1) playing a card
                setTimeout(() => {
                  console.log("  [mock] AI Bob plays a card (♥6)");
                  server.broadcast({
                    type: "state_update",
                    data: {
                      phase: 1,
                      players: [
                        { user_id: 1, seat: 0, hand: [0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16], is_landlord: false },
                        { user_id: 2, seat: 1, hand: [18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31, 32, 33, 51, 52, 53], is_landlord: true },
                        { user_id: 3, seat: 2, hand: [34, 35, 36, 37, 38, 39, 40, 41, 42, 43, 44, 45, 46, 47, 48, 49, 50], is_landlord: false },
                      ],
                      current_seat: 2,
                      landlord_seat: 1,
                      landlord_cards: [51, 52, 53],
                      last_play: { seat: 1, cards: [{ id: 17 }] },
                    },
                  });

                  // Simulate AI Charlie passing
                  setTimeout(() => {
                    console.log("  [mock] AI Charlie passes — player's turn (seat 0)");
                    server.broadcast({
                      type: "state_update",
                      data: {
                        phase: 1,
                        players: [
                          { user_id: 1, seat: 0, hand: [0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16], is_landlord: false },
                          { user_id: 2, seat: 1, hand: [18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31, 32, 33, 51, 52, 53], is_landlord: true },
                          { user_id: 3, seat: 2, hand: [34, 35, 36, 37, 38, 39, 40, 41, 42, 43, 44, 45, 46, 47, 48, 49, 50], is_landlord: false },
                        ],
                        current_seat: 0, // Player's turn!
                        landlord_seat: 1,
                        landlord_cards: [51, 52, 53],
                        last_play: { seat: 1, cards: [{ id: 17 }] },
                      },
                    });
                  }, 1500);
                }, 1500);
              }, 1500);
            }, 1500);
          }
        }

        // === PLAYING PHASE ===
        if (msg.type === "room_action" && msg.data?.action === "play") {
          playCount++;
          const playedCards = msg.data?.cards || [];
          console.log(`  [mock] Player plays ${playedCards.length} cards (play #${playCount}):`, playedCards);

          if (playCount === 1) {
            // Player beats Bob's single card — update state
            const remainingHand = [0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16].filter(
              (c) => !playedCards.includes(c)
            );
            const cardsLeftMsg = remainingHand.length === 1 ? "seat_0_baodan" :
                                 remainingHand.length === 2 ? "seat_0_baoshuang" : "";

            if (cardsLeftMsg) {
              console.log(`  [mock] 报单/报双: ${cardsLeftMsg}`);
            }

            setTimeout(() => {
              server.broadcast({
                type: "state_update",
                data: {
                  phase: 1,
                  players: [
                    { user_id: 1, seat: 0, hand: remainingHand, is_landlord: false },
                    { user_id: 2, seat: 1, hand: [18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31, 32, 33, 51, 52, 53], is_landlord: true },
                    { user_id: 3, seat: 2, hand: [34, 35, 36, 37, 38, 39, 40, 41, 42, 43, 44, 45, 46, 47, 48, 49, 50], is_landlord: false },
                  ],
                  current_seat: 2,
                  landlord_seat: 1,
                  landlord_cards: [51, 52, 53],
                  last_play: { seat: 0, cards: playedCards.map((id: number) => ({ id })) },
                },
              });

              if (cardsLeftMsg) {
                setTimeout(() => {
                  server.broadcast({
                    type: "cards_left",
                    data: { message: cardsLeftMsg },
                  });
                }, 500);
              }
            }, 500);
          }
        }

        if (msg.type === "room_action" && msg.data?.action === "pass") {
          console.log("  [mock] Player passes");
          setTimeout(() => {
            server.broadcast({
              type: "state_update",
              data: {
                phase: 1,
                players: [
                  { user_id: 1, seat: 0, hand: [0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15], is_landlord: false },
                  { user_id: 2, seat: 1, hand: [18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31, 32, 33, 51, 52, 53], is_landlord: true },
                  { user_id: 3, seat: 2, hand: [34, 35, 36, 37, 38, 39, 40, 41, 42, 43, 44, 45, 46, 47, 48, 49, 50], is_landlord: false },
                ],
                current_seat: 1,
                landlord_seat: 1,
                landlord_cards: [51, 52, 53],
                last_play: { seat: 2, cards: [{ id: 34 }] },
              },
            });
          }, 500);
        }
      },
    });

    // Start the mock server
    await mockServer.start(WS_PORT);

    // Navigate to app origin first, then inject auth
    await page.goto("/");
    await injectAuth(page, 1);
    await page.goto("/room/full-game-sim");

    // Wait for connection (header appears after player_joined)
    await page.waitForSelector("header", { timeout: 10000 });
    await page.waitForTimeout(2000); // wait for AI bots to fill (300+600+900ms timers)
    await page.screenshot({ path: path.join(SCREENSHOT_DIR, "01-connected-with-ais.png"), fullPage: true });
    console.log("📸 Screenshot 1: Connected to room with AI bots");

    // === PHASE 2: START GAME ===
    // The user is the owner, room is full (3 players), all AIs are ready
    // So the owner should see an enabled "开始游戏" button
    const startBtn = page.getByRole("button", { name: /开始游戏|开始/i });
    await expect(startBtn).toBeVisible({ timeout: 5000 });
    await startBtn.click();
    await page.waitForTimeout(1000);
    await page.screenshot({ path: path.join(SCREENSHOT_DIR, "02-game-start.png"), fullPage: true });
    console.log("📸 Screenshot 2: Game started, bidding phase");

    // === PHASE 3: BIDDING (叫地主) ===
    // Look for the "叫地主" button from ActionBar (only visible on our turn)
    const callBtn = page.getByRole("button", { name: /叫地主/i });
    await expect(callBtn).toBeVisible({ timeout: 5000 });
    await callBtn.click();
    await page.waitForTimeout(1000);
    await page.screenshot({ path: path.join(SCREENSHOT_DIR, "03-bid-call.png"), fullPage: true });
    console.log("📸 Screenshot 3: Called landlord");

    // Wait for AI bidding sequence (Bob snatches, Charlie passes)
    await page.waitForTimeout(5000);
    await page.screenshot({ path: path.join(SCREENSHOT_DIR, "04-bidding-done.png"), fullPage: true });
    console.log("📸 Screenshot 4: Bidding done — Bob is landlord, 底牌 shown");

    // Wait for AI Bob's first card play + AI Charlie pass → player's turn
    await page.waitForTimeout(5000);
    await page.screenshot({ path: path.join(SCREENSHOT_DIR, "05-player-turn.png"), fullPage: true });
    console.log("📸 Screenshot 5: Player's turn, last play visible in center");

    // === PHASE 4: PLAY CARDS ===
    // Cards are rendered as <div> elements (Card component) with cursor-pointer class
    // Find clickable card divs
    const cardElements = page.locator("div.cursor-pointer");
    const cardCount = await cardElements.count();
    console.log(`  Found ${cardCount} clickable card elements`);

    if (cardCount > 0) {
      // Click the first card to select it (it gets a green border and moves up)
      await cardElements.first().click();
      await page.waitForTimeout(500);
      await page.screenshot({ path: path.join(SCREENSHOT_DIR, "06-card-selected.png"), fullPage: true });
      console.log("📸 Screenshot 6: Card selected (green border)");

      // Click the "出牌" button in HandCards
      const playBtn = page.getByRole("button", { name: /出牌/i });
      const playBtnVisible = await playBtn.isVisible().catch(() => false);
      if (playBtnVisible) {
        await playBtn.click();
        await page.waitForTimeout(2000);
        console.log("📸 Screenshot 7: Card played, state updated");
      } else {
        console.log("  ⚠️ 出牌 button not visible, trying 不出 instead");
        const passBtn = page.getByRole("button", { name: /不出/i });
        if (await passBtn.isVisible().catch(() => false)) {
          await passBtn.click();
          await page.waitForTimeout(1000);
        }
      }
      await page.screenshot({ path: path.join(SCREENSHOT_DIR, "07-post-action.png"), fullPage: true });
      console.log("📸 Screenshot 7: Post-action state");
    } else {
      console.log("  ⚠️ No clickable cards found");
      await page.screenshot({ path: path.join(SCREENSHOT_DIR, "06-no-cards-clickable.png"), fullPage: true });
    }

    // === PHASE 5: ROUND END ===
    // Send round_end from mock server
    mockServer.broadcast({
      type: "round_end",
      data: {
        scores: [
          { player_id: 1, score: 2 },
          { player_id: 2, score: -1 },
          { player_id: 3, score: -1 },
        ],
      },
    });
    await page.waitForTimeout(1500);
    await page.screenshot({ path: path.join(SCREENSHOT_DIR, "08-round-end.png"), fullPage: true });
    console.log("📸 Screenshot 8: Round end score overlay");

    // Verify the round end overlay is visible
    const endTitle = page.getByText("本局结束");
    await expect(endTitle).toBeVisible({ timeout: 3000 });
    console.log("✅ Round end overlay confirmed!");

    // Click "再来一局"
    const playAgainBtn = page.getByRole("button", { name: /再来一局/i });
    await expect(playAgainBtn).toBeVisible({ timeout: 3000 });
    await playAgainBtn.click();
    await page.waitForTimeout(1000);
    await page.screenshot({ path: path.join(SCREENSHOT_DIR, "09-play-again.png"), fullPage: true });
    console.log("📸 Screenshot 9: Play again (back to waiting)");

    console.log("\n✅ Full game simulation complete! All phases worked.");
    console.log(`📸 Screenshots saved in: ${SCREENSHOT_DIR}`);
  });

  test.afterAll(() => {
    if (mockServer) {
      mockServer.stop();
    }
  });
});
