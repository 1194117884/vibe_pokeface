import { Room, RoomEvent, RemoteParticipant } from "livekit-client";

export type VoiceEventCallback = (identity: string, enabled: boolean) => void;

export class LiveKitClient {
  private room: Room | null = null;
  private url: string = "";
  private token: string = "";
  private roomName: string = "";
  private micEnabled = false;
  private onParticipantToggle: VoiceEventCallback | null = null;

  async connect(wsUrl: string, token: string, roomName: string): Promise<void> {
    this.url = wsUrl;
    this.token = token;
    this.roomName = roomName;

    const room = new Room({
      adaptiveStream: true,
      dynacast: true,
    });

    room.on(RoomEvent.ParticipantConnected, (participant: RemoteParticipant) => {
      participant.on(RoomEvent.TrackSubscribed, () => {
        this.onParticipantToggle?.(participant.identity, true);
      });
    });

    room.on(RoomEvent.ParticipantDisconnected, (participant: RemoteParticipant) => {
      this.onParticipantToggle?.(participant.identity, false);
    });

    await room.connect(wsUrl, token);
    this.room = room;
  }

  async toggleMic(): Promise<boolean> {
    if (!this.room) return false;

    try {
      if (this.micEnabled) {
        await this.room.localParticipant.setMicrophoneEnabled(false);
        this.micEnabled = false;
      } else {
        await this.room.localParticipant.setMicrophoneEnabled(true);
        this.micEnabled = true;
      }
    } catch (e) {
      console.error("Failed to toggle microphone:", e);
    }
    return this.micEnabled;
  }

  isMicEnabled(): boolean {
    return this.micEnabled;
  }

  disconnect(): void {
    this.micEnabled = false;
    this.room?.disconnect();
    this.room = null;
  }

  onParticipantVoiceToggled(callback: VoiceEventCallback): void {
    this.onParticipantToggle = callback;
  }
}
