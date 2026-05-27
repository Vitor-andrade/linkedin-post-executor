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

interface LinkedInStatus {
  configured: boolean;
  connected: boolean;
  expiresAt?: string;
}

interface Metrics {
  draftsTotal: number;
  draftsByStatus: Record<string, number>;
  scheduled: { pending: number; published: number; failed: number };
  publishedTotal: number;
  lastPublishedAt?: string;
}

interface ScheduledPost {
  id: number;
  draftId: number | null;
  content: string;
  scheduledFor: string;
  status: string;
  linkedinUrn: string;
  error: string;
  attempts: number;
  nextAttemptAt?: string;
  createdAt: string;
}

export default function App() {
  const [health, setHealth] = useState<Health | null>(null);
  const [title, setTitle] = useState("");
  const [description, setDescription] = useState("");
  const [content, setContent] = useState("");
  const [loading, setLoading] = useState(false);
  const [saving, setSaving] = useState(false);
  const [publishing, setPublishing] = useState(false);
  const [error, setError] = useState("");
  const [notice, setNotice] = useState("");
  const [drafts, setDrafts] = useState<Draft[]>([]);
  const [editingId, setEditingId] = useState<number | null>(null);
  const [linkedin, setLinkedin] = useState<LinkedInStatus | null>(null);
  const [scheduledFor, setScheduledFor] = useState("");
  const [scheduling, setScheduling] = useState(false);
  const [scheduled, setScheduled] = useState<ScheduledPost[]>([]);
  const [metrics, setMetrics] = useState<Metrics | null>(null);

  const loadDrafts = useCallback(async () => {
    try {
      const res = await fetch("/api/drafts");
      if (!res.ok) return;
      setDrafts(await res.json());
    } catch {
      /* listing is best-effort; ignore */
    }
  }, []);

  const loadLinkedIn = useCallback(async () => {
    try {
      const res = await fetch("/api/linkedin/status");
      if (!res.ok) return;
      setLinkedin(await res.json());
    } catch {
      /* ignore */
    }
  }, []);

  const loadScheduled = useCallback(async () => {
    try {
      const res = await fetch("/api/schedule");
      if (!res.ok) return;
      setScheduled(await res.json());
    } catch {
      /* ignore */
    }
  }, []);

  const loadMetrics = useCallback(async () => {
    try {
      const res = await fetch("/api/metrics");
      if (!res.ok) return;
      setMetrics(await res.json());
    } catch {
      /* ignore */
    }
  }, []);

  useEffect(() => {
    fetch("/api/health")
      .then((r) => r.json())
      .then(setHealth)
      .catch(() => setHealth(null));
    void loadDrafts();
    void loadLinkedIn();
    void loadScheduled();
    void loadMetrics();

    // Surface the result of the OAuth callback redirect, then clean the URL.
    const params = new URLSearchParams(window.location.search);
    const li = params.get("linkedin");
    if (li === "connected") setNotice("LinkedIn connected ✅");
    else if (li === "error") setError("LinkedIn connection failed. Please try again.");
    if (li) window.history.replaceState({}, "", window.location.pathname);
  }, [loadDrafts, loadLinkedIn, loadScheduled, loadMetrics]);

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
      await loadMetrics();
    } catch (err) {
      setError(err instanceof Error ? err.message : String(err));
    } finally {
      setSaving(false);
    }
  }

  async function publishNow() {
    setPublishing(true);
    setError("");
    setNotice("");
    try {
      const res = await fetch("/api/linkedin/publish", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ content, draftId: editingId }),
      });
      const data = await res.json();
      if (!res.ok) throw new Error(data.error ?? "Failed to publish");
      setNotice(`Published to LinkedIn ✅ (${data.urn})`);
      await loadDrafts();
      await loadMetrics();
    } catch (err) {
      setError(err instanceof Error ? err.message : String(err));
    } finally {
      setPublishing(false);
    }
  }

  async function schedulePost() {
    if (!scheduledFor) {
      setError("Pick a date and time to schedule.");
      return;
    }
    setScheduling(true);
    setError("");
    setNotice("");
    try {
      const res = await fetch("/api/schedule", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({
          content,
          draftId: editingId,
          // datetime-local is local time; send an absolute (UTC) instant.
          scheduledFor: new Date(scheduledFor).toISOString(),
        }),
      });
      const data = await res.json();
      if (!res.ok) throw new Error(data.error ?? "Failed to schedule");
      setNotice(`Scheduled for ${new Date(data.scheduledFor).toLocaleString()}`);
      setScheduledFor("");
      await loadScheduled();
      await loadMetrics();
    } catch (err) {
      setError(err instanceof Error ? err.message : String(err));
    } finally {
      setScheduling(false);
    }
  }

  async function cancelScheduled(id: number) {
    await fetch(`/api/schedule/${id}`, { method: "DELETE" });
    await loadScheduled();
    await loadMetrics();
  }

  async function disconnect() {
    await fetch("/api/linkedin/disconnect", { method: "POST" });
    setNotice("LinkedIn disconnected");
    await loadLinkedIn();
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

      <section className="linkedin">
        {linkedin?.connected ? (
          <>
            <span className="badge ok">LinkedIn connected</span>
            <button className="link" onClick={disconnect}>
              disconnect
            </button>
          </>
        ) : linkedin?.configured ? (
          <>
            <span className="badge down">LinkedIn not connected</span>
            <a className="connect" href="/api/linkedin/login">
              Connect LinkedIn
            </a>
          </>
        ) : (
          <span className="meta">
            LinkedIn publishing disabled — set LPE_LINKEDIN_CLIENT_ID/SECRET to enable.
          </span>
        )}
      </section>

      {notice && <p className="notice">{notice}</p>}

      {metrics && (metrics.draftsTotal > 0 || metrics.publishedTotal > 0) && (
        <section className="card stats">
          <div className="stat">
            <span className="stat-num">{metrics.draftsTotal}</span>
            <span className="stat-label">drafts</span>
          </div>
          <div className="stat">
            <span className="stat-num">{metrics.scheduled.pending}</span>
            <span className="stat-label">scheduled</span>
          </div>
          <div className="stat">
            <span className="stat-num">{metrics.publishedTotal}</span>
            <span className="stat-label">published</span>
          </div>
          {metrics.scheduled.failed > 0 && (
            <div className="stat">
              <span className="stat-num failed">{metrics.scheduled.failed}</span>
              <span className="stat-label">failed</span>
            </div>
          )}
          {metrics.lastPublishedAt && (
            <div className="stat last">
              <span className="stat-label">
                last published {new Date(metrics.lastPublishedAt).toLocaleDateString()}
              </span>
            </div>
          )}
        </section>
      )}

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
          <div className="actions">
            <button onClick={saveDraft} disabled={saving || !title.trim()}>
              {saving ? "Saving..." : editingId ? "Update draft" : "Save draft"}
            </button>
            <button
              className="secondary"
              onClick={publishNow}
              disabled={publishing || !linkedin?.connected || !content.trim()}
              title={
                linkedin?.connected
                  ? "Publish this post to your LinkedIn profile now"
                  : "Connect LinkedIn first"
              }
            >
              {publishing ? "Publishing..." : "Publish to LinkedIn"}
            </button>
          </div>
          <div className="schedule-row">
            <input
              type="datetime-local"
              value={scheduledFor}
              onChange={(e) => setScheduledFor(e.target.value)}
            />
            <button
              className="secondary"
              onClick={schedulePost}
              disabled={scheduling || !content.trim() || !scheduledFor}
            >
              {scheduling ? "Scheduling..." : "Schedule"}
            </button>
          </div>
        </section>
      )}

      {scheduled.length > 0 && (
        <section className="card">
          <h2>Scheduled posts</h2>
          <ul className="drafts">
            {scheduled.map((p) => (
              <li key={p.id}>
                <span className="sched-line">
                  <strong>{new Date(p.scheduledFor).toLocaleString()}</strong>
                  <span className={`pill ${p.status}`}>{p.status}</span>
                  {p.status === "pending" && (
                    <button className="link" onClick={() => cancelScheduled(p.id)}>
                      cancel
                    </button>
                  )}
                </span>
                <span className="meta">
                  {p.content.slice(0, 80)}
                  {p.content.length > 80 ? "…" : ""}
                </span>
                {p.attempts > 0 && p.status === "pending" && (
                  <span className="meta">
                    retrying — attempt {p.attempts}
                    {p.nextAttemptAt
                      ? ` · next ${new Date(p.nextAttemptAt).toLocaleTimeString()}`
                      : ""}
                  </span>
                )}
                {p.error && <span className="error">{p.error}</span>}
              </li>
            ))}
          </ul>
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
