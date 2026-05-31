const getToken = () => {
  if (typeof window === "undefined") return null;
  return localStorage.getItem("token");
};

const getRole = () => {
  if (typeof window === "undefined") return null;
  return localStorage.getItem("role");
};

export function requireAdmin(): boolean {
  const role = getRole();
  return role === "admin";
}

export async function adminFetch(
  input: RequestInfo | URL,
  init?: RequestInit
): Promise<Response> {
  const token = getToken();

  if (!token) {
    redirectToLogin();
    throw new Error("Not authenticated");
  }

  if (getRole() !== "admin") {
    redirectToLobby();
    throw new Error("Not authorized");
  }

  const res = await fetch(input, {
    ...init,
    headers: {
      ...init?.headers,
      Authorization: `Bearer ${token}`,
    },
  });

  if (res.status === 401) {
    redirectToLogin();
    throw new Error("Session expired");
  }

  if (res.status === 403) {
    redirectToLobby();
    throw new Error("Not authorized");
  }

  return res;
}

function redirectToLogin() {
  if (typeof window === "undefined") return;
  const current = window.location.pathname + window.location.search;
  window.location.href = `/auth/login?redirect=${encodeURIComponent(current)}`;
}

function redirectToLobby() {
  if (typeof window === "undefined") return;
  window.location.href = "/lobby";
}

const API_BASE = process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080";

export interface RegistrationCode {
  id: number;
  code: string;
  is_used: boolean;
  is_disabled: boolean;
  used_by_user_id: number | null;
  used_at: string | null;
  created_by: number;
  note: string;
  created_at: string;
  updated_at: string;
  used_by_nickname: string | null;
}

export interface CodesListResponse {
  codes: RegistrationCode[];
  total: number;
}

export async function fetchRegistrationCodes(
  page: number,
  size: number,
  used?: boolean
): Promise<CodesListResponse> {
  const params = new URLSearchParams({ page: String(page), size: String(size) });
  if (used !== undefined) params.set("used", String(used));
  const res = await adminFetch(`${API_BASE}/api/admin/registration-codes?${params}`);
  return res.json();
}

export async function generateRegistrationCodes(
  count: number,
  note?: string
): Promise<{ codes: RegistrationCode[]; count: number }> {
  const res = await adminFetch(`${API_BASE}/api/admin/registration-codes`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ count, note }),
  });
  return res.json();
}

export async function disableRegistrationCode(id: number): Promise<void> {
  const res = await adminFetch(`${API_BASE}/api/admin/registration-codes/${id}/disable`, {
    method: "PUT",
  });
  if (!res.ok) {
    const data = await res.json();
    throw new Error(data.error || "Disable failed");
  }
}
