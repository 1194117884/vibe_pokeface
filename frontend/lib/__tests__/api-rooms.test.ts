import { beforeEach, describe, expect, it, vi } from "vitest";
import { createRoom, listRooms } from "../api-rooms";

const mockFetch = vi.fn();
globalThis.fetch = mockFetch;

const localStorageStore = new Map<string, string>();
if (!("localStorage" in globalThis)) {
  Object.defineProperty(globalThis, "localStorage", {
    configurable: true,
    value: {
      getItem: vi.fn((key: string) => localStorageStore.get(key) ?? null),
      setItem: vi.fn((key: string, value: string) => localStorageStore.set(key, value)),
      removeItem: vi.fn((key: string) => localStorageStore.delete(key)),
      clear: vi.fn(() => localStorageStore.clear()),
    },
  });
}

describe("room API helpers", () => {
  beforeEach(() => {
    mockFetch.mockReset();
    localStorage.clear();
    localStorage.setItem("token", "jwt");
  });

  it("sends private room passwords when creating a room", async () => {
    mockFetch.mockResolvedValue({
      ok: true,
      json: async () => ({ room_id: "ROOM123" }),
    });

    await createRoom({
      name: "private",
      gameType: "doudizhu",
      maxPlayers: 3,
      isOpen: false,
      password: "secret",
    });

    expect(mockFetch).toHaveBeenCalledWith(
      "http://localhost:8080/api/rooms",
      expect.objectContaining({
        method: "POST",
        body: JSON.stringify({
          name: "private",
          game_type: "doudizhu",
          max_players: 3,
          is_open: false,
          password: "secret",
        }),
      })
    );
  });

  it("maps room list snake_case fields to the frontend shape", async () => {
    mockFetch.mockResolvedValue({
      ok: true,
      json: async () => ([
        {
          id: "ROOM123",
          name: "open",
          game_type: "dashengji",
          status: "waiting",
          max_players: 4,
          player_count: 2,
          is_open: true,
          has_password: false,
          owner_id: 9,
        },
      ]),
    });

    await expect(listRooms()).resolves.toEqual([
      {
        id: "ROOM123",
        name: "open",
        game_type: "dashengji",
        status: "waiting",
        maxPlayers: 4,
        playerCount: 2,
        isOpen: true,
        hasPassword: false,
        ownerId: 9,
      },
    ]);
  });
});
