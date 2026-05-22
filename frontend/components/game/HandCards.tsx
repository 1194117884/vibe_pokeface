"use client";

import { useState, useMemo } from "react";
import clsx from "clsx";
import { Card } from "./Card";

interface HandCardsProps {
  cards: number[];
  onPlayCards?: (cardIds: number[]) => void;
  disabled?: boolean;
  compact?: boolean;
}

export function HandCards({ cards, onPlayCards, disabled, compact }: HandCardsProps) {
  const [selected, setSelected] = useState<Set<number>>(new Set());

  // Dynamic sizing based on card count
  const { cardSize, overlap } = useMemo(() => {
    const n = cards.length;
    if (compact || n > 30) return { cardSize: "tiny" as const, overlap: "-space-x-5" };
    if (n > 20) return { cardSize: "tiny" as const, overlap: "-space-x-3" };
    if (n > 12) return { cardSize: "small" as const, overlap: "-space-x-8" };
    return { cardSize: "default" as const, overlap: "-space-x-12" };
  }, [cards.length, compact]);

  const toggleCard = (cardId: number) => {
    if (disabled) return;
    const next = new Set(selected);
    if (next.has(cardId)) {
      next.delete(cardId);
    } else {
      next.add(cardId);
    }
    setSelected(next);
  };

  const handlePlay = () => {
    if (disabled || selected.size === 0) return;
    onPlayCards?.(Array.from(selected));
    setSelected(new Set());
  };

  const handlePass = () => {
    if (disabled) return;
    onPlayCards?.([]);
    setSelected(new Set());
  };

  const isTiny = cardSize === "tiny";
  const isSmall = cardSize === "small";

  return (
    <div className="space-y-3">
      {/* Hand cards with dynamic overlap */}
      <div className={clsx(
        "flex justify-center px-8 overflow-visible items-end pb-1",
        overlap,
        isTiny ? "min-h-[70px]" : isSmall ? "min-h-[100px]" : "min-h-[144px]"
      )}>
        {cards.map((cardId) => (
          <div
            key={cardId}
            onClick={() => toggleCard(cardId)}
            className={clsx(
              "relative rounded-md flex flex-col items-center justify-center border border-black/20 bg-white shadow-md",
              isTiny
                ? "w-7 h-10 text-[8px] leading-none"
                : isSmall
                ? "w-12 h-16 text-xs"
                : "w-24 h-36 text-card-number",
              isTiny && "p-[1px]",
              isSmall && "p-0.5",
              selected.has(cardId)
                ? "ring-2 ring-blue-400 -translate-y-1.5"
                : !isTiny && !disabled && "hover:-translate-y-2.5 cursor-pointer",
            )}
          >
            <MiniCardFace cardId={cardId} tiny={isTiny} small={isSmall} />
          </div>
        ))}
      </div>

      {onPlayCards && (
        <div className="flex justify-center gap-6">
          <button
            onClick={handlePlay}
            disabled={disabled || selected.size === 0}
            className="px-10 py-3 rounded-full gold-button text-button-text font-button-text hover:brightness-110 active:scale-95 transition-all disabled:opacity-40"
          >
            出牌
          </button>
          <button
            onClick={handlePass}
            disabled={disabled}
            className="px-10 py-3 rounded-full emerald-button text-button-text font-button-text hover:brightness-110 active:scale-95 transition-all disabled:opacity-40"
          >
            不出
          </button>
        </div>
      )}
    </div>
  );
}

function MiniCardFace({ cardId, tiny, small }: { cardId: number; tiny: boolean; small: boolean }) {
  const suitChars = ["♠", "♥", "♣", "♦"];
  const rankChars = ["3", "4", "5", "6", "7", "8", "9", "10", "J", "Q", "K", "A", "2"];

  if (cardId < 0) {
    // Face-down card (not used here but kept for compatibility)
    return (
      <div
        className="w-full h-full rounded"
        style={{
          backgroundImage: "repeating-linear-gradient(45deg, #1d2021, #1d2021 2px, #323536 2px, #323536 4px)",
        }}
      />
    );
  }

  if (cardId >= 52) {
    const label = cardId === 52 ? "小" : "大";
    const color = cardId === 52 ? "text-[#1a1a1a]" : "text-[#c82014]";
    return (
      <>
        <span className={clsx(tiny ? "text-[7px]" : small ? "text-[11px]" : "text-xs", color, "font-bold")}>{label}</span>
        <span className={clsx(tiny ? "text-[5px]" : small ? "text-[8px]" : "text-[10px]", color)}>王</span>
      </>
    );
  }

  const face = cardId % 54;
  const suit = Math.floor(face / 13);
  const rank = face % 13;
  const isRed = suit === 1 || suit === 3;
  const colorClass = isRed ? "text-[#c82014]" : "text-[#1a1a1a]";

  return (
    <>
      <span className={clsx(tiny ? "text-[7px]" : small ? "text-[11px]" : "text-sm", colorClass, "font-bold leading-none")}>
        {rankChars[rank]}
      </span>
      <span className={clsx(tiny ? "text-[5px]" : small ? "text-[7px]" : "text-xs", colorClass, "leading-none")}>
        {suitChars[suit]}
      </span>
    </>
  );
}
