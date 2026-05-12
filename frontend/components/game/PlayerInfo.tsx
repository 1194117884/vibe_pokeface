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
  seat,
  isLandlord,
  isCurrentTurn,
  isSelf,
}: PlayerInfoProps) {
  return (
    <div
      className={clsx(
        "flex items-center gap-2 px-3 py-2 rounded-[12px] transition-all duration-200",
        isCurrentTurn && "bg-green-accent/10 ring-2 ring-green-accent shadow-card",
        isSelf ? "flex-row" : "flex-row-reverse"
      )}
    >
      <div className="relative">
        <div className={clsx(
          "w-10 h-10 rounded-full flex items-center justify-center text-sm font-bold text-white shadow-sm",
          isLandlord ? "bg-house-green" : "bg-green-accent"
        )}>
          {name[0]}
        </div>
        {isLandlord && (
          <span className="absolute -top-1 -right-1 text-xs drop-shadow-sm">👑</span>
        )}
      </div>
      <div className="text-sm">
        <p className="font-semibold text-text-black tracking-tight">{name}</p>
        <p className="text-xs text-text-black-soft tracking-tight">
          {cardCount} cards {isLandlord ? "· Landlord" : ""}
        </p>
      </div>
    </div>
  );
}
