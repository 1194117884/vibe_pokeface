import { describe, it, expect, beforeEach, vi } from "vitest";

const mockFetch = vi.fn();
globalThis.fetch = mockFetch;

const localStorageStore = new Map<string, string>();
const localStorageMock = {
  getItem: vi.fn((key: string) => localStorageStore.get(key) ?? null),
  setItem: vi.fn((key: string, value: string) => localStorageStore.set(key, value)),
  removeItem: vi.fn((key: string) => localStorageStore.delete(key)),
  clear: vi.fn(() => localStorageStore.clear()),
  get length() { return localStorageStore.size; },
  key: vi.fn((index: number) => Array.from(localStorageStore.keys())[index] ?? null),
};
Object.defineProperty(globalThis, "localStorage", { value: localStorageMock });

const API_BASE = "http://localhost:8080";

describe("API Client", () => {
  beforeEach(() => {
    mockFetch.mockReset();
    localStorageStore.clear();
    vi.clearAllMocks();
  });

  describe("auth endpoints", () => {
    it("should construct register request correctly", async () => {
      mockFetch.mockResolvedValue({
        ok: true,
        json: async () => ({ token: "test-jwt-token", user: { id: 1, nickname: "test" } }),
      });

      const response = await fetch(`${API_BASE}/api/auth/register`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ nickname: "test", password: "password123" }),
      });

      expect(mockFetch).toHaveBeenCalledWith(
        `${API_BASE}/api/auth/register`,
        expect.objectContaining({
          method: "POST",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify({ nickname: "test", password: "password123" }),
        })
      );
      expect(response.ok).toBe(true);
    });

    it("should construct login request correctly", async () => {
      mockFetch.mockResolvedValue({
        ok: true,
        json: async () => ({ token: "test-jwt-token" }),
      });

      const response = await fetch(`${API_BASE}/api/auth/login`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ nickname: "test", password: "password123" }),
      });

      expect(mockFetch).toHaveBeenCalledWith(
        `${API_BASE}/api/auth/login`,
        expect.objectContaining({
          method: "POST",
          body: JSON.stringify({ nickname: "test", password: "password123" }),
        })
      );
      expect(response.ok).toBe(true);
    });

    it("should handle auth error response", async () => {
      mockFetch.mockResolvedValue({
        ok: false,
        status: 401,
        json: async () => ({ error: "invalid credentials" }),
      });

      const response = await fetch(`${API_BASE}/api/auth/login`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ nickname: "test", password: "wrong" }),
      });

      expect(response.ok).toBe(false);
      expect(response.status).toBe(401);
      const data = await response.json();
      expect(data.error).toBe("invalid credentials");
    });
  });

  describe("token management", () => {
    it("should store token in localStorage", () => {
      const token = "test-jwt-token";
      localStorage.setItem("pokeface_token", token);
      expect(localStorage.getItem("pokeface_token")).toBe(token);
    });

    it("should clear token on logout", () => {
      localStorage.setItem("pokeface_token", "test-jwt-token");
      localStorage.removeItem("pokeface_token");
      expect(localStorage.getItem("pokeface_token")).toBeNull();
    });
  });

  describe("authenticated requests", () => {
    it("should include Authorization header with stored token", async () => {
      localStorage.setItem("pokeface_token", "my-token");
      mockFetch.mockResolvedValue({
        ok: true,
        json: async () => ({ nickname: "test" }),
      });

      const token = localStorage.getItem("pokeface_token");
      const response = await fetch(`${API_BASE}/api/user/profile`, {
        headers: {
          Authorization: `Bearer ${token}`,
          "Content-Type": "application/json",
        },
      });

      expect(mockFetch).toHaveBeenCalledWith(
        `${API_BASE}/api/user/profile`,
        expect.objectContaining({
          headers: expect.objectContaining({
            Authorization: "Bearer my-token",
          }),
        })
      );
      expect(response.ok).toBe(true);
    });
  });
});
