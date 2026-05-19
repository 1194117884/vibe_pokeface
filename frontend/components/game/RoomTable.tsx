"use client";

import { SeatPosition } from "./SeatPosition";
import { Card } from "./Card";

export interface TablePlayer {
  userId: string;
  name: string;
  nickname: string;
  characterId: string;
  seat: number;
  isBot: boolean;
  isOwner: boolean;
  isReady: boolean;
  isCurrentTurn?: boolean;
  isLandlord?: boolean;
  cardCount: number;
}

interface RoomTableProps {
  players: TablePlayer[];
  mySeat: number;
  phase: string;
  onSitDown: (seat: number) => void;
  onAddBot: () => void;
  landlordCards?: number[];
  landlordSeat?: number;
  lastPlay?: { seat: number; cards: number[] } | null;
  cardsLeftMessage?: string | null;
  maxPlayers?: number;
  tableSize?: "sm" | "lg";
  speechBubbles?: Record<number, string>;
}

export function RoomTable({
  players,
  mySeat,
  phase,
  onSitDown,
  onAddBot,
  landlordCards = [],
  landlordSeat = undefined,
  lastPlay = null,
  cardsLeftMessage = null,
  maxPlayers = 3,
  speechBubbles = {},
  compact = false,
}: RoomTableProps & { compact?: boolean }) {
  const seatMap = new Map<number, TablePlayer>();
  players.forEach((p) => seatMap.set(p.seat, p));

  const isGamePhase = phase !== "waiting" && phase !== "ended";

  const cardsLeftInfo = (() => {
    if (!cardsLeftMessage) return null;
    const match = cardsLeftMessage.match(/^seat_(\d+)_(baodan|baoshuang)$/);
    if (!match) return null;
    return { seat: parseInt(match[1]), type: match[2] as "baodan" | "baoshuang" };
  })();

  // During game phases: left opponent = next seat, right opponent = seat after that
  const leftSeat = (mySeat + 1) % maxPlayers;
  const rightSeat = (mySeat + 2) % maxPlayers;

  function renderSeat(seatNum: number, position?: "left" | "right") {
    const player = seatMap.get(seatNum);
    const isMine = seatNum === mySeat;
    const isLandlord = player?.isLandlord ?? false;

    return (
      <SeatPosition
        seatNumber={seatNum}
        player={
          player
            ? {
                userId: player.userId,
                name: player.name,
                nickname: player.nickname,
                characterId: player.characterId,
                isBot: player.isBot,
                isOwner: player.isOwner,
                isReady: player.isReady,
                isCurrentTurn: player.isCurrentTurn,
                isLandlord: player.isLandlord,
                cardCount: player.cardCount,
              }
            : null
        }
        isMySeat={isMine}
        position={position}
        cardsLeft={cardsLeftInfo?.seat === seatNum ? cardsLeftInfo.type : null}
        action={speechBubbles[seatNum] ?? null}
        landlordCards={isLandlord && landlordSeat === seatNum ? landlordCards : undefined}
        compact={compact}
        onChangeSeat={() => {
          if (!player && seatNum !== mySeat) {
            onSitDown(seatNum);
          }
        }}
        onAddBot={!player && seatNum !== mySeat ? onAddBot : undefined}
      />
    );
  }

  return (
    <div className="flex-1 flex flex-col items-center justify-between py-2 px-6 relative w-full">
      {/* Kitty / Landlord Cards (top center) — only during game phases */}
      {isGamePhase && (
        <div className="flex flex-col items-center gap-1 mt-2">
          <div className="flex gap-1.5">
            {landlordCards.length > 0
              ? landlordCards.map((cardId, i) => (
                  <Card key={i} cardId={cardId} small />
                ))
              : Array.from({ length: 3 }).map((_, i) => (
                  <div
                    key={i}
                    className="w-8 h-11 rounded shadow-md border border-white/10"
                    style={{
                      backgroundImage:
                        "repeating-linear-gradient(45deg, #1d2021, #1d2021 3px, #323536 3px, #323536 6px)",
                      transform: i === 0 ? "rotate(-4deg)" : i === 2 ? "rotate(4deg)" : "none",
                    }}
                  />
                ))}
          </div>
          <span className="text-[8px] font-black text-secondary-fixed/40 uppercase tracking-[0.2em]">
            底牌
          </span>
        </div>
      )}

      {/* Opponents + Timer Row (middle) */}
      {isGamePhase ? (
        <div className="w-full flex justify-between items-center px-4">
          {/* Left opponent */}
          {renderSeat(leftSeat, "left")}

          {/* Center: Timer or last play */}
          <div className="flex flex-col items-center gap-2">
            {lastPlay && phase === "playing" ? (
              <div className="flex flex-col items-center gap-1">
                <span className="text-on-surface-variant text-xs font-medium opacity-60">
                  {players.find((p) => p.seat === lastPlay.seat)?.nickname || `Player ${lastPlay.seat}`}
                </span>
                <div className="flex gap-1">
                  {lastPlay.cards.map((cardId, i) => (
                    <Card key={i} cardId={cardId} medium />
                  ))}
                </div>
              </div>
            ) : (
              <>
                {/* Circular timer */}
                <div className="relative w-16 h-16 flex items-center justify-center">
                  <svg className="absolute inset-0 w-full h-full -rotate-90">
                    <circle
                      cx="32"
                      cy="32"
                      r="28"
                      fill="transparent"
                      stroke="white"
                      strokeWidth="2"
                      className="opacity-10"
                    />
                    <circle
                      cx="32"
                      cy="32"
                      r="28"
                      fill="transparent"
                      stroke="#ffe16d"
                      strokeDasharray="176"
                      strokeDashoffset="0"
                      strokeLinecap="round"
                      strokeWidth="3"
                    />
                  </svg>
                  <span className="text-xl font-black text-secondary-fixed font-display-gold">
                    ∞
                  </span>
                </div>
              </>
            )}
          </div>

          {/* Right opponent */}
          {renderSeat(rightSeat, "right")}
        </div>
      ) : (
        /* Waiting phase: show all seats in a row */
        <div className="w-full flex justify-center gap-6 items-center">
          {Array.from({ length: maxPlayers }, (_, i) => (
            <div key={i}>{renderSeat(i)}</div>
          ))}
        </div>
      )}

      {/* Center status during waiting phase */}
      {!isGamePhase && (
        <div className="text-center mb-4">
          <span className="text-4xl select-none drop-shadow-lg opacity-50">🃏</span>
          <p className="text-white/40 text-sm font-medium mt-1">
            {phase === "ended" ? "已结束" : "等待中"}
          </p>
        </div>
      )}
    </div>
  );
}
