// frontend/lib/api-rooms.ts
const API_BASE = process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080";

export interface CreateRoomParams {
  name: string;
  gameType: string;
  maxPlayers: number;
  isOpen: boolean;
  password?: string;
}

export interface RoomInfo {
  id: string;
  name: string;
  gameType: string;
  status: string;
  maxPlayers: number;
  playerCount: number;
  isOpen: boolean;
  hasPassword: boolean;
  ownerId: number;
}

export async function createRoom(params: CreateRoomParams): Promise<string> {
  const token = localStorage.getItem("token");
  const res = await fetch(`${API_BASE}/api/rooms`, {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
      Authorization: `Bearer ${token}`,
    },
    body: JSON.stringify({
      name: params.name,
      game_type: params.gameType,
      max_players: params.maxPlayers,
      is_open: params.isOpen,
      password: params.password || undefined,
    }),
  });
  if (!res.ok) throw new Error("Failed to create room");
  const data = await res.json();
  return data.room_id;
}

export interface AICharacterInfo {
  id: number;
  name: string;
  avatar_url?: string;
  personality?: string;
  play_style: string;
  catchphrase?: string;
  occupation?: string;
  greeting?: string;
  enabled: boolean;
}

export async function listAICharacters(): Promise<AICharacterInfo[]> {
  try {
    const token = localStorage.getItem("token");
    const res = await fetch(`${API_BASE}/api/ai-characters`, {
      headers: { Authorization: `Bearer ${token}` },
    });
    if (!res.ok) return [];
    return res.json();
  } catch {
    return [];
  }
}

export async function listRooms(): Promise<RoomInfo[]> {
  try {
    const token = localStorage.getItem("token");
    const res = await fetch(`${API_BASE}/api/rooms`, {
      headers: { Authorization: `Bearer ${token}` },
    });
    if (!res.ok) return [];
    return res.json();
  } catch {
    return [];
  }
}
