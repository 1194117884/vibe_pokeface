"use client";

import { useEffect, useMemo, useState } from "react";
import { Card } from "@/components/ui/Card";
import { Button } from "@/components/ui/Button";
import { adminFetch } from "@/lib/admin-fetch";

interface PromptTemplate {
  id?: number;
  game_type: string;
  phase: string;
  template_type: "system" | "user";
  name: string;
  content: string;
  variables_json?: string;
  status?: string;
  version?: number;
}

interface ToolSchema {
  function: {
    name: string;
    description: string;
    parameters: unknown;
  };
}

interface ToolTemplate {
  id?: number;
  game_type: string;
  phase: string;
  tool_name: string;
  enabled: boolean;
  description?: string;
  parameters_json?: string;
  status?: string;
  version?: number;
}

const phases: Record<string, string[]> = {
  dashengji: ["set_trump", "counter_trump", "take_bottom", "discard_bottom", "playing"],
  doudizhu: ["calling", "snatching", "revealing", "doubling", "playing"],
};

function pretty(value: unknown) {
  return JSON.stringify(value, null, 2);
}

export default function AITemplatesPage() {
  const [gameType, setGameType] = useState("dashengji");
  const [phase, setPhase] = useState("playing");
  const [templates, setTemplates] = useState<PromptTemplate[]>([]);
  const [defaults, setDefaults] = useState<ToolSchema[]>([]);
  const [toolTemplates, setToolTemplates] = useState<ToolTemplate[]>([]);
  const [message, setMessage] = useState("");
  const [preview, setPreview] = useState("");
  const [form, setForm] = useState<PromptTemplate>({
    game_type: "dashengji",
    phase: "playing",
    template_type: "system",
    name: "Dashengji system prompt",
    content: "${default_content}",
    status: "draft",
  });
  const [varsText, setVarsText] = useState('{\n  "default_content": "原始提示词",\n  "character_name": "小李",\n  "phase": "playing"\n}');

  const currentPhases = useMemo(() => phases[gameType] || ["playing"], [gameType]);

  const load = async () => {
    const promptRes = await adminFetch(`/api/admin/ai-templates?game_type=${gameType}&phase=${phase}`);
    const promptData = await promptRes.json();
    setTemplates(Array.isArray(promptData) ? promptData : []);

    const toolRes = await adminFetch(`/api/admin/ai-tool-templates?game_type=${gameType}&phase=${phase}`);
    const toolData = await toolRes.json();
    setDefaults(Array.isArray(toolData.defaults) ? toolData.defaults : []);
    setToolTemplates(Array.isArray(toolData.templates) ? toolData.templates : []);
  };

  useEffect(() => {
    if (!currentPhases.includes(phase)) {
      setPhase(currentPhases[0]);
      return;
    }
    setForm((prev) => ({ ...prev, game_type: gameType, phase }));
    load().catch(() => setMessage("Failed to load templates"));
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [gameType, phase, currentPhases]);

  const savePrompt = async () => {
    const url = form.id ? `/api/admin/ai-templates/${form.id}` : "/api/admin/ai-templates";
    await adminFetch(url, {
      method: form.id ? "PUT" : "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ ...form, game_type: gameType, phase, status: "draft" }),
    });
    setMessage("Draft saved");
    load();
  };

  const publishPrompt = async (id?: number) => {
    if (!id) return;
    await adminFetch(`/api/admin/ai-templates/${id}/publish`, { method: "POST" });
    setMessage("Template published");
    load();
  };

  const previewPrompt = async () => {
    let vars: Record<string, string> = {};
    try {
      vars = JSON.parse(varsText);
    } catch {
      setMessage("Preview variables must be valid JSON");
      return;
    }
    const res = await adminFetch("/api/admin/ai-templates/preview", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ content: form.content, vars }),
    });
    const data = await res.json();
    setPreview(data.rendered || "");
  };

  const editTemplate = (tmpl: PromptTemplate) => {
    setForm({
      id: tmpl.id,
      game_type: tmpl.game_type,
      phase: tmpl.phase,
      template_type: tmpl.template_type,
      name: tmpl.name,
      content: tmpl.content,
      status: tmpl.status || "draft",
    });
    setPreview("");
  };

  const saveTool = async (schema: ToolSchema, existing?: ToolTemplate) => {
    const description = window.prompt("Tool description", existing?.description || schema.function.description);
    if (description == null) return;
    const enabled = window.confirm("Enable this tool for the selected phase?");
    const payload: ToolTemplate = {
      id: existing?.id,
      game_type: gameType,
      phase,
      tool_name: schema.function.name,
      enabled,
      description,
      parameters_json: existing?.parameters_json || pretty(schema.function.parameters),
      status: "draft",
    };
    await adminFetch(existing?.id ? `/api/admin/ai-tool-templates/${existing.id}` : "/api/admin/ai-tool-templates", {
      method: existing?.id ? "PUT" : "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(payload),
    });
    setMessage("Tool draft saved");
    load();
  };

  const publishTool = async (id?: number) => {
    if (!id) return;
    await adminFetch(`/api/admin/ai-tool-templates/${id}/publish`, { method: "POST" });
    setMessage("Tool template published");
    load();
  };

  return (
    <div>
      <div className="mb-6">
        <h1 className="text-2xl font-bold text-starbucks tracking-tight">AI Templates</h1>
        <p className="text-sm text-text-black-soft mt-0.5">Edit prompt templates, preview rendering, and publish controlled tool descriptions.</p>
      </div>

      {message && <div className="mb-4 rounded-[4px] border border-green-accent/30 bg-green-light px-4 py-3 text-sm text-starbucks">{message}</div>}

      <Card padding="md" className="mb-5">
        <div className="grid grid-cols-1 gap-3 md:grid-cols-4">
          <select className="rounded-[4px] border border-gray-300 px-3 py-2 text-sm" value={gameType} onChange={(e) => setGameType(e.target.value)}>
            <option value="dashengji">Dashengji</option>
            <option value="doudizhu">Dou Dizhu</option>
          </select>
          <select className="rounded-[4px] border border-gray-300 px-3 py-2 text-sm" value={phase} onChange={(e) => setPhase(e.target.value)}>
            {currentPhases.map((p) => <option key={p} value={p}>{p}</option>)}
          </select>
          <select className="rounded-[4px] border border-gray-300 px-3 py-2 text-sm" value={form.template_type} onChange={(e) => setForm({ ...form, template_type: e.target.value as "system" | "user" })}>
            <option value="system">System Prompt</option>
            <option value="user">User Context</option>
          </select>
          <Button className="min-h-10 py-2 text-sm" variant="dark-outlined" onClick={load}>Refresh</Button>
        </div>
      </Card>

      <div className="grid grid-cols-1 gap-5 xl:grid-cols-[0.8fr_1.2fr]">
        <div className="space-y-5">
          <Card padding="md">
            <h2 className="mb-3 text-lg font-bold text-text-black">Prompt Versions</h2>
            <div className="space-y-2">
              {templates.map((tmpl) => (
                <button key={tmpl.id} onClick={() => editTemplate(tmpl)} className="w-full rounded-[4px] border border-cream p-3 text-left hover:bg-cream/50">
                  <div className="flex items-center justify-between gap-2">
                    <p className="font-semibold text-text-black">{tmpl.name}</p>
                    <span className={`rounded-[4px] px-2 py-1 text-xs ${tmpl.status === "published" ? "bg-green-light text-starbucks" : "bg-ceramic text-text-black-soft"}`}>
                      {tmpl.status} v{tmpl.version}
                    </span>
                  </div>
                  <p className="mt-1 text-xs text-text-black-soft">{tmpl.template_type} · {tmpl.game_type}/{tmpl.phase}</p>
                </button>
              ))}
              {templates.length === 0 && <p className="text-sm text-text-black-soft">No templates yet. Save a draft to start.</p>}
            </div>
          </Card>

          <Card padding="md">
            <h2 className="mb-3 text-lg font-bold text-text-black">Tool Templates</h2>
            <div className="space-y-3">
              {defaults.map((schema) => {
                const existing = toolTemplates.find((t) => t.tool_name === schema.function.name);
                return (
                  <div key={schema.function.name} className="rounded-[4px] border border-cream p-3">
                    <div className="flex flex-wrap items-center justify-between gap-2">
                      <div>
                        <p className="font-semibold text-text-black">{schema.function.name}</p>
                        <p className="text-xs text-text-black-soft">{existing?.description || schema.function.description}</p>
                      </div>
                      <div className="flex gap-2">
                        <button className="text-sm font-semibold text-green-accent" onClick={() => saveTool(schema, existing)}>Edit</button>
                        {existing?.id && <button className="text-sm font-semibold text-starbucks" onClick={() => publishTool(existing.id)}>Publish</button>}
                      </div>
                    </div>
                    {existing && <p className="mt-2 text-xs text-text-black-soft">Draft/status: {existing.status} · enabled: {String(existing.enabled)}</p>}
                  </div>
                );
              })}
            </div>
          </Card>
        </div>

        <div className="space-y-5">
          <Card padding="md">
            <div className="mb-3 flex flex-wrap items-center justify-between gap-3">
              <h2 className="text-lg font-bold text-text-black">Editor</h2>
              <div className="flex gap-2">
                <Button className="min-h-10 py-2 text-sm" variant="dark-outlined" onClick={() => setForm({ game_type: gameType, phase, template_type: "system", name: "New template", content: "${default_content}", status: "draft" })}>New</Button>
                <Button className="min-h-10 py-2 text-sm" variant="outlined" onClick={previewPrompt}>Preview</Button>
                <Button className="min-h-10 py-2 text-sm" onClick={savePrompt}>Save Draft</Button>
                {form.id && <Button className="min-h-10 py-2 text-sm" variant="black-fill" onClick={() => publishPrompt(form.id)}>Publish</Button>}
              </div>
            </div>
            <label className="mb-1 block text-xs font-semibold uppercase text-text-black-soft">Name</label>
            <input className="mb-3 w-full rounded-[4px] border border-gray-300 px-3 py-2 text-sm" value={form.name} onChange={(e) => setForm({ ...form, name: e.target.value })} />
            <label className="mb-1 block text-xs font-semibold uppercase text-text-black-soft">Template Content</label>
            <textarea className="min-h-[360px] w-full rounded-[4px] border border-gray-300 p-3 font-mono text-sm" value={form.content} onChange={(e) => setForm({ ...form, content: e.target.value })} />
          </Card>

          <Card padding="md">
            <h3 className="mb-3 font-semibold text-text-black">Preview Variables</h3>
            <textarea className="min-h-[140px] w-full rounded-[4px] border border-gray-300 p-3 font-mono text-sm" value={varsText} onChange={(e) => setVarsText(e.target.value)} />
          </Card>

          <Card padding="md">
            <h3 className="mb-3 font-semibold text-text-black">Rendered Preview</h3>
            <pre className="max-h-[520px] overflow-auto rounded-[4px] bg-black/90 p-4 text-xs leading-relaxed text-white">{preview || "Click Preview to render this draft without publishing."}</pre>
          </Card>
        </div>
      </div>
    </div>
  );
}
