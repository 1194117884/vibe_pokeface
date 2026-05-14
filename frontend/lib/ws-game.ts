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
  | "cards_left";

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

export class WSGameClient {
  private ws: WebSocket | null = null;
  private url: string;
  private handlers: Map<GameMessageType, (msg: GameMessage) => void> = new Map();
  private token: string;
  private reconnectTimer: ReturnType<typeof setTimeout> | null = null;
  private roomId: string | null = null;
  private autoJoinRoomId: string | null = null;
  private reconnectAttempts = 0;
  private maxReconnectAttempts = 10;
  private destroyed = false;

  constructor(userId: number, token: string, autoJoinRoomId?: string) {
    const baseUrl = process.env.NEXT_PUBLIC_WS_URL || "ws://localhost:8080";
    this.url = `${baseUrl}/ws?user_id=${userId}`;
    this.token = token;
    this.autoJoinRoomId = autoJoinRoomId || null;
  }

  connect() {
    if (this.destroyed) return;
    this.ws = new WebSocket(this.url);

    this.ws.onopen = () => {
      console.log("WS connected");
      this.reconnectAttempts = 0;
      // Auto-join room if specified in constructor
      if (this.autoJoinRoomId) {
        this.joinRoom(this.autoJoinRoomId);
      }
      // Re-join room if reconnecting
      if (this.roomId && this.roomId !== this.autoJoinRoomId) {
        this.send("join_room", this.roomId);
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

  joinRoom(roomId: string) {
    this.roomId = roomId;
    this.send("join_room", roomId);
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
