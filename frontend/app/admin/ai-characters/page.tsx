"use client";

import { useState, useEffect } from "react";
import { Card } from "@/components/ui/Card";
import { Button } from "@/components/ui/Button";
import { adminFetch } from "@/lib/admin-fetch";

interface AICharacter {
  id: number;
  name: string;
  avatar_url?: string;
  personality?: string;
  play_style: string;
  catchphrase?: string;
  occupation?: string;
  voice?: string;
  greeting?: string;
  enabled: boolean;
}

const PLAY_STYLES = ["aggressive", "conservative", "balanced", "unpredictable"];

const playStyleLabels: Record<string, string> = {
  aggressive: "Aggressive",
  conservative: "Conservative",
  balanced: "Balanced",
  unpredictable: "Unpredictable",
};

const emptyForm = {
  name: "",
  play_style: "balanced",
  personality: "",
  catchphrase: "",
  occupation: "",
  greeting: "",
  enabled: true,
};

export default function AICharactersPage() {
  const [characters, setCharacters] = useState<AICharacter[]>([]);
  const [loading, setLoading] = useState(true);
  const [showForm, setShowForm] = useState(false);
  const [editingId, setEditingId] = useState<number | null>(null);
  const [form, setForm] = useState(emptyForm);
  const [saving, setSaving] = useState(false);

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

  const openCreate = () => {
    setEditingId(null);
    setForm(emptyForm);
    setShowForm(true);
  };

  const openEdit = (char: AICharacter) => {
    setEditingId(char.id);
    setForm({
      name: char.name,
      play_style: char.play_style,
      personality: char.personality || "",
      catchphrase: char.catchphrase || "",
      occupation: char.occupation || "",
      greeting: char.greeting || "",
      enabled: char.enabled,
    });
    setShowForm(true);
  };

  const saveChar = async () => {
    if (!form.name) return;
    setSaving(true);
    try {
      const url = editingId
        ? `/api/admin/ai-characters/${editingId}`
        : "/api/admin/ai-characters";
      const method = editingId ? "PUT" : "POST";
      const body = editingId
        ? { ...form, id: editingId }
        : form;

      await adminFetch(url, {
        method,
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(body),
      });
      setShowForm(false);
      setEditingId(null);
      setForm(emptyForm);
      fetchChars();
    } catch { /* ignore */ }
    setSaving(false);
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

  const updateField = (field: string, value: string) => {
    setForm({ ...form, [field]: value });
  };

  return (
    <div>
      <div className="flex items-center justify-between mb-6">
        <div>
          <h1 className="text-2xl font-bold text-starbucks tracking-tight">AI Characters</h1>
          <p className="text-sm text-text-black-soft mt-0.5">Manage bot personalities and traits</p>
        </div>
        <Button
          variant={showForm ? "dark-outlined" : "primary"}
          onClick={() => showForm ? setShowForm(false) : openCreate()}
        >
          {showForm ? "Cancel" : "+ Add Character"}
        </Button>
      </div>

      {showForm && (
        <Card padding="md" className="mb-6">
          <h3 className="text-sm font-semibold text-text-black mb-4">
            {editingId ? "Edit Character" : "New Character"}
          </h3>
          <div className="space-y-4">
            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
              <div>
                <label className="block text-xs font-semibold uppercase tracking-wide text-text-black-soft mb-1.5">Name *</label>
                <input
                  className="w-full px-3 py-2 border border-gray-300 rounded-[4px] text-sm text-text-black outline-none transition-all duration-200 focus:border-green-accent"
                  value={form.name}
                  onChange={(e) => updateField("name", e.target.value)}
                  placeholder="e.g. 老王"
                />
              </div>
              <div>
                <label className="block text-xs font-semibold uppercase tracking-wide text-text-black-soft mb-1.5">Play Style</label>
                <select
                  className="w-full px-3 py-2 border border-gray-300 rounded-[4px] text-sm text-text-black outline-none transition-all duration-200 focus:border-green-accent bg-white"
                  value={form.play_style}
                  onChange={(e) => updateField("play_style", e.target.value)}
                >
                  {PLAY_STYLES.map((s) => (
                    <option key={s} value={s}>{playStyleLabels[s]}</option>
                  ))}
                </select>
              </div>
            </div>
            <div>
              <label className="block text-xs font-semibold uppercase tracking-wide text-text-black-soft mb-1.5">
                Personality
                <span className="font-normal normal-case ml-1 text-text-black-soft/60">
                  — injected into LLM system prompt
                </span>
              </label>
              <input
                className="w-full px-3 py-2 border border-gray-300 rounded-[4px] text-sm text-text-black outline-none transition-all duration-200 focus:border-green-accent"
                value={form.personality}
                onChange={(e) => updateField("personality", e.target.value)}
                placeholder="e.g. 冷静分析、喜欢冒险、稳健保守"
              />
            </div>
            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
              <div>
                <label className="block text-xs font-semibold uppercase tracking-wide text-text-black-soft mb-1.5">Catchphrase</label>
                <input
                  className="w-full px-3 py-2 border border-gray-300 rounded-[4px] text-sm text-text-black outline-none transition-all duration-200 focus:border-green-accent"
                  value={form.catchphrase}
                  onChange={(e) => updateField("catchphrase", e.target.value)}
                  placeholder="e.g. 这把稳了！"
                />
              </div>
              <div>
                <label className="block text-xs font-semibold uppercase tracking-wide text-text-black-soft mb-1.5">Occupation</label>
                <input
                  className="w-full px-3 py-2 border border-gray-300 rounded-[4px] text-sm text-text-black outline-none transition-all duration-200 focus:border-green-accent"
                  value={form.occupation}
                  onChange={(e) => updateField("occupation", e.target.value)}
                  placeholder="e.g. 退休工程师"
                />
              </div>
            </div>
            <div>
              <label className="block text-xs font-semibold uppercase tracking-wide text-text-black-soft mb-1.5">
                Greeting
                <span className="font-normal normal-case ml-1 text-text-black-soft/60">
                  — shown when bot joins room
                </span>
              </label>
              <input
                className="w-full px-3 py-2 border border-gray-300 rounded-[4px] text-sm text-text-black outline-none transition-all duration-200 focus:border-green-accent"
                value={form.greeting}
                onChange={(e) => updateField("greeting", e.target.value)}
                placeholder="e.g. 大家好！今天运气不错~"
              />
            </div>
            <div className="flex gap-3 pt-2">
              <Button variant="primary" onClick={saveChar} disabled={saving}>
                {saving ? "Saving..." : editingId ? "Save Changes" : "Create"}
              </Button>
              <Button variant="outlined" onClick={() => setShowForm(false)}>
                Cancel
              </Button>
            </div>
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
            <table className="w-full min-w-[600px]">
            <thead>
              <tr className="border-b border-cream">
                <th className="text-left p-4 text-sm font-semibold text-text-black tracking-tight">Name</th>
                <th className="text-left p-4 text-sm font-semibold text-text-black tracking-tight">Personality</th>
                <th className="text-left p-4 text-sm font-semibold text-text-black tracking-tight">Play Style</th>
                <th className="text-left p-4 text-sm font-semibold text-text-black tracking-tight">Status</th>
                <th className="text-right p-4 text-sm font-semibold text-text-black tracking-tight">Actions</th>
              </tr>
            </thead>
            <tbody>
              {characters.map((c) => (
                <tr key={c.id} className="border-b border-cream last:border-b-0 hover:bg-cream/50 transition-colors">
                  <td className="p-4 text-sm text-text-black">
                    <button
                      className="font-medium hover:text-green-accent transition-colors text-left"
                      onClick={() => openEdit(c)}
                    >
                      {c.name}
                    </button>
                    {c.occupation && (
                      <div className="text-xs text-text-black-soft mt-0.5">{c.occupation}</div>
                    )}
                  </td>
                  <td className="p-4 text-sm text-text-black-soft">
                    {c.personality ? (
                      <span className="text-xs">{c.personality}</span>
                    ) : (
                      <span className="text-xs text-text-black-soft/40 italic">—</span>
                    )}
                  </td>
                  <td className="p-4 text-sm text-text-black">
                    <span className="text-xs px-2 py-0.5 rounded-full bg-cream text-text-black-soft">
                      {playStyleLabels[c.play_style] || c.play_style}
                    </span>
                  </td>
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
                    <div className="flex items-center justify-end gap-3">
                      <button
                        className="text-sm font-medium text-text-black-soft hover:text-green-accent transition-colors"
                        onClick={() => openEdit(c)}
                      >
                        Edit
                      </button>
                      <button
                        className="text-sm font-semibold text-red-error hover:underline transition-colors"
                        onClick={() => deleteChar(c.id)}
                      >
                        Delete
                      </button>
                    </div>
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
