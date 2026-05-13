"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";
import { createRoom } from "@/lib/api-rooms";
import { Button } from "@/components/ui/Button";
import { Input } from "@/components/ui/Input";

export default function CreateRoomPage() {
  const router = useRouter();
  const [name, setName] = useState("");
  const [gameType, setGameType] = useState("doudizhu");
  const [maxPlayers, setMaxPlayers] = useState(3);
  const [isOpen, setIsOpen] = useState(true);
  const [password, setPassword] = useState("");
  const [error, setError] = useState("");
  const [loading, setLoading] = useState(false);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError("");
    setLoading(true);

    try {
      const roomId = await createRoom({
        name: name || `${gameType === "doudizhu" ? "斗地主" : gameType} 房间`,
        gameType,
        maxPlayers,
        isOpen,
        password: isOpen ? undefined : password,
      });
      router.push(`/room/${roomId}`);
    } catch {
      setError("创建房间失败，请重试");
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="min-h-screen bg-cream flex items-center justify-center p-4">
      <div className="bg-white rounded-2xl shadow-frap p-8 w-full max-w-md">
        <h1 className="text-2xl font-bold text-text-black-strong mb-6 text-center">创建房间</h1>
        <form onSubmit={handleSubmit} className="space-y-5">
          <Input
            label="房间名称"
            value={name}
            onChange={(e) => setName(e.target.value)}
          />

          <div>
            <label className="block text-sm font-medium text-text-black-soft mb-1">玩法</label>
            <select
              value={gameType}
              onChange={(e) => setGameType(e.target.value)}
              className="w-full rounded-lg border border-ceramic px-3 py-2.5 text-sm focus:outline-none focus:border-green-accent"
            >
              <option value="doudizhu">斗地主</option>
            </select>
          </div>

          <div>
            <label className="block text-sm font-medium text-text-black-soft mb-1">人数</label>
            <div className="flex gap-2">
              {[2, 3, 4].map((n) => (
                <button
                  key={n}
                  type="button"
                  onClick={() => setMaxPlayers(n)}
                  className={`flex-1 py-2 rounded-lg text-sm font-medium border transition-colors ${
                    maxPlayers === n
                      ? "bg-green-accent text-white border-green-accent"
                      : "bg-white text-text-black-soft border-ceramic hover:border-green-accent"
                  }`}
                >
                  {n}人
                </button>
              ))}
            </div>
          </div>

          <div className="flex items-center justify-between">
            <label className="text-sm font-medium text-text-black-soft">开放房间</label>
            <button
              type="button"
              onClick={() => setIsOpen(!isOpen)}
              className={`relative w-11 h-6 rounded-full transition-colors ${
                isOpen ? "bg-green-accent" : "bg-gray-300"
              }`}
            >
              <span
                className={`absolute top-0.5 left-0.5 w-5 h-5 bg-white rounded-full shadow transition-transform ${
                  isOpen ? "translate-x-5" : ""
                }`}
              />
            </button>
          </div>

          {!isOpen && (
            <Input
              label="房间密码"
              type="password"
              value={password}
              onChange={(e) => setPassword(e.target.value)}
            />
          )}

          {error && <p className="text-red-500 text-sm">{error}</p>}

          <Button type="submit" variant="primary" fullWidth disabled={loading}>
            {loading ? "创建中..." : "创建房间"}
          </Button>

          <div className="text-center">
            <a onClick={() => router.back()} className="text-sm text-text-black-soft hover:text-green-accent cursor-pointer">
              返回
            </a>
          </div>
        </form>
      </div>
    </div>
  );
}
