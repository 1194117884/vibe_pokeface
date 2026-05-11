import { Card } from "@/components/ui/Card";
import { Button } from "@/components/ui/Button";
import Link from "next/link";

export default function LobbyPage() {
  return (
    <div className="max-w-4xl mx-auto py-8 px-4">
      <div className="flex justify-between items-center mb-6">
        <h1 className="text-2xl font-semibold text-starbucks">Game Lobby</h1>
        <Link href="/room/create">
          <Button variant="primary">Create Room</Button>
        </Link>
      </div>
      <Card>
        <div className="text-center py-8 text-text-black-soft">
          <p>No active rooms yet.</p>
          <p className="text-sm mt-1">Create a room to start playing!</p>
        </div>
      </Card>
    </div>
  );
}
