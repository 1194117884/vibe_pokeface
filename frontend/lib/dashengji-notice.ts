export interface DashengjiNotice {
  seq: number;
  kind: string;
  seat?: number;
  action?: string;
}

export interface DashengjiNoticePlayer {
  seat?: number;
  nickname?: string;
  name?: string;
}

const SUIT_NAMES = ["黑桃", "红桃", "梅花", "方块"] as const;

function playerName(players: DashengjiNoticePlayer[] | undefined, seat: number | undefined): string {
  const player = players?.find((p) => (p.seat ?? -1) === seat);
  return player?.nickname ?? player?.name ?? `玩家${(seat ?? 0) + 1}`;
}

function trumpSuitName(trumpSuit: number): string {
  return SUIT_NAMES[trumpSuit] ?? "未知";
}

export function mergeDashengjiNoticePlayers(
  current: DashengjiNoticePlayer[] | undefined,
  cached: DashengjiNoticePlayer[] | undefined,
): DashengjiNoticePlayer[] | undefined {
  if (!current) return cached;
  if (!cached) return current;

  return current.map((player) => {
    const cachedPlayer = cached.find((p) => p.seat === player.seat);
    return {
      ...player,
      name: player.name ?? cachedPlayer?.name,
      nickname: player.nickname ?? cachedPlayer?.nickname ?? cachedPlayer?.name,
    };
  });
}

export function formatDashengjiNoticeToast(
  notice: DashengjiNotice,
  players: DashengjiNoticePlayer[] | undefined,
  trumpSuit: number,
): string | null {
  const name = playerName(players, notice.seat);

  if (notice.kind === "trump_decision") {
    if (notice.action === "set_trump") return `${name}:定主-${trumpSuitName(trumpSuit)}`;
    if (notice.action === "pass_trump") return `${name}:不定主`;
  }

  if (notice.kind === "counter_decision") {
    if (notice.action === "counter_trump") return `${name}:反主-${trumpSuitName(trumpSuit)}`;
    if (notice.action === "pass_counter") return `${name}:不反主`;
  }

  if (notice.kind === "bottom_taken") return `${name}:起底了`;
  if (notice.kind === "bottom_discarded") return `${name}:完成扣底`;
  if (notice.kind === "trick_winner") return `${name} 牌最大`;

  return null;
}
