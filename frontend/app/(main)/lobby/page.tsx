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
    <Link href={`/room/${room.id}/${room.game_type || "doudizhu"}`} className="block">
      <div className="bg-white rounded-xl p-4 shadow-md active:scale-[0.98] transition-transform border border-ceramic/30 cursor-pointer">
        <div className="flex items-start justify-between gap-3 mb-3">
          <div className="min-w-0">
            <h3 className="truncate text-xl font-black text-text-black-strong">
              {room.name || `Room ${room.id.slice(0, 4)}`}
            </h3>
            <p className="text-base font-bold text-text-black-soft">
              {room.game_type === "doudizhu" ? "斗地主" : room.game_type}
            </p>
          </div>
          <div className="flex shrink-0 items-center gap-1.5">
            {room.hasPassword && (
              <span className="text-xs bg-yellow-100 text-yellow-800 px-2 py-0.5 rounded-full">
                锁
              </span>
            )}
            <span
              className={clsx(
                "text-sm px-2.5 py-1 rounded-full font-bold",
                room.status === "waiting"
                  ? "bg-green-100 text-green-700"
                  : "bg-gray-100 text-gray-500"
              )}
            >
              {room.status === "waiting" ? "等待中" : "游戏中"}
            </span>
          </div>
        </div>
        <div className="flex min-w-0 items-center gap-2 text-base font-bold text-text-black-soft">
          <span>&#x1f464; {room.playerCount}/{room.maxPlayers}</span>
          <span>·</span>
          <span className="truncate">&#x1f194; {room.id}</span>
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
    <div className="mobile-page bg-cream pb-6 pt-[max(14px,var(--safe-area-top))]">
      <header className="sticky top-0 z-20 -mx-4 border-b border-ceramic bg-white/95 px-4 py-3 backdrop-blur">
        <div className="mx-auto flex max-w-[430px] items-center justify-between gap-3">
          <h1 className="text-2xl font-bold text-text-black-strong">PokeFace</h1>
          <div className="flex items-center gap-3">
            <Link href="/room/create">
              <Button variant="primary">创建</Button>
            </Link>
          </div>
        </div>
      </header>

      <main className="mx-auto max-w-[430px] py-5">
        <div className="mb-4 flex items-end justify-between gap-3">
          <h2 className="text-2xl font-black text-text-black-strong">游戏房间</h2>
          <Link href="/room/create">
            <Button variant="outlined">新房间</Button>
          </Link>
        </div>
        {loading ? (
          <div className="text-center py-12 text-text-black-soft">加载中...</div>
        ) : rooms.length === 0 ? (
          <div className="rounded-xl bg-white px-4 py-10 text-center shadow-card">
            <p className="mb-4 text-lg font-bold text-text-black-soft">暂无开放房间</p>
            <Link href="/room/create">
              <Button variant="primary" fullWidth>创建房间</Button>
            </Link>
          </div>
        ) : (
          <div className="grid grid-cols-1 gap-3">
            {rooms.map((room) => (
              <RoomCard key={room.id} room={room} />
            ))}
          </div>
        )}
      </main>
    </div>
  );
}
