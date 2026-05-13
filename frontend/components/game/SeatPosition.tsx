"use client";

import clsx from "clsx";
import { getCharacterStyle } from "@/themes";

interface SeatPositionProps {
  seatNumber: number;
  player?: {
    userId: string;
    name: string;
    nickname: string;
    characterId: string;
    isBot: boolean;
    isOwner: boolean;
    isReady: boolean;
    isCurrentTurn?: boolean;
    cardCount: number;
  } | null;
  position: "top" | "bottom" | "left" | "right";
  isMySeat?: boolean;
  onChangeSeat?: () => void;
  onAddBot?: () => void;
}

export function SeatPosition({
  seatNumber,
  player,
  position,
  isMySeat,
  onChangeSeat,
  onAddBot,
}: SeatPositionProps) {
  if (!player) {
    return (
      <div
        className={clsx(
          "relative flex flex-col items-center gap-2 p-3 rounded-xl border-2 border-dashed border-ceramic/50",
          "bg-white/40 min-w-[100px]",
        )}
      >
        <div className="w-12 h-12 rounded-full bg-gray-200 flex items-center justify-center text-gray-400 text-lg">
          ?
        </div>
        <p className="text-xs text-text-black-soft">空座位</p>
        {isMySeat ? (
          <button
            onClick={onChangeSeat}
            className="text-xs text-green-accent hover:underline"
          >
            坐下
          </button>
        ) : onAddBot ? (
          <button
            onClick={onAddBot}
            className="text-xs text-green-accent hover:underline"
          >
            添加AI
          </button>
        ) : null}
      </div>
    );
  }

  const charStyle = !player.isBot
    ? getCharacterStyle(player.characterId || "panda")
    : undefined;

  return (
    <div
      className={clsx(
        "relative flex flex-col items-center gap-2 p-3 rounded-xl border-2 min-w-[120px] transition-all",
        player.isCurrentTurn
          ? "border-yellow-400 bg-yellow-50 shadow-md"
          : "border-ceramic/30 bg-white",
        isMySeat && "ring-2 ring-green-accent/30",
      )}
    >
      {player.isOwner && (
        <span className="absolute -top-2 text-lg" title="房主">
          👑
        </span>
      )}

      {player.isBot ? (
        <div className="w-12 h-12 rounded-full bg-purple-400 flex items-center justify-center text-white font-bold text-lg">
          AI
        </div>
      ) : (
        <div
          className="w-12 h-12 rounded-full flex items-center justify-center text-2xl"
          style={{
            backgroundColor: charStyle?.backgroundColor ?? "#374151",
            borderColor: charStyle?.borderColor ?? "#4B5563",
            borderWidth: "2px",
            borderStyle: "solid",
          }}
        >
          {charStyle?.emoji ?? "🐼"}
        </div>
      )}

      <p className="text-sm font-medium text-text-black-strong truncate max-w-[100px]">
        {player.nickname || player.name}
      </p>

      <p className="text-xs text-text-black-soft">
        {player.cardCount} 张牌
      </p>

      {player.isReady && (
        <span className="text-xs bg-green-100 text-green-700 px-2 py-0.5 rounded-full">
          ✓ 已准备
        </span>
      )}

      {isMySeat && onChangeSeat && (
        <button
          onClick={onChangeSeat}
          className="text-xs text-text-black-soft hover:text-green-accent"
        >
          换座
        </button>
      )}
    </div>
  );
}
