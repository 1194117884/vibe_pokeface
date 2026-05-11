"use client";

import { useEffect, useState } from "react";
import { useParams } from "next/navigation";
import { GameTable } from "@/components/game/GameTable";
import { WSGameClient } from "@/lib/ws-game";

export default function RoomPage() {
  const params = useParams();
  const roomId = params.id as string;
  const [players, setPlayers] = useState<Array<{
    userId: number; name: string; seat: number; cardCount: number; isLandlord?: boolean; hand?: number[];
  }>>([]);
  const [mySeat, setMySeat] = useState<number>(0);
  const [currentSeat, setCurrentSeat] = useState<number>(0);
  const [phase, setPhase] = useState<"bidding" | "playing" | "ended">("bidding");
  const [plays, setPlays] = useState<Array<{ seat: number; cards: number[] }>>([]);
  const [landlordCards, setLandlordCards] = useState<number[]>([]);
  const [connected, setConnected] = useState(false);
  const [timer, setTimer] = useState<number | undefined>(undefined);

  useEffect(() => {
    const userId = 1; // TODO: decode from JWT properly
    const token = typeof window !== "undefined" ? localStorage.getItem("token") : null;
    if (!token) return;

    const client = new WSGameClient(userId, token);

    client.on("joined", () => {
      setConnected(true);
      client.joinRoom(roomId);
    });

    client.on("state_update", (msg) => {
      const data = msg.data as any;
      if (data?.players) {
        setPlayers(data.players.map((p: any) => ({
          userId: p.user_id,
          name: p.name || `Player ${p.seat + 1}`,
          seat: p.seat,
          cardCount: p.hand?.length || p.card_count || 0,
          isLandlord: p.is_landlord,
          hand: p.hand,
        })));
      }
      if (data?.current_seat !== undefined) setCurrentSeat(data.current_seat);
      if (data?.phase !== undefined) {
        setPhase(data.phase === 2 ? "ended" : data.phase === 1 ? "playing" : "bidding");
      }
      if (data?.landlord_cards) setLandlordCards(data.landlord_cards);
    });

    client.on("player_joined", (msg) => {
      const data = msg.data as any;
      if (data?.players) {
        setPlayers(data.players.map((p: any) => ({
          userId: p.user_id,
          name: `Player ${p.seat + 1}`,
          seat: p.seat,
          cardCount: 0,
        })));
      }
    });

    client.on("round_end", (msg) => {
      setPhase("ended");
    });

    client.connect();

    return () => {
      client.disconnect();
    };
  }, [roomId]);

  if (!connected) {
    return (
      <div className="flex items-center justify-center min-h-screen bg-cream">
        <p className="text-text-black-soft">Connecting to room...</p>
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-cream">
      <GameTable
        players={players}
        mySeat={mySeat}
        currentSeat={currentSeat}
        phase={phase}
        plays={plays}
        landlordCards={landlordCards}
      />
    </div>
  );
}
