"use client";

import { useState, useEffect } from "react";

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
    const token = localStorage.getItem("token");
    if (!token) return;
    try {
      const res = await fetch("/api/admin/ai-characters", {
        headers: { Authorization: `Bearer ${token}` },
      });
      const data = await res.json();
      setCharacters(Array.isArray(data) ? data : []);
    } catch { /* ignore */ }
    setLoading(false);
  };

  useEffect(() => { fetchChars(); }, []);

  const createChar = async () => {
    if (!form.name) return;
    const token = localStorage.getItem("token");
    if (!token) return;
    await fetch("/api/admin/ai-characters", {
      method: "POST",
      headers: { "Content-Type": "application/json", Authorization: `Bearer ${token}` },
      body: JSON.stringify(form),
    });
    setShowForm(false);
    setForm({ name: "", play_style: "balanced", enabled: true });
    fetchChars();
  };

  const deleteChar = async (id: number) => {
    const token = localStorage.getItem("token");
    if (!token || !confirm("Delete this character?")) return;
    await fetch(`/api/admin/ai-characters/${id}`, {
      method: "DELETE",
      headers: { Authorization: `Bearer ${token}` },
    });
    fetchChars();
  };

  const toggleEnabled = async (char: AICharacter) => {
    const token = localStorage.getItem("token");
    if (!token) return;
    await fetch(`/api/admin/ai-characters/${char.id}`, {
      method: "PUT",
      headers: { "Content-Type": "application/json", Authorization: `Bearer ${token}` },
      body: JSON.stringify({ ...char, enabled: !char.enabled }),
    });
    fetchChars();
  };

  return (
    <div>
      <div className="flex justify-between items-center mb-6">
        <h1 className="text-2xl font-semibold text-starbucks">AI Characters</h1>
        <button
          className="px-4 py-2 bg-green-600 text-white rounded-full text-sm font-semibold hover:bg-green-700"
          onClick={() => setShowForm(!showForm)}
        >
          {showForm ? "Cancel" : "+ Add Character"}
        </button>
      </div>

      {showForm && (
        <div className="bg-white rounded-xl shadow p-4 mb-6 flex gap-3 items-end">
          <div>
            <label className="block text-xs text-gray-500 mb-1">Name</label>
            <input
              className="px-3 py-2 border border-gray-300 rounded-lg text-sm"
              value={form.name}
              onChange={(e) => setForm({ ...form, name: e.target.value })}
              placeholder="Character name"
            />
          </div>
          <div>
            <label className="block text-xs text-gray-500 mb-1">Play Style</label>
            <select
              className="px-3 py-2 border border-gray-300 rounded-lg text-sm"
              value={form.play_style}
              onChange={(e) => setForm({ ...form, play_style: e.target.value })}
            >
              <option value="aggressive">Aggressive</option>
              <option value="conservative">Conservative</option>
              <option value="balanced">Balanced</option>
              <option value="unpredictable">Unpredictable</option>
            </select>
          </div>
          <button
            className="px-4 py-2 bg-blue-600 text-white rounded-lg text-sm hover:bg-blue-700"
            onClick={createChar}
          >
            Create
          </button>
        </div>
      )}

      {loading ? (
        <p className="text-gray-500 py-4 text-center">Loading...</p>
      ) : characters.length === 0 ? (
        <div className="bg-white rounded-xl shadow p-8 text-center">
          <p className="text-gray-500">No AI characters yet.</p>
          <p className="text-sm text-gray-400 mt-1">Create characters to populate bot players.</p>
        </div>
      ) : (
        <div className="bg-white rounded-xl shadow overflow-hidden">
          <table className="w-full">
            <thead>
              <tr className="border-b border-gray-200">
                <th className="text-left p-4 text-sm font-semibold text-gray-700">Name</th>
                <th className="text-left p-4 text-sm font-semibold text-gray-700">Play Style</th>
                <th className="text-left p-4 text-sm font-semibold text-gray-700">Status</th>
                <th className="text-right p-4 text-sm font-semibold text-gray-700">Actions</th>
              </tr>
            </thead>
            <tbody>
              {characters.map((c) => (
                <tr key={c.id} className="border-b border-gray-100">
                  <td className="p-4 text-sm">{c.name}</td>
                  <td className="p-4 text-sm capitalize">{c.play_style}</td>
                  <td className="p-4 text-sm">
                    <button onClick={() => toggleEnabled(c)}>
                      <span className={`px-2 py-1 rounded-full text-xs cursor-pointer ${c.enabled ? "bg-green-100 text-green-700" : "bg-gray-100 text-gray-500"}`}>
                        {c.enabled ? "Active" : "Disabled"}
                      </span>
                    </button>
                  </td>
                  <td className="p-4 text-right">
                    <button className="text-red-600 text-sm hover:underline" onClick={() => deleteChar(c.id)}>
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
