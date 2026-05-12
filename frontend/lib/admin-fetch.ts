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
