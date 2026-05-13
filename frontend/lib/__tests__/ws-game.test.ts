import { describe, it, expect, beforeEach, afterEach, vi } from "vitest";

class MockWebSocket {
  onopen: (() => void) | null = null;
  onclose: ((event: { code: number; reason: string }) => void) | null = null;
  onmessage: ((event: { data: string }) => void) | null = null;
  onerror: ((event: Event) => void) | null = null;
  readyState: number = WebSocket.CONNECTING;
  url: string;

  constructor(url: string) {
    this.url = url;
  }

  // eslint-disable-next-line @typescript-eslint/no-unused-vars
  send(_data: string) {}

  close(code?: number, reason?: string) {
    this.readyState = WebSocket.CLOSED;
    if (this.onclose) {
      this.onclose({ code: code ?? 1000, reason: reason ?? "" });
    }
  }

  _open() {
    this.readyState = WebSocket.OPEN;
    if (this.onopen) this.onopen();
  }

  _receive(data: object) {
    if (this.onmessage) {
      this.onmessage({ data: JSON.stringify(data) });
    }
  }
}

let mockWsInstances: MockWebSocket[] = [];
const originalWebSocket = globalThis.WebSocket;

describe("WebSocket Game Client", () => {
  beforeEach(() => {
    mockWsInstances = [];

    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    (globalThis as any).WebSocket = class extends MockWebSocket {
      constructor(url: string | URL) {
        super(url.toString());
        mockWsInstances.push(this);
      }
    };
  });

  afterEach(() => {
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    (globalThis as any).WebSocket = originalWebSocket;
  });

  function createTestClient() {
    const messages: object[] = [];

    const client = {
      messages,
      connectionState: "disconnected",
      lastError: null as string | null,

      connect(url: string, token: string) {
        const ws = new WebSocket(`${url}?token=${token}`);
        client.connectionState = "connecting";

        ws.onopen = () => {
          client.connectionState = "connected";
        };

        ws.onmessage = (event) => {
          const data = JSON.parse(event.data);
          messages.push(data);
        };

        ws.onerror = () => {
          client.lastError = "WebSocket error";
        };

        ws.onclose = (event) => {
          client.connectionState = "disconnected";
          if (event.code !== 1000) {
            client.lastError = `closed: ${event.reason}`;
          }
        };

        return ws;
      },

      sendAction(ws: WebSocket, action: string, payload: object) {
        ws.send(JSON.stringify({ type: action, payload }));
      },
    };

    return client;
  }

  it("should connect with token in URL", () => {
    const client = createTestClient();
    const ws = client.connect("ws://localhost:8080/ws", "test-token");

    expect(ws.url).toBe("ws://localhost:8080/ws?token=test-token");
  });

  it("should update state on connection open", () => {
    const client = createTestClient();
    client.connect("ws://localhost:8080/ws", "token");

    expect(mockWsInstances.length).toBe(1);
    expect(client.connectionState).toBe("connecting");

    mockWsInstances[0]._open();
    expect(client.connectionState).toBe("connected");
  });

  it("should handle incoming messages", () => {
    const client = createTestClient();
    client.connect("ws://localhost:8080/ws", "token");
    mockWsInstances[0]._open();

    mockWsInstances[0]._receive({ type: "state_update", payload: { room_id: "123" } });
    mockWsInstances[0]._receive({ type: "chat_message", payload: { content: "hello" } });

    expect(client.messages).toHaveLength(2);
    expect(client.messages[0]).toEqual({ type: "state_update", payload: { room_id: "123" } });
    expect(client.messages[1]).toEqual({ type: "chat_message", payload: { content: "hello" } });
  });

  it("should handle connection close", () => {
    const client = createTestClient();
    client.connect("ws://localhost:8080/ws", "token");
    mockWsInstances[0]._open();

    mockWsInstances[0].close(1001, "room closed");
    expect(client.connectionState).toBe("disconnected");
    expect(client.lastError).toBe("closed: room closed");
  });

  it("should send room_action messages correctly", () => {
    const client = createTestClient();
    const ws = client.connect("ws://localhost:8080/ws", "token");

    const sendSpy = vi.spyOn(ws, "send");
    client.sendAction(ws, "room_action", { action: "play", data: [3, 4, 5] });

    expect(sendSpy).toHaveBeenCalledWith(
      JSON.stringify({ type: "room_action", payload: { action: "play", data: [3, 4, 5] } })
    );
  });

  it("should handle join_room flow", () => {
    const client = createTestClient();
    const ws = client.connect("ws://localhost:8080/ws", "token");
    mockWsInstances[0]._open();

    const sendSpy = vi.spyOn(ws, "send");

    client.sendAction(ws, "join_room", { room_id: "abc-123" });
    expect(sendSpy).toHaveBeenCalledWith(
      JSON.stringify({ type: "join_room", payload: { room_id: "abc-123" } })
    );

    mockWsInstances[0]._receive({
      type: "state_update",
      payload: { room_id: "abc-123", players: [{ id: "me", nickname: "test" }] },
    });

    expect(client.messages).toHaveLength(1);
    expect(client.messages[0]).toEqual({
      type: "state_update",
      payload: { room_id: "abc-123", players: [{ id: "me", nickname: "test" }] },
    });
  });
});
