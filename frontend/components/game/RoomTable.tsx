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

interface SeatLayout {
  seatNum: number;
  left: string;
  top: string;
}

const TABLE_SIZES: Record<string, string> = {
  sm: "max-w-lg",
  lg: "max-w-5xl",
};

const SEAT_RADIUS: Record<string, number> = {
  sm: 32,
  lg: 36,
};

function calcSeatPositions(numSeats: number, mySeat: number, radius: number): SeatLayout[] {
  return Array.from({ length: numSeats }, (_, i) => {
    const seatNum = (mySeat + i) % numSeats;
    const angleDeg = 90 + (i / numSeats) * 360;
    const angleRad = (angleDeg * Math.PI) / 180;
    return {
      seatNum,
      left: `${50 + radius * Math.cos(angleRad)}%`,
      top: `${50 + radius * Math.sin(angleRad)}%`,
    };
  });
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
  tableSize = "lg",
  speechBubbles = {},
  compact = false,
}: RoomTableProps & { compact?: boolean }) {
  const seatMap = new Map<number, TablePlayer>();
  players.forEach((p) => seatMap.set(p.seat, p));

  const radius = compact ? 24 : SEAT_RADIUS[tableSize];
  const seatLayouts = calcSeatPositions(maxPlayers, mySeat, radius);

  const centerText = phase === "playing" ? "游戏中" : phase === "ended" ? "已结束" : "等待中";

  const cardsLeftInfo = (() => {
    if (!cardsLeftMessage) return null;
    const match = cardsLeftMessage.match(/^seat_(\d+)_(baodan|baoshuang)$/);
    if (!match) return null;
    return { seat: parseInt(match[1]), type: match[2] as "baodan" | "baoshuang" };
  })();

  return (
    <div className={`room-table relative w-full ${TABLE_SIZES[tableSize]} mx-auto aspect-[4/3]`}>
      {/* Stitch-style dark green poker table */}
      <div
        className="absolute inset-0 rounded-[48px]"
        style={{
          background: "radial-gradient(circle, #226a4b 0%, #003824 100%)",
          boxShadow:
            "0 0 6px rgba(0,0,0,0.24), 0 8px 12px rgba(0,0,0,0.14), 0 20px 60px rgba(0,0,0,0.5)",
          borderWidth: "12px",
          borderStyle: "solid",
          borderColor: "#1a1a1a",
        }}
      >
        {/* Inner felt rim */}
        <div
          className="absolute inset-4 rounded-[36px]"
          style={{ backgroundColor: "#003824", opacity: 0.4 }}
        />
      </div>

      {/* Center: The Kitty (landlord cards) or last play */}
      <div className="absolute top-1/2 left-1/2 -translate-x-1/2 -translate-y-1/2 text-center z-10">
        {lastPlay && phase === "playing" ? (
          <div className="flex flex-col items-center gap-2">
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
            <span className="text-4xl select-none drop-shadow-lg opacity-50">🃏</span>
            <p className="text-white/40 text-sm font-medium mt-1">{centerText}</p>
          </>
        )}
      </div>

      {/* Seats positioned around the table */}
      {seatLayouts.map(({ seatNum, left, top }) => (
        <div
          key={seatNum}
          className="absolute z-10 pointer-events-none"
          style={{ left, top, transform: "translate(-50%, -50%)" }}
        >
          <div className="pointer-events-auto">
            {renderSeat(seatNum)}
          </div>
        </div>
      ))}
    </div>
  );

  function renderSeat(seatNum: number) {
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
}
