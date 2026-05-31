"use client";

import { useCallback, useEffect, useState, useRef, useMemo } from "react";
import { useParams, useRouter, useSearchParams } from "next/navigation";
import { WSGameClient, formatError, type ErrorData } from "@/lib/ws-game";
import { formatDashengjiNoticeToast, mergeDashengjiNoticePlayers } from "@/lib/dashengji-notice";
import { RoomTable, TablePlayer } from "@/components/game/RoomTable";
import { ReadyBar } from "@/components/game/ReadyBar";
import { HandCards } from "@/components/game/HandCards";
import { DashengjiActionBar } from "@/components/game/dashengji/DashengjiActionBar";
import { GamePhasePrompt } from "@/components/game/GamePhasePrompt";
import { PasswordPrompt } from "@/components/game/PasswordPrompt";

/** Sort hand for Dashengji display: when trump is set, main cards (trump) first, then side cards by suit. */
function sortDashengjiHand(cardIds: number[], trumpSuit: number, currentLevel: number): number[] {
  const levelRank = currentLevel >= 3 && currentLevel <= 14 ? currentLevel : 3;

  const isMain = (face: number): boolean => {
    if (face >= 52) return true; // jokers
    const suit = Math.floor(face / 13);
    const rank = face % 13;
    const baseRank = rank + 3; // 0→3, 1→4, ..., 11→A(14), 12→2(15)
    if (baseRank === 15 || baseRank === levelRank) return true; // 2 or level card
    if (suit === trumpSuit) return true; // trump suit
    return false;
  };

  return [...cardIds].sort((a, b) => {
    const fa = a % 54, fb = b % 54;
    const ma = isMain(fa), mb = isMain(fb);
    // Main cards before side cards
    if (ma && !mb) return -1;
    if (!ma && mb) return 1;
    // Both main: jokers first, then by suit, then rank desc
    if (ma && mb) {
      if (fa === 53 && fb !== 53) return -1;
      if (fb === 53 && fa !== 53) return 1;
      if (fa === 52 && fb !== 52 && fb !== 53) return -1;
      if (fb === 52 && fa !== 52 && fa !== 53) return 1;
      if (fa >= 52 && fb >= 52) return fb - fa;
      const sa = Math.floor(fa / 13);
      const sb = Math.floor(fb / 13);
      if (sa !== sb) return sa - sb;
      return (fb % 13) - (fa % 13);
    }
    // Both side: group by suit, then rank desc
    const sa = Math.floor(fa / 13);
    const sb = Math.floor(fb / 13);
    if (sa !== sb) return sa - sb;
    return (fb % 13) - (fa % 13);
  });
}
import { DashengjiInfoPanel } from "@/components/game/dashengji/DashengjiInfoPanel";

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
  is_dead_trump?: boolean;
  current_level?: number;
  team_levels?: [number, number];
  dealer_seats?: [number, number];
  bottom_cards?: Array<{ id: number } | number>;
  discarded_cards?: Array<{ id: number } | number>;
  take_bottom_seat?: number;
  round_points?: number;
  last_play?: {
    seat: number;
    cards: Array<{ id: number } | number>;
  } | null;
  play_history?: Array<{
    seat: number;
    cards: Array<{ id: number } | number>;
  }>;
  scores?: Array<{ player_id: number; score: number }>;
  notices?: Array<{
    seq: number;
    kind: string;
    seat?: number;
    action?: string;
  }>;
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

function getDashengjiPrompt(
  phase: string,
  isMyTurn: boolean,
  isDealerTeam: boolean,
  mySeat: number | null,
  takeBottomSeat: number,
  currentPlayerName?: string,
  errorMessage?: string | null,
): { title: string; detail?: string; tone: "action" | "waiting" | "error" } {
  if (errorMessage) return { title: errorMessage, detail: "请重新选择或等待下一步", tone: "error" };
  if (phase === "waiting") return { title: "等待开始", detail: "先点准备，四人到齐后房主开始游戏", tone: "waiting" };
  if (phase === "ended") return { title: "本轮结束", detail: "查看得分后可以再来一局", tone: "waiting" };
  if (phase === "set_trump") {
    if (!isDealerTeam) return { title: "请稍等", detail: "庄家队正在选择是否定主", tone: "waiting" };
    return { title: "轮到庄家队定主", detail: "选中可定主的牌后点“定主”，也可以点“不定”", tone: "action" };
  }
  if (phase === "counter_trump") {
    if (isDealerTeam) return { title: "请稍等", detail: "闲家队正在选择是否反主", tone: "waiting" };
    return { title: "轮到闲家队反主", detail: "选中可反主的牌后点“反主”，也可以点“不反”", tone: "action" };
  }
  if (phase === "take_bottom") {
    if (!isDealerTeam) return { title: "请稍等", detail: "庄家队正在选择谁起底", tone: "waiting" };
    if (!isMyTurn) return { title: "请稍等", detail: `正在等待 ${currentPlayerName ?? "队友"} 操作`, tone: "waiting" };
    return { title: "轮到你起底", detail: "可以自己起底，也可以交给队友起底", tone: "action" };
  }
  if (phase === "discard_bottom") {
    if (mySeat !== takeBottomSeat) return { title: "请稍等", detail: "正在等待起底玩家扣底", tone: "waiting" };
    return { title: "轮到你扣底", detail: "选好要扣的底牌后，点“扣底”", tone: "action" };
  }
  if (phase === "playing") {
    if (!isMyTurn) return { title: "请稍等", detail: `正在等待 ${currentPlayerName ?? "其他玩家"} 出牌`, tone: "waiting" };
    return { title: "轮到你出牌", detail: "先点选手牌，再点“出牌”", tone: "action" };
  }
  return { title: "请查看当前提示", tone: "waiting" };
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

function extractHandCards(p: ServerPlayer): number[] {
  if (!p?.hand) return [];
  return p.hand.map((c) => (typeof c === "number" ? c : c.id));
}

function toTablePlayer(p: ServerPlayer, currentSeat: number | undefined): TablePlayer {
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
    isDealerTeam: false, // only set after game_start reveals actual dealer seats
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
  const searchParams = useSearchParams();
  const roomId = params.id as string;
  const roomPassword = searchParams.get("password") || "";

  const [players, setPlayers] = useState<TablePlayer[]>([]);
  const playersRef = useRef<TablePlayer[]>([]);
  const setPlayersWithRef = (next: TablePlayer[] | ((current: TablePlayer[]) => TablePlayer[])) => {
    setPlayers((current) => {
      const resolved = typeof next === "function" ? next(current) : next;
      playersRef.current = resolved;
      return resolved;
    });
  };
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
  const [teamLevels, setTeamLevels] = useState<[number, number]>([3, 3]);
  const [bottomCards, setBottomCards] = useState<number[]>([]);
  const [discardedCards, setDiscardedCards] = useState<number[]>([]);
  const [takeBottomSeat, setTakeBottomSeat] = useState<number>(-1);
  const [lastPlay, setLastPlay] = useState<{ seat: number; cards: number[] } | null>(null);
  const [trickPlays, setTrickPlays] = useState<Record<number, number[]>>({});
  const [roundPoints, setRoundPoints] = useState(0);
  const [roundResult, setRoundResult] = useState<{ scores: Array<{ player_id: number; score: number }> } | null>(null);
  const [selectedCards, setSelectedCards] = useState<number[]>([]);
  const [errorMessage, setErrorMessage] = useState<string | null>(null);
  const [toasts, setToasts] = useState<Array<{ id: number; text: string }>>([]);
  const [inviteCopied, setInviteCopied] = useState(false);
  const [showPasswordPrompt, setShowPasswordPrompt] = useState(false);
  const [passwordError, setPasswordError] = useState("");
  const toastIdRef = useRef(0);
  const noticeSeqRef = useRef(0);
  const [dealerSeats, setDealerSeats] = useState<[number, number]>([-1, -1]);
  const dealerSeatsRef = useRef<[number, number]>([-1, -1]);
  const setDealerSeatsWithRef = (seats: [number, number]) => {
    dealerSeatsRef.current = seats;
    setDealerSeats(seats);
  };

  const wsRef = useRef<WSGameClient | null>(null);

  useEffect(() => {
    const token = typeof window !== "undefined" ? localStorage.getItem("token") : null;
    if (!token) {
      const currentUrl = window.location.pathname + window.location.search;
      router.push(`/auth/login?redirect=${encodeURIComponent(currentUrl)}`);
      return;
    }

    let userIdStr = "";
    try {
      const payload = JSON.parse(atob(token.split(".")[1]));
      userIdStr = String(payload.user_id ?? payload.sub ?? "");
    } catch {
      const currentUrl = window.location.pathname + window.location.search;
      router.push(`/auth/login?redirect=${encodeURIComponent(currentUrl)}`);
      return;
    }

    const client = new WSGameClient(Number(userIdStr), token, roomId, "dashengji", roomPassword);
    wsRef.current = client;
    const uid = userIdStr;
    const showToast = (text: string) => {
      const id = ++toastIdRef.current;
      setToasts((current) => [...current, { id, text }]);
      window.setTimeout(() => setToasts((current) => current.filter((toast) => toast.id !== id)), 4000);
    };
    const applyNotices = (data: ServerData) => {
      const noticePlayers = mergeDashengjiNoticePlayers(data.players, playersRef.current);
      for (const notice of data.notices ?? []) {
        if (notice.seq <= noticeSeqRef.current) continue;
        noticeSeqRef.current = notice.seq;
        if (notice.kind === "bottom_delegated") {
          if (notice.seat === mySeatRef.current) showToast("队友将起底让与您操作");
        } else if (notice.kind === "dealer_swapped") {
          showToast("庄方无法定主，局内换庄");
        } else if (notice.kind === "redeal") {
          showToast("双方无法定主，流局重新发牌");
        } else {
          const text = formatDashengjiNoticeToast(notice, noticePlayers, data.trump_suit ?? trumpSuit);
          if (text) showToast(text);
        }
      }
    };

    client.on("player_joined", (msg) => {
      const data = msg.data as ServerData;
      if (data?.players) {
        setPlayersWithRef(data.players.map((p) => toTablePlayer(p, undefined)));
      }
      if (data?.seat !== undefined && String(data.user_id) === uid) setMySeatWithRef(data.seat);
    });

    client.on("player_ready", (msg) => {
      const data = msg.data as ServerData;
      if (data?.players) {
        setPlayersWithRef(data.players.map((p) => toTablePlayer(p, currentSeat)));
      }
    });

    client.on("state_update", (msg) => {
      const data = msg.data as ServerData;
      applyNotices(data);

      if (data?.phase !== undefined) {
        setPhase(phaseMap[data.phase] ?? "waiting");
        setSelectedCards([]);
      }
      if (data?.current_seat !== undefined) setCurrentSeat(data.current_seat);
      if (data?.trump_suit !== undefined) {
        setTrumpSuit(data.trump_suit);
      }
      if (data?.current_level !== undefined) setCurrentLevel(data.current_level);
      if (data?.team_levels) setTeamLevels(data.team_levels);
      if (data?.dealer_seats) setDealerSeatsWithRef(data.dealer_seats);
      if (data?.bottom_cards) setBottomCards(extractCards(data.bottom_cards));
      if (data?.discarded_cards) setDiscardedCards(extractCards(data.discarded_cards));
      if (data?.take_bottom_seat !== undefined) setTakeBottomSeat(data.take_bottom_seat);
      if (data?.round_points !== undefined) setRoundPoints(data.round_points);
      if (data?.last_play) {
        setLastPlay({ seat: data.last_play.seat, cards: extractCards(data.last_play.cards) });
      }
      if (data?.play_history) {
        const history = data.play_history;
        const n = history.length;
        const trickSize = n % 4 === 0 ? 4 : n % 4;
        const start = n - trickSize;
        const next: Record<number, number[]> = {};
        for (let i = start; i < n; i++) {
          next[history[i].seat] = extractCards(history[i].cards);
        }
        setTrickPlays(next);
      }
      if (data?.players) {
        const ds = data.dealer_seats || dealerSeatsRef.current;
        setPlayersWithRef((prev) => {
          const newPlayers = data.players!.map((p) => toTablePlayer(p, data.current_seat));
          // Merge with previous to preserve known nicknames and apply dealer team
          return newPlayers.map((np) => {
            const existing = prev.find((pp) => pp.seat === np.seat);
            const merged = { ...np };
            if (existing && existing.nickname && existing.nickname !== String(existing.seat)) {
              merged.nickname = existing.nickname;
              merged.name = existing.nickname;
              merged.characterId = existing.characterId || np.characterId;
            }
            merged.isDealerTeam = ds.includes(np.seat);
            return merged;
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
      noticeSeqRef.current = 0;
      if (data?.dealer_seats) setDealerSeatsWithRef(data.dealer_seats);
      const ds = data.dealer_seats || dealerSeatsRef.current;
      if (data?.players) {
        setPlayersWithRef((prev) => {
          const newPlayers = data.players!.map((p) => toTablePlayer(p, data.current_seat));
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
      setSelectedCards([]);
      setTrickPlays({});
      setRoundResult(null);
      if (data?.current_seat !== undefined) setCurrentSeat(data.current_seat);
      if (data?.trump_suit !== undefined) setTrumpSuit(data.trump_suit);
      if (data?.current_level !== undefined) setCurrentLevel(data.current_level);
      if (data?.team_levels) setTeamLevels(data.team_levels);
      // Mark dealer team players after dealer_seats is known
      if (ds) {
        setPlayersWithRef((prev) => prev.map((p) => ({ ...p, isDealerTeam: ds.includes(p.seat) })));
      }
      const ownSeat = mySeatRef.current ?? data.players?.find((p) => String(p.user_id ?? p.userId) === uid)?.seat;
      if (ownSeat !== undefined && ownSeat !== null) {
        showToast(ds.includes(ownSeat) ? "您是庄家" : "您是闲家");
      }
    });

    client.on("round_end", (msg) => {
      const state = msg.data as ServerData;
      applyNotices(state);
      setPhase("ended");
      setSelectedCards([]);
      setHand([]);
      setLastPlay(null);
      setTrickPlays({});
      if (state?.scores) {
        setRoundResult({ scores: state.scores });
      }
      if (state?.team_levels) setTeamLevels(state.team_levels);
      if (state?.dealer_seats) setDealerSeatsWithRef(state.dealer_seats);
      if (state?.round_points !== undefined) setRoundPoints(state.round_points);
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
      if (errMsg.indexOf("room password required") !== -1) {
        setPasswordError(roomPassword ? "密码错误，请重试" : "");
        setShowPasswordPrompt(true);
        return;
      }
    });

    client.on("player_left", (msg) => {
      const data = msg.data as ServerData;
      if (data?.players) {
        setPlayersWithRef(data.players.map((p) => toTablePlayer(p, currentSeat)));
      }
    });

    client.connect();
    return () => {
      wsRef.current = null;
      client.disconnect();
    };
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [roomId, roomPassword, router]);

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

  const handleCopyInvite = useCallback(async () => {
    if (typeof window === "undefined") return;
    const inviteUrl = new URL(`/room/${roomId}/dashengji`, window.location.origin);
    if (roomPassword) inviteUrl.searchParams.set("password", roomPassword);
    await navigator.clipboard.writeText(inviteUrl.toString());
    setInviteCopied(true);
    window.setTimeout(() => setInviteCopied(false), 1800);
  }, [roomId, roomPassword]);


  const myUserId = getUserIdFromToken();
  const myPlayer = players.find((p) => p.userId === myUserId);
  const amIOwner = myPlayer?.isOwner ?? false;
  const amIReady = myPlayer?.isReady ?? false;
  const allReady = players.length >= 2 && players.every((p) => p.isReady);
  const canStart = amIOwner && players.length >= GAME_CONFIG.maxPlayers && allReady;
  const isMyTurn = currentSeat !== undefined && mySeat !== null && mySeat === currentSeat;
  const isDealerTeam = mySeat !== null && dealerSeats.includes(mySeat);
  const showPhaseActions = phase !== "waiting" && phase !== "ended";
  const sortedHand = useMemo(() => sortDashengjiHand(hand, trumpSuit, currentLevel), [hand, trumpSuit, currentLevel]);
  const teamNames = useMemo<[string, string]>(() => {
    const bySeat = (seat: number) => players.find((p) => p.seat === seat)?.nickname ?? `玩家${seat + 1}`;
    return [`${bySeat(0)} + ${bySeat(2)}`, `${bySeat(1)} + ${bySeat(3)}`];
  }, [players]);
  const currentPlayerName = players.find((p) => p.seat === currentSeat)?.nickname;
  const prompt = getDashengjiPrompt(phase, isMyTurn, isDealerTeam, mySeat, takeBottomSeat, currentPlayerName, errorMessage);
  const showPrompt = phase === "waiting" || phase === "ended" || prompt.tone !== "waiting";

  // Count main cards (trump set) to split display into main/secondary rows
  const mainCount = useMemo(() => {
    if (trumpSuit < 0) return 0;
    const levelRank = currentLevel >= 3 && currentLevel <= 14 ? currentLevel : 3;
    let count = 0;
    for (const id of sortedHand) {
      const face = id % 54;
      if (face >= 52) { count++; continue; }
      const suit = Math.floor(face / 13);
      const rank = face % 13;
      const baseRank = rank + 3;
      if (baseRank === 15 || baseRank === levelRank || suit === trumpSuit) {
        count++;
      } else {
        break; // side cards start here
      }
    }
    return count;
  }, [sortedHand, trumpSuit, currentLevel]);

  return (
    <div className="mobile-game-shell relative bg-gradient-to-b from-green-900 via-green-800 to-green-950">
      <div className="fixed right-3 top-[max(12px,var(--safe-area-top))] z-50">
        <button
          type="button"
          onClick={() => { void handleCopyInvite(); }}
          className="min-h-10 rounded-full bg-black/40 px-3 text-sm font-black text-white backdrop-blur border border-white/15"
        >
          {inviteCopied ? "已复制" : "邀请"}
        </button>
      </div>
      <div className="flex h-full flex-col pt-4">
        <RoomTable
          players={players}
          mySeat={mySeat ?? 0}
          phase={phase}
          onSitDown={handleSitDown}
          onAddBot={handleAddBot}
          lastPlay={lastPlay}
          trickPlays={trickPlays}
          bottomCards={takeBottomSeat >= 0 ? bottomCards : undefined}
          discardedCards={takeBottomSeat >= 0 && discardedCards.length > 0 ? discardedCards : undefined}
          bottomSeat={takeBottomSeat >= 0 ? takeBottomSeat : undefined}
          maxPlayers={GAME_CONFIG.maxPlayers}
          waitingLayout="row"
          hideCardCount
        />
      </div>
      <DashengjiInfoPanel trumpSuit={trumpSuit} roundPoints={roundPoints} teamLevels={teamLevels} dealerSeats={dealerSeats} teamNames={teamNames} />
      {showPrompt && (
        <GamePhasePrompt
          title={prompt.title}
          detail={prompt.detail}
          tone={prompt.tone}
          position={phase === "waiting" || phase === "ended" ? "top" : "tableCenter"}
          compact={phase !== "waiting" && phase !== "ended"}
        />
      )}

      {errorMessage && (
        <div className="fixed left-0 right-0 top-[18vh] z-50 flex justify-center px-4 pointer-events-none">
          <div className="max-w-[min(92vw,28rem)] px-4 py-3 bg-red-500/20 border border-red-500/40 rounded-lg text-base font-bold text-red-200 text-center animate-pulse">
          {errorMessage}
          </div>
        </div>
      )}

      <div className="fixed left-0 right-0 top-[18vh] z-50 flex flex-col gap-2 items-center px-4 pointer-events-none">
        {toasts.map((toast) => (
          <div key={toast.id} className="max-w-[min(92vw,32rem)] px-5 py-3 bg-amber-500/20 border border-amber-500/40 rounded-lg text-base text-amber-300 text-center font-bold animate-toast-in whitespace-normal break-words">
            {toast.text}
          </div>
        ))}
      </div>

      {showPhaseActions && (
        <div className="mobile-bottom-controls fixed bottom-0 w-full z-30 bg-gradient-to-t from-black/90 via-black/60 to-transparent pt-5">
          {sortedHand.length > 0 && (
            /* Drop internal card selection when discarded bottom cards leave the hand. */
            <HandCards
              key={phase}
              cards={sortedHand}
              onPlayCards={phase === "playing" ? handlePlayCards : undefined}
              onSelectionChange={setSelectedCards}
              disabled={!isMyTurn}
              mainCount={trumpSuit >= 0 ? mainCount : undefined}
              hidePass
            />
          )}
          <DashengjiActionBar phase={phase} isMyTurn={isMyTurn} isDealerTeam={isDealerTeam} selectedCards={selectedCards} mySeat={mySeat} takeBottomSeat={takeBottomSeat} onAction={handleAction} />
        </div>
      )}

      {phase === "waiting" && (
        <div className="mobile-bottom-controls fixed bottom-0 left-0 right-0 z-40 bg-gradient-to-t from-black/80 to-transparent pt-4">
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
        <div className="fixed inset-0 bg-black/70 flex items-center justify-center z-50 px-4">
          <div className="w-full max-w-[390px] bg-gray-900 rounded-2xl p-5 text-white text-center">
            <h2 className="text-3xl font-black mb-4">本轮结束</h2>
            {players.map((p) => {
              const score = roundResult.scores.find((s) => s.player_id === p.seat);
              return (
                <div key={p.userId} className="flex items-center justify-between px-4 py-2 bg-white/5 rounded-xl mb-2">
                  <span className="text-lg font-bold">{p.nickname}</span>
                  <span className={`text-xl font-black ${score && score.score > 0 ? "text-amber-400" : "text-red-400"}`}>
                    {score ? (score.score > 0 ? "+" : "") + score.score : "0"}
                  </span>
                </div>
              );
            })}
            <div className="grid grid-cols-2 gap-2 mt-6">
              <button
                onClick={() => {
                  setPhase("waiting");
                  setSelectedCards([]);
                  setRoundResult(null);
                  setLastPlay(null);
                  setCurrentSeat(undefined);
                  handleReady();
                }}
                className="min-h-14 px-4 py-3 rounded-full gold-button text-lg font-black"
              >
                再来一局
              </button>
              <button
                onClick={() => router.push("/lobby")}
                className="min-h-14 px-4 py-3 rounded-full bg-white/10 text-lg font-black text-white"
              >
                返回大厅
              </button>
            </div>
          </div>
        </div>
      )}

      <PasswordPrompt
        open={showPasswordPrompt}
        error={passwordError}
        loading={false}
        onSubmit={(pw) => {
          setPasswordError("");
          wsRef.current?.rejoinRoom(pw);
          setShowPasswordPrompt(false);
        }}
        onCancel={() => {
          setShowPasswordPrompt(false);
          router.push("/lobby");
        }}
      />
    </div>
  );
}
