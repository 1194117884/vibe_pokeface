"use client";

import { useEffect, useState, useRef } from "react";
import { useParams, useRouter } from "next/navigation";
import clsx from "clsx";
import { WSGameClient } from "@/lib/ws-game";
import { RoomTable, TablePlayer } from "@/components/game/RoomTable";
import { ReadyBar } from "@/components/game/ReadyBar";
import { HandCards } from "@/components/game/HandCards";
import { ActionBar } from "@/components/game/ActionBar";
import { ChatPanel } from "@/components/chat/ChatPanel";
import { VoiceButton } from "@/components/chat/VoiceButton";
import { LiveKitClient } from "@/lib/livekit-client";
import { RoomThemeProvider } from "@/themes";
import { NPCWalker } from "@/components/game/NPCWalker";

interface ChatMessage {
  userId: string;
  content: string;
  type: "text" | "emoji";
  timestamp: number;
}

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

interface RoundResult {
  scores: Array<{ player_id: number; score: number }>;
}

interface ServerData {
  players?: ServerPlayer[];
  seat?: number;
  user_id?: number | string;
  new_seat?: number;
  current_seat?: number;
  phase?: number;
  landlord_cards?: Array<{ id: number } | number>;
  landlord_seat?: number;
  bid_history?: Array<{ seat: number; called: boolean }>;
  last_play?: {
    seat: number;
    play?: { type: number; main_rank: number; length: number };
    cards: Array<{ id: number } | number>;
  };
  timer?: number;
  content?: string;
  type?: string;
  timestamp?: number;
  error?: string;
  theme?: string;
}

function toTablePlayer(p: ServerPlayer): TablePlayer {
  return {
    userId: String(p.user_id ?? p.userId ?? ""),
    name: String(p.user_id ?? p.userId ?? "").replace(/^ai:bot:/, "AI "),
    nickname: p.nickname ?? String(p.user_id ?? p.userId ?? "").replace(/^ai:bot:/, "AI "),
    characterId: p.character_id ?? p.characterId ?? "",
    seat: p.seat ?? 0,
    isBot: p.is_bot ?? p.isBot ?? false,
    isOwner: p.is_owner ?? p.isOwner ?? false,
    isReady: p.ready ?? p.isReady ?? false,
    cardCount: Array.isArray(p.hand) ? p.hand.length : (p.card_count ?? p.cardCount ?? 0),
  };
}

function extractHand(p: ServerPlayer): number[] {
  if (!p?.hand) return [];
  return p.hand.map((c) => (typeof c === "number" ? c : c.id));
}

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

export default function RoomPage() {
  const params = useParams();
  const router = useRouter();
  const roomId = params.id as string;

  const [players, setPlayers] = useState<TablePlayer[]>([]);
  const [mySeat, setMySeat] = useState<number | null>(null);
  const [phase, setPhase] = useState<"waiting" | "bidding" | "playing" | "ended">("waiting");
  const [roomTheme, setRoomTheme] = useState("classic-poker");
  const [connected, setConnected] = useState(false);
  const [currentSeat, setCurrentSeat] = useState<number | undefined>(undefined);
  const [hand, setHand] = useState<number[]>([]);
  const [chatMessages, setChatMessages] = useState<ChatMessage[]>([]);
  const [chatOpen, setChatOpen] = useState(false);
  const [micEnabled, setMicEnabled] = useState(false);
  const [landlordCards, setLandlordCards] = useState<number[]>([]);
  const [lastPlay, setLastPlay] = useState<{ seat: number; cards: number[] } | null>(null);
  const [roundResult, setRoundResult] = useState<RoundResult | null>(null);
  const [cardsLeftMessage, setCardsLeftMessage] = useState<string | null>(null);

  const wsClientRef = useRef<WSGameClient | null>(null);
  const voiceClientRef = useRef<LiveKitClient | null>(null);

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

    const client = new WSGameClient(Number(userIdStr), token, roomId);
    wsClientRef.current = client;
    const uid = userIdStr;
    let joined = false;

    client.on("player_joined", (msg) => {
      const data = msg.data as ServerData;
      if (data?.players) {
        setPlayers(data.players.map(toTablePlayer));
      }
      if (data?.seat !== undefined) setMySeat(data.seat);
      if (data?.theme) setRoomTheme(data.theme);
      setConnected(true);
      joined = true;
    });

    client.on("player_left", (msg) => {
      const data = msg.data as ServerData;
      if (data?.players) {
        setPlayers(data.players.map(toTablePlayer));
      }
    });

    client.on("state_update", (msg) => {
      const data = msg.data as ServerData;
      if (data?.players) {
        setPlayers(data.players.map(toTablePlayer));
        const me = data.players.find(
          (p) => String(p.user_id ?? p.userId) === uid,
        );
        if (me) setHand(extractHand(me));
      }
      if (data?.current_seat !== undefined) setCurrentSeat(data.current_seat);
      if (data?.phase !== undefined) {
        const p = data.phase;
        setPhase(p === 2 ? "ended" : p === 1 ? "playing" : "bidding");
      }
      // Extract landlord cards (revealed after bidding ends)
      if (data?.landlord_cards && Array.isArray(data.landlord_cards)) {
        setLandlordCards(data.landlord_cards.map((c: { id: number } | number) => typeof c === "number" ? c : c.id));
      }
      // Extract last play
      if (data?.last_play && data?.last_play?.cards) {
        const lp = data.last_play as { seat: number; cards: Array<{ id: number } | number> };
        setLastPlay({
          seat: lp.seat,
          cards: lp.cards.map((c: { id: number } | number) => typeof c === "number" ? c : c.id),
        });
      } else {
        setLastPlay(null);
      }
      setCardsLeftMessage(null);
      setConnected(true);
    });

    client.on("player_ready", (msg) => {
      const data = msg.data as ServerData;
      if (data?.players) {
        setPlayers(data.players.map(toTablePlayer));
      }
    });

    client.on("seat_changed", (msg) => {
      const data = msg.data as ServerData;
      if (data?.players) {
        setPlayers(data.players.map(toTablePlayer));
      }
      if (data?.new_seat !== undefined && String(data.user_id) === uid) {
        setMySeat(data.new_seat);
      }
    });

    client.on("game_start", (msg) => {
      const data = msg.data as ServerData;
      if (data?.players) {
        setPlayers(data.players.map(toTablePlayer));
        const me = data.players.find(
          (p) => String(p.user_id ?? p.userId) === uid,
        );
        if (me) setHand(extractHand(me));
      }
      setPhase("bidding");
      setRoundResult(null);
      setCardsLeftMessage(null);
      if (data?.current_seat !== undefined) setCurrentSeat(data.current_seat);
      if (data?.landlord_cards && Array.isArray(data.landlord_cards)) {
        setLandlordCards(data.landlord_cards.map((c: { id: number } | number) => typeof c === "number" ? c : c.id));
      }
    });

    client.on("round_end", (msg) => {
      setPhase("ended");
      setHand([]);
      setLandlordCards([]);
      setLastPlay(null);
      setCardsLeftMessage(null);
      const data = msg.data as { scores?: Array<{ player_id: number; score: number }> };
      if (data?.scores) {
        setRoundResult({ scores: data.scores });
      }
    });

    client.on("chat", (msg) => {
      const data = msg.data as ServerData;
      if (data?.content) {
        setChatMessages((prev) => [
          ...prev,
          {
            userId: String(data.user_id ?? "unknown"),
            content: data.content ?? "",
            type: data.type === "emoji" ? "emoji" : "text",
            timestamp: data.timestamp ?? Date.now(),
          },
        ]);
      }
    });

    client.on("error", (msg) => {
      const errMsg = msg.data as string | undefined ?? msg.error ?? "";
      console.error("GAME ERROR RAW:", JSON.stringify(msg));
      if (errMsg.indexOf("room is full") !== -1 || errMsg.indexOf("room is closed") !== -1) {
        console.error("REDIRECTING TO LOBBY");
        window.location.replace("/lobby");
        return;
      }
      console.error("SETTING CONNECTED, errMsg:", errMsg);
      setConnected(true);
    });

    client.on("theme_changed", (msg) => {
      const data = msg.data as { theme?: string };
      if (data?.theme) setRoomTheme(data.theme);
    });

    client.on("cards_left", (msg) => {
      const data = msg.data as { message?: string };
      if (data?.message) {
        setCardsLeftMessage(data.message);
      }
    });

    client.connect();

    return () => {
      voiceClientRef.current?.disconnect();
      voiceClientRef.current = null;
      wsClientRef.current = null;
      client.disconnect();
    };
  }, [roomId, router]);

  const myUserId = getUserIdFromToken();
  const myPlayer = players.find((p) => p.userId === myUserId);
  const amIOwner = players.some((p) => p.userId === myUserId && p.isOwner);
  const amIReady = myPlayer?.isReady ?? false;
  const allReady = players.length >= 2 && players.every((p) => p.isReady);
  const canStart = players.length >= 3 && allReady;
  const isMyTurn = currentSeat !== undefined && myPlayer?.seat === currentSeat;
  const displayPlayers = players.map((p) => ({
    ...p,
    isCurrentTurn: p.seat === currentSeat,
  }));

  const handleSitDown = (seat: number) => {
    wsClientRef.current?.changeSeat(seat);
  };

  const handleAddBot = () => {
    wsClientRef.current?.addBot();
  };

  const handleReady = () => {
    wsClientRef.current?.sendReady();
  };

  const handleStartGame = () => {
    wsClientRef.current?.startGame();
  };

  const handlePlayCards = (cards: number[]) => {
    const action = cards.length === 0 ? "pass" : "play";
    wsClientRef.current?.sendAction(action, cards);
  };

  const handleBidCall = () => {
    wsClientRef.current?.sendAction("bid_call");
  };

  const handleBidPass = () => {
    wsClientRef.current?.sendAction("bid_pass");
  };

  const handleSendChat = (content: string, type: "text" | "emoji") => {
    wsClientRef.current?.sendChat(content, type);
  };

  const handleVoiceToggle = async (enabled: boolean) => {
    if (enabled) {
      const token = typeof window !== "undefined" ? localStorage.getItem("token") : null;
      if (!token) return;
      voiceClientRef.current?.disconnect();
      try {
        const res = await fetch(`/api/livekit/token?room=${roomId}`, {
          headers: { Authorization: `Bearer ${token}` },
        });
        const json = await res.json();
        if (json.success && json.token) {
          const vc = new LiveKitClient();
          await vc.connect(json.url, json.token, json.room);
          await vc.toggleMic();
          voiceClientRef.current = vc;
          setMicEnabled(true);
        }
      } catch (e) {
        console.error("Failed to connect voice:", e);
      }
    } else {
      try {
        voiceClientRef.current?.toggleMic();
      } catch (e) {
        console.error("Failed to toggle mic:", e);
      }
      setMicEnabled(false);
    }
  };

  if (!connected) {
    return (
      <div className="flex items-center justify-center min-h-screen bg-cream">
        <div className="text-center">
          <div className="text-4xl mb-4 animate-pulse">🎴</div>
          <p className="text-text-black-soft">连接房间中...</p>
        </div>
      </div>
    );
  }

  return (
    <RoomThemeProvider themeId={roomTheme}>
      <NPCWalker />
      <div
        className="min-h-screen flex flex-col"
        style={{
          backgroundImage: "var(--bg-image)",
          backgroundColor: "var(--bg-color, #f2f0eb)",
          backgroundSize: "cover",
          backgroundPosition: "center",
          backgroundAttachment: "fixed",
        }}
      >
        {/* Dark overlay for readability */}
        <div
          className="fixed inset-0 pointer-events-none"
          style={{ background: "var(--bg-overlay, none)", zIndex: 1 }}
        />

        {/* Room Header */}
        <header className="bg-white/90 backdrop-blur-sm border-b border-ceramic px-6 py-3 flex items-center justify-between shrink-0 relative z-20">
          <div className="flex items-center gap-3">
            <button
              onClick={() => router.push("/lobby")}
              className="text-sm text-text-black-soft hover:text-green-accent"
            >
              ← 退出
            </button>
            <h1 className="font-bold text-text-black-strong">
              房间 {roomId.slice(0, 6)}
            </h1>
            <span className="text-xs bg-green-100 text-green-700 px-2 py-0.5 rounded-full">
              斗地主
            </span>
          </div>
          <div className="flex items-center gap-2">
            <VoiceButton onToggle={handleVoiceToggle} disabled={!connected} />
            <span className="text-xs text-text-black-soft">
              {micEnabled ? "Mic on" : "Mic off"}
            </span>
          </div>
        </header>

        {/* Main Content */}
        <div className="flex-1 flex min-h-0 relative z-10">
          {/* Game Area */}
          <div className="flex-1 flex flex-col items-center justify-center p-4 overflow-auto">
            {players.length === 0 ? (
              <div className="text-center text-text-black-soft">房间是空的</div>
            ) : (
              <>
                <RoomTable
                  players={displayPlayers}
                  mySeat={mySeat ?? 0}
                  phase={phase}
                  onSitDown={handleSitDown}
                  onAddBot={handleAddBot}
                  landlordCards={landlordCards}
                  lastPlay={lastPlay}
                  cardsLeftMessage={cardsLeftMessage}
                />

                {/* Waiting phase: Ready/Start controls */}
                {phase === "waiting" && (
                  <ReadyBar
                    amIOwner={amIOwner}
                    isReady={amIReady}
                    allReady={allReady}
                    playerCount={players.length}
                    maxPlayers={3}
                    canStart={canStart}
                    onReady={handleReady}
                    onStartGame={handleStartGame}
                    onAddBot={handleAddBot}
                  />
                )}

                {/* Playing/Bidding phase: hand cards and action buttons */}
                {(phase === "bidding" || phase === "playing") && (
                  <div className="w-full max-w-3xl mt-4 space-y-2">
                    <HandCards
                      cards={hand}
                      onPlayCards={
                        phase === "playing" ? handlePlayCards : undefined
                      }
                      disabled={!isMyTurn}
                    />
                    {phase === "bidding" && (
                      <ActionBar
                        phase={phase}
                        isMyTurn={isMyTurn}
                        onBidCall={handleBidCall}
                        onBidPass={handleBidPass}
                      />
                    )}
                  </div>
                )}
              </>
            )}
          </div>

          {/* Desktop Chat Sidebar */}
          <div className="hidden lg:flex w-80 p-4 border-l border-white/20 flex-col shrink-0 relative z-10">
            <div className="flex-1 min-h-0">
              <ChatPanel
                messages={chatMessages}
                onSendMessage={handleSendChat}
                disabled={!connected}
              />
            </div>
          </div>
        </div>

        {/* Mobile Chat FAB */}
        <button
          onClick={() => setChatOpen(true)}
          className="fixed bottom-6 right-6 lg:hidden z-30 w-14 h-14 rounded-full bg-green-accent text-white text-2xl shadow-frap flex items-center justify-center active:scale-[0.95] transition-transform"
          aria-label="打开聊天"
        >
          💬
        </button>

        {/* Mobile Chat Sheet */}
        {chatOpen && (
          <div className="fixed inset-0 z-50 lg:hidden flex flex-col">
            <div
              className="absolute inset-0 bg-black/40"
              onClick={() => setChatOpen(false)}
            />
            <div
              className={clsx(
                "absolute bottom-0 left-0 right-0 bg-white rounded-t-xl shadow-frap",
                "flex flex-col max-h-[70vh] transition-transform duration-300",
                "pb-[var(--safe-area-bottom,0px)]",
              )}
            >
              <div className="flex items-center justify-between px-4 py-3 border-b border-ceramic">
                <div className="flex items-center gap-2">
                  <VoiceButton onToggle={handleVoiceToggle} disabled={!connected} />
                  <span className="text-sm text-text-black-soft">
                    {micEnabled ? "Mic on" : "Mic off"}
                  </span>
                </div>
                <button
                  onClick={() => setChatOpen(false)}
                  className="w-8 h-8 rounded-full bg-cream flex items-center justify-center text-text-black-soft hover:bg-ceramic transition-colors"
                  aria-label="关闭聊天"
                >
                  ✕
                </button>
              </div>
              <div className="flex-1 min-h-0 p-4">
                <ChatPanel
                  messages={chatMessages}
                  onSendMessage={handleSendChat}
                  disabled={!connected}
                />
              </div>
            </div>
          </div>
        )}

        {/* Round end overlay */}
        {phase === "ended" && roundResult && (
          <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50">
            <div className="bg-white rounded-2xl shadow-frap p-8 max-w-sm w-full mx-4 text-center">
              <h2 className="text-2xl font-bold text-text-black-strong mb-2">
                🎉 本局结束
              </h2>
              <div className="space-y-3 my-6">
                {players.map((p) => {
                  const score = roundResult.scores.find(s => String(s.player_id) === p.userId);
                  const isPositive = score && score.score > 0;
                  return (
                    <div key={p.userId} className="flex items-center justify-between px-4 py-2 bg-cream rounded-xl">
                      <span className="font-medium text-text-black">{p.nickname || p.name}</span>
                      <span className={`font-bold text-lg ${isPositive ? 'text-green-accent' : 'text-red-400'}`}>
                        {score ? (score.score > 0 ? "+" : "") + score.score : "0"}
                      </span>
                    </div>
                  );
                })}
              </div>
              <div className="flex gap-3 justify-center">
                <button
                  onClick={() => {
                    setPhase("waiting");
                    setRoundResult(null);
                    setLandlordCards([]);
                    setLastPlay(null);
                    setCardsLeftMessage(null);
                    setCurrentSeat(undefined);
                    handleReady();
                  }}
                  className="px-6 py-2 bg-green-accent text-white rounded-pill font-semibold hover:bg-green-accent/90 transition-colors"
                >
                  再来一局
                </button>
                <button
                  onClick={() => router.push("/lobby")}
                  className="px-6 py-2 bg-white text-text-black border border-ceramic rounded-pill font-semibold hover:border-green-accent/50 transition-colors"
                >
                  返回大厅
                </button>
              </div>
            </div>
          </div>
        )}
      </div>
    </RoomThemeProvider>
  );
}
