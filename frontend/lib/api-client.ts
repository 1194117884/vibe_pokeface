const API_BASE = process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080";

interface ApiResponse<T = unknown> {
  data?: T;
  error?: string;
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
      }
    }
  }

  getToken(): string | null {
    return this.token;
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

  register(nickname: string, password: string) {
    return this.request<{ token: string; user: unknown }>("POST", "/api/auth/register", {
      nickname,
      password,
    });
  }

  login(password: string, providerUid?: string) {
    return this.request<{ token: string; user: unknown }>("POST", "/api/auth/login", {
      password,
      provider_uid: providerUid,
    });
  }

  guestLogin(deviceId: string) {
    return this.request<{ token: string; user: unknown }>("POST", "/api/auth/guest", {
      device_id: deviceId,
    });
  }
}

export const apiClient = new ApiClient();
