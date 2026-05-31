"use client";

import clsx from "clsx";

interface ReadyBarProps {
  amIOwner: boolean;
  isReady: boolean;
  allReady: boolean;
  playerCount: number;
  maxPlayers: number;
  canStart: boolean;
  onReady: () => void;
  onStartGame: () => void;
  onAddBot: () => void;
}

export function ReadyBar({
  amIOwner,
  isReady,
  playerCount,
  maxPlayers,
  canStart,
  onReady,
  onStartGame,
  onAddBot,
}: ReadyBarProps) {
  const roomFull = playerCount >= maxPlayers;

  return (
    <div className="grid w-full grid-cols-2 gap-2 px-3 py-3">
      {amIOwner && !roomFull && (
        <button
          onClick={onAddBot}
          className="min-h-14 rounded-full bg-white/5 px-4 py-3 text-lg font-black text-on-surface backdrop-blur-md border border-white/20 active:scale-95 transition-all"
        >
          + 添加AI ({playerCount}/{maxPlayers})
        </button>
      )}

      <button
        onClick={onReady}
        className={`min-h-14 rounded-full px-4 py-3 text-lg font-black active:scale-95 transition-all ${!amIOwner ? "col-span-2" : ""} ${
          isReady
            ? "bg-white/5 backdrop-blur-md border border-white/20 text-on-surface hover:bg-white/10"
            : "bg-gradient-to-b from-secondary-container to-on-secondary-container text-on-secondary-fixed hover:brightness-110 shadow-[inset_0_2px_0_rgba(255,255,255,0.4),0_4px_6px_rgba(0,0,0,0.2)]"
        }`}
      >
        {isReady ? "取消准备" : "准备"}
      </button>

      {amIOwner && (
        <button
          onClick={onStartGame}
          disabled={!canStart}
          className={clsx(
            roomFull ? "col-span-1" : "col-span-2",
            "min-h-14 rounded-full px-4 py-3 text-lg font-black hover:brightness-110 active:scale-95 transition-all",
            canStart
              ? "gold-button animate-pulse"
              : "bg-white/5 backdrop-blur-md border border-white/20 text-on-surface-variant disabled:opacity-60"
          )}
        >
          {canStart ? "开始游戏" : !roomFull ? "等待更多玩家..." : "等待准备..."}
        </button>
      )}
    </div>
  );
}
