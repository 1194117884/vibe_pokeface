"use client";

import { SeatPosition } from "./SeatPosition";
import { useRoomTheme } from "@/themes";

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
  cardCount: number;
}

interface RoomTableProps {
  players: TablePlayer[];
  mySeat: number;
  phase: string;
  onSitDown: (seat: number) => void;
  onAddBot: () => void;
  landlordCards?: number[];
  lastPlay?: { seat: number; cards: number[] } | null;
  cardsLeftMessage?: string | null;
}

export function RoomTable({
  players,
  mySeat,
  phase,
  onSitDown,
  onAddBot,
  landlordCards = [],
  lastPlay = null,
  cardsLeftMessage = null,
}: RoomTableProps) {
  const theme = useRoomTheme();

  const seatMap = new Map<number, TablePlayer>();
  players.forEach((p) => seatMap.set(p.seat, p));

  // For 3-player doudizhu: seats 0, 1, 2
  const activeSeats = [0, 1, 2];

  // Center text based on phase
  const centerText = phase === "playing" ? "游戏中" : phase === "ended" ? "已结束" : "等待中";

  // Parse cards left message for seat badge
  const cardsLeftInfo = (() => {
    if (!cardsLeftMessage) return null;
    const match = cardsLeftMessage.match(/^seat_(\d+)_(baodan|baoshuang)$/);
    if (!match) return null;
    return { seat: parseInt(match[1]), type: match[2] as "baodan" | "baoshuang" };
  })();

  return (
    <div className="relative w-full max-w-3xl mx-auto aspect-[4/3]">
      {/* Theme-aware felt table */}
      <div
        className="absolute inset-0 rounded-[40%] shadow-xl"
        style={{
          backgroundColor: "var(--felt-color, #1B5E20)",
          boxShadow: "var(--felt-shadow, 0 20px 60px rgba(0,0,0,0.5))",
          borderWidth: "var(--table-border-width, 8px)",
          borderStyle: "solid",
          borderColor: "var(--table-border-color, #8B4513)",
        }}
      >
        <div
          className="absolute inset-4 rounded-[35%]"
          style={{ backgroundColor: "var(--felt-color, #1B5E20)", opacity: 0.3 }}
        />
      </div>

      {/* Center decoration — show landlord cards, last play, or default */}
      <div className="absolute top-1/2 left-1/2 -translate-x-1/2 -translate-y-1/2 text-center z-10">
        {lastPlay && phase === "playing" ? (
          <div className="flex flex-col items-center gap-1">
            <span className="text-xs text-white/70 font-medium">
              {players.find(p => p.seat === lastPlay.seat)?.nickname || `Player ${lastPlay.seat}`}
            </span>
            <div className="flex gap-0.5">
              {lastPlay.cards.map((cardId, i) => (
                <div key={i} className="w-7 h-10 bg-white rounded shadow-md flex items-center justify-center text-xs font-bold"
                  style={{
                    color: cardId >= 52 ? "#d32f2f" : [1, 2, 3].includes(Math.floor(cardId / 13)) ? "#d32f2f" : "#333",
                  }}>
                  {cardDisplay(cardId)}
                </div>
              ))}
            </div>
          </div>
        ) : landlordCards && landlordCards.length > 0 && (phase === "playing" || phase === "bidding") ? (
          <div className="flex flex-col items-center gap-1">
            <span className="text-xs text-white/70 font-medium">底牌</span>
            <div className="flex gap-0.5">
              {landlordCards.map((cardId, i) => (
                <div key={i} className="w-7 h-10 bg-white rounded shadow-md flex items-center justify-center text-xs font-bold"
                  style={{
                    color: cardId >= 52 ? "#d32f2f" : [1, 2, 3].includes(Math.floor(cardId / 13)) ? "#d32f2f" : "#333",
                  }}>
                  {cardDisplay(cardId)}
                </div>
              ))}
            </div>
          </div>
        ) : (
          <>
            <span className="text-4xl select-none drop-shadow-lg">
              {theme.table.decoration}
            </span>
            <p className="text-white/60 text-sm font-medium mt-1">{centerText}</p>
          </>
        )}
      </div>

      {/* Seats positioned around the table */}
      <div className="absolute inset-0 grid grid-cols-3 grid-rows-3 p-4">
        <div className="col-start-2 row-start-1 flex justify-center items-start pt-2">
          {renderSeat(activeSeats[0])}
        </div>
        <div className="col-start-1 row-start-2 flex items-center justify-start pl-2">
          {renderSeat(activeSeats[1])}
        </div>
        <div className="col-start-3 row-start-2 flex items-center justify-end pr-2">
          {renderSeat(activeSeats[2])}
        </div>
        <div className="col-start-2 row-start-3 flex justify-center items-end pb-2">
          {renderSeat(mySeat)}
        </div>
      </div>
    </div>
  );

  function renderSeat(seatNum: number) {
    const player = seatMap.get(seatNum);
    const isMine = seatNum === mySeat;
    const posLabel = seatNum === 0 ? "top" : seatNum === 1 ? "left" : "bottom";

    return (
      <SeatPosition
        seatNumber={seatNum}
        player={player ? {
          userId: player.userId,
          name: player.name,
          nickname: player.nickname,
          characterId: player.characterId,
          isBot: player.isBot,
          isOwner: player.isOwner,
          isReady: player.isReady,
          isCurrentTurn: player.isCurrentTurn,
          cardCount: player.cardCount,
        } : null}
        position={posLabel}
        isMySeat={isMine}
        cardsLeft={cardsLeftInfo?.seat === seatNum ? cardsLeftInfo.type : null}
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

function cardDisplay(cardId: number): string {
  const suits = ["♠", "♥", "♣", "♦"];
  const ranks = ["3", "4", "5", "6", "7", "8", "9", "10", "J", "Q", "K", "A", "2"];
  if (cardId === 52) return "🃏";
  if (cardId === 53) return "👑";
  return suits[Math.floor(cardId / 13)] + ranks[cardId % 13];
}
