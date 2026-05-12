"use client";

import clsx from "clsx";

interface CardProps {
  cardId: number;
  selected?: boolean;
  onClick?: () => void;
  faceDown?: boolean;
  small?: boolean;
}

const SUITS = ["♠", "♥", "♣", "♦"];
const RANKS = ["3", "4", "5", "6", "7", "8", "9", "10", "J", "Q", "K", "A", "2"];

const SUIT_COLORS: Record<string, string> = {
  "♠": "text-text-black",
  "♥": "text-red-error",
  "♣": "text-text-black",
  "♦": "text-red-error",
};

export function Card({ cardId, selected, onClick, faceDown, small }: CardProps) {
  if (faceDown) {
    return (
      <div
        className={clsx(
          "bg-house-green border-2 border-white/30 rounded-lg shadow-md",
          "flex items-center justify-center",
          "font-sans select-none",
          small ? "w-8 h-12 text-xs" : "w-12 h-16",
          onClick && "cursor-pointer hover:scale-105 transition-transform"
        )}
      >
        <span className="text-white/50 text-lg font-bold">?</span>
      </div>
    );
  }

  const isJoker = cardId >= 52;
  const suit = isJoker ? "" : SUITS[Math.floor(cardId / 13)];
  const rank = isJoker ? (cardId === 52 ? "🃏" : "👑") : RANKS[cardId % 13];
  const color = isJoker ? "text-red-600" : SUIT_COLORS[suit] || "text-gray-900";

  return (
    <div
      onClick={onClick}
      className={clsx(
        "bg-white border-2 rounded-lg shadow-md",
        "flex flex-col items-center justify-center",
        "font-sans font-semibold select-none",
        "transition-all duration-150",
        small ? "w-8 h-12 text-xs px-0.5" : "w-12 h-16 px-1",
        selected ? "border-green-accent -translate-y-3 shadow-lg" : "border-gray-200",
        onClick && "cursor-pointer hover:border-green-accent/50"
      )}
    >
      <span className={clsx("leading-none", small ? "text-xs" : "text-sm")}>{rank}</span>
      {!isJoker && (
        <span className={clsx("leading-none", color, small ? "text-[8px]" : "text-[10px]")}>
          {suit}
        </span>
      )}
    </div>
  );
}
