"use client";

import { useCallback, useEffect, useState, useRef } from "react";
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
import { AICharacterPicker } from "@/components/game/AICharacterPicker";
import { useGameAudio } from "@/hooks/useGameAudio";
import { RoomThemeProvider } from "@/themes";

interface ChatMessage {
  userId: string;
  nickname: string;
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
  is_landlord?: boolean;
  isLandlord?: boolean;
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
  } | null;
  timer?: number;
  nickname?: string;
  content?: string;
  type?: string;
  timestamp?: number;
  error?: string;
  theme?: string;
  game_type?: string;
  max_players?: number;
  // Full GameState fields available at runtime
  reveal_count?: number;
  double_count?: number;
  consecutive_passes?: number;
  revealed?: Record<string, boolean>;
  doubled?: Record<string, boolean>;
  has_passed?: Record<string, boolean>;
  snatch_count?: number;
  multiplier?: number;
  round_num?: number;
  winner_seat?: number;
}

type ActionType =
  | "called_landlord" | "passed_calling"
  | "snatched" | "passed_snatching"
  | "revealed" | "passed_reveal"
  | "doubled" | "passed_double"
  | "played_cards" | "passed_play";

interface DetectedAction {
  seat: number;
  type: ActionType;
  bubbleText: string;
  toastText: string;
  speechText: string;
}

function getPhrases(
  type: ActionType,
  ctx?: { hasPrevPlay?: boolean; isBomb?: boolean; isRocket?: boolean },
): { bubbleText: string; toastText: string; speechText: string } {
  switch (type) {
    case "called_landlord": return { bubbleText: "叫地主!", toastText: "叫地主!", speechText: "叫地主" };
    case "passed_calling": return { bubbleText: "不叫", toastText: "不叫", speechText: "不叫" };
    case "snatched": return { bubbleText: "抢!", toastText: "抢地主!", speechText: "抢地主" };
    case "passed_snatching": return { bubbleText: "不抢", toastText: "不抢", speechText: "不抢" };
    case "revealed": return { bubbleText: "明牌!", toastText: "明牌!", speechText: "明牌" };
    case "passed_reveal": return { bubbleText: "不明牌", toastText: "不明牌", speechText: "不明牌" };
    case "doubled": return { bubbleText: "加倍!", toastText: "加倍!", speechText: "加倍" };
    case "passed_double": return { bubbleText: "不加倍", toastText: "不加倍", speechText: "不加倍" };
    case "played_cards": {
      if (ctx?.isRocket) return { bubbleText: "火箭!", toastText: "王炸!", speechText: "压死" };
      if (ctx?.isBomb) return { bubbleText: "炸弹!", toastText: "炸弹!", speechText: "压死" };
      if (ctx?.hasPrevPlay) return { bubbleText: "大你～", toastText: "大你～", speechText: "大你" };
      return { bubbleText: "出牌!", toastText: "打出了一手牌", speechText: "出牌" };
    }
    case "passed_play": return { bubbleText: "不出", toastText: "过牌", speechText: "要不起" };
  }
}

function isBomb(cards: number[]): boolean {
  if (cards.length !== 4) return false;
  const ranks = cards.map((c) => (c >= 52 ? c : c % 13));
  return new Set(ranks).size === 1;
}
function isRocket(cards: number[]): boolean {
  return cards.length === 2 && cards.includes(52) && cards.includes(53);
}

function detectAction(prev: ServerData | null, curr: ServerData): DetectedAction | null {
  if (!prev) return null;

  const prevBids = prev.bid_history ?? [];
  const currBids = curr.bid_history ?? [];
  if (currBids.length > prevBids.length) {
    const entry = currBids[currBids.length - 1];
    const isSnatching = curr.phase === 1;
    const type: ActionType = entry.called
      ? (isSnatching ? "snatched" : "called_landlord")
      : (isSnatching ? "passed_snatching" : "passed_calling");
    const phrases = getPhrases(type);
    return { seat: entry.seat, type, ...phrases };
  }

  const crc = curr.reveal_count ?? 0;
  const prc = prev.reveal_count ?? 0;
  if (crc > prc) {
    const currRv: Record<string, boolean> = (curr.revealed ?? {}) as Record<string, boolean>;
    const prevRv: Record<string, boolean> = (prev.revealed ?? {}) as Record<string, boolean>;
    for (const key of Object.keys(currRv)) {
      if (currRv[key] && !prevRv[key]) {
        const phrases = getPhrases("revealed");
        return { seat: Number(key), type: "revealed", ...phrases };
      }
    }
    if (prev.current_seat !== undefined) {
      const phrases = getPhrases("passed_reveal");
      return { seat: prev.current_seat, type: "passed_reveal", ...phrases };
    }
  }

  const cdc = curr.double_count ?? 0;
  const pdc = prev.double_count ?? 0;
  if (cdc > pdc) {
    const currDb: Record<string, boolean> = (curr.doubled ?? {}) as Record<string, boolean>;
    const prevDb: Record<string, boolean> = (prev.doubled ?? {}) as Record<string, boolean>;
    for (const key of Object.keys(currDb)) {
      if (currDb[key] && !prevDb[key]) {
        const phrases = getPhrases("doubled");
        return { seat: Number(key), type: "doubled", ...phrases };
      }
    }
    if (prev.current_seat !== undefined) {
      const phrases = getPhrases("passed_double");
      return { seat: prev.current_seat, type: "passed_double", ...phrases };
    }
  }

  const pcp = prev.consecutive_passes ?? 0;
  const ccp = curr.consecutive_passes ?? 0;
  if (ccp > pcp) {
    if (prev.current_seat !== undefined) {
      const phrases = getPhrases("passed_play");
      return { seat: prev.current_seat, type: "passed_play", ...phrases };
    }
  }

  const prevLP = prev.last_play;
  const currLP = curr.last_play;
  if (currLP && (!prevLP || currLP.seat !== prevLP.seat)) {
    const cards: number[] = Array.isArray(currLP.cards)
      ? currLP.cards.map((c: number | { id: number }) => (typeof c === "number" ? c : c.id))
      : [];
    const hasPrevPlay = !!prevLP;
    const rocket = isRocket(cards);
    const bomb = !rocket && isBomb(cards);
    const phrases = getPhrases("played_cards", { hasPrevPlay, isBomb: bomb, isRocket: rocket });
    return { seat: currLP.seat, type: "played_cards", ...phrases };
  }

  return null;
}

const GAME_CONFIG: Record<string, { maxPlayers: number; tableSize: "sm" | "lg" }> = {
  doudizhu: { maxPlayers: 3, tableSize: "lg" },
};

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
    isLandlord: p.is_landlord ?? false,
    cardCount: Array.isArray(p.hand) ? p.hand.length : (p.card_count ?? p.cardCount ?? 0),
  };
}

function mergePlayers(existing: TablePlayer[], incoming: TablePlayer[]): TablePlayer[] {
  const existingBySeat = new Map<number, TablePlayer>();
  for (const p of existing) {
    existingBySeat.set(p.seat, p);
  }
  return incoming.map((p) => {
    const old = existingBySeat.get(p.seat);
    if (old) {
      return {
        ...p,
        isBot: old.isBot,
        nickname: old.nickname || p.nickname,
        characterId: old.characterId,
        isOwner: old.isOwner,
      };
    }
    return p;
  });
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
  const mySeatRef = useRef<number | null>(null);
  const playersRef = useRef<TablePlayer[]>([]);
  const setMySeatWithRef = (seat: number | null) => {
    mySeatRef.current = seat;
    setMySeat(seat);
  };
  const [phase, setPhase] = useState<"waiting" | "calling" | "snatching" | "revealing" | "doubling" | "playing" | "ended">("waiting");
  const [roomTheme, setRoomTheme] = useState("classic-poker");
  const [connected, setConnected] = useState(false);
  const [currentSeat, setCurrentSeat] = useState<number | undefined>(undefined);
  const [hand, setHand] = useState<number[]>([]);
  const [chatMessages, setChatMessages] = useState<ChatMessage[]>([]);
  const [chatOpen, setChatOpen] = useState(false);
  const [micEnabled, setMicEnabled] = useState(false);
  const [landlordCards, setLandlordCards] = useState<number[]>([]);
  const [landlordSeat, setLandlordSeat] = useState<number | undefined>(undefined);
  const [lastPlay, setLastPlay] = useState<{ seat: number; cards: number[] } | null>(null);
  const [roundResult, setRoundResult] = useState<RoundResult | null>(null);
  const [cardsLeftMessage, setCardsLeftMessage] = useState<string | null>(null);
  const [gameType, setGameType] = useState("doudizhu");
  const [showAIPicker, setShowAIPicker] = useState(false);
  const [multiplier, setMultiplier] = useState(1);
  const [roomScores, setRoomScores] = useState<Record<string, number>>({});
  const [showSettings, setShowSettings] = useState(false);
  const gameConfig = GAME_CONFIG[gameType] || GAME_CONFIG.doudizhu;
  const [speechBubbles, setSpeechBubbles] = useState<Record<number, string>>({});
  const [audioMuted, setAudioMuted] = useState<boolean>(() => {
    if (typeof window === "undefined") return false;
    return localStorage.getItem("audio_muted") === "true";
  });
  const [compactUI, setCompactUI] = useState(false);
  const prevDataRef = useRef<ServerData | null>(null);
  const bubbleTimersRef = useRef<Record<number, ReturnType<typeof setTimeout>>>({});
  const { speak } = useGameAudio(audioMuted);

  const wsClientRef = useRef<WSGameClient | null>(null);
  const voiceClientRef = useRef<LiveKitClient | null>(null);

  const showSpeechBubble = useCallback((seat: number, text: string) => {
    if (bubbleTimersRef.current[seat]) {
      clearTimeout(bubbleTimersRef.current[seat]);
    }
    setSpeechBubbles((prev) => ({ ...prev, [seat]: text }));
    bubbleTimersRef.current[seat] = setTimeout(() => {
      setSpeechBubbles((prev) => {
        const next = { ...prev };
        delete next[seat];
        return next;
      });
      delete bubbleTimersRef.current[seat];
    }, 3000);
  }, []);

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
      if (data?.seat !== undefined && String(data.user_id) === uid) setMySeatWithRef(data.seat);
      if (data?.theme) setRoomTheme(data.theme);
      if (data?.game_type) setGameType(data.game_type);
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

      // Detect and announce player actions via diff
      const prev = prevDataRef.current;
      const action = detectAction(prev, data);
      if (action) {
        showSpeechBubble(action.seat, action.bubbleText);
        speak(action.speechText);
      }
      prevDataRef.current = data;

      if (data?.players) {
        setPlayers((prev) => mergePlayers(prev, data.players!.map(toTablePlayer)));
        if (mySeatRef.current !== null) {
          const me = data.players.find(
            (p) => (p.seat ?? 0) === mySeatRef.current,
          );
          if (me) setHand(extractHand(me));
        }
      }
      if (data?.current_seat !== undefined) setCurrentSeat(data.current_seat);
      if (data?.phase !== undefined) {
        const p = data.phase;
        setPhase(p === 5 ? "ended" : p === 4 ? "playing" : p === 3 ? "doubling" : p === 2 ? "revealing" : p === 1 ? "snatching" : "calling");
      }
      // Extract landlord cards (revealed after bidding ends)
      if (data?.landlord_cards && Array.isArray(data.landlord_cards)) {
        setLandlordCards(data.landlord_cards.map((c: { id: number } | number) => typeof c === "number" ? c : c.id));
      }
      if (data?.landlord_seat !== undefined) setLandlordSeat(data.landlord_seat);
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
      if (data?.multiplier !== undefined) setMultiplier(data.multiplier);
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
        setMySeatWithRef(data.new_seat);
      }
    });

    client.on("game_start", (msg) => {
      const data = msg.data as ServerData;
      if (data?.players) {
        setPlayers((prev) => mergePlayers(prev, data.players!.map(toTablePlayer)));
        // Find my hand by matching seat (engine user_id is seat index, not real user ID)
        if (mySeatRef.current !== null) {
          const me = data.players.find(
            (p) => (p.seat ?? 0) === mySeatRef.current,
          );
          if (me) setHand(extractHand(me));
        }
      }
      setPhase("calling");
      setRoundResult(null);
      setCardsLeftMessage(null);
      if (data?.current_seat !== undefined) setCurrentSeat(data.current_seat);
      if (data?.landlord_cards && Array.isArray(data.landlord_cards)) {
        setLandlordCards(data.landlord_cards.map((c: { id: number } | number) => typeof c === "number" ? c : c.id));
      }
      if (data?.landlord_seat !== undefined) setLandlordSeat(data.landlord_seat);
    });

    client.on("round_end", (msg) => {
      setPhase("ended");
      setHand([]);
      setLandlordCards([]);
      setLandlordSeat(undefined);
      setLastPlay(null);
      setCardsLeftMessage(null);
      const data = msg.data as { scores?: Array<{ player_id: number; score: number }> };
      if (data?.scores) {
        setRoundResult({ scores: data.scores });
        // Map seat-based player_id to actual userId
        const seatToUser: Record<string, string> = {};
        for (const p of playersRef.current) {
          seatToUser[String(p.seat)] = p.userId;
        }
        setRoomScores((prev) => {
          const next = { ...prev };
          for (const s of data.scores!) {
            const key = seatToUser[String(s.player_id)] || String(s.player_id);
            next[key] = (next[key] || 0) + s.score;
          }
          return next;
        });
      }
    });

    client.on("chat", (msg) => {
      const data = msg.data as ServerData;
      if (data?.content) {
        setChatMessages((prev) => [
          ...prev,
          {
            userId: String(data.user_id ?? "unknown"),
            nickname: String(data.nickname ?? data.user_id ?? "unknown"),
            content: data.content ?? "",
            type: data.type === "emoji" ? "emoji" : "text",
            timestamp: data.timestamp ?? Date.now(),
          },
        ]);

        // Show chat as speech bubble near sender's seat
        const senderId = String(data.user_id ?? "");
        const sender = playersRef.current.find(p => p.userId === senderId);
        if (sender) {
          showSpeechBubble(sender.seat, data.content);
        }
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

  useEffect(() => {
    playersRef.current = players;
  }, [players]);

  useEffect(() => {
    const check = () => {
      setCompactUI(window.innerWidth > window.innerHeight && window.innerHeight < 500);
    };
    check();
    window.addEventListener("resize", check);
    return () => window.removeEventListener("resize", check);
  }, []);

  const myUserId = getUserIdFromToken();
  const myPlayer = players.find((p) => p.seat === mySeat);
  const amIOwner = players.some((p) => p.userId === myUserId && p.isOwner);
  const amIReady = myPlayer?.isReady ?? false;
  const allReady = players.length >= 2 && players.every((p) => p.isReady);
  const canStart = players.length >= gameConfig.maxPlayers && allReady;
  const isMyTurn = currentSeat !== undefined && mySeat !== null && mySeat === currentSeat;
  const myCumulativeScore = roomScores[String(mySeat ?? -1)] ?? 0;
  const displayPlayers = players.map((p) => ({
    ...p,
    isCurrentTurn: p.seat === currentSeat,
  }));

  const handleSitDown = (seat: number) => {
    wsClientRef.current?.changeSeat(seat);
  };

  const handleAddBot = () => {
    setShowAIPicker(true);
  };

  const handleSelectAICharacter = (characterId: number) => {
    setShowAIPicker(false);
    wsClientRef.current?.addBot(characterId);
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

  const handleReveal = () => {
    wsClientRef.current?.sendAction("reveal_all");
  };

  const handleRevealPass = () => {
    wsClientRef.current?.sendAction("pass");
  };

  const handleDouble = () => {
    wsClientRef.current?.sendAction("double");
  };

  const handleNoDouble = () => {
    wsClientRef.current?.sendAction("no_double");
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
    <div className="min-h-screen bg-background text-on-background flex flex-col overflow-hidden">
      {/* Top Navigation — Stitch compact header */}
      <header className={clsx(
        "fixed top-0 left-0 w-full z-50 flex items-center justify-between px-4 h-10 bg-gradient-to-b from-black/40 to-transparent",
        compactUI && "landscape-nav"
      )}>
        <div className="flex items-center gap-3">
          <span className="text-secondary-fixed text-sm font-black tracking-tight uppercase">
            房间 {roomId.slice(0, 4)}
          </span>
          <div className="flex items-center bg-black/30 rounded-full px-3 py-0.5 border border-outline-variant/30">
            <span className="text-secondary-fixed text-xs font-bold mr-1">$</span>
            <span className={clsx(
              "text-xs font-bold",
              myCumulativeScore >= 0 ? "text-secondary-fixed" : "text-error"
            )}>
              {myCumulativeScore >= 0 ? "+" : ""}{myCumulativeScore}
            </span>
          </div>
          <div className="bg-black/30 px-2 py-0.5 rounded border border-outline-variant/30 flex items-center gap-2">
            <span className="text-[10px] font-bold text-on-surface-variant uppercase">倍数</span>
            <span className="text-secondary-fixed text-xs font-bold">x{multiplier}</span>
          </div>
        </div>
        <div className="relative">
          <button
            onClick={() => setShowSettings(!showSettings)}
            className="text-on-surface-variant hover:text-secondary-fixed transition-colors text-xl leading-none"
            aria-label="Settings"
          >
            ⚙
          </button>
          {showSettings && (
            <>
              <div className="fixed inset-0 z-40" onClick={() => setShowSettings(false)} />
              <div className="absolute right-0 top-8 z-50 bg-surface-container-high rounded-xl border border-outline-variant shadow-frap p-1.5 min-w-[140px] flex flex-col">
                <button
                  onClick={() => { handleVoiceToggle(!micEnabled); }}
                  className="flex items-center gap-2 px-3 py-2 rounded-lg hover:bg-surface-container text-sm transition-colors w-full text-left"
                  style={{ color: micEnabled ? "#8ed5af" : undefined }}
                >
                  <span className="text-base">{micEnabled ? "🎤" : "🤐"}</span>
                  <span className="text-on-surface">麦克风</span>
                </button>
                <button
                  onClick={() => {
                    const next = !audioMuted;
                    setAudioMuted(next);
                    localStorage.setItem("audio_muted", String(next));
                  }}
                  className="flex items-center gap-2 px-3 py-2 rounded-lg hover:bg-surface-container text-sm transition-colors w-full text-left"
                  style={{ color: audioMuted ? undefined : "#8ed5af" }}
                >
                  <span className="text-base">{audioMuted ? "🔇" : "🔊"}</span>
                  <span className="text-on-surface">音效</span>
                </button>
                <hr className="border-outline-variant my-0.5" />
                <button
                  onClick={() => { setShowSettings(false); setChatOpen(true); }}
                  className="flex items-center gap-2 px-3 py-2 rounded-lg hover:bg-surface-container text-on-surface text-sm transition-colors w-full text-left"
                >
                  <span className="text-base">💬</span>
                  聊天
                </button>
                <hr className="border-outline-variant my-0.5" />
                <button
                  onClick={() => router.push("/lobby")}
                  className="flex items-center gap-2 px-3 py-2 rounded-lg hover:bg-error-container text-error text-sm transition-colors w-full text-left"
                >
                  <span className="text-base">🚪</span>
                  退出
                </button>
              </div>
            </>
          )}
        </div>
      </header>

      {/* Main Game Canvas — Imperial Emerald table */}
      <RoomThemeProvider themeId={roomTheme || "imperial-emerald"}>
        <main className={clsx("flex-grow flex flex-col items-center justify-center pt-10 relative", compactUI && "landscape-main")}
          style={{
            background: "radial-gradient(circle, #1a7452 0%, #063a27 100%)",
          }}
        >
          {/* Lattice pattern overlay */}
          <div className="absolute inset-0 pointer-events-none lattice-overlay" />
          {players.length === 0 ? (
            <div className="text-center text-on-surface-variant text-lg">房间是空的</div>
          ) : (
            <>
              <div className={clsx("w-full flex-1 flex flex-col", compactUI && "landscape-table")}>
                <RoomTable
                  players={displayPlayers}
                  mySeat={mySeat ?? 0}
                  phase={phase}
                  onSitDown={handleSitDown}
                  onAddBot={handleAddBot}
                  landlordCards={landlordCards}
                  landlordSeat={landlordSeat}
                  lastPlay={lastPlay}
                  cardsLeftMessage={cardsLeftMessage}
                  maxPlayers={gameConfig.maxPlayers}
                  tableSize={gameConfig.tableSize}
                  speechBubbles={speechBubbles}
                  compact={compactUI}
                />
              </div>

              {/* Waiting phase: Ready/Start controls */}
              {phase === "waiting" && (
                <ReadyBar
                  amIOwner={amIOwner}
                  isReady={amIReady}
                  allReady={allReady}
                  playerCount={players.length}
                  maxPlayers={gameConfig.maxPlayers}
                  canStart={canStart}
                  onReady={handleReady}
                  onStartGame={handleStartGame}
                  onAddBot={handleAddBot}
                />
              )}

            </>
          )}
        </main>
      </RoomThemeProvider>

      {/* Bottom area: bidding controls + hand cards */}
      {(phase !== "waiting" && phase !== "ended") && (
        <div className={clsx("fixed bottom-0 w-full flex flex-col items-center z-30 bg-gradient-to-t from-black/90 via-black/60 to-transparent pt-6 pb-8", compactUI && "landscape-bottom-bar")}>
          {/* Bidding action buttons — above the hand cards */}
          {(phase === "calling" || phase === "snatching" || phase === "revealing" || phase === "doubling") && (
            <div className="w-full max-w-3xl space-y-2 mb-2">
              <ActionBar
                phase={phase}
                isMyTurn={isMyTurn}
                onBidCall={handleBidCall}
                onBidPass={handleBidPass}
                onReveal={handleReveal}
                onRevealPass={handleRevealPass}
                onDouble={handleDouble}
                onNoDouble={handleNoDouble}
              />
              {(phase === "calling" || phase === "snatching") && !isMyTurn && (
                <p className="text-center text-sm text-white/60 animate-pulse">
                  {phase === "calling" ? "等待其他玩家叫地主..." : "等待其他玩家抢地主..."}
                </p>
              )}
              {phase === "revealing" && !isMyTurn && (
                <p className="text-center text-sm text-white/60 animate-pulse">
                  等待其他玩家明牌...
                </p>
              )}
              {phase === "doubling" && !isMyTurn && (
                <p className="text-center text-sm text-white/60 animate-pulse">
                  等待其他玩家加倍...
                </p>
              )}
            </div>
          )}

          {/* Player hand cards */}
          <HandCards
            cards={hand}
            onPlayCards={phase === "playing" ? handlePlayCards : undefined}
            disabled={!isMyTurn}
            compact={compactUI}
          />
        </div>
      )}

      {/* Chat Sheet */}
      {chatOpen && (
        <div className="fixed inset-0 z-50 flex flex-col">
          <div className="absolute inset-0 bg-black/70 backdrop-blur-sm" onClick={() => setChatOpen(false)} />
          <div className="absolute bottom-0 left-0 right-0 bg-surface-container-high rounded-t-2xl shadow-frap flex flex-col max-h-[70vh] pb-[var(--safe-area-bottom,0px)] border-t border-outline-variant">
            <div className="flex items-center justify-between px-4 py-3 border-b border-outline-variant/50">
              <div className="flex items-center gap-2">
                <span className="text-sm font-semibold text-on-surface">聊天</span>
              </div>
              <div className="flex items-center gap-1">
                <VoiceButton onToggle={handleVoiceToggle} disabled={!connected} />
                <button
                  onClick={() => setChatOpen(false)}
                  className="w-8 h-8 rounded-full bg-surface-container flex items-center justify-center text-on-surface-variant hover:bg-surface-container-highest transition-colors"
                >
                  ✕
                </button>
              </div>
            </div>
            <div className="flex-1 min-h-0">
              <ChatPanel messages={chatMessages} onSendMessage={handleSendChat} disabled={!connected} />
            </div>
          </div>
        </div>
      )}

      {/* AI character picker */}
      <AICharacterPicker
        open={showAIPicker}
        onClose={() => setShowAIPicker(false)}
        onSelect={handleSelectAICharacter}
      />

      {/* Round end overlay */}
      {phase === "ended" && roundResult && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/60 backdrop-blur-sm">
          <div className="bg-surface-container-high rounded-2xl shadow-frap p-8 max-w-sm w-full mx-4 text-center border border-outline-variant">
            <h2 className="text-2xl font-bold text-on-surface mb-2">🎉 本局结束</h2>
            <div className="space-y-3 my-6">
              {players.map((p) => {
                const score = roundResult.scores.find((s) => s.player_id === p.seat);
                const isPositive = score && score.score > 0;
                return (
                  <div key={p.userId} className="flex items-center justify-between px-4 py-2 bg-surface-container rounded-xl">
                    <span className="font-medium text-on-surface">{p.nickname || p.name}</span>
                    <span className={`font-bold text-lg ${isPositive ? "text-primary" : "text-error"}`}>
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
                  setLandlordSeat(undefined);
                  setLastPlay(null);
                  setCardsLeftMessage(null);
                  setCurrentSeat(undefined);
                  handleReady();
                }}
                className="px-6 py-2 rounded-full gold-button font-bold hover:brightness-110 active:scale-95 transition-all"
              >
                再来一局
              </button>
              <button
                onClick={() => router.push("/lobby")}
                className="px-6 py-2 rounded-full emerald-button font-bold hover:brightness-110 active:scale-95 transition-all"
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
