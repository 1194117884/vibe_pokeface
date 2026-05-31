import { describe, it, expect, vi, beforeEach } from "vitest";
import React from "react";
import { renderToStaticMarkup } from "react-dom/server";
import type { ReadonlyURLSearchParams } from "next/navigation";

// Mock next/navigation hooks so the login page renders without a Next.js runtime.
vi.mock("next/navigation", () => ({
  useRouter: vi.fn(() => ({ push: vi.fn(), back: vi.fn(), replace: vi.fn(), prefetch: vi.fn() })),
  useSearchParams: vi.fn(() => new URLSearchParams() as unknown as ReadonlyURLSearchParams),
}));

// Mock next/link so it renders a plain <a> in the node environment.
vi.mock("next/link", () => ({
  default: ({ href, children, ...props }: { href: string; children: React.ReactNode; [key: string]: unknown }) =>
    React.createElement("a", { href, ...props }, children),
}));

// Mock the api-client so we control login return values.
vi.mock("@/lib/api-client", () => ({
  apiClient: {
    login: vi.fn(),
    setToken: vi.fn(),
    setUser: vi.fn(),
    getToken: vi.fn(() => null),
    getRole: vi.fn(() => null),
    isAdmin: vi.fn(() => false),
    clearSession: vi.fn(),
  },
}));

import LoginPage from "@/app/auth/login/page";
import * as navigation from "next/navigation";

describe("Login Page Redirect", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it("renders sign in form with expected fields", () => {
    const html = renderToStaticMarkup(React.createElement(LoginPage));
    expect(html).toContain("Sign In");
    expect(html).toContain("Nickname");
    expect(html).toContain("Password");
    expect(html).toContain('type="submit"');
  });

  it("reads redirect parameter from search params for post-login navigation", () => {
    const mockGet = vi.fn((key: string) => {
      if (key === "redirect") return "/room/456/doudizhu";
      return null;
    });

    vi.mocked(navigation.useSearchParams).mockReturnValue({
      get: mockGet,
      has: vi.fn(),
      getAll: vi.fn(),
      forEach: vi.fn(),
      entries: function* () { yield* []; },
      keys: function* () { yield* []; },
      values: function* () { yield* []; },
      toString: () => "redirect=%2Froom%2F456%2Fdoudizhu",
      [Symbol.iterator]: function* () { yield* []; },
      size: 1,
    } as unknown as ReadonlyURLSearchParams);

    renderToStaticMarkup(React.createElement(LoginPage));

    expect(mockGet).toHaveBeenCalledWith("redirect");
  });
});
