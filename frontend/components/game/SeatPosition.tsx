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
    isLandlord?: boolean;
    cardCount: number;
  } | null;
  position?: "top" | "bottom" | "left" | "right";
  isMySeat?: boolean;
  onChangeSeat?: () => void;
  onAddBot?: () => void;
  cardsLeft?: "baodan" | "baoshuang" | null;
}

export function SeatPosition({
  player,
  isMySeat,
  onChangeSeat,
  onAddBot,
  cardsLeft,
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
        "relative flex flex-col items-center gap-2 p-3 rounded-xl min-w-[120px] transition-all",
        // Landlord gets red theme; current turn just enhances glow
        player.isLandlord
          ? "border-[2.5px] border-amber-400 bg-gradient-to-br from-red-950 to-red-900 shadow-[0_0_20px_rgba(251,191,36,0.4)]"
          : clsx(
              "border-2",
              player.isCurrentTurn
                ? "border-yellow-400 bg-yellow-50 shadow-md"
                : "border-ceramic/30 bg-white"
            ),
        // Landlord + current turn = even stronger glow
        player.isLandlord && player.isCurrentTurn && "shadow-[0_0_28px_rgba(251,191,36,0.6)]",
        isMySeat && !player.isLandlord && "ring-2 ring-green-accent/30",
      )}
    >
      {player.isOwner && (
        <span className="absolute -top-2 -left-1 text-base" title="房主">
          ⭐
        </span>
      )}
      {player.isLandlord && (
        <div className="absolute -top-3.5 left-1/2 -translate-x-1/2 bg-amber-400 text-red-950 text-[11px] font-extrabold px-3 py-0.5 rounded tracking-wider whitespace-nowrap">
          👑 地主
        </div>
      )}

      {player.isCurrentTurn && (
        <div className="absolute -bottom-2 left-1/2 -translate-x-1/2 bg-amber-400 text-amber-950 text-[10px] font-bold px-2.5 py-0.5 rounded-full whitespace-nowrap shadow-[0_0_8px_rgba(250,204,21,0.5)]">
          ⚡ 出牌中
        </div>
      )}

      {player.isBot ? (
        <div className={clsx(
          "w-12 h-12 rounded-full bg-purple-400 flex items-center justify-center text-white font-bold text-lg",
          player.isLandlord && "ring-[2.5px] ring-amber-400"
        )}>
          AI
        </div>
      ) : (
        <div
          className={clsx(
            "w-12 h-12 rounded-full flex items-center justify-center text-2xl",
            player.isLandlord && "ring-[2.5px] ring-amber-400"
          )}
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

      <p className={clsx(
        "text-sm font-medium truncate max-w-[100px]",
        player.isLandlord ? "text-white" : "text-text-black-strong"
      )}>
        {player.nickname || player.name}
      </p>

      <p className={clsx(
        "text-xs",
        player.isLandlord ? "text-amber-300" : "text-text-black-soft"
      )}>
        {player.cardCount} 张牌
      </p>

      {player.isReady && (
        <span className={clsx(
          "text-xs px-2 py-0.5 rounded-full",
          player.isLandlord
            ? "bg-green-700 text-green-100"
            : "bg-green-100 text-green-700"
        )}>
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

      {cardsLeft === "baodan" && (
        <span className={player.isLandlord
          ? "absolute -bottom-2 left-0 bg-red-500 text-white text-[10px] px-1.5 py-0.5 rounded-full font-bold animate-pulse"
          : "absolute -top-2 right-0 bg-red-500 text-white text-[10px] px-1.5 py-0.5 rounded-full font-bold animate-pulse"
        }>
          报单
        </span>
      )}
      {cardsLeft === "baoshuang" && (
        <span className={player.isLandlord
          ? "absolute -bottom-2 left-0 bg-orange-400 text-white text-[10px] px-1.5 py-0.5 rounded-full font-bold animate-pulse"
          : "absolute -top-2 right-0 bg-orange-400 text-white text-[10px] px-1.5 py-0.5 rounded-full font-bold animate-pulse"
        }>
          报双
        </span>
      )}
    </div>
  );
}
