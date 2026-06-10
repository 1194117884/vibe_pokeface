export type GameMessageType =
  | "join_room"
  | "leave_room"
  | "room_action"
  | "state_update"
  | "round_end"
  | "player_joined"
  | "player_left"
  | "round_end"
  | "chat"
  | "error"
  | "joined"
  | "left"
  | "action_received"
  | "change_seat"
  | "seat_changed"
  | "ready"
  | "player_ready"
  | "start_game"
  | "game_start"
  | "add_bot"
  | "room_info"
  | "theme_changed"
  | "cards_left"
  | "ai_status";

export interface GameMessage {
  type: GameMessageType;
  room_id?: string;
  data?: unknown;
  error?: string;
}

export interface RoomAction {
  action: string;
  cards?: number[];
}

const phaseNames: Record<string, string> = {
  calling: "叫地主阶段",
  snatching: "抢地主阶段",
  revealing: "明牌阶段",
  doubling: "加倍阶段",
  playing: "出牌阶段",
};

const actionNames: Record<string, string> = {
  play: "出牌",
  pass: "不出",
  bid_call: "叫地主",
  bid_pass: "不叫",
  reveal_all: "明牌",
  double: "加倍",
  no_double: "不加倍",
};

export interface ErrorData {
  code: string;
  phase?: string;
  action?: string;
}

export function formatError(data: ErrorData): string {
  switch (data.code) {
    case "PHASE_MISMATCH": {
      const p = phaseNames[data.phase ?? ""] ?? data.phase ?? "未知阶段";
      const a = actionNames[data.action ?? ""] ?? data.action ?? "未知动作";
      return `此阶段是"${p}"，不能"${a}"`;
    }
    case "NOT_YOUR_TURN":
      return "不是你的回合";
    case "INVALID_ACTION":
      return "无效操作";
    case "INVALID_CARDS":
      return "无效牌型";
    case "CANNOT_PASS":
      return "当前必须出牌，不能不出";
    case "CANNOT_BEAT":
      return "打不过上家的牌";
    case "WAIT_TEAMMATE":
      return "等待队友操作";
    case "ALREADY_ACTED":
      return "您已操作过，等待队友操作";
    default:
      return `未知错误: ${data.code}`;
  }
}

export class WSGameClient {
  private ws: WebSocket | null = null;
  private url: string;
  private handlers: Map<GameMessageType, (msg: GameMessage) => void> = new Map();
  private token: string;
  private reconnectTimer: ReturnType<typeof setTimeout> | null = null;
  private roomId: string | null = null;
  private autoJoinRoomId: string | null = null;
  private gameType: string;
  private roomPassword: string;
  private reconnectAttempts = 0;
  private maxReconnectAttempts = 10;
  private destroyed = false;

  constructor(userId: number, token: string, autoJoinRoomId?: string, gameType?: string, roomPassword?: string) {
    const baseUrl = process.env.NEXT_PUBLIC_WS_URL || "ws://localhost:8080";
    this.url = `${baseUrl}/ws?user_id=${userId}`;
    this.token = token;
    this.autoJoinRoomId = autoJoinRoomId || null;
    this.gameType = gameType || "doudizhu";
    this.roomPassword = roomPassword || "";
  }

  connect() {
    if (this.destroyed) return;
    this.ws = new WebSocket(this.url);

    this.ws.onopen = () => {
      console.log("WS connected");
      this.reconnectAttempts = 0;
      // Auto-join room if specified in constructor
      if (this.autoJoinRoomId) {
        this.joinRoom(this.autoJoinRoomId, this.gameType);
      }
      // Re-join room if reconnecting
      if (this.roomId && this.roomId !== this.autoJoinRoomId) {
        this.joinRoom(this.roomId, this.gameType);
      }
    };

    this.ws.onmessage = (event) => {
      try {
        const msg: GameMessage = JSON.parse(event.data);
        const handler = this.handlers.get(msg.type);
        if (handler) {
          handler(msg);
        }
      } catch (e) {
        console.error("WS parse error:", e);
      }
    };

    this.ws.onclose = (event: CloseEvent) => {
      console.log(
        `WS disconnected: code=${event.code} reason="${event.reason}" wasClean=${event.wasClean}`,
      );
      if (this.destroyed) return;
      if (this.reconnectAttempts >= this.maxReconnectAttempts) {
        console.error("WS max reconnection attempts reached");
        return;
      }
      this.reconnectAttempts++;
      const delay = Math.min(1000 * Math.pow(2, this.reconnectAttempts - 1), 30000);
      console.log(`WS reconnecting in ${delay}ms (attempt ${this.reconnectAttempts})`);
      this.reconnectTimer = setTimeout(() => this.connect(), delay);
    };

    this.ws.onerror = () => {
      // onerror always fires before onclose; details come in the CloseEvent
    };
  }

  disconnect() {
    this.destroyed = true;
    if (this.reconnectTimer) {
      clearTimeout(this.reconnectTimer);
      this.reconnectTimer = null;
    }
    if (this.ws) {
      this.ws.onclose = null; // prevent reconnect handler from firing
      this.ws.close();
      this.ws = null;
    }
  }

  send(type: GameMessageType, roomId?: string, data?: unknown) {
    if (!this.ws || this.ws.readyState !== WebSocket.OPEN) {
      console.warn("WS not connected");
      return;
    }
    const msg: GameMessage = { type, room_id: roomId, data };
    this.ws.send(JSON.stringify(msg));
  }

  joinRoom(roomId: string, gameType?: string, roomPassword?: string) {
    this.roomId = roomId;
    const password = roomPassword ?? this.roomPassword;
    this.send("join_room", roomId, {
      game_type: gameType || "doudizhu",
      ...(password ? { password } : {}),
    });
  }

  leaveRoom() {
    this.roomId = null;
    this.send("leave_room");
  }

  sendAction(action: string, cards?: number[]) {
    this.send("room_action", this.roomId || undefined, { action, cards } as RoomAction);
  }

  sendChat(content: string, type: "text" | "emoji" = "text") {
    this.send("chat", this.roomId || undefined, { content, type });
  }

  changeSeat(seat: number) {
    this.send("change_seat", this.roomId || undefined, { seat });
  }

  sendReady() {
    this.send("ready", this.roomId || undefined);
  }

  startGame() {
    this.send("start_game", this.roomId || undefined);
  }

  rejoinRoom(password: string): void {
    this.send("join_room", this.roomId || undefined, {
      game_type: this.gameType,
      password,
    });
  }

  addBot(characterId?: number) {
    this.send("add_bot", this.roomId || undefined, characterId != null ? { character_id: characterId } : undefined);
  }

  on(type: GameMessageType, handler: (msg: GameMessage) => void) {
    this.handlers.set(type, handler);
  }

  off(type: GameMessageType) {
    this.handlers.delete(type);
  }

  isConnected(): boolean {
    return this.ws?.readyState === WebSocket.OPEN || false;
  }
}
