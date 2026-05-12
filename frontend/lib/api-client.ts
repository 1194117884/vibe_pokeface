const API_BASE = process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080";

interface ApiResponse<T = unknown> {
  data?: T;
  error?: string;
}

interface LoginResult {
  token: string;
  user: { id: number; nickname: string; role: string };
}

class ApiClient {
  private token: string | null = null;

  constructor() {
    if (typeof window !== "undefined") {
      this.token = localStorage.getItem("token");
    }
  }

  setToken(token: string | null) {
    this.token = token;
    if (typeof window !== "undefined") {
      if (token) {
        localStorage.setItem("token", token);
      } else {
        localStorage.removeItem("token");
        localStorage.removeItem("role");
      }
    }
  }

  getToken(): string | null {
    return this.token;
  }

  getRole(): string | null {
    if (typeof window === "undefined") return null;
    return localStorage.getItem("role");
  }

  isAdmin(): boolean {
    return this.getRole() === "admin";
  }

  setUser(user: { role: string }) {
    if (typeof window !== "undefined") {
      localStorage.setItem("role", user.role);
    }
  }

  clearSession() {
    this.token = null;
    if (typeof window !== "undefined") {
      localStorage.removeItem("token");
      localStorage.removeItem("role");
    }
  }

  private async request<T>(
    method: string,
    path: string,
    body?: unknown
  ): Promise<ApiResponse<T>> {
    const headers: Record<string, string> = {
      "Content-Type": "application/json",
    };
    if (this.token) {
      headers["Authorization"] = `Bearer ${this.token}`;
    }

    try {
      const res = await fetch(`${API_BASE}${path}`, {
        method,
        headers,
        body: body ? JSON.stringify(body) : undefined,
      });

      const data = await res.json();
      if (!res.ok) {
        return { error: data.error || `HTTP ${res.status}` };
      }
      return { data };
    } catch (err) {
      return { error: err instanceof Error ? err.message : "Network error" };
    }
  }

  async register(nickname: string, password: string) {
    const result = await this.request<LoginResult>("POST", "/api/auth/register", {
      nickname,
      password,
    });
    if (result.data) {
      this.setUser(result.data.user);
    }
    return result;
  }

  async login(password: string, nickname: string) {
    const result = await this.request<LoginResult>("POST", "/api/auth/login", {
      password,
      provider_uid: "password:" + nickname,
    });
    if (result.data) {
      this.setUser(result.data.user);
    }
    return result;
  }
}

export const apiClient = new ApiClient();
