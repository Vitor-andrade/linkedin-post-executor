import { useCallback, useEffect, useState } from "react";

interface Health {
  status: string;
  aiProvider: string;
}

interface Draft {
  id: number;
  title: string;
  sourceDescription: string;
  content: string;
  status: string;
  createdAt: string;
  updatedAt: string;
}

export default function App() {
  const [health, setHealth] = useState<Health | null>(null);
  const [title, setTitle] = useState("");
  const [description, setDescription] = useState("");
  const [content, setContent] = useState("");
  const [loading, setLoading] = useState(false);
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState("");
  const [drafts, setDrafts] = useState<Draft[]>([]);
  const [editingId, setEditingId] = useState<number | null>(null);

  const loadDrafts = useCallback(async () => {
    try {
      const res = await fetch("/api/drafts");
      if (!res.ok) return;
      setDrafts(await res.json());
    } catch {
      /* listing is best-effort; ignore */
    }
  }, []);

  useEffect(() => {
    fetch("/api/health")
      .then((r) => r.json())
      .then(setHealth)
      .catch(() => setHealth(null));
    void loadDrafts();
  }, [loadDrafts]);

  async function generate(e: React.FormEvent) {
    e.preventDefault();
    setLoading(true);
    setError("");
    try {
      const res = await fetch("/api/generate", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ title, description }),
      });
      const data = await res.json();
      if (!res.ok) throw new Error(data.error ?? "Failed to generate");
      setContent(data.content);
    } catch (err) {
      setError(err instanceof Error ? err.message : String(err));
    } finally {
      setLoading(false);
    }
  }

  async function saveDraft() {
    setSaving(true);
    setError("");
    try {
      const res = await fetch(
        editingId ? `/api/drafts/${editingId}` : "/api/drafts",
        {
          method: editingId ? "PUT" : "POST",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify({
            title,
            sourceDescription: description,
            content,
          }),
        },
      );
      const data = await res.json();
      if (!res.ok) throw new Error(data.error ?? "Failed to save draft");
      setEditingId(data.id);
      await loadDrafts();
    } catch (err) {
      setError(err instanceof Error ? err.message : String(err));
    } finally {
      setSaving(false);
    }
  }

  function loadDraft(d: Draft) {
    setTitle(d.title);
    setDescription(d.sourceDescription);
    setContent(d.content);
    setEditingId(d.id);
    setError("");
  }

  function newDraft() {
    setTitle("");
    setDescription("");
    setContent("");
    setEditingId(null);
    setError("");
  }

  return (
    <main className="container">
      <header>
        <h1>LinkedIn Post Executor</h1>
        <span className={`badge ${health ? "ok" : "down"}`}>
          {health ? `online · AI: ${health.aiProvider}` : "backend offline"}
        </span>
      </header>

      {editingId && (
        <p className="editing">
          Editing draft #{editingId}{" "}
          <button className="link" onClick={newDraft}>
            start a new one
          </button>
        </p>
      )}

      <form onSubmit={generate} className="card">
        <label>
          Title / Topic
          <input
            value={title}
            onChange={(e) => setTitle(e.target.value)}
            placeholder="e.g. 5 lessons I learned scaling an API in Go"
            required
          />
        </label>
        <label>
          Description / Context (optional)
          <textarea
            value={description}
            onChange={(e) => setDescription(e.target.value)}
            rows={4}
            placeholder="Details, key points, desired tone..."
          />
        </label>
        <button type="submit" disabled={loading}>
          {loading ? "Generating..." : "Generate post"}
        </button>
        {error && <p className="error">{error}</p>}
      </form>

      {content && (
        <section className="card">
          <div className="section-head">
            <h2>Generated draft</h2>
            <span className="count">{content.length} chars</span>
          </div>
          <textarea
            className="output"
            value={content}
            onChange={(e) => setContent(e.target.value)}
            rows={16}
          />
          <button onClick={saveDraft} disabled={saving || !title.trim()}>
            {saving ? "Saving..." : editingId ? "Update draft" : "Save draft"}
          </button>
        </section>
      )}

      {drafts.length > 0 && (
        <section className="card">
          <h2>Saved drafts</h2>
          <ul className="drafts">
            {drafts.map((d) => (
              <li key={d.id}>
                <button className="link" onClick={() => loadDraft(d)}>
                  {d.title || "(untitled)"}
                </button>
                <span className="meta">
                  {new Date(d.createdAt).toLocaleString()} · {d.status}
                </span>
              </li>
            ))}
          </ul>
        </section>
      )}
    </main>
  );
}
