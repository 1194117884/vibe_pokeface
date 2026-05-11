"use client";

import { useState, useEffect } from "react";

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
    const token = localStorage.getItem("token");
    if (!token) return;
    try {
      const res = await fetch("/api/admin/llm-configs", {
        headers: { Authorization: `Bearer ${token}` },
      });
      const data = await res.json();
      setConfigs(Array.isArray(data) ? data : []);
    } catch { /* ignore */ }
    setLoading(false);
  };

  useEffect(() => { fetchConfigs(); }, []);

  const createConfig = async () => {
    if (!form.model || !form.api_key) return;
    const token = localStorage.getItem("token");
    if (!token) return;
    await fetch("/api/admin/llm-configs", {
      method: "POST",
      headers: { "Content-Type": "application/json", Authorization: `Bearer ${token}` },
      body: JSON.stringify(form),
    });
    setShowForm(false);
    setForm({ provider: "openai", model: "", api_key: "", temperature: 0.7, max_tokens: 2048, is_active: true });
    fetchConfigs();
  };

  const deleteConfig = async (id: number) => {
    const token = localStorage.getItem("token");
    if (!token || !confirm("Delete this config?")) return;
    await fetch(`/api/admin/llm-configs/${id}`, {
      method: "DELETE",
      headers: { Authorization: `Bearer ${token}` },
    });
    fetchConfigs();
  };

  return (
    <div>
      <div className="flex justify-between items-center mb-6">
        <h1 className="text-2xl font-semibold text-starbucks">LLM Configuration</h1>
        <button
          className="px-4 py-2 bg-green-600 text-white rounded-full text-sm font-semibold hover:bg-green-700"
          onClick={() => setShowForm(!showForm)}
        >
          {showForm ? "Cancel" : "+ Add Config"}
        </button>
      </div>

      {showForm && (
        <div className="bg-white rounded-xl shadow p-4 mb-6 flex gap-3 items-end flex-wrap">
          <div>
            <label className="block text-xs text-gray-500 mb-1">Provider</label>
            <select
              className="px-3 py-2 border border-gray-300 rounded-lg text-sm"
              value={form.provider}
              onChange={(e) => setForm({ ...form, provider: e.target.value })}
            >
              <option value="openai">OpenAI</option>
              <option value="anthropic">Anthropic</option>
              <option value="custom">Custom</option>
            </select>
          </div>
          <div>
            <label className="block text-xs text-gray-500 mb-1">Model</label>
            <input
              className="px-3 py-2 border border-gray-300 rounded-lg text-sm"
              value={form.model}
              onChange={(e) => setForm({ ...form, model: e.target.value })}
              placeholder="gpt-4o"
            />
          </div>
          <div>
            <label className="block text-xs text-gray-500 mb-1">API Key</label>
            <input
              type="password"
              className="px-3 py-2 border border-gray-300 rounded-lg text-sm"
              value={form.api_key}
              onChange={(e) => setForm({ ...form, api_key: e.target.value })}
              placeholder="sk-..."
            />
          </div>
          <button
            className="px-4 py-2 bg-blue-600 text-white rounded-lg text-sm hover:bg-blue-700"
            onClick={createConfig}
          >
            Create
          </button>
        </div>
      )}

      {loading ? (
        <p className="text-gray-500 py-4 text-center">Loading...</p>
      ) : configs.length === 0 ? (
        <div className="bg-white rounded-xl shadow p-8 text-center">
          <p className="text-gray-500">No LLM configurations yet.</p>
          <p className="text-sm text-gray-400 mt-1">Add an OpenAI or Anthropic API configuration for AI bot integration.</p>
        </div>
      ) : (
        <div className="bg-white rounded-xl shadow overflow-hidden">
          <table className="w-full">
            <thead>
              <tr className="border-b border-gray-200">
                <th className="text-left p-4 text-sm font-semibold text-gray-700">Provider</th>
                <th className="text-left p-4 text-sm font-semibold text-gray-700">Model</th>
                <th className="text-left p-4 text-sm font-semibold text-gray-700">API Key</th>
                <th className="text-left p-4 text-sm font-semibold text-gray-700">Status</th>
                <th className="text-right p-4 text-sm font-semibold text-gray-700">Actions</th>
              </tr>
            </thead>
            <tbody>
              {configs.map((c) => (
                <tr key={c.id} className="border-b border-gray-100">
                  <td className="p-4 text-sm capitalize">{c.provider}</td>
                  <td className="p-4 text-sm font-mono">{c.model}</td>
                  <td className="p-4 text-sm text-gray-400">{c.api_key || "---"}</td>
                  <td className="p-4 text-sm">
                    <span className={`px-2 py-1 rounded-full text-xs ${c.is_active ? "bg-green-100 text-green-700" : "bg-gray-100 text-gray-500"}`}>
                      {c.is_active ? "Active" : "Inactive"}
                    </span>
                  </td>
                  <td className="p-4 text-right">
                    <button className="text-red-600 text-sm hover:underline" onClick={() => deleteConfig(c.id)}>
                      Delete
                    </button>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}
    </div>
  );
}
