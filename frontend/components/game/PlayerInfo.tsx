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
        "flex items-center gap-2 px-3 py-2 rounded-lg transition-colors",
        isCurrentTurn && "bg-green-accent/10 ring-2 ring-green-accent",
        isSelf ? "flex-row" : "flex-row-reverse"
      )}
    >
      <div className="relative">
        <div className={clsx(
          "w-10 h-10 rounded-full flex items-center justify-center text-sm font-bold text-white",
          isLandlord ? "bg-house-green" : "bg-green-accent"
        )}>
          {name[0]}
        </div>
        {isLandlord && (
          <span className="absolute -top-1 -right-1 text-xs">👑</span>
        )}
      </div>
      <div className="text-sm">
        <p className="font-semibold text-text-black">{name}</p>
        <p className="text-xs text-text-black-soft">
          {cardCount} cards {isLandlord ? "· Landlord" : ""}
        </p>
      </div>
    </div>
  );
}
