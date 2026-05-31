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
  isDealerTeam?: boolean;
  cardCount: number;
  hand?: number[];
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
  trickPlays?: Record<number, number[]>;
  bottomCards?: number[];
  bottomSeat?: number;
  discardedCards?: number[];
  cardsLeftMessage?: string | null;
  maxPlayers?: number;
  tableSize?: "sm" | "lg";
  speechBubbles?: Record<number, string>;
  waitingLayout?: "grid" | "row";
  hideCardCount?: boolean;
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
  trickPlays = {},
  bottomCards = [],
  bottomSeat = undefined,
  discardedCards = [],
  cardsLeftMessage = null,
  maxPlayers = 3,
  speechBubbles = {},
  waitingLayout = "grid",
  compact = true,
  hideCardCount = false,
}: RoomTableProps & { compact?: boolean }) {
  const seatMap = new Map<number, TablePlayer>();
  players.forEach((p) => seatMap.set(p.seat, p));

  const isGamePhase = phase !== "waiting" && phase !== "ended";

  const cardsLeftInfo = (() => {
    if (!cardsLeftMessage) return null;
    const match = cardsLeftMessage.match(/^seat_(\d+)_(baodan|baoshuang)$/);
    if (!match) return null;
    return {
      seat: parseInt(match[1]),
      type: match[2] as "baodan" | "baoshuang",
    };
  })();

  // During game phases: seat positions depend on player count
  const leftSeat = (mySeat + 1) % maxPlayers;
  const rightSeat =
    maxPlayers === 4 ? (mySeat + 3) % maxPlayers : (mySeat + 2) % maxPlayers;
  const partnerSeat = maxPlayers === 4 ? (mySeat + 2) % maxPlayers : -1;

  function renderSeat(seatNum: number, position?: "left" | "right" | "top") {
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
                isDealerTeam: player.isDealerTeam,
                cardCount: player.cardCount,
                hand: player.hand,
              }
            : null
        }
        isMySeat={isMine}
        position={position}
        phase={phase}
        cardsLeft={cardsLeftInfo?.seat === seatNum ? cardsLeftInfo.type : null}
        action={speechBubbles[seatNum] ?? null}
        landlordCards={
          isLandlord && landlordSeat === seatNum ? landlordCards : undefined
        }
        bottomCards={bottomSeat === seatNum ? bottomCards : undefined}
        discardedCards={bottomSeat === seatNum ? discardedCards : undefined}
        compact={compact}
        hideCardCount={hideCardCount}
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
    <div className="flex-1 flex flex-col items-center justify-between px-2 pb-44 relative w-full min-h-0">
      {/* Opponents + Timer Row (middle) */}
      {isGamePhase ? (
        <div className="w-full flex flex-col items-center gap-1">
          {partnerSeat >= 0 && renderSeat(partnerSeat, "top")}
          <div className="w-full grid grid-cols-[1fr_minmax(70px,auto)_1fr] justify-items-center items-center gap-y-1 px-1">
            {/* Row 1: top trick cards */}
            <div />
            <div>
              {trickPlays[partnerSeat] && (
                <div className="flex -space-x-5">
                  {trickPlays[partnerSeat].map((cardId, i) => (
                    <Card key={i} cardId={cardId} small />
                  ))}
                </div>
              )}
            </div>
            <div />

            {/* Row 2: left seat+trick, center, right seat+trick */}
            <div className="flex items-center gap-1 justify-self-start">
              {renderSeat(leftSeat, "left")}
              {trickPlays[leftSeat] && (
                <div className="flex -space-x-5">
                  {trickPlays[leftSeat].map((cardId, i) => (
                    <Card key={i} cardId={cardId} small />
                  ))}
                </div>
              )}
            </div>
            <div className="flex flex-col items-center gap-2"></div>
            <div className="flex items-center gap-1 justify-self-end">
              {trickPlays[rightSeat] && (
                <div className="flex -space-x-5">
                  {trickPlays[rightSeat].map((cardId, i) => (
                    <Card key={i} cardId={cardId} small />
                  ))}
                </div>
              )}
              {renderSeat(rightSeat, "right")}
            </div>

            {/* Row 3: bottom (my) trick cards */}
            <div />
            <div>
              {trickPlays[mySeat] && (
                <div className="flex -space-x-5">
                  {trickPlays[mySeat].map((cardId, i) => (
                    <Card key={i} cardId={cardId} small />
                  ))}
                </div>
              )}
            </div>
            <div />
          </div>
        </div>
      ) : (
        /* Waiting phase: show all seats in a row */
        <div
          className={
            waitingLayout === "row"
              ? "flex w-full max-w-[430px] items-start justify-between gap-1 px-1"
              : "grid w-full max-w-[430px] grid-cols-2 gap-2 px-3"
          }
        >
          {Array.from({ length: maxPlayers }, (_, i) => (
            <div
              key={i}
              className={
                waitingLayout === "row"
                  ? "flex min-w-0 flex-1 justify-center"
                  : maxPlayers === 3 && i === 2
                    ? "col-span-2 flex justify-center"
                    : "flex justify-center"
              }
            >
              {renderSeat(i)}
            </div>
          ))}
        </div>
      )}

      {/* Center status during waiting phase */}
      {!isGamePhase && (
        <div className="text-center mb-4">
          <span className="text-5xl select-none drop-shadow-lg opacity-50">
            🃏
          </span>
          <p className="text-white/60 text-lg font-bold mt-1">
            {phase === "ended" ? "已结束" : "等待中"}
          </p>
        </div>
      )}
    </div>
  );
}
