import { Card } from "@/components/ui/Card";
import { Button } from "@/components/ui/Button";
import Link from "next/link";

export default function LobbyPage() {
  return (
    <div className="min-h-screen bg-cream">
      {/* Top bar */}
      <header className="bg-white border-b border-ceramic">
        <div className="max-w-5xl mx-auto px-4 h-16 flex items-center justify-between">
          <div className="flex items-center gap-2">
            <span className="text-2xl">🎴</span>
            <span className="text-lg font-bold text-starbucks tracking-tight">PokeFace</span>
          </div>
          <Link href="/auth/login">
            <Button variant="dark-outlined" className="text-sm">
              Sign Out
            </Button>
          </Link>
        </div>
      </header>

      {/* Main content */}
      <main className="max-w-5xl mx-auto py-6 lg:py-10 px-4">
        <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between mb-6 lg:mb-8 gap-4">
          <div>
            <h1 className="text-2xl lg:text-3xl font-bold text-starbucks tracking-tight">Game Lobby</h1>
            <p className="text-sm text-text-black-soft mt-1">Choose a room or create your own</p>
          </div>
          <Link href="/room/create" className="w-full sm:w-auto">
            <Button variant="primary" className="w-full sm:w-auto justify-center">+ Create Room</Button>
          </Link>
        </div>

        <Card padding="lg">
          <div className="text-center py-12">
            <div className="text-5xl mb-4 opacity-40">🃏</div>
            <p className="text-lg font-semibold text-text-black-soft">No active rooms yet</p>
            <p className="text-sm text-text-black-soft mt-1">
              Create a room to start playing!
            </p>
          </div>
        </Card>
      </main>
    </div>
  );
}
