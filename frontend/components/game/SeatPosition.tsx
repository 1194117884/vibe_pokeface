"use client";

import clsx from "clsx";
import { getCharacterStyle } from "@/themes";

const SUITS = ["♠", "♥", "♣", "♦"] as const;
const RANKS = ["3", "4", "5", "6", "7", "8", "9", "10", "J", "Q", "K", "A", "2"] as const;

const phaseLabels: Record<string, string> = {
  set_trump: "定主中",
  counter_trump: "反主中",
  take_bottom: "起底中",
  discard_bottom: "扣底中",
  playing: "出牌中",
};

function miniCardFace(cardId: number): number { return cardId % 54; }
function miniCardColor(cardId: number): string {
  const face = cardId % 54;
  if (face >= 52) return face === 52 ? "text-[#1a1a1a]" : "text-[#c82014]";
  const suit = Math.floor(face / 13);
  return suit === 1 || suit === 3 ? "text-[#c82014]" : "text-[#1a1a1a]";
}
function miniCardSuit(cardId: number): number {
  const face = cardId % 54;
  return face >= 52 ? 4 : Math.floor(face / 13);
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
    isDealerTeam?: boolean;
    cardCount: number;
    hand?: number[];
  } | null;
  position?: "top" | "bottom" | "left" | "right";
  isMySeat?: boolean;
  onChangeSeat?: () => void;
  onAddBot?: () => void;
  cardsLeft?: "baodan" | "baoshuang" | null;
  action?: string | null;
  landlordCards?: number[];
  bottomCards?: number[];
  discardedCards?: number[];
  compact?: boolean;
  phase?: string;
}

export function SeatPosition({
  player,
  isMySeat,
  onChangeSeat,
  onAddBot,
  cardsLeft,
  action,
  landlordCards,
  bottomCards,
  discardedCards,
  compact = false,
  position,
  phase,
}: SeatPositionProps) {
  const showTurnBadge = player?.isCurrentTurn;
  const turnLabel = (phase && phaseLabels[phase]) || "出牌中";

  // Stitch opponent mode: compact display for left/right opponents during game phases
  if (position && player) {
    const charStyle = !player.isBot
      ? getCharacterStyle(player.characterId || "panda")
      : undefined;

    return (
      <div
        className={clsx(
          "relative flex rounded-xl transition-all",
          position === "top"
            ? "flex-row items-center gap-2 p-1.5"
            : compact ? "flex-col items-center gap-1 p-2"
            : "flex-col items-center gap-2 p-3",
          compact ? "gap-1 p-2" : "gap-2 p-3",
          "bg-surface-container-low/40 backdrop-blur-md border border-white/5 shadow-lg",
          player.isCurrentTurn && "border-primary/50 ring-2 ring-primary/20 shadow-[0_0_16px_rgba(142,213,175,0.2)]",
          player.isLandlord && "border-secondary-container/50 ring-1 ring-secondary-container/20",
          player.isDealerTeam && "!border-amber-500/60 ring-1 ring-amber-500/30 bg-amber-500/10",
        )}
      >
        {/* Speech bubble — positioned toward table center */}
        {action && (
          <div
            key={action}
            className={clsx(
              "absolute z-20 animate-bubble-in",
              position === "left" && "right-0 translate-x-[calc(100%+8px)]",
              position === "right" && "left-0 -translate-x-[calc(100%+8px)]",
            )}
          >
            <div className="bg-black/85 backdrop-blur-sm text-white text-sm font-bold px-3.5 py-1.5 rounded-xl whitespace-nowrap shadow-lg border border-white/10">
              {action}
            </div>
            <div className={clsx(
              "absolute top-1/2 -translate-y-1/2 w-3 h-3 bg-black/85 rotate-45",
              position === "left" && "-left-1.5",
              position === "right" && "-right-1.5",
            )} />
          </div>
        )}

        {/* Owner badge */}
        {player.isOwner && (
          <span
            className="absolute -top-2 -left-1 text-base drop-shadow-md z-10"
            title="房主"
            style={{ color: "#e9c400" }}
          >
            ⭐
          </span>
        )}

        {/* Landlord badge */}
        {player.isLandlord && (
          <div className="absolute -top-3.5 left-1/2 -translate-x-1/2 bg-secondary-container text-on-secondary-container text-[11px] font-extrabold px-3 py-0.5 rounded-full whitespace-nowrap shadow-md z-10">
            👑 地主
          </div>
        )}

        {/* Dealer team badge */}
        {player.isDealerTeam && (
          <div className="absolute -top-3.5 left-1/2 -translate-x-1/2 bg-amber-500 text-white text-[11px] font-extrabold px-3 py-0.5 rounded-full whitespace-nowrap shadow-md z-10">
            🏠 庄
          </div>
        )}

        {/* Phase turn badge */}
        {showTurnBadge && (
          <div className="absolute -bottom-2 left-1/2 -translate-x-1/2 bg-primary text-on-primary text-[10px] font-bold px-2.5 py-0.5 rounded-full whitespace-nowrap shadow-[0_0_8px_rgba(142,213,175,0.3)] z-10">
            ⚡ {turnLabel}
          </div>
        )}

        {/* Baodan / Baoshuang badges */}
        {cardsLeft === "baodan" && (
          <span className="absolute -top-2 right-0 bg-red-500 text-white text-[10px] px-1.5 py-0.5 rounded-full font-bold animate-pulse z-10">
            报单
          </span>
        )}
        {cardsLeft === "baoshuang" && (
          <span className="absolute -top-2 right-0 bg-orange-400 text-white text-[10px] px-1.5 py-0.5 rounded-full font-bold animate-pulse z-10">
            报双
          </span>
        )}

        {/* Revealed hand (only when cards are exposed) */}
        {player.hand && player.hand.length > 0 && (
          <div className="flex flex-wrap justify-center gap-0.5 max-w-[100px] mb-0.5">
            {player.hand.map((cardId, i) => (
              <div
                key={i}
                className="w-3.5 h-5.5 bg-white rounded-[2px] flex flex-col items-center justify-center text-[7px] leading-none shadow-sm border border-black/10"
              >
                <span className={miniCardColor(cardId)}>
                  {cardId % 54 >= 52 ? (cardId % 54 === 52 ? "小" : "大") : RANKS[cardId % 54 % 13]}
                </span>
                <span className={miniCardColor(cardId)}>
                  {cardId % 54 >= 52 ? "王" : SUITS[Math.floor((cardId % 54) / 13)]}
                </span>
              </div>
            ))}
          </div>
        )}

        {/* Avatar */}
        {/* <div className="relative -mt-3">
          <div
            className={clsx(
              "rounded-full border-2 border-surface-container-highest overflow-hidden",
              compact ? "w-10 h-10" : "w-12 h-12",
              player.isLandlord && "border-secondary-container/50",
              player.isDealerTeam && "!border-amber-500",
            )}
          >
            {player.isBot ? (
              <div className="w-full h-full bg-purple-500 flex items-center justify-center text-white font-bold text-lg">
                AI
              </div>
            ) : (
              <div
                className="w-full h-full flex items-center justify-center text-2xl"
                style={{
                  backgroundColor: charStyle?.backgroundColor ?? "#374151",
                }}
              >
                {charStyle?.emoji ?? "🐼"}
              </div>
            )}
          </div>
          {player.cardCount > 0 && (
            <div className="absolute -top-1 -right-1 bg-secondary-container text-on-secondary-container rounded-full w-5 h-5 flex items-center justify-center text-[9px] font-bold border border-black/20">
              {player.cardCount}
            </div>
          )}
        </div> */}

        {/* Name */}
        <p className={clsx("font-player-name text-on-surface text-center truncate", compact ? "text-[10px] max-w-[70px]" : "text-xs max-w-[90px]")}>
          {player.nickname || player.name}
        </p>

        {/* Landlord cards (底牌) */}
        {player.isLandlord && landlordCards && landlordCards.length > 0 && (
          <div className="flex gap-0.5 mt-1">
            {landlordCards.map((cardId, i) => (
              <div
                key={i}
                className="w-3 h-5 bg-white rounded-md flex flex-col items-center justify-center text-[8px] leading-none shadow"
              >
                <span className={miniCardColor(cardId)}>
                  {cardId % 54 >= 52 ? (cardId % 54 === 52 ? "小" : "大") : RANKS[cardId % 54 % 13]}
                </span>
                <span className={miniCardColor(cardId)}>
                  {cardId % 54 >= 52 ? "王" : SUITS[Math.floor((cardId % 54) / 13)]}
                </span>
              </div>
            ))}
          </div>
        )}

        {/* Dashengji bottom cards (底牌) */}
        {bottomCards && bottomCards.length > 0 && (
          <div className="flex gap-0.5 mt-1">
            <span className="text-[8px] text-amber-300/60 mr-0.5">底</span>
            {bottomCards.map((cardId, i) => (
              <div
                key={i}
                className="w-3 h-5 bg-amber-100 rounded-md flex flex-col items-center justify-center text-[8px] leading-none shadow"
              >
                <span className={miniCardColor(cardId)}>
                  {cardId % 54 >= 52 ? (cardId % 54 === 52 ? "小" : "大") : RANKS[cardId % 54 % 13]}
                </span>
                <span className={miniCardColor(cardId)}>
                  {cardId % 54 >= 52 ? "王" : SUITS[Math.floor((cardId % 54) / 13)]}
                </span>
              </div>
            ))}
          </div>
        )}

        {/* Dashengji discarded cards (扣底) */}
        {discardedCards && discardedCards.length > 0 && (
          <div className="flex gap-0.5 mt-0.5">
            <span className="text-[8px] text-red-300/60 mr-0.5">扣</span>
            {discardedCards.map((cardId, i) => (
              <div
                key={i}
                className="w-3 h-5 bg-red-100 rounded-md flex flex-col items-center justify-center text-[8px] leading-none shadow"
              >
                <span className={miniCardColor(cardId)}>
                  {cardId % 54 >= 52 ? (cardId % 54 === 52 ? "小" : "大") : RANKS[cardId % 54 % 13]}
                </span>
                <span className={miniCardColor(cardId)}>
                  {cardId % 54 >= 52 ? "王" : SUITS[Math.floor((cardId % 54) / 13)]}
                </span>
              </div>
            ))}
          </div>
        )}
      </div>
    );
  }

  // Empty seat (waiting phase)
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

  // Full SeatPosition (waiting phase, no position)
  const charStyle = !player.isBot
    ? getCharacterStyle(player.characterId || "panda")
    : undefined;

  return (
    <div
      className={clsx(
        "relative flex flex-col items-center rounded-xl transition-all",
        compact ? "gap-1.5 p-2 min-w-[80px]" : "gap-3 p-4 min-w-[120px]",
        "bg-surface-container-low/40 backdrop-blur-md border border-white/5 shadow-lg",
        player.isCurrentTurn && "border-primary/50 ring-2 ring-primary/20 shadow-[0_0_16px_rgba(142,213,175,0.2)]",
        player.isLandlord && "border-secondary-container/50 ring-1 ring-secondary-container/20",
        player.isDealerTeam && "!border-amber-500/60 ring-1 ring-amber-500/30 bg-amber-500/10",
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

      {/* Dealer team badge */}
      {player.isDealerTeam && (
        <div className="absolute -top-3.5 left-1/2 -translate-x-1/2 bg-amber-500 text-white text-[11px] font-extrabold px-3 py-0.5 rounded-full whitespace-nowrap shadow-md">
          🏠 庄
        </div>
      )}

      {/* Phase turn badge */}
      {showTurnBadge && (
        <div className="absolute -bottom-2 left-1/2 -translate-x-1/2 bg-primary text-on-primary text-[10px] font-bold px-2.5 py-0.5 rounded-full whitespace-nowrap shadow-[0_0_8px_rgba(142,213,175,0.3)]">
          ⚡ {turnLabel}
        </div>
      )}

      {/* Avatar */}
      <div className="relative">
        <div
          className={clsx(
            "rounded-full border-surface-container-highest overflow-hidden",
            compact ? "w-10 h-10 border-2" : "w-16 h-16 border-4",
            player.isLandlord && "border-secondary-container/50",
            player.isDealerTeam && "!border-amber-500",
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
      <p className={clsx("font-player-name text-on-surface text-center truncate", compact ? "text-[10px] max-w-[70px]" : "text-player-name max-w-[110px]")}>
        {player.nickname || player.name}
      </p>

      {/* Status */}
      <p className={clsx("text-primary font-bold", compact ? "text-[10px]" : "text-label-md")}>
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

      {/* Dashengji bottom cards (底牌) */}
      {bottomCards && bottomCards.length > 0 && (
        <div className="flex gap-0.5 mt-1">
          <span className="text-[10px] text-amber-300/60 mr-0.5">底</span>
          {bottomCards.map((cardId, i) => (
            <div
              key={i}
              className="w-7 h-10 bg-amber-100 rounded-md flex flex-col items-center justify-center text-[8px] leading-none shadow"
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

      {/* Dashengji discarded cards (扣底) */}
      {discardedCards && discardedCards.length > 0 && (
        <div className="flex gap-0.5 mt-0.5">
          <span className="text-[10px] text-red-300/60 mr-0.5">扣</span>
          {discardedCards.map((cardId, i) => (
            <div
              key={i}
              className="w-7 h-10 bg-red-100 rounded-md flex flex-col items-center justify-center text-[8px] leading-none shadow"
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
