import { WebSocketServer, WebSocket } from "ws";

interface MockPlayer {
  userId: string;
  ws: WebSocket;
  seat: number;
}

export interface GameScenario {
  onMessage(msg: any, player: MockPlayer, server: MockGameServer): void;
}

export class MockGameServer {
  private wss: WebSocketServer;
  private connections: MockPlayer[] = [];
  private scenario: GameScenario;

  constructor(scenario: GameScenario) {
    this.scenario = scenario;
    this.wss = new WebSocketServer({ port: 0 }); // random port
  }

  start(port?: number): Promise<number> {
    return new Promise((resolve) => {
      if (port !== undefined) {
        // Close and reopen on specific port
        this.wss.close();
        this.wss = new WebSocketServer({ port });
      }

      this.wss.removeAllListeners();
      this.wss.on("connection", (ws, req) => {
        const url = new URL(req.url || "", "http://localhost");
        const userId = url.searchParams.get("user_id") || "unknown";
        const seat = this.connections.length;

        const player: MockPlayer = { userId, ws, seat };
        this.connections.push(player);

        ws.on("message", (raw) => {
          try {
            const msg = JSON.parse(raw.toString());
            this.scenario.onMessage(msg, player, this);
          } catch (e) {
            console.error("Mock server parse error:", e);
          }
        });

        ws.on("close", () => {
          this.connections = this.connections.filter((p) => p.ws !== ws);
        });

        ws.on("error", () => {});

        // Send join confirmation
        ws.send(JSON.stringify({
          type: "player_joined",
          data: {
            user_id: userId,
            seat,
            players: this.connections.map((p) => ({
              user_id: p.userId,
              seat: p.seat,
              ready: false,
              is_bot: false,
              is_owner: p.seat === 0,
              nickname: `Player ${p.seat}`,
            })),
            theme: "classic-poker",
            game_type: "doudizhu",
            max_players: 3,
          },
        }));
      });

      this.wss.on("listening", () => {
        const addr = this.wss.address();
        if (typeof addr === "object" && addr) {
          resolve(addr.port);
        }
      });
    });
  }

  stop() {
    this.wss.close();
  }

  broadcast(msg: object) {
    const data = JSON.stringify(msg);
    this.connections.forEach((p) => {
      if (p.ws.readyState === WebSocket.OPEN) {
        p.ws.send(data);
      }
    });
  }

  sendTo(userId: string, msg: object) {
    const p = this.connections.find((c) => c.userId === userId);
    if (p && p.ws.readyState === WebSocket.OPEN) {
      p.ws.send(JSON.stringify(msg));
    }
  }
}
