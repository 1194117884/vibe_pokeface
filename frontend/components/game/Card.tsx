"use client";

import clsx from "clsx";

interface CardProps {
  cardId: number;
  selected?: boolean;
  onClick?: () => void;
  faceDown?: boolean;
  small?: boolean;
  medium?: boolean;
}

const SUITS = ["♠", "♥", "♣", "♦"] as const;
const RANKS = ["3", "4", "5", "6", "7", "8", "9", "10", "J", "Q", "K", "A", "2"] as const;

function suitColor(suit: string): string {
  return suit === "♥" || suit === "♦" ? "text-[#c82014]" : "text-[#1a1a1a]";
}

export function Card({ cardId, selected, onClick, faceDown, small, medium }: CardProps) {
  if (faceDown) {
    return (
      <div
        className={clsx(
          "bg-surface-container-highest border border-white/20 rounded-xl",
          "flex items-center justify-center overflow-hidden select-none",
          "shadow-[0_0_0.5px_rgba(0,0,0,0.14),0_1px_1px_rgba(0,0,0,0.24)]",
          small ? "w-8 h-12" : medium ? "w-10 h-14" : "w-12 h-16",
          onClick && "cursor-pointer hover:scale-105 transition-transform"
        )}
      >
        <div
          className="w-full h-full m-[3px] rounded-[8px] border-2 border-primary-container/30"
          style={{
            backgroundImage:
              "repeating-linear-gradient(45deg, #1d2021, #1d2021 3px, #323536 3px, #323536 6px)",
          }}
        />
      </div>
    );
  }

  const face = cardId % 54;
  const isJoker = face >= 52;
  const suit = isJoker ? "" : SUITS[Math.floor(face / 13)];
  const rank = isJoker ? "J" : RANKS[face % 13];

  return (
    <div
      onClick={onClick}
      className={clsx(
        "rounded-xl flex flex-col p-2 select-none relative overflow-hidden",
        "transition-all duration-300",
        small
          ? "w-12 h-16"
          : medium
            ? "w-16 h-24"
            : "w-24 h-36 cursor-pointer",
        "shadow-[0_2px_4px_rgba(0,0,0,0.2),inset_0_2px_4px_rgba(0,0,0,0.05)]",
        small && "shadow-[0_1px_2px_rgba(0,0,0,0.15)]",
        selected
          ? "border-2 border-primary ring-4 ring-primary/20"
          : "border border-gray-200",
      )}
      style={{
        background: "linear-gradient(135deg, #fff 0%, #f3f3f3 100%)",
      }}
    >
      {/* Linen texture overlay */}
      <div
        className="absolute inset-0 pointer-events-none opacity-[0.04]"
        style={{
          backgroundImage:
            "repeating-linear-gradient(0deg, transparent, transparent 2px, rgba(0,0,0,0.35) 2px, rgba(0,0,0,0.35) 3px), repeating-linear-gradient(90deg, transparent, transparent 2px, rgba(0,0,0,0.25) 2px, rgba(0,0,0,0.25) 3px)",
        }}
      />
      {isJoker ? (
        <>
          <span
            className={clsx(
              "leading-none self-start",
              small ? "text-[11px]" : medium ? "text-[15px]" : "text-card-number",
              face === 53 ? "text-[#c82014]" : "text-[#1a1a1a]"
            )}
          >
            {face === 53 ? "大" : "小"}
          </span>
          <span
            className={clsx(
              "leading-none self-start",
              small ? "text-[7px]" : medium ? "text-[11px]" : "text-[20px]",
              face === 53 ? "text-[#c82014]" : "text-[#1a1a1a]"
            )}
          >
            王
          </span>
        </>
      ) : (
        <>
          <span
            className={clsx(
              "leading-none self-start",
              small ? "text-[11px]" : medium ? "text-[16px]" : "text-card-number",
              suitColor(suit)
            )}
          >
            {rank}
          </span>
          <span
            className={clsx(
              "leading-none self-start",
              small ? "text-[7px]" : medium ? "text-[12px]" : "text-[20px]",
              suitColor(suit)
            )}
          >
            {suit}
          </span>
        </>
      )}
    </div>
  );
}
