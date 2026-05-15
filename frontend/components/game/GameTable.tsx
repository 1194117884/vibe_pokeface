"use client";

import { PlayerInfo } from "./PlayerInfo";
import { HandCards } from "./HandCards";
import { PlayArea } from "./PlayArea";
import { ActionBar } from "./ActionBar";

interface PlayerData {
  userId: number;
  name: string;
  seat: number;
  cardCount: number;
  isLandlord?: boolean;
  hand?: number[];
}

interface GameTableProps {
  players: PlayerData[];
  currentSeat?: number;
  mySeat?: number;
  phase: "calling" | "snatching" | "playing" | "ended";
  plays?: Array<{ seat: number; cards: number[] }>;
  landlordCards?: number[];
  onPlayCards?: (cardIds: number[]) => void;
  onBidCall?: () => void;
  onBidPass?: () => void;
  timer?: number;
}

export function GameTable({
  players,
  currentSeat,
  mySeat,
  phase,
  plays = [],
  landlordCards = [],
  onPlayCards,
  onBidCall,
  onBidPass,
  timer,
}: GameTableProps) {
  const me = players.find(p => p.seat === mySeat);
  const others = players.filter(p => p.seat !== mySeat);
  const isMyTurn = currentSeat === mySeat;

  return (
    <div className="max-w-2xl mx-auto p-4">
      {/* Opponents */}
      <div className="flex justify-between mb-4">
        {others.map((p) => (
          <PlayerInfo
            key={p.seat}
            name={p.name}
            cardCount={p.cardCount}
            seat={p.seat}
            isLandlord={p.isLandlord}
            isCurrentTurn={currentSeat === p.seat}
          />
        ))}
      </div>

      {/* Landlord cards */}
      {landlordCards.length > 0 && (
        <div className="flex justify-center gap-1 mb-4">
          {landlordCards.map((id, i) => (
            <div
              key={i}
              className="bg-house-green border-2 border-white/30 rounded-lg shadow-md
                         w-8 h-12 flex items-center justify-center"
            >
              <span className="text-white/50 text-lg font-bold">?</span>
            </div>
          ))}
        </div>
      )}

      {/* Play area */}
      <PlayArea
        plays={plays.map(p => ({
          ...p,
          playerName: players.find(pl => pl.seat === p.seat)?.name,
        }))}
      />

      {/* My cards */}
      {me && (
        <div className="mt-4">
          <PlayerInfo
            name={me.name}
            cardCount={me.cardCount}
            seat={me.seat}
            isLandlord={me.isLandlord}
            isCurrentTurn={isMyTurn}
            isSelf
          />
          <HandCards cards={me.hand || []} onPlayCards={onPlayCards} disabled={!isMyTurn} />
        </div>
      )}

      {/* Action bar */}
      <ActionBar
        phase={phase}
        isMyTurn={isMyTurn}
        onBidCall={onBidCall}
        onBidPass={onBidPass}
        timer={timer}
      />
    </div>
  );
}
