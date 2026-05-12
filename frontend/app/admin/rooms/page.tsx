"use client";

import { useEffect, useState } from "react";
import { Card } from "@/components/ui/Card";
import { adminFetch } from "@/lib/admin-fetch";

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
  waiting: "bg-gold-lightest text-gold",
  playing: "bg-green-light text-starbucks",
  ended: "bg-ceramic text-text-black-soft",
};

export default function AdminRoomsPage() {
  const [rooms, setRooms] = useState<Room[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    adminFetch("/api/admin/rooms")
      .then((r) => r.json())
      .then((data) => { setRooms(Array.isArray(data) ? data : []); setLoading(false); })
      .catch(() => setLoading(false));
  }, []);

  return (
    <div>
      <div className="mb-6">
        <h1 className="text-2xl font-bold text-starbucks tracking-tight">Room Monitor</h1>
        <p className="text-sm text-text-black-soft mt-0.5">Live game room status</p>
      </div>
      <Card padding="md" className="overflow-hidden">
        {loading ? (
          <p className="text-text-black-soft text-center py-4">Loading...</p>
        ) : rooms.length === 0 ? (
          <p className="text-text-black-soft text-center py-4">No active rooms.</p>
        ) : (
          <div className="overflow-x-auto">
            <table className="w-full min-w-[500px]">
            <thead>
              <tr className="border-b border-cream">
                <th className="text-left p-3 text-sm font-semibold text-text-black tracking-tight">Room ID</th>
                <th className="text-left p-3 text-sm font-semibold text-text-black tracking-tight">Game</th>
                <th className="text-left p-3 text-sm font-semibold text-text-black tracking-tight">Status</th>
                <th className="text-left p-3 text-sm font-semibold text-text-black tracking-tight">Max Players</th>
                <th className="text-left p-3 text-sm font-semibold text-text-black tracking-tight">Bot</th>
                <th className="text-left p-3 text-sm font-semibold text-text-black tracking-tight">Created</th>
              </tr>
            </thead>
            <tbody>
              {rooms.map((room) => (
                <tr key={room.id} className="border-b border-cream last:border-b-0 hover:bg-cream/50 transition-colors">
                  <td className="p-3 text-sm font-mono text-text-black">{room.id.slice(0, 8)}...</td>
                  <td className="p-3 text-sm text-text-black capitalize">{room.game_type}</td>
                  <td className="p-3 text-sm">
                    <span className={`inline-block px-3 py-1 rounded-pill text-xs font-semibold tracking-tight ${statusColors[room.status] || "bg-ceramic text-text-black-soft"}`}>
                      {room.status}
                    </span>
                  </td>
                  <td className="p-3 text-sm text-text-black">{room.max_players}</td>
                  <td className="p-3 text-sm text-text-black">{room.bot_enabled ? "Yes" : "No"}</td>
                  <td className="p-3 text-sm text-text-black-soft">
                    {new Date(room.created_at).toLocaleDateString()}
                  </td>
                </tr>
              ))}
            </tbody>
            </table>
          </div>
        )}
      </Card>
    </div>
  );
}
