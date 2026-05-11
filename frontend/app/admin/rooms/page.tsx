"use client";

import { useEffect, useState } from "react";
import { Card } from "@/components/ui/Card";

interface Room {
  id: string;
  game_type: string;
  owner_id: number;
  status: string;
  max_players: number;
  bot_enabled: boolean;
  created_at: string;
}

const statusColors: Record<string, string> = {
  waiting: "bg-yellow-100 text-yellow-700",
  playing: "bg-green-100 text-green-700",
  ended: "bg-gray-100 text-gray-500",
};

export default function AdminRoomsPage() {
  const [rooms, setRooms] = useState<Room[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    const token = localStorage.getItem("token");
    if (!token) return;
    fetch("/api/admin/rooms", {
      headers: { Authorization: `Bearer ${token}` },
    })
      .then((r) => r.json())
      .then((data) => { setRooms(Array.isArray(data) ? data : []); setLoading(false); })
      .catch(() => setLoading(false));
  }, []);

  return (
    <div>
      <h1 className="text-2xl font-semibold text-starbucks mb-6">Room Monitor</h1>
      <Card>
        {loading ? (
          <p className="text-gray-500 py-4 text-center">Loading...</p>
        ) : rooms.length === 0 ? (
          <p className="text-gray-500 py-4 text-center">No active rooms.</p>
        ) : (
          <table className="w-full">
            <thead>
              <tr className="border-b border-gray-200">
                <th className="text-left p-3 text-sm font-semibold">Room ID</th>
                <th className="text-left p-3 text-sm font-semibold">Game</th>
                <th className="text-left p-3 text-sm font-semibold">Status</th>
                <th className="text-left p-3 text-sm font-semibold">Max Players</th>
                <th className="text-left p-3 text-sm font-semibold">Bot</th>
                <th className="text-left p-3 text-sm font-semibold">Created</th>
              </tr>
            </thead>
            <tbody>
              {rooms.map((room) => (
                <tr key={room.id} className="border-b border-gray-100 hover:bg-gray-50">
                  <td className="p-3 text-sm font-mono">{room.id.slice(0, 8)}...</td>
                  <td className="p-3 text-sm capitalize">{room.game_type}</td>
                  <td className="p-3 text-sm">
                    <span className={`px-2 py-1 rounded-full text-xs ${statusColors[room.status] || "bg-gray-100 text-gray-500"}`}>
                      {room.status}
                    </span>
                  </td>
                  <td className="p-3 text-sm">{room.max_players}</td>
                  <td className="p-3 text-sm">{room.bot_enabled ? "Yes" : "No"}</td>
                  <td className="p-3 text-sm text-gray-500">
                    {new Date(room.created_at).toLocaleDateString()}
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        )}
      </Card>
    </div>
  );
}
