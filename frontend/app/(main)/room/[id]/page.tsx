"use client";

import { useEffect, useState, useRef } from "react";
import { useParams } from "next/navigation";
import clsx from "clsx";
import { GameTable } from "@/components/game/GameTable";
import { ChatPanel } from "@/components/chat/ChatPanel";
import { VoiceButton } from "@/components/chat/VoiceButton";
import { LiveKitClient } from "@/lib/livekit-client";
import { WSGameClient } from "@/lib/ws-game";

interface ChatMessage {
  userId: string;
  content: string;
  type: "text" | "emoji";
  timestamp: number;
}

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
  const [chatMessages, setChatMessages] = useState<ChatMessage[]>([]);
  const wsClientRef = useRef<WSGameClient | null>(null);
  const voiceClientRef = useRef<LiveKitClient | null>(null);
  const [micEnabled, setMicEnabled] = useState(false);
  const [chatOpen, setChatOpen] = useState(false);

  useEffect(() => {
    const userId = 1; // TODO: decode from JWT properly
    const token = typeof window !== "undefined" ? localStorage.getItem("token") : null;
    if (!token) return;

    const client = new WSGameClient(userId, token);
    wsClientRef.current = client;

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

    client.on("chat", (msg) => {
      const data = msg.data as any;
      if (data?.content) {
        setChatMessages((prev) => [...prev, {
          userId: String(data.user_id || "unknown"),
          content: data.content,
          type: data.type === "emoji" ? "emoji" : "text",
          timestamp: data.timestamp || Date.now(),
        }]);
      }
    });

    client.connect();

    return () => {
      voiceClientRef.current?.disconnect();
      voiceClientRef.current = null;
      wsClientRef.current = null;
      client.disconnect();
    };
  }, [roomId]);

  const handleVoiceToggle = async (enabled: boolean) => {
    if (enabled) {
      const token = typeof window !== "undefined" ? localStorage.getItem("token") : null;
      if (!token) return;
      // Disconnect existing client before creating a new one
      voiceClientRef.current?.disconnect();
      try {
        const res = await fetch(`/api/livekit/token?room=${roomId}`, {
          headers: { Authorization: `Bearer ${token}` },
        });
        const data = await res.json();
        if (data.success && data.token) {
          const vc = new LiveKitClient();
          await vc.connect(data.url, data.token, data.room);
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

  const handleSendChat = (content: string, type: "text" | "emoji") => {
    wsClientRef.current?.sendChat(content, type);
  };

  if (!connected) {
    return (
      <div className="flex items-center justify-center min-h-screen bg-cream">
        <div className="text-center">
          <div className="text-4xl mb-4 animate-pulse">🎴</div>
          <p className="text-text-black-soft">Connecting to room...</p>
        </div>
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-cream flex">
      <div className="flex-1 relative">
        <GameTable
          players={players}
          mySeat={mySeat}
          currentSeat={currentSeat}
          phase={phase}
          plays={plays}
          landlordCards={landlordCards}
        />

        {/* Chat FAB — visible on mobile only */}
        <button
          onClick={() => setChatOpen(true)}
          className="fixed bottom-6 right-6 lg:hidden z-30 w-14 h-14 rounded-full bg-green-accent text-white text-2xl shadow-frap flex items-center justify-center active:scale-[0.95] transition-transform"
          aria-label="Open chat"
        >
          💬
        </button>
      </div>

      {/* Desktop chat sidebar */}
      <div className="hidden lg:flex w-80 p-4 space-y-4 border-l border-ceramic flex-col">
        <div className="flex items-center gap-2">
          <VoiceButton onToggle={handleVoiceToggle} disabled={!connected} />
          <span className="text-sm text-text-black-soft">{micEnabled ? "Mic on" : "Mic off"}</span>
        </div>
        <div className="flex-1 min-h-0">
          <ChatPanel
            messages={chatMessages}
            onSendMessage={handleSendChat}
            disabled={!connected}
          />
        </div>
      </div>

      {/* Mobile chat bottom sheet */}
      {chatOpen && (
        <div className="fixed inset-0 z-50 lg:hidden flex flex-col">
          {/* Backdrop */}
          <div className="absolute inset-0 bg-black/40" onClick={() => setChatOpen(false)} />
          {/* Sheet */}
          <div
            className={clsx(
              "absolute bottom-0 left-0 right-0 bg-white rounded-t-xl shadow-frap",
              "flex flex-col max-h-[70vh] transition-transform duration-300",
              "pb-[var(--safe-area-bottom,0px)]"
            )}
          >
            {/* Handle */}
            <div className="flex items-center justify-between px-4 py-3 border-b border-ceramic">
              <div className="flex items-center gap-2">
                <VoiceButton onToggle={handleVoiceToggle} disabled={!connected} />
                <span className="text-sm text-text-black-soft">{micEnabled ? "Mic on" : "Mic off"}</span>
              </div>
              <button
                onClick={() => setChatOpen(false)}
                className="w-8 h-8 rounded-full bg-cream flex items-center justify-center text-text-black-soft hover:bg-ceramic transition-colors"
                aria-label="Close chat"
              >
                ✕
              </button>
            </div>
            {/* Chat panel */}
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
    </div>
  );
}
