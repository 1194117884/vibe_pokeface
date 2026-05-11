"use client";

import { useEffect, useState } from "react";
import { Card } from "@/components/ui/Card";

interface LLMStats {
  total_calls: number;
  total_tokens: number;
  avg_latency_ms: number;
  success_rate: number;
}

export default function LLMStatsPage() {
  const [stats, setStats] = useState<LLMStats | null>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    const token = localStorage.getItem("token");
    if (!token) return;
    fetch("/api/admin/llm-stats", {
      headers: { Authorization: `Bearer ${token}` },
    })
      .then((r) => r.json())
      .then((d: LLMStats) => { setStats(d); setLoading(false); })
      .catch(() => setLoading(false));
  }, []);

  return (
    <div>
      <h1 className="text-2xl font-semibold text-starbucks mb-6">LLM Call Statistics</h1>
      {loading ? (
        <p className="text-gray-500 py-4 text-center">Loading...</p>
      ) : !stats ? (
        <p className="text-gray-500 py-4 text-center">Unable to load stats.</p>
      ) : (
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
          <Card>
            <p className="text-sm text-text-black-soft">Total Calls</p>
            <p className="text-3xl font-semibold mt-1">{stats.total_calls}</p>
          </Card>
          <Card>
            <p className="text-sm text-text-black-soft">Total Tokens</p>
            <p className="text-3xl font-semibold mt-1">{stats.total_tokens.toLocaleString()}</p>
          </Card>
          <Card>
            <p className="text-sm text-text-black-soft">Avg Latency</p>
            <p className="text-3xl font-semibold mt-1">{stats.avg_latency_ms}ms</p>
          </Card>
          <Card>
            <p className="text-sm text-text-black-soft">Success Rate</p>
            <p className="text-3xl font-semibold mt-1">{stats.success_rate}%</p>
          </Card>
        </div>
      )}
    </div>
  );
}
