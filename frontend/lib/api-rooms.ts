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

export async function listRooms(): Promise<RoomInfo[]> {
  const token = localStorage.getItem("token");
  const res = await fetch(`${API_BASE}/api/rooms`, {
    headers: { Authorization: `Bearer ${token}` },
  });
  if (!res.ok) throw new Error("Failed to list rooms");
  return res.json();
}
