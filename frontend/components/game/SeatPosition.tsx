"use client";

import clsx from "clsx";
import { getCharacterStyle } from "@/themes";

const SUITS = ["♠", "♥", "♣", "♦"] as const;
const RANKS = ["3", "4", "5", "6", "7", "8", "9", "10", "J", "Q", "K", "A", "2"] as const;

function miniCardColor(cardId: number): string {
  if (cardId >= 52) return cardId === 52 ? "text-green-700" : "text-[#c82014]";
  const suit = Math.floor(cardId / 13);
  return suit === 1 || suit === 3 ? "text-[#c82014]" : "text-[#1a1a1a]";
}

interface SeatPositionProps {
  seatNumber: number;
  player?: {
    userId: string;
    name: string;
    nickname: string;
    characterId: string;
    isBot: boolean;
    isOwner: boolean;
    isReady: boolean;
    isCurrentTurn?: boolean;
    isLandlord?: boolean;
    cardCount: number;
  } | null;
  position?: "top" | "bottom" | "left" | "right";
  isMySeat?: boolean;
  onChangeSeat?: () => void;
  onAddBot?: () => void;
  cardsLeft?: "baodan" | "baoshuang" | null;
  action?: string | null;
  landlordCards?: number[];
}

export function SeatPosition({
  player,
  isMySeat,
  onChangeSeat,
  onAddBot,
  cardsLeft,
  action,
  landlordCards,
}: SeatPositionProps) {
  if (!player) {
    return (
      <div
        className={clsx(
          "relative flex flex-col items-center gap-2 p-4 rounded-xl border-2 border-dashed border-white/10",
          "bg-white/5 backdrop-blur-sm min-w-[110px]",
        )}
      >
        <div className="w-14 h-14 rounded-full border-2 border-white/10 flex items-center justify-center text-white/30 text-lg">
          ?
        </div>
        <p className="text-xs text-white/40">空座位</p>
        {isMySeat ? (
          <button
            onClick={onChangeSeat}
            className="text-xs text-primary hover:underline font-medium"
          >
            坐下
          </button>
        ) : onAddBot ? (
          <button
            onClick={onAddBot}
            className="text-xs text-primary hover:underline font-medium"
          >
            添加AI
          </button>
        ) : null}
      </div>
    );
  }

  const charStyle = !player.isBot
    ? getCharacterStyle(player.characterId || "panda")
    : undefined;

  return (
    <div
      className={clsx(
        "relative flex flex-col items-center gap-3 p-4 rounded-xl min-w-[120px] transition-all",
        "bg-surface-container-low/40 backdrop-blur-md border border-white/5 shadow-lg",
        // Current turn: gold glow
        player.isCurrentTurn && "border-primary/50 ring-2 ring-primary/20 shadow-[0_0_16px_rgba(142,213,175,0.2)]",
        // Landlord: gold border accent
        player.isLandlord && "border-secondary-container/50 ring-1 ring-secondary-container/20",
        isMySeat && "ring-1 ring-primary/30",
      )}
    >
      {/* Speech bubble */}
      {action && (
        <div
          key={action}
          className="absolute -top-14 left-1/2 -translate-x-1/2 z-20 animate-bubble-in"
        >
          <div className="bg-black/85 backdrop-blur-sm text-white text-sm font-bold px-3.5 py-1.5 rounded-xl whitespace-nowrap shadow-lg border border-white/10">
            {action}
          </div>
          <div className="absolute -bottom-1.5 left-1/2 -translate-x-1/2 w-3 h-3 bg-black/85 rotate-45" />
        </div>
      )}
      {/* Owner badge */}
      {player.isOwner && (
        <span
          className="absolute -top-2 -left-1 text-base drop-shadow-md"
          title="房主"
          style={{ color: "#e9c400" }}
        >
          ⭐
        </span>
      )}

      {/* Landlord badge */}
      {player.isLandlord && (
        <div className="absolute -top-3.5 left-1/2 -translate-x-1/2 bg-secondary-container text-on-secondary-container text-[11px] font-extrabold px-3 py-0.5 rounded-full whitespace-nowrap shadow-md">
          👑 地主
        </div>
      )}

      {/* Current turn badge */}
      {player.isCurrentTurn && (
        <div className="absolute -bottom-2 left-1/2 -translate-x-1/2 bg-primary text-on-primary text-[10px] font-bold px-2.5 py-0.5 rounded-full whitespace-nowrap shadow-[0_0_8px_rgba(142,213,175,0.3)]">
          ⚡ 出牌中
        </div>
      )}

      {/* Avatar */}
      <div className="relative">
        <div
          className={clsx(
            "w-16 h-16 rounded-full border-4 border-surface-container-highest overflow-hidden",
            player.isLandlord && "border-secondary-container/50",
          )}
        >
          {player.isBot ? (
            <div className="w-full h-full bg-purple-500 flex items-center justify-center text-white font-bold text-xl">
              AI
            </div>
          ) : (
            <div
              className="w-full h-full flex items-center justify-center text-3xl"
              style={{
                backgroundColor: charStyle?.backgroundColor ?? "#374151",
              }}
            >
              {charStyle?.emoji ?? "🐼"}
            </div>
          )}
        </div>
        {player.cardCount > 0 && (
          <div className="absolute -top-1 -right-1 bg-secondary-container text-on-secondary-container rounded-full px-2 py-0.5 text-[10px] font-bold shadow-md">
            {player.cardCount}
          </div>
        )}
      </div>

      {/* Name */}
      <p className="text-player-name font-player-name text-on-surface text-center truncate max-w-[110px]">
        {player.nickname || player.name}
      </p>

      {/* Status */}
      <p className="text-label-md text-primary font-bold">
        {player.isReady ? "已准备" : "等待中..."}
      </p>

      {/* Ready badge */}
      {player.isReady && (
        <span className="text-xs px-2 py-0.5 rounded-full bg-primary-container text-on-primary-container font-semibold">
          ✓ 已准备
        </span>
      )}

      {/* Change seat button */}
      {isMySeat && onChangeSeat && (
        <button
          onClick={onChangeSeat}
          className="text-xs text-on-surface-variant hover:text-primary font-medium"
        >
          换座
        </button>
      )}

      {/* Cards left badges */}
      {cardsLeft === "baodan" && (
        <span className="absolute -top-2 right-0 bg-red-500 text-white text-[10px] px-1.5 py-0.5 rounded-full font-bold animate-pulse">
          报单
        </span>
      )}
      {cardsLeft === "baoshuang" && (
        <span className="absolute -top-2 right-0 bg-orange-400 text-white text-[10px] px-1.5 py-0.5 rounded-full font-bold animate-pulse">
          报双
        </span>
      )}

      {/* Landlord cards (底牌) */}
      {player.isLandlord && landlordCards && landlordCards.length > 0 && (
        <div className="flex gap-0.5 mt-1">
          {landlordCards.map((cardId, i) => (
            <div
              key={i}
              className="w-7 h-10 bg-white rounded-md flex flex-col items-center justify-center text-[8px] leading-none shadow"
            >
              <span className={miniCardColor(cardId)}>
                {cardId >= 52 ? (cardId === 52 ? "小" : "大") : RANKS[cardId % 13]}
              </span>
              <span className={miniCardColor(cardId)}>
                {cardId >= 52 ? "王" : SUITS[Math.floor(cardId / 13)]}
              </span>
            </div>
          ))}
        </div>
      )}
    </div>
  );
}
