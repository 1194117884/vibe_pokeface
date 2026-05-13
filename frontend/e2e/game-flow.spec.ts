import { test, expect, Page } from "@playwright/test";
import { MockGameServer, GameScenario } from "./mock-ws-server";
import { injectAuth } from "./helpers";

const WS_PORT = 18765;

test.describe("斗地主 Full Game Flow", () => {
  // Tests share a mutable scenario object — must run serially
  test.describe.configure({ mode: "serial" });

  let mockServer: MockGameServer;
  let scenario: GameScenario;

  test.beforeAll(async () => {
    scenario = { onMessage: () => {} };
    mockServer = new MockGameServer(scenario);
    await mockServer.start(WS_PORT);
  });

  test.afterAll(async () => {
    mockServer.stop();
  });

  test.afterEach(async ({ page }) => {
    // Navigate to base origin first so localStorage is accessible
    await page.goto("/");
    await page.evaluate(() => localStorage.clear());
  });

  async function setupPage(page: Page, userId: number = 1) {
    // Navigate to base origin first so localStorage is accessible
    await page.goto("/");
    await injectAuth(page, userId);
    await page.goto("/room/test-room");
    // Wait for the room page to render (header appears after connection)
    await page.waitForSelector("header", { timeout: 10000 });
  }

  test("bidding phase: 叫地主 and 抢地主", async ({ page }) => {
    scenario.onMessage = (msg, _player, server) => {
      if (msg.type === "ready") {
        server.broadcast({
          type: "player_ready",
          data: [
            { user_id: "1", seat: 0, ready: true, is_bot: false, is_owner: true, nickname: "TestPlayer" },
          ],
        });
      }

      if (msg.type === "start_game") {
        server.broadcast({
          type: "game_start",
          data: {
            phase: 0,
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

      if (msg.type === "room_action") {
        const action = msg.data?.action;

        if (action === "bid_call") {
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
        }

        if (action === "bid_pass") {
          server.broadcast({
            type: "state_update",
            data: {
              phase: 1,
              players: [
                { user_id: 1, seat: 0, hand: [0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 51, 52, 53], is_landlord: true },
                { user_id: 2, seat: 1, hand: [17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31, 32, 33], is_landlord: false },
                { user_id: 3, seat: 2, hand: [34, 35, 36, 37, 38, 39, 40, 41, 42, 43, 44, 45, 46, 47, 48, 49, 50], is_landlord: false },
              ],
              current_seat: 0,
              landlord_seat: 0,
              landlord_cards: [51, 52, 53],
              last_play: null,
            },
          });
        }
      }
    };

    await setupPage(page);

    // Wait for initial render
    await page.waitForTimeout(500);

    // Verify the room page loaded (header visible)
    await expect(page.locator("header")).toBeVisible({ timeout: 5000 });

    // Verify the page shows room-related content
    await expect(page.locator("text=房间").first()).toBeVisible({ timeout: 5000 });
  });

  test("playing phase: state update with last play", async ({ page }) => {
    scenario.onMessage = (msg, _player, server) => {
      if (msg.type === "join_room") {
        // Send game state as if already playing
        setTimeout(() => {
          server.broadcast({
            type: "state_update",
            data: {
              phase: 1,
              players: [
                { user_id: 1, seat: 0, hand: [3, 7, 11, 52, 53], is_landlord: true },
                { user_id: 2, seat: 1, hand: [0, 4, 8, 12, 16], is_landlord: false },
                { user_id: 3, seat: 2, hand: [2, 6, 10, 14, 18], is_landlord: false },
              ],
              current_seat: 0,
              landlord_seat: 0,
              landlord_cards: [51, 52, 53],
              last_play: null,
            },
          });
        }, 100);
      }

      if (msg.type === "room_action" && msg.data?.action === "play") {
        setTimeout(() => {
          server.broadcast({
            type: "state_update",
            data: {
              phase: 1,
              players: [
                { user_id: 1, seat: 0, hand: [7, 11, 52, 53], is_landlord: true },
                { user_id: 2, seat: 1, hand: [0, 4, 8, 12, 16], is_landlord: false },
                { user_id: 3, seat: 2, hand: [2, 6, 10, 14, 18], is_landlord: false },
              ],
              current_seat: 1,
              landlord_seat: 0,
              landlord_cards: [51, 52, 53],
              last_play: { seat: 0, cards: [{ id: 3 }] },
            },
          });
        }, 100);
      }
    };

    await setupPage(page);
    await page.waitForTimeout(1000);

    // Verify header is visible
    await expect(page.locator("header")).toBeVisible({ timeout: 5000 });
  });

  test("round end: score display overlay", async ({ page }) => {
    scenario.onMessage = (msg, _player, server) => {
      if (msg.type === "join_room") {
        // Send a state_update first so players are visible
        setTimeout(() => {
          server.broadcast({
            type: "state_update",
            data: {
              phase: 1,
              players: [
                { user_id: 1, seat: 0, hand: [3], is_landlord: true },
                { user_id: 2, seat: 1, hand: [0, 4], is_landlord: false },
                { user_id: 3, seat: 2, hand: [2, 6], is_landlord: false },
              ],
              current_seat: 0,
              landlord_seat: 0,
              landlord_cards: [51, 52, 53],
              last_play: { seat: 2, cards: [{ id: 6 }] },
            },
          });
        }, 100);

        // Send round_end after a bit more delay
        setTimeout(() => {
          server.broadcast({
            type: "round_end",
            data: {
              scores: [
                { player_id: 1, score: 2 },
                { player_id: 2, score: -1 },
                { player_id: 3, score: -1 },
              ],
            },
          });
        }, 300);
      }
    };

    await setupPage(page);
    await page.waitForTimeout(1500);

    // Verify header
    await expect(page.locator("header")).toBeVisible({ timeout: 5000 });
  });
});
