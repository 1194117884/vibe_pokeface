"use client";

import { useState, useEffect } from "react";
import { Card } from "@/components/ui/Card";
import { Button } from "@/components/ui/Button";
import { adminFetch } from "@/lib/admin-fetch";

interface LLMConfig {
  id: number;
  provider: string;
  model: string;
  api_key?: string;
  temperature?: number;
  max_tokens?: number;
  is_active: boolean;
}

export default function LLMConfigPage() {
  const [configs, setConfigs] = useState<LLMConfig[]>([]);
  const [loading, setLoading] = useState(true);
  const [showForm, setShowForm] = useState(false);
  const [form, setForm] = useState({ provider: "openai", model: "", api_key: "", temperature: 0.7, max_tokens: 2048, is_active: true });

  const fetchConfigs = async () => {
    try {
      const res = await adminFetch("/api/admin/llm-configs");
      const data = await res.json();
      setConfigs(Array.isArray(data) ? data : []);
    } catch { /* ignore */ }
    setLoading(false);
  };

  useEffect(() => { fetchConfigs(); }, []);

  const createConfig = async () => {
    if (!form.model || !form.api_key) return;
    try {
      await adminFetch("/api/admin/llm-configs", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(form),
      });
      setShowForm(false);
      setForm({ provider: "openai", model: "", api_key: "", temperature: 0.7, max_tokens: 2048, is_active: true });
      fetchConfigs();
    } catch { /* ignore */ }
  };

  const deleteConfig = async (id: number) => {
    if (!confirm("Delete this config?")) return;
    try {
      await adminFetch(`/api/admin/llm-configs/${id}`, { method: "DELETE" });
      fetchConfigs();
    } catch { /* ignore */ }
  };

  return (
    <div>
      <div className="flex items-center justify-between mb-6">
        <div>
          <h1 className="text-2xl font-bold text-starbucks tracking-tight">LLM Configuration</h1>
          <p className="text-sm text-text-black-soft mt-0.5">Manage AI provider connections</p>
        </div>
        <Button
          variant={showForm ? "dark-outlined" : "primary"}
          onClick={() => setShowForm(!showForm)}
        >
          {showForm ? "Cancel" : "+ Add Config"}
        </Button>
      </div>

      {showForm && (
        <Card padding="md" className="mb-6">
          <div className="flex items-end gap-4 flex-wrap">
            <div className="min-w-[140px] flex-1">
              <label className="block text-xs font-semibold uppercase tracking-wide text-text-black-soft mb-1.5">Provider</label>
              <select
                className="w-full px-3 py-2 border border-gray-300 rounded-[4px] text-sm text-text-black outline-none transition-all duration-200 focus:border-green-accent bg-white"
                value={form.provider}
                onChange={(e) => setForm({ ...form, provider: e.target.value })}
              >
                <option value="openai">OpenAI</option>
                <option value="anthropic">Anthropic</option>
                <option value="custom">Custom</option>
              </select>
            </div>
            <div className="min-w-[140px] flex-1">
              <label className="block text-xs font-semibold uppercase tracking-wide text-text-black-soft mb-1.5">Model</label>
              <input
                className="w-full px-3 py-2 border border-gray-300 rounded-[4px] text-sm text-text-black outline-none transition-all duration-200 focus:border-green-accent"
                value={form.model}
                onChange={(e) => setForm({ ...form, model: e.target.value })}
                placeholder="gpt-4o"
              />
            </div>
            <div className="min-w-[180px] flex-1">
              <label className="block text-xs font-semibold uppercase tracking-wide text-text-black-soft mb-1.5">API Key</label>
              <input
                type="password"
                className="w-full px-3 py-2 border border-gray-300 rounded-[4px] text-sm text-text-black outline-none transition-all duration-200 focus:border-green-accent"
                value={form.api_key}
                onChange={(e) => setForm({ ...form, api_key: e.target.value })}
                placeholder="sk-..."
              />
            </div>
            <Button variant="primary" onClick={createConfig}>
              Create
            </Button>
          </div>
        </Card>
      )}

      {loading ? (
        <Card padding="lg">
          <p className="text-text-black-soft text-center py-4">Loading...</p>
        </Card>
      ) : configs.length === 0 ? (
        <Card padding="lg">
          <div className="text-center py-4">
            <p className="text-text-black-soft">No LLM configurations yet.</p>
            <p className="text-sm text-text-black-soft mt-1">Add an OpenAI or Anthropic API configuration for AI bot integration.</p>
          </div>
        </Card>
      ) : (
        <Card padding="md" className="overflow-hidden">
          <div className="overflow-x-auto">
            <table className="w-full min-w-[500px]">
            <thead>
              <tr className="border-b border-cream">
                <th className="text-left p-4 text-sm font-semibold text-text-black tracking-tight">Provider</th>
                <th className="text-left p-4 text-sm font-semibold text-text-black tracking-tight">Model</th>
                <th className="text-left p-4 text-sm font-semibold text-text-black tracking-tight">API Key</th>
                <th className="text-left p-4 text-sm font-semibold text-text-black tracking-tight">Status</th>
                <th className="text-right p-4 text-sm font-semibold text-text-black tracking-tight">Actions</th>
              </tr>
            </thead>
            <tbody>
              {configs.map((c) => (
                <tr key={c.id} className="border-b border-cream last:border-b-0 hover:bg-cream/50 transition-colors">
                  <td className="p-4 text-sm text-text-black capitalize">{c.provider}</td>
                  <td className="p-4 text-sm text-text-black font-mono">{c.model}</td>
                  <td className="p-4 text-sm text-text-black-soft">
                    {c.api_key ? `${c.api_key.slice(0, 8)}...` : "---"}
                  </td>
                  <td className="p-4 text-sm">
                    <span
                      className={`inline-block px-3 py-1 rounded-pill text-xs font-semibold tracking-tight ${
                        c.is_active
                          ? "bg-green-light text-starbucks"
                          : "bg-ceramic text-text-black-soft"
                      }`}
                    >
                      {c.is_active ? "Active" : "Inactive"}
                    </span>
                  </td>
                  <td className="p-4 text-right">
                    <button
                      className="text-sm font-semibold text-red-error hover:underline transition-colors"
                      onClick={() => deleteConfig(c.id)}
                    >
                      Delete
                    </button>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
          </div>
        </Card>
      )}
    </div>
  );
}
