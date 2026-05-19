"use client";

import { useState } from "react";
import clsx from "clsx";
import { Card } from "./Card";

interface HandCardsProps {
  cards: number[];
  onPlayCards?: (cardIds: number[]) => void;
  disabled?: boolean;
  compact?: boolean;
}

export function HandCards({ cards, onPlayCards, disabled, compact = false }: HandCardsProps) {
  const [selected, setSelected] = useState<Set<number>>(new Set());

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
    <div className="space-y-4">
      {/* Hand cards with overlap — matching Stitch design */}
      <div className={clsx("flex justify-center px-12 overflow-visible items-end pb-2",
        compact ? "-space-x-8 min-h-[100px]" : "-space-x-12 min-h-[144px]"
      )}>
        {cards.map((cardId) => (
          <Card
            key={cardId}
            cardId={cardId}
            selected={selected.has(cardId)}
            onClick={() => toggleCard(cardId)}
            medium={compact}
          />
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
