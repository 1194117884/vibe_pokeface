import { Card } from "@/components/ui/Card";

export default function DashboardPage() {
  return (
    <div>
      <h1 className="text-2xl font-semibold text-starbucks mb-6">Dashboard</h1>
      <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
        <Card>
          <p className="text-sm text-text-black-soft">Online Players</p>
          <p className="text-3xl font-semibold mt-1">--</p>
        </Card>
        <Card>
          <p className="text-sm text-text-black-soft">Active Rooms</p>
          <p className="text-3xl font-semibold mt-1">--</p>
        </Card>
        <Card>
          <p className="text-sm text-text-black-soft">Total Users</p>
          <p className="text-3xl font-semibold mt-1">--</p>
        </Card>
      </div>
    </div>
  );
}
