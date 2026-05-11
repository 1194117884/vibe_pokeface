export type GameMessageType =
  | "join_room"
  | "leave_room"
  | "room_action"
  | "state_update"
  | "round_end"
  | "player_joined"
  | "player_left"
  | "chat"
  | "error"
  | "joined"
  | "left"
  | "action_received";

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

  constructor(userId: number, token: string) {
    const baseUrl = process.env.NEXT_PUBLIC_WS_URL || "ws://localhost:8080";
    this.url = `${baseUrl}/ws?user_id=${userId}`;
    this.token = token;
  }

  connect() {
    this.ws = new WebSocket(this.url);

    this.ws.onopen = () => {
      console.log("WS connected");
      // Re-join room if reconnecting
      if (this.roomId) {
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

    this.ws.onclose = () => {
      console.log("WS disconnected, reconnecting in 3s...");
      this.reconnectTimer = setTimeout(() => this.connect(), 3000);
    };

    this.ws.onerror = (err) => {
      console.error("WS error:", err);
    };
  }

  disconnect() {
    if (this.reconnectTimer) {
      clearTimeout(this.reconnectTimer);
    }
    if (this.ws) {
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
