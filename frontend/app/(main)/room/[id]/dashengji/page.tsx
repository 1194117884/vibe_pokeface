"use client";

import { useCallback, useEffect, useState, useRef } from "react";
import { useParams, useRouter } from "next/navigation";
import clsx from "clsx";
import { WSGameClient, formatError, type ErrorData } from "@/lib/ws-game";
import { RoomTable, TablePlayer } from "@/components/game/RoomTable";
import { ReadyBar } from "@/components/game/ReadyBar";
import { HandCards } from "@/components/game/HandCards";
import { DashengjiActionBar } from "@/components/game/dashengji/DashengjiActionBar";
import { ScoreBoard } from "@/components/game/dashengji/ScoreBoard";

interface ServerPlayer {
  user_id?: number | string;
  userId?: number | string;
  seat?: number;
  is_bot?: boolean;
  isBot?: boolean;
  is_owner?: boolean;
  isOwner?: boolean;
  ready?: boolean;
  isReady?: boolean;
  hand?: Array<{ id: number } | number>;
  card_count?: number;
  cardCount?: number;
  nickname?: string;
  character_id?: string;
  characterId?: string;
}

interface ServerData {
  players?: ServerPlayer[];
  seat?: number;
  user_id?: number | string;
  phase?: number;
  current_seat?: number;
  trump_suit?: number;
  current_level?: number;
  dealer_seats?: [number, number];
  bottom_cards?: Array<{ id: number } | number>;
  round_points?: number;
  last_play?: {
    seat: number;
    cards: Array<{ id: number } | number>;
  } | null;
  scores?: Array<{ player_id: number; score: number }>;
}

const phaseMap: Record<number, string> = {
  0: "set_trump",
  1: "counter_trump",
  2: "take_bottom",
  3: "discard_bottom",
  4: "playing",
  5: "ended",
};

const GAME_CONFIG = { maxPlayers: 4, tableSize: "sm" as const };

function getUserIdFromToken(): string {
  if (typeof window === "undefined") return "";
  const token = localStorage.getItem("token");
  if (!token) return "";
  try {
    const payload = JSON.parse(atob(token.split(".")[1]));
    return String(payload.user_id ?? payload.sub ?? "");
  } catch {
    return "";
  }
}

function extractHandCards(p: ServerPlayer): number[] {
  if (!p?.hand) return [];
  return p.hand.map((c) => (typeof c === "number" ? c : c.id));
}

function toTablePlayer(p: ServerPlayer, currentSeat: number | undefined, dealerSeats?: [number, number]): TablePlayer {
  const seat = p.seat ?? 0;
  return {
    userId: String(p.user_id ?? p.userId ?? ""),
    name: p.nickname ?? String(p.user_id ?? ""),
    nickname: p.nickname ?? String(p.user_id ?? "").replace(/^ai:bot:/, "AI "),
    characterId: p.character_id ?? p.characterId ?? "",
    seat,
    isBot: p.is_bot ?? p.isBot ?? false,
    isOwner: p.is_owner ?? p.isOwner ?? false,
    isReady: p.ready ?? p.isReady ?? false,
    isCurrentTurn: p.seat === currentSeat,
    isDealerTeam: dealerSeats ? dealerSeats.includes(seat) : undefined,
    cardCount: Array.isArray(p.hand) ? p.hand.length : (p.card_count ?? p.cardCount ?? 0),
    hand: extractHandCards(p).length > 0 ? extractHandCards(p) : undefined,
  };
}

function extractCards(arr: Array<{ id: number } | number> | undefined): number[] {
  if (!arr) return [];
  return arr.map((c: { id: number } | number) => (typeof c === "number" ? c : c.id));
}

export default function DashengjiRoomPage() {
  const params = useParams();
  const router = useRouter();
  const roomId = params.id as string;

  const [players, setPlayers] = useState<TablePlayer[]>([]);
  const [mySeat, setMySeat] = useState<number | null>(null);
  const mySeatRef = useRef<number | null>(null);
  const setMySeatWithRef = (seat: number | null) => {
    mySeatRef.current = seat;
    setMySeat(seat);
  };
  const [phase, setPhase] = useState<string>("waiting");
  const [hand, setHand] = useState<number[]>([]);
  const [currentSeat, setCurrentSeat] = useState<number | undefined>(undefined);
  const [trumpSuit, setTrumpSuit] = useState(-1);
  const [currentLevel, setCurrentLevel] = useState(3);
  const [bottomCards, setBottomCards] = useState<number[]>([]);
  const [lastPlay, setLastPlay] = useState<{ seat: number; cards: number[] } | null>(null);
  const [roundPoints, setRoundPoints] = useState(0);
  const [roundResult, setRoundResult] = useState<{ scores: Array<{ player_id: number; score: number }> } | null>(null);
  const [errorMessage, setErrorMessage] = useState<string | null>(null);
  const [dealerSeats, setDealerSeats] = useState<[number, number]>([0, 2]);
  const dealerSeatsRef = useRef<[number, number]>([0, 2]);
  const setDealerSeatsWithRef = (seats: [number, number]) => {
    dealerSeatsRef.current = seats;
    setDealerSeats(seats);
  };

  const wsRef = useRef<WSGameClient | null>(null);

  useEffect(() => {
    const token = typeof window !== "undefined" ? localStorage.getItem("token") : null;
    if (!token) {
      router.push("/auth/login");
      return;
    }

    let userIdStr = "";
    try {
      const payload = JSON.parse(atob(token.split(".")[1]));
      userIdStr = String(payload.user_id ?? payload.sub ?? "");
    } catch {
      router.push("/auth/login");
      return;
    }

    const client = new WSGameClient(Number(userIdStr), token, roomId, "dashengji");
    wsRef.current = client;
    const uid = userIdStr;

    client.on("player_joined", (msg) => {
      const data = msg.data as ServerData;
      if (data?.players) {
        setPlayers(data.players.map((p) => toTablePlayer(p, undefined, undefined)));
      }
      if (data?.seat !== undefined && String(data.user_id) === uid) setMySeatWithRef(data.seat);
    });

    client.on("player_ready", (msg) => {
      const data = msg.data as ServerData;
      if (data?.players) {
        setPlayers(data.players.map((p) => toTablePlayer(p, currentSeat, dealerSeatsRef.current)));
      }
    });

    client.on("state_update", (msg) => {
      const data = msg.data as ServerData;

      if (data?.phase !== undefined) {
        setPhase(phaseMap[data.phase] ?? "waiting");
      }
      if (data?.current_seat !== undefined) setCurrentSeat(data.current_seat);
      if (data?.trump_suit !== undefined) setTrumpSuit(data.trump_suit);
      if (data?.current_level !== undefined) setCurrentLevel(data.current_level);
      if (data?.dealer_seats) setDealerSeatsWithRef(data.dealer_seats);
      if (data?.bottom_cards) setBottomCards(extractCards(data.bottom_cards));
      if (data?.round_points !== undefined) setRoundPoints(data.round_points);
      if (data?.last_play) {
        setLastPlay({ seat: data.last_play.seat, cards: extractCards(data.last_play.cards) });
      }
      if (data?.players) {
        const ds = data.dealer_seats || dealerSeatsRef.current;
        setPlayers((prev) => {
          const newPlayers = data.players!.map((p) => toTablePlayer(p, data.current_seat, ds));
          // Merge with previous to preserve known nicknames
          return newPlayers.map((np) => {
            const existing = prev.find((pp) => pp.seat === np.seat);
            if (existing && existing.nickname && existing.nickname !== String(existing.seat)) {
              return { ...np, nickname: existing.nickname, name: existing.nickname, characterId: existing.characterId || np.characterId };
            }
            return np;
          });
        });
      }
      if (mySeatRef.current !== null && data?.players) {
        const me = data.players.find(
          (p) => (p.seat ?? 0) === mySeatRef.current,
        );
        if (me) setHand(extractHandCards(me));
      }
    });

    client.on("game_start", (msg) => {
      const data = msg.data as ServerData;
      if (data?.dealer_seats) setDealerSeatsWithRef(data.dealer_seats);
      const ds = data.dealer_seats || dealerSeatsRef.current;
      if (data?.players) {
        setPlayers((prev) => {
          const newPlayers = data.players!.map((p) => toTablePlayer(p, data.current_seat, ds));
          return newPlayers.map((np) => {
            const existing = prev.find((pp) => pp.seat === np.seat);
            if (existing && existing.nickname && existing.nickname !== String(existing.seat)) {
              return { ...np, nickname: existing.nickname, name: existing.nickname, characterId: existing.characterId || np.characterId };
            }
            return np;
          });
        });
        if (mySeatRef.current !== null) {
          const me = data.players.find(
            (p) => (p.seat ?? 0) === mySeatRef.current,
          );
          if (me) setHand(extractHandCards(me));
        }
      }
      setPhase("set_trump");
      setRoundResult(null);
      if (data?.current_seat !== undefined) setCurrentSeat(data.current_seat);
      if (data?.trump_suit !== undefined) setTrumpSuit(data.trump_suit);
      if (data?.current_level !== undefined) setCurrentLevel(data.current_level);
    });

    client.on("round_end", (msg) => {
      setPhase("ended");
      setHand([]);
      setLastPlay(null);
      const data = msg.data as { scores?: Array<{ player_id: number; score: number }> };
      if (data?.scores) {
        setRoundResult({ scores: data.scores });
      }
    });

    client.on("error", (msg) => {
      const data = msg.data;
      if (typeof data === "object" && data !== null && "code" in data) {
        const errData = data as ErrorData;
        const formatted = formatError(errData);
        setErrorMessage(formatted);
        setTimeout(() => setErrorMessage(null), 4000);
        return;
      }
      const errMsg = (msg.data as string) ?? msg.error ?? "";
      if (errMsg.indexOf("room is full") !== -1 || errMsg.indexOf("room is closed") !== -1) {
        window.location.replace("/lobby");
        return;
      }
    });

    client.on("player_left", (msg) => {
      const data = msg.data as ServerData;
      if (data?.players) {
        setPlayers(data.players.map((p) => toTablePlayer(p, currentSeat, dealerSeatsRef.current)));
      }
    });

    client.connect();
    return () => {
      wsRef.current = null;
      client.disconnect();
    };
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [roomId, router]);

  const handleAction = useCallback((action: string, cards?: number[]) => {
    wsRef.current?.sendAction(action, cards);
  }, []);

  const handlePlayCards = useCallback((cards: number[]) => {
    const action = cards.length === 0 ? "pass" : "play";
    wsRef.current?.sendAction(action, cards);
  }, []);

  const handleReady = useCallback(() => {
    wsRef.current?.sendReady();
  }, []);

  const handleStartGame = useCallback(() => {
    wsRef.current?.startGame();
  }, []);

  const handleAddBot = useCallback(() => {
    wsRef.current?.addBot();
  }, []);

  const handleSitDown = useCallback((seat: number) => {
    wsRef.current?.changeSeat(seat);
  }, []);

  const myUserId = getUserIdFromToken();
  const myPlayer = players.find((p) => p.userId === myUserId);
  const amIOwner = myPlayer?.isOwner ?? false;
  const amIReady = myPlayer?.isReady ?? false;
  const allReady = players.length >= 2 && players.every((p) => p.isReady);
  const canStart = amIOwner && players.length >= GAME_CONFIG.maxPlayers && allReady;
  const isMyTurn = currentSeat !== undefined && mySeat !== null && mySeat === currentSeat;
  const showPhaseActions = phase !== "waiting" && phase !== "ended";

  return (
    <div className="min-h-screen bg-gradient-to-b from-green-900 via-green-800 to-green-950 relative overflow-hidden">
      <RoomTable
        players={players}
        mySeat={mySeat ?? 0}
        phase={phase}
        onSitDown={handleSitDown}
        onAddBot={handleAddBot}
        lastPlay={lastPlay}
        maxPlayers={GAME_CONFIG.maxPlayers}
      />
      <ScoreBoard roundPoints={roundPoints} dealerLevel={currentLevel} />

      {trumpSuit >= 0 && (
        <div className="fixed top-4 left-4 bg-black/60 backdrop-blur rounded-xl p-3 text-white z-40">
          <div className="text-lg">{["♠", "♥", "♣", "♦"][trumpSuit]} 主花色</div>
        </div>
      )}

      {errorMessage && (
        <div className="fixed top-20 left-1/2 -translate-x-1/2 z-50 px-4 py-3 bg-red-500/20 border border-red-500/40 rounded-lg text-sm text-red-300 text-center animate-pulse">
          {errorMessage}
        </div>
      )}

      {showPhaseActions && (
        <DashengjiActionBar phase={phase} isMyTurn={isMyTurn} onAction={handleAction} />
      )}

      {showPhaseActions && hand.length > 0 && (
        <div className="fixed bottom-0 w-full z-30 bg-gradient-to-t from-black/90 via-black/60 to-transparent pt-6 pb-8">
          <HandCards
            cards={hand}
            onPlayCards={handlePlayCards}
            disabled={!isMyTurn}
          />
        </div>
      )}

      {phase === "waiting" && (
        <div className="fixed bottom-0 left-0 right-0 z-40">
          <ReadyBar
            amIOwner={amIOwner}
            isReady={amIReady}
            allReady={allReady}
            playerCount={players.length}
            maxPlayers={GAME_CONFIG.maxPlayers}
            canStart={canStart}
            onReady={handleReady}
            onStartGame={handleStartGame}
            onAddBot={handleAddBot}
          />
        </div>
      )}

      {phase === "ended" && roundResult && (
        <div className="fixed inset-0 bg-black/70 flex items-center justify-center z-50">
          <div className="bg-gray-900 rounded-2xl p-8 text-white text-center min-w-[300px]">
            <h2 className="text-2xl font-bold mb-4">本轮结束</h2>
            {players.map((p) => {
              const score = roundResult.scores.find((s) => s.player_id === p.seat);
              return (
                <div key={p.userId} className="flex items-center justify-between px-4 py-2 bg-white/5 rounded-xl mb-2">
                  <span className="font-medium">{p.nickname}</span>
                  <span className={`font-bold text-lg ${score && score.score > 0 ? "text-amber-400" : "text-red-400"}`}>
                    {score ? (score.score > 0 ? "+" : "") + score.score : "0"}
                  </span>
                </div>
              );
            })}
            <div className="flex gap-3 justify-center mt-6">
              <button
                onClick={() => {
                  setPhase("waiting");
                  setRoundResult(null);
                  setLastPlay(null);
                  setCurrentSeat(undefined);
                  handleReady();
                }}
                className="px-6 py-2 rounded-full gold-button font-bold"
              >
                再来一局
              </button>
              <button
                onClick={() => router.push("/lobby")}
                className="px-6 py-2 rounded-full bg-white/10 text-white"
              >
                返回大厅
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
