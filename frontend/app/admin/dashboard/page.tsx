"use client";

import { useEffect, useState } from "react";
import { Card } from "@/components/ui/Card";

interface DashboardData {
  online_players: number;
  active_rooms: number;
  total_users: number;
}

export default function DashboardPage() {
  const [data, setData] = useState<DashboardData | null>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    const token = localStorage.getItem("token");
    if (!token) return;
    fetch("/api/admin/dashboard", {
      headers: { Authorization: `Bearer ${token}` },
    })
      .then((r) => r.json())
      .then((d: DashboardData) => { setData(d); setLoading(false); })
      .catch(() => setLoading(false));
  }, []);

  return (
    <div>
      <h1 className="text-2xl font-semibold text-starbucks mb-6">Dashboard</h1>
      <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
        <Card>
          <p className="text-sm text-text-black-soft">Online Players</p>
          <p className="text-3xl font-semibold mt-1">{loading ? "--" : data?.online_players ?? 0}</p>
        </Card>
        <Card>
          <p className="text-sm text-text-black-soft">Active Rooms</p>
          <p className="text-3xl font-semibold mt-1">{loading ? "--" : data?.active_rooms ?? 0}</p>
        </Card>
        <Card>
          <p className="text-sm text-text-black-soft">Total Users</p>
          <p className="text-3xl font-semibold mt-1">{loading ? "--" : data?.total_users ?? 0}</p>
        </Card>
      </div>
    </div>
  );
}
