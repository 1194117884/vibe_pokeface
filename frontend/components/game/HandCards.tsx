"use client";

import { useState } from "react";
import { Card } from "./Card";

interface HandCardsProps {
  cards: number[];
  onPlayCards?: (cardIds: number[]) => void;
  disabled?: boolean;
}

export function HandCards({ cards, onPlayCards, disabled }: HandCardsProps) {
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

  return (
    <div className="space-y-2">
      <div className="flex justify-center gap-1 px-4 py-2 min-h-[72px] flex-wrap">
        {cards.map((cardId) => (
          <Card
            key={cardId}
            cardId={cardId}
            selected={selected.has(cardId)}
            onClick={() => toggleCard(cardId)}
          />
        ))}
      </div>
      {onPlayCards && (
        <div className="flex justify-center gap-2">
          <button
            onClick={handlePlay}
            disabled={disabled || selected.size === 0}
            className="px-4 py-1.5 bg-green-accent text-white rounded-pill text-sm font-semibold
                       disabled:opacity-40 transition-all active:scale-95"
          >
            出牌
          </button>
          <button
            onClick={() => { onPlayCards([]); setSelected(new Set()); }}
            disabled={disabled}
            className="px-4 py-1.5 bg-white text-gray-700 border border-gray-300 rounded-pill text-sm font-semibold
                       disabled:opacity-40 transition-all active:scale-95"
          >
            不出
          </button>
        </div>
      )}
    </div>
  );
}
