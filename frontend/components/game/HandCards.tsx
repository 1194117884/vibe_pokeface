"use client";

import { useState, useMemo } from "react";
import clsx from "clsx";

interface HandCardsProps {
  cards: number[];
  onPlayCards?: (cardIds: number[]) => void;
  disabled?: boolean;
  compact?: boolean;
}

export function HandCards({ cards, onPlayCards, disabled, compact }: HandCardsProps) {
  const [selected, setSelected] = useState<Set<number>>(new Set());

  // Split into rows for large hands (>24 cards = 2 rows)
  const rows = useMemo(() => {
    const n = cards.length;
    if (n > 24) {
      const mid = Math.ceil(n / 2);
      return [cards.slice(0, mid), cards.slice(mid)];
    }
    return [cards];
  }, [cards]);

  const cardSize = useMemo(() => {
    const n = cards.length;
    if (n > 24) return "small" as const;     // 2 rows, small cards
    if (n > 12) return "small" as const;      // 1 row, small cards
    return "default" as const;                 // ≤12, full size
  }, [cards.length]);

  const overlap = cardSize === "small" ? "-space-x-8" : "-space-x-12";
  const isSmall = cardSize === "small";
  const isMultiRow = rows.length > 1;

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

  return (
    <div>
      {/* Hand cards — rows overlap: second row behind first, cards staggered */}
      {rows.map((row, ri) => (
        <div
          key={ri}
          className={clsx(
            "flex justify-center px-8 overflow-visible items-end relative",
            overlap,
            ri > 0 && "z-10",
            ri > 0 && isSmall && "-mt-10",
            ri > 0 && !isSmall && "-mt-24",
            isSmall ? "min-h-[70px] pb-0" : "min-h-[144px] pb-2",
          )}
          style={ri > 0 ? { paddingLeft: "1.5rem" } : undefined}
        >
          {row.map((cardId) => (
            <div
              key={cardId}
              onClick={() => toggleCard(cardId)}
              className={clsx(
                "relative rounded-md flex flex-col items-center justify-center border border-black/20 bg-white shadow-md",
                isSmall
                  ? "w-12 h-16 p-0.5 text-[11px] leading-none"
                  : "w-24 h-36 text-card-number",
                selected.has(cardId)
                  ? "ring-2 ring-blue-400 -translate-y-1"
                  : !isSmall && !disabled && "cursor-pointer",
                "transition-transform duration-150 ease-out",
              )}
            >
              <MiniCardFace cardId={cardId} small={isSmall} />
            </div>
          ))}
        </div>
      ))}

      {onPlayCards && (
        <div className={clsx("flex justify-center gap-6", isMultiRow && "pt-1")}>
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

function MiniCardFace({ cardId, small }: { cardId: number; small: boolean }) {
  const suitChars = ["♠", "♥", "♣", "♦"];
  const rankChars = ["3", "4", "5", "6", "7", "8", "9", "10", "J", "Q", "K", "A", "2"];

  // Use face (0-53) to identify the card, since dashengji uses 3-deck IDs 0-161
  const face = cardId % 54;

  if (face >= 52) {
    const label = face === 52 ? "小" : "大";
    const color = face === 52 ? "text-[#1a1a1a]" : "text-[#c82014]";
    return (
      <div className="absolute top-0.5 left-1 flex flex-col items-center">
        <span className={clsx(small ? "text-[11px]" : "text-xs", color, "font-bold leading-none")}>{label}</span>
        <span className={clsx(small ? "text-[7px]" : "text-[10px]", color, "leading-none")}>王</span>
      </div>
    );
  }

  const suit = Math.floor(face / 13);
  const rank = face % 13;
  const isRed = suit === 1 || suit === 3;
  const colorClass = isRed ? "text-[#c82014]" : "text-[#1a1a1a]";

  return (
    <>
      <div className={clsx("absolute top-0.5 left-1 flex flex-col items-center", small ? "gap-[-1px]" : "gap-0")}>
        <span className={clsx(small ? "text-[11px]" : "text-sm", colorClass, "font-bold leading-none")}>
          {rankChars[rank]}
        </span>
        <span className={clsx(small ? "text-[7px]" : "text-xs", colorClass, "leading-none")}>
          {suitChars[suit]}
        </span>
      </div>
    </>
  );
}
