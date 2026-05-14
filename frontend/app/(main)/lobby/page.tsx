"use client";

import { useEffect, useState } from "react";
import Link from "next/link";
import clsx from "clsx";
import { listRooms, RoomInfo } from "@/lib/api-rooms";
import { Button } from "@/components/ui/Button";

interface RoomCardProps {
  room: RoomInfo;
}

function RoomCard({ room }: RoomCardProps) {
  return (
    <Link href={`/room/${room.id}`}>
      <div className="bg-white rounded-xl p-5 shadow-md hover:shadow-lg transition-shadow border border-ceramic/30 cursor-pointer">
        <div className="flex items-start justify-between mb-3">
          <div>
            <h3 className="font-bold text-text-black-strong text-lg">
              {room.name || `Room ${room.id.slice(0, 4)}`}
            </h3>
            <p className="text-sm text-text-black-soft">
              {room.gameType === "doudizhu" ? "斗地主" : room.gameType}
            </p>
          </div>
          <div className="flex items-center gap-1.5">
            {room.hasPassword && (
              <span className="text-xs bg-yellow-100 text-yellow-800 px-2 py-0.5 rounded-full">
                &#x1f512;
              </span>
            )}
            <span
              className={clsx(
                "text-xs px-2 py-0.5 rounded-full font-medium",
                room.status === "waiting"
                  ? "bg-green-100 text-green-700"
                  : "bg-gray-100 text-gray-500"
              )}
            >
              {room.status === "waiting" ? "等待中" : "游戏中"}
            </span>
          </div>
        </div>
        <div className="flex items-center gap-2 text-sm text-text-black-soft">
          <span>&#x1f464; {room.playerCount}/{room.maxPlayers}</span>
          <span>·</span>
          <span>&#x1f194; {room.id}</span>
        </div>
      </div>
    </Link>
  );
}

export default function LobbyPage() {
  const [rooms, setRooms] = useState<RoomInfo[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    listRooms().then(setRooms).finally(() => setLoading(false));
  }, []);

  return (
    <div className="min-h-screen bg-cream">
      <header className="bg-white border-b border-ceramic px-6 py-4">
        <div className="max-w-6xl mx-auto flex items-center justify-between">
          <h1 className="text-2xl font-bold text-text-black-strong">PokeFace</h1>
          <div className="flex items-center gap-3">
            <Link href="/room/create">
              <Button variant="primary">+ 创建房间</Button>
            </Link>
          </div>
        </div>
      </header>

      <main className="max-w-6xl mx-auto px-6 py-8">
        <h2 className="text-lg font-bold text-text-black-strong mb-4">游戏房间</h2>
        {loading ? (
          <div className="text-center py-12 text-text-black-soft">加载中...</div>
        ) : rooms.length === 0 ? (
          <div className="text-center py-12">
            <p className="text-text-black-soft mb-4">暂无开放房间</p>
            <Link href="/room/create">
              <Button variant="primary">创建房间</Button>
            </Link>
          </div>
        ) : (
          <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4">
            {rooms.map((room) => (
              <RoomCard key={room.id} room={room} />
            ))}
          </div>
        )}
      </main>
    </div>
  );
}
