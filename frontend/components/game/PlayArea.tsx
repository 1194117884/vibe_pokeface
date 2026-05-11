"use client";

import { Card } from "./Card";

interface PlayAreaProps {
  plays: Array<{
    seat: number;
    cards: number[];
    playerName?: string;
  }>;
}

export function PlayArea({ plays }: PlayAreaProps) {
  if (plays.length === 0) {
    return (
      <div className="flex items-center justify-center min-h-[100px]">
        <p className="text-text-black-soft text-sm">等待出牌...</p>
      </div>
    );
  }

  return (
    <div className="flex items-center justify-center gap-4 min-h-[100px]">
      {plays.map((play, i) => (
        <div key={i} className="flex flex-col items-center gap-1">
          {play.playerName && (
            <span className="text-xs text-text-black-soft">{play.playerName}</span>
          )}
          <div className="flex gap-0.5">
            {play.cards.map((cardId, j) => (
              <Card key={j} cardId={cardId} small />
            ))}
          </div>
        </div>
      ))}
    </div>
  );
}
