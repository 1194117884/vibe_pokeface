"use client";

import { Card } from "@/components/game/Card";

export interface DashengjiPlayer {
  userId: string;
  seat: number;
  nickname: string;
  cardCount: number;
  isBot: boolean;
  isOwner: boolean;
  characterId?: string;
  isCurrentTurn: boolean;
  isDealerTeam: boolean;
  revealedHand?: number[];
}

interface DashengjiTableProps {
  players: DashengjiPlayer[];
  mySeat: number;
  trumpSuit: number;
  bottomCards: number[];
  lastPlay: { seat: number; cards: number[] } | null;
}

export function DashengjiTable({ players, mySeat, trumpSuit, bottomCards, lastPlay }: DashengjiTableProps) {
  const suitSymbols = ["♠", "♥", "♣", "♦"];

  const partner = players.find((p) => p.seat === (mySeat + 2) % 4);
  const left = players.find((p) => p.seat === (mySeat + 1) % 4);
  const right = players.find((p) => p.seat === (mySeat + 3) % 4);
  const me = players.find((p) => p.seat === mySeat);

  const renderSeat = (player: DashengjiPlayer | undefined, position: string) => {
    if (!player) return null;
    const posClass =
      position === "top" ? "top-4 left-1/2 -translate-x-1/2"
      : position === "left" ? "left-4 top-1/2 -translate-y-1/2"
      : position === "right" ? "right-4 top-1/2 -translate-y-1/2"
      : "bottom-4 left-1/2 -translate-x-1/2";

    return (
      <div className={`absolute flex flex-col items-center gap-2 ${posClass}`}>
        <div className="flex items-center gap-2">
          <span className="text-white text-sm font-medium">{player.nickname}</span>
          {player.isDealerTeam && <span className="text-xs px-1.5 py-0.5 rounded bg-amber-600/80 text-white">庄</span>}
          {player.isCurrentTurn && <span className="w-2 h-2 rounded-full bg-green-400 animate-pulse" />}
        </div>
        <div className="flex -space-x-2">
          {player.revealedHand
            ? player.revealedHand.map((id, i) => <Card key={i} cardId={id} small />)
            : Array.from({ length: Math.min(player.cardCount, 10) }).map((_, i) => <Card key={i} cardId={-1} faceDown small />)}
        </div>
        <span className="text-white/60 text-xs">{player.cardCount}张</span>
      </div>
    );
  };

  return (
    <div className="relative w-full h-full min-h-[500px]">
      {trumpSuit >= 0 && (
        <div className="absolute top-1/2 left-1/2 -translate-x-1/2 -translate-y-1/2 text-white/80 text-center">
          <div className="text-4xl">{suitSymbols[trumpSuit]}</div>
          <div className="text-sm">主花色</div>
        </div>
      )}
      {bottomCards.length > 0 && (
        <div className="absolute top-1/2 left-1/2 -translate-x-1/2 translate-y-8 flex -space-x-1">
          {bottomCards.map((id, i) => <Card key={i} cardId={id} small />)}
        </div>
      )}
      {lastPlay && (
        <div className="absolute top-1/2 left-1/2 -translate-x-1/2 -translate-y-16 flex -space-x-1">
          {lastPlay.cards.map((id, i) => <Card key={i} cardId={id} medium />)}
        </div>
      )}
      {renderSeat(partner, "top")}
      {renderSeat(left, "left")}
      {renderSeat(right, "right")}
      {me && <div className="absolute bottom-4 left-1/2 -translate-x-1/2 text-white/40 text-sm">{me.nickname} (你)</div>}
    </div>
  );
}
