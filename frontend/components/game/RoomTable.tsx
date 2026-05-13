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
}

export function RoomTable({
  players,
  mySeat,
  phase,
  onSitDown,
  onAddBot,
}: RoomTableProps) {
  const theme = useRoomTheme();

  const seatMap = new Map<number, TablePlayer>();
  players.forEach((p) => seatMap.set(p.seat, p));

  // For 3-player doudizhu: seats 0, 1, 2
  const activeSeats = [0, 1, 2];

  // Center text based on phase
  const centerText = phase === "playing" ? "游戏中" : phase === "ended" ? "已结束" : "等待中";

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

      {/* Center decoration */}
      <div className="absolute top-1/2 left-1/2 -translate-x-1/2 -translate-y-1/2 text-center">
        <span className="text-4xl select-none drop-shadow-lg">
          {theme.table.decoration}
        </span>
        <p className="text-white/60 text-sm font-medium mt-1">{centerText}</p>
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
