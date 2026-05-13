"use client";

import { useState, useEffect } from "react";
import { Card } from "@/components/ui/Card";
import { Button } from "@/components/ui/Button";
import { adminFetch } from "@/lib/admin-fetch";

interface AICharacter {
  id: number;
  name: string;
  play_style: string;
  enabled: boolean;
}

export default function AICharactersPage() {
  const [characters, setCharacters] = useState<AICharacter[]>([]);
  const [loading, setLoading] = useState(true);
  const [showForm, setShowForm] = useState(false);
  const [form, setForm] = useState({ name: "", play_style: "balanced", enabled: true });

  const fetchChars = async () => {
    try {
      const res = await adminFetch("/api/admin/ai-characters");
      const data = await res.json();
      setCharacters(Array.isArray(data) ? data : []);
    } catch { /* ignore */ }
    setLoading(false);
  };

  useEffect(() => {
    // eslint-disable-next-line react-hooks/set-state-in-effect
    fetchChars();
  }, []);

  const createChar = async () => {
    if (!form.name) return;
    try {
      await adminFetch("/api/admin/ai-characters", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(form),
      });
      setShowForm(false);
      setForm({ name: "", play_style: "balanced", enabled: true });
      fetchChars();
    } catch { /* ignore */ }
  };

  const deleteChar = async (id: number) => {
    if (!confirm("Delete this character?")) return;
    try {
      await adminFetch(`/api/admin/ai-characters/${id}`, { method: "DELETE" });
      fetchChars();
    } catch { /* ignore */ }
  };

  const toggleEnabled = async (char: AICharacter) => {
    try {
      await adminFetch(`/api/admin/ai-characters/${char.id}`, {
        method: "PUT",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ ...char, enabled: !char.enabled }),
      });
      fetchChars();
    } catch { /* ignore */ }
  };

  return (
    <div>
      <div className="flex items-center justify-between mb-6">
        <div>
          <h1 className="text-2xl font-bold text-starbucks tracking-tight">AI Characters</h1>
          <p className="text-sm text-text-black-soft mt-0.5">Manage bot personalities</p>
        </div>
        <Button
          variant={showForm ? "dark-outlined" : "primary"}
          onClick={() => setShowForm(!showForm)}
        >
          {showForm ? "Cancel" : "+ Add Character"}
        </Button>
      </div>

      {showForm && (
        <Card padding="md" className="mb-6">
          <div className="flex items-end gap-4">
            <div className="flex-1">
              <label className="block text-xs font-semibold uppercase tracking-wide text-text-black-soft mb-1.5">Name</label>
              <input
                className="w-full px-3 py-2 border border-gray-300 rounded-[4px] text-sm text-text-black outline-none transition-all duration-200 focus:border-green-accent"
                value={form.name}
                onChange={(e) => setForm({ ...form, name: e.target.value })}
                placeholder="Character name"
              />
            </div>
            <div className="flex-1">
              <label className="block text-xs font-semibold uppercase tracking-wide text-text-black-soft mb-1.5">Play Style</label>
              <select
                className="w-full px-3 py-2 border border-gray-300 rounded-[4px] text-sm text-text-black outline-none transition-all duration-200 focus:border-green-accent bg-white"
                value={form.play_style}
                onChange={(e) => setForm({ ...form, play_style: e.target.value })}
              >
                <option value="aggressive">Aggressive</option>
                <option value="conservative">Conservative</option>
                <option value="balanced">Balanced</option>
                <option value="unpredictable">Unpredictable</option>
              </select>
            </div>
            <Button variant="primary" onClick={createChar}>
              Create
            </Button>
          </div>
        </Card>
      )}

      {loading ? (
        <Card padding="lg">
          <p className="text-text-black-soft text-center py-4">Loading...</p>
        </Card>
      ) : characters.length === 0 ? (
        <Card padding="lg">
          <div className="text-center py-4">
            <p className="text-text-black-soft">No AI characters yet.</p>
            <p className="text-sm text-text-black-soft mt-1">Create characters to populate bot players.</p>
          </div>
        </Card>
      ) : (
        <Card padding="md" className="overflow-hidden">
          <div className="overflow-x-auto">
            <table className="w-full min-w-[500px]">
            <thead>
              <tr className="border-b border-cream">
                <th className="text-left p-4 text-sm font-semibold text-text-black tracking-tight">Name</th>
                <th className="text-left p-4 text-sm font-semibold text-text-black tracking-tight">Play Style</th>
                <th className="text-left p-4 text-sm font-semibold text-text-black tracking-tight">Status</th>
                <th className="text-right p-4 text-sm font-semibold text-text-black tracking-tight">Actions</th>
              </tr>
            </thead>
            <tbody>
              {characters.map((c) => (
                <tr key={c.id} className="border-b border-cream last:border-b-0 hover:bg-cream/50 transition-colors">
                  <td className="p-4 text-sm text-text-black">{c.name}</td>
                  <td className="p-4 text-sm text-text-black capitalize">{c.play_style}</td>
                  <td className="p-4 text-sm">
                    <button onClick={() => toggleEnabled(c)}>
                      <span
                        className={`inline-block px-3 py-1 rounded-pill text-xs font-semibold tracking-tight cursor-pointer transition-colors ${
                          c.enabled
                            ? "bg-green-light text-starbucks"
                            : "bg-ceramic text-text-black-soft"
                        }`}
                      >
                        {c.enabled ? "Active" : "Disabled"}
                      </span>
                    </button>
                  </td>
                  <td className="p-4 text-right">
                    <button
                      className="text-sm font-semibold text-red-error hover:underline transition-colors"
                      onClick={() => deleteChar(c.id)}
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
