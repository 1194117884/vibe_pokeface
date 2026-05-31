"use client";

import { useState, useMemo } from "react";
import clsx from "clsx";

interface HandCardsProps {
  cards: number[];
  onPlayCards?: (cardIds: number[]) => void;
  onSelectionChange?: (cardIds: number[]) => void;
  disabled?: boolean;
  compact?: boolean;
  mainCount?: number;
  hidePass?: boolean;
}

export function HandCards({ cards, onPlayCards, onSelectionChange, disabled, mainCount, hidePass }: HandCardsProps) {
  const [selected, setSelected] = useState<Set<number>>(new Set());

  // Split into rows: when mainCount is set, split at the main/secondary boundary.
  // Otherwise, for large hands (>24 cards), split at midpoint.
  const rows = useMemo(() => {
    if (mainCount != null && mainCount > 0 && mainCount < cards.length) {
      const first = cards.slice(0, mainCount);
      const second = cards.slice(mainCount);
      if (first.length > 18 || second.length > 18) {
        const mid = Math.ceil(cards.length / 2);
        return [cards.slice(0, mid), cards.slice(mid)];
      }
      return [first, second];
    }
    const n = cards.length;
    if (n > 24) {
      const mid = Math.ceil(n / 2);
      return [cards.slice(0, mid), cards.slice(mid)];
    }
    return [cards];
  }, [cards, mainCount]);

  const cardSize = useMemo(() => {
    const maxRow = Math.max(...rows.map((row) => row.length), 0);
    if (maxRow <= 8) return { width: 78, height: 112, margin: -30, rank: "text-2xl", suit: "text-lg" };
    if (maxRow <= 12) return { width: 68, height: 98, margin: -30, rank: "text-xl", suit: "text-base" };
    if (maxRow <= 18) return { width: 62, height: 90, margin: -36, rank: "text-lg", suit: "text-base" };
    if (maxRow <= 24) return { width: 60, height: 88, margin: -40, rank: "text-lg", suit: "text-base" };
    return { width: 58, height: 84, margin: -41, rank: "text-lg", suit: "text-base" };
  }, [rows]);

  const isSmall = cardSize.width < 60;
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
    onSelectionChange?.(Array.from(next));
  };

  const handlePlay = () => {
    if (disabled || selected.size === 0) return;
    const cardsArray = Array.from(selected);
    console.log("[HANDPLAY] Sending cards:", cardsArray.map(id => ({
      id,
      face: id % 54,
      suit: Math.floor((id % 54) / 13),
      suitChar: ["♠", "♥", "♣", "♦"][Math.floor((id % 54) / 13)],
      rank: ["3", "4", "5", "6", "7", "8", "9", "10", "J", "Q", "K", "A", "2"][(id % 54) % 13],
    })));
    onPlayCards?.(cardsArray);
    setSelected(new Set());
    onSelectionChange?.([]);
  };

  const handlePass = () => {
    if (disabled) return;
    onPlayCards?.([]);
    setSelected(new Set());
    onSelectionChange?.([]);
  };

  return (
    <div className="w-full overflow-hidden">
      {/* Hand cards — rows overlap: second row behind first, cards staggered */}
      {rows.map((row, ri) => (
        <div
          key={ri}
          className={clsx(
            "flex justify-center px-2 overflow-visible items-end relative",
            ri > 0 && "z-10",
            ri > 0 && (isSmall ? "-mt-11" : "-mt-14"),
            isSmall ? "min-h-[92px] pb-0" : "min-h-[112px] pb-1",
          )}
          style={ri > 0 ? { paddingLeft: 18 } : undefined}
        >
          {row.map((cardId, index) => (
            <div
              key={cardId}
              onClick={() => toggleCard(cardId)}
              className={clsx(
                "relative shrink-0 rounded-md flex flex-col items-center justify-center border border-black/20 bg-white shadow-md",
                selected.has(cardId)
                  ? "ring-4 ring-amber-300 -translate-y-3"
                  : !disabled && "cursor-pointer",
                "transition-transform duration-150 ease-out",
              )}
              style={{
                width: cardSize.width,
                height: cardSize.height,
                marginLeft: index === 0 ? 0 : cardSize.margin,
              }}
            >
              <MiniCardFace cardId={cardId} rankClass={cardSize.rank} suitClass={cardSize.suit} />
            </div>
          ))}
        </div>
      ))}

      {onPlayCards && !disabled && (
        <div className={clsx("grid grid-cols-2 gap-2 px-3", isMultiRow ? "pt-2" : "")}>
          <button
            onClick={handlePlay}
            disabled={disabled || selected.size === 0}
            className={clsx(
              "min-h-14 rounded-full gold-button px-4 py-3 text-lg font-black hover:brightness-110 active:scale-95 transition-all disabled:opacity-40",
              hidePass && "col-span-2"
            )}
          >
            出牌
          </button>
          {!hidePass && (
            <button
              onClick={handlePass}
              disabled={disabled}
              className="min-h-14 rounded-full emerald-button px-4 py-3 text-lg font-black hover:brightness-110 active:scale-95 transition-all disabled:opacity-40"
            >
              不出
            </button>
          )}
        </div>
      )}
    </div>
  );
}

function MiniCardFace({ cardId, rankClass, suitClass }: { cardId: number; rankClass: string; suitClass: string }) {
  const suitChars = ["♠", "♥", "♣", "♦"];
  const rankChars = ["3", "4", "5", "6", "7", "8", "9", "10", "J", "Q", "K", "A", "2"];

  // Use face (0-53) to identify the card, since dashengji uses 3-deck IDs 0-161
  const face = cardId % 54;

  if (face >= 52) {
    const label = face === 52 ? "小" : "大";
    const color = face === 52 ? "text-[#1a1a1a]" : "text-[#c82014]";
    return (
      <div className="absolute left-1.5 top-1 flex flex-col items-center">
        <span className={clsx(rankClass, color, "font-bold leading-none")}>{label}</span>
        <span className={clsx(suitClass, color, "leading-none")}>王</span>
      </div>
    );
  }

  const suit = Math.floor(face / 13);
  const rank = face % 13;
  const isRed = suit === 1 || suit === 3;
  const colorClass = isRed ? "text-[#c82014]" : "text-[#1a1a1a]";

  return (
    <>
      <div className="absolute left-1.5 top-1 flex flex-col items-center gap-0">
        <span className={clsx(rankClass, colorClass, "font-bold leading-none")}>
          {rankChars[rank]}
        </span>
        <span className={clsx(suitClass, colorClass, "leading-none")}>
          {suitChars[suit]}
        </span>
      </div>
    </>
  );
}
