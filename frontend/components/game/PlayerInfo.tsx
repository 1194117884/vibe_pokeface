"use client";

import clsx from "clsx";

interface PlayerInfoProps {
  name: string;
  cardCount: number;
  seat: number;
  isLandlord?: boolean;
  isCurrentTurn?: boolean;
  isSelf?: boolean;
}

export function PlayerInfo({
  name,
  cardCount,
  isLandlord,
  isCurrentTurn,
  isSelf,
}: PlayerInfoProps) {
  return (
    <div
      className={clsx(
        "flex items-center gap-2 px-3 py-2 rounded-[12px] transition-all duration-200",
        // Self + landlord: full red theme
        isSelf && isLandlord && "bg-gradient-to-br from-red-950 to-red-900 border-2 border-amber-400 shadow-[0_0_16px_rgba(251,191,36,0.4)]",
        // Opponent landlord: gold border on white bg
        !isSelf && isLandlord && "bg-white border-2 border-amber-400 shadow-[0_0_8px_rgba(251,191,36,0.2)]",
        // Normal current turn (only when NOT landlord)
        !isLandlord && isCurrentTurn && "bg-green-accent/10 ring-2 ring-green-accent shadow-card",
        isSelf ? "flex-row" : "flex-row-reverse"
      )}
    >
      <div className="relative">
        <div className={clsx(
          "w-10 h-10 rounded-full flex items-center justify-center text-sm font-bold text-white shadow-sm",
          isLandlord
            ? "bg-gradient-to-br from-red-950 to-red-900 ring-[2px] ring-amber-400"
            : "bg-green-accent"
        )}>
          {name[0]}
        </div>
        {isLandlord && (
          <span className="absolute -top-1 -right-1 text-xs drop-shadow-sm">👑</span>
        )}
      </div>
      <div className="text-sm">
        <p className={clsx(
          "font-semibold tracking-tight",
          isSelf && isLandlord ? "text-white" : "text-text-black"
        )}>
          {name}
        </p>
        <p className={clsx(
          "text-xs tracking-tight",
          isSelf && isLandlord
            ? "text-amber-300 font-semibold"
            : isLandlord
              ? "text-amber-700 font-semibold"
              : "text-text-black-soft"
        )}>
          {cardCount} cards
          {isLandlord && (
            <span className={clsx(
              "ml-1 px-1.5 py-0.5 rounded text-[10px] font-bold",
              isSelf
                ? "bg-amber-400 text-red-950"
                : "bg-amber-100 text-amber-800"
            )}>
              地主
            </span>
          )}
        </p>
        {isCurrentTurn && (
          <p className={clsx(
            "text-[10px] font-bold mt-0.5",
            isLandlord
              ? "text-amber-400"
              : "text-green-accent"
          )}>
            ⚡ 出牌
          </p>
        )}
      </div>
    </div>
  );
}
