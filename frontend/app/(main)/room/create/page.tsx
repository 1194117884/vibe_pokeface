"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";
import { createRoom } from "@/lib/api-rooms";
import { Button } from "@/components/ui/Button";
import { Input } from "@/components/ui/Input";

const generateRoomPassword = () => Math.floor(100000 + Math.random() * 900000).toString();

function DiceIcon() {
  return (
    <svg
      aria-hidden="true"
      className="h-5 w-5"
      fill="none"
      viewBox="0 0 24 24"
      stroke="currentColor"
      strokeWidth="2"
    >
      <rect x="4" y="4" width="16" height="16" rx="4" />
      <circle cx="8.5" cy="8.5" r="1" fill="currentColor" stroke="none" />
      <circle cx="15.5" cy="8.5" r="1" fill="currentColor" stroke="none" />
      <circle cx="12" cy="12" r="1" fill="currentColor" stroke="none" />
      <circle cx="8.5" cy="15.5" r="1" fill="currentColor" stroke="none" />
      <circle cx="15.5" cy="15.5" r="1" fill="currentColor" stroke="none" />
    </svg>
  );
}

export default function CreateRoomPage() {
  const router = useRouter();
  const [name, setName] = useState("");
  const [gameType, setGameType] = useState("doudizhu");
  const [maxPlayers, setMaxPlayers] = useState(3);
  const [isOpen, setIsOpen] = useState(true);
  const [password, setPassword] = useState("");
  const [error, setError] = useState("");
  const [loading, setLoading] = useState(false);

  const chooseGameType = (nextGameType: string) => {
    setGameType(nextGameType);
    setMaxPlayers(nextGameType === "dashengji" ? 4 : 3);
  };

  const closeRoomWithPassword = () => {
    setIsOpen(false);
    setPassword((currentPassword) => currentPassword.trim() || generateRoomPassword());
  };

  const toggleOpenRoom = () => {
    if (isOpen) {
      closeRoomWithPassword();
      return;
    }
    setIsOpen(true);
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError("");
    const trimmedPassword = password.trim();
    if (!isOpen && !trimmedPassword) {
      setError("请设置房间密码");
      return;
    }
    setLoading(true);

    try {
      const roomId = await createRoom({
        name: name || `${gameType === "doudizhu" ? "斗地主" : gameType} 房间`,
        gameType,
        maxPlayers,
        isOpen,
        password: isOpen ? undefined : trimmedPassword,
      });
      const invitePath = `/room/${roomId}/${gameType}`;
      const passwordQuery = !isOpen ? `?password=${encodeURIComponent(trimmedPassword)}` : "";
      router.push(`${invitePath}${passwordQuery}`);
    } catch (err) {
      setError(err instanceof Error ? err.message : "创建房间失败，请重试");
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="mobile-page flex items-center justify-center bg-cream py-5">
      <div className="w-full max-w-[430px] rounded-2xl bg-white p-5 shadow-frap">
        <h1 className="text-3xl font-black text-text-black-strong mb-5 text-center">创建房间</h1>
        <form onSubmit={handleSubmit} className="space-y-5">
          <Input
            label="房间名称"
            value={name}
            onChange={(e) => setName(e.target.value)}
          />

          <div>
            <label className="block text-base font-bold text-text-black-soft mb-1.5">玩法</label>
            <div className="grid grid-cols-2 gap-2">
              {[
                { value: "doudizhu", title: "斗地主", detail: "3人" },
                { value: "dashengji", title: "打升级", detail: "4人" },
              ].map((game) => {
                const selected = gameType === game.value;
                return (
                  <button
                    key={game.value}
                    type="button"
                    onClick={() => chooseGameType(game.value)}
                    className={`min-h-[72px] rounded-2xl border px-3 py-3 text-left transition-all active:scale-[0.97] ${
                      selected
                        ? "border-green-accent bg-green-accent text-white shadow-md"
                        : "border-ceramic bg-white text-text-black-soft"
                    }`}
                    aria-pressed={selected}
                  >
                    <span className="block text-xl font-black leading-6">{game.title}</span>
                    <span className={`mt-1 block text-base font-bold ${selected ? "text-white/85" : "text-text-black-soft"}`}>
                      {game.detail}
                    </span>
                  </button>
                );
              })}
            </div>
          </div>

          <div>
            <label className="block text-base font-bold text-text-black-soft mb-1.5">人数</label>
            <div className="flex gap-2">
              {[2, 3, 4].map((n) => (
                <button
                  key={n}
                  type="button"
                  onClick={() => setMaxPlayers(n)}
                  className={`min-h-12 flex-1 rounded-lg border py-2 text-base font-bold transition-colors ${
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
            <label className="text-base font-bold text-text-black-soft">开放房间</label>
            <button
              type="button"
              onClick={toggleOpenRoom}
              className={`relative h-8 w-14 rounded-full transition-colors ${
                isOpen ? "bg-green-accent" : "bg-gray-300"
              }`}
            >
              <span
                className={`absolute left-1 top-1 h-6 w-6 rounded-full bg-white shadow transition-transform ${
                  isOpen ? "translate-x-6" : ""
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
              inputMode="numeric"
              pattern="[0-9]*"
              rightElement={
                <button
                  type="button"
                  onClick={() => setPassword(generateRoomPassword())}
                  className="flex h-9 w-9 items-center justify-center rounded-lg text-text-black-soft transition-colors hover:bg-cream hover:text-green-accent active:scale-95"
                  aria-label="随机生成 6 位密码"
                  title="随机生成 6 位密码"
                >
                  <DiceIcon />
                </button>
              }
            />
          )}

          {error && <p className="text-base font-bold text-red-500">{error}</p>}

          <Button type="submit" variant="primary" fullWidth disabled={loading}>
            {loading ? "创建中..." : "创建房间"}
          </Button>

          <div className="text-center">
            <a onClick={() => router.back()} className="inline-flex min-h-12 items-center text-base font-bold text-text-black-soft hover:text-green-accent cursor-pointer">
              返回
            </a>
          </div>
        </form>
      </div>
    </div>
  );
}
