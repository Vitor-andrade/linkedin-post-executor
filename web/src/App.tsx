import { useCallback, useEffect, useState } from "react";
import { ConfirmDialog, type ConfirmRequest } from "./ConfirmDialog";

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

interface Idea {
  title: string;
  description: string;
}

interface Category {
  name: string;
  ideas: Idea[];
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
  const [categories, setCategories] = useState<Category[]>([]);
  const [activeCategory, setActiveCategory] = useState("");
  const [suggestions, setSuggestions] = useState<Idea[]>([]);
  const [searchQuery, setSearchQuery] = useState("");
  const [loadingIdeas, setLoadingIdeas] = useState(false);
  const [confirmReq, setConfirmReq] = useState<ConfirmRequest | null>(null);
  const [images, setImages] = useState<{ id: number; filename: string }[]>([]);
  const [uploading, setUploading] = useState(false);

  // confirm shows the modal and resolves to the user's choice, replacing the
  // native window.confirm for a consistent look.
  const confirm = useCallback(
    (opts: Omit<ConfirmRequest, "resolve">) =>
      new Promise<boolean>((resolve) => {
        setConfirmReq({ ...opts, resolve });
      }),
    [],
  );

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

  const loadCategories = useCallback(async () => {
    try {
      const res = await fetch("/api/suggestions");
      if (!res.ok) return;
      const cats: Category[] = await res.json();
      setCategories(cats);
      if (cats.length > 0) {
        setActiveCategory(cats[0].name);
        setSuggestions(cats[0].ideas);
      }
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
    void loadCategories();

    // Surface the result of the OAuth callback redirect, then clean the URL.
    const params = new URLSearchParams(window.location.search);
    const li = params.get("linkedin");
    if (li === "connected") setNotice("LinkedIn connected ✅");
    else if (li === "error") setError("LinkedIn connection failed. Please try again.");
    if (li) window.history.replaceState({}, "", window.location.pathname);
  }, [loadDrafts, loadLinkedIn, loadScheduled, loadMetrics, loadCategories]);

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

  async function uploadImages(files: FileList | null) {
    if (!files || files.length === 0) return;
    setUploading(true);
    setError("");
    try {
      const form = new FormData();
      Array.from(files).forEach((f) => form.append("files", f));
      const res = await fetch("/api/images", { method: "POST", body: form });
      const data = await res.json();
      if (!res.ok) throw new Error(data.error ?? "Failed to upload image");
      setImages((prev) => [...prev, ...data]);
    } catch (err) {
      setError(err instanceof Error ? err.message : String(err));
    } finally {
      setUploading(false);
    }
  }

  function removeImage(id: number) {
    setImages((prev) => prev.filter((img) => img.id !== id));
  }

  async function publishNow() {
    const ok = await confirm({
      message: "Publish this post to your LinkedIn profile now? This can't be undone.",
      confirmLabel: "Publish",
    });
    if (!ok) return;
    setPublishing(true);
    setError("");
    setNotice("");
    try {
      const res = await fetch("/api/linkedin/publish", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({
          content,
          draftId: editingId,
          imageIds: images.map((img) => img.id),
        }),
      });
      const data = await res.json();
      if (!res.ok) throw new Error(data.error ?? "Failed to publish");
      setNotice(`Published to LinkedIn ✅ (${data.urn})`);
      setImages([]);
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
    const ok = await confirm({
      message: `Schedule this post for ${new Date(scheduledFor).toLocaleString()}?`,
      confirmLabel: "Schedule",
    });
    if (!ok) return;
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
    const ok = await confirm({
      message: "Cancel this scheduled post?",
      confirmLabel: "Cancel post",
      danger: true,
    });
    if (!ok) return;
    await fetch(`/api/schedule/${id}`, { method: "DELETE" });
    await loadScheduled();
    await loadMetrics();
  }

  async function disconnect() {
    const ok = await confirm({
      message:
        "Disconnect LinkedIn? You'll need to authorize again to publish or schedule.",
      confirmLabel: "Disconnect",
      danger: true,
    });
    if (!ok) return;
    await fetch("/api/linkedin/disconnect", { method: "POST" });
    setNotice("LinkedIn disconnected");
    await loadLinkedIn();
  }

  function selectCategory(cat: Category) {
    setActiveCategory(cat.name);
    setSuggestions(cat.ideas);
    setSearchQuery("");
  }

  async function generateIdeas(query: string) {
    setLoadingIdeas(true);
    setError("");
    try {
      const res = await fetch("/api/suggestions/generate", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ category: activeCategory, query }),
      });
      const data = await res.json();
      if (!res.ok) throw new Error(data.error ?? "Failed to get ideas");
      // Prepend the fresh ideas, keeping the list unique by title.
      setSuggestions((prev) => {
        const merged = [...(data.suggestions ?? []), ...prev] as Idea[];
        const seen = new Set<string>();
        return merged.filter((i) =>
          seen.has(i.title) ? false : (seen.add(i.title), true),
        );
      });
    } catch (err) {
      setError(err instanceof Error ? err.message : String(err));
    } finally {
      setLoadingIdeas(false);
    }
  }

  function searchIdeas(e: React.FormEvent) {
    e.preventDefault();
    if (searchQuery.trim()) void generateIdeas(searchQuery.trim());
  }

  function pickSuggestion(idea: Idea) {
    setTitle(idea.title);
    setDescription(idea.description);
    setError("");
    window.scrollTo({ top: 0, behavior: "smooth" });
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
    setImages([]);
    setError("");
  }

  async function deleteDraft(id: number) {
    const ok = await confirm({
      message: "Delete this draft? This cannot be undone.",
      confirmLabel: "Delete",
      danger: true,
    });
    if (!ok) return;
    try {
      const res = await fetch(`/api/drafts/${id}`, { method: "DELETE" });
      if (!res.ok) {
        const data = await res.json();
        throw new Error(data.error ?? "Failed to delete draft");
      }
      if (editingId === id) newDraft();
      await loadDrafts();
      await loadMetrics();
    } catch (err) {
      setError(err instanceof Error ? err.message : String(err));
    }
  }

  return (
    <div className="container">
      <header>
        <h1>LinkedIn Post Executor</h1>
        <span className={`badge ${health ? "ok" : "down"}`}>
          {health ? `online · AI: ${health.aiProvider}` : "backend offline"}
        </span>
      </header>

      <div className="layout">
        <main className="main-col">
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

          <div className="images">
            {images.map((img) => (
              <div className="thumb" key={img.id}>
                <img src={`/api/images/${img.id}`} alt={img.filename} />
                <button
                  className="thumb-remove"
                  onClick={() => removeImage(img.id)}
                  title="Remove image"
                >
                  ×
                </button>
              </div>
            ))}
            <label className="add-image" title="Attach images">
              <input
                type="file"
                accept="image/*"
                multiple
                onChange={(e) => {
                  void uploadImages(e.target.files);
                  e.target.value = "";
                }}
              />
              {uploading ? "…" : "+ Image"}
            </label>
          </div>
          {images.length > 0 && (
            <span className="meta">
              {images.length} image{images.length > 1 ? "s" : ""} attached · included when
              publishing now (scheduling is text-only)
            </span>
          )}

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
                <span className="draft-line">
                  <button className="link" onClick={() => loadDraft(d)}>
                    {d.title || "(untitled)"}
                  </button>
                  <button
                    className="link danger"
                    onClick={() => deleteDraft(d.id)}
                    title="Delete draft"
                  >
                    delete
                  </button>
                </span>
                <span className="meta">
                  {new Date(d.createdAt).toLocaleString()} · {d.status}
                </span>
              </li>
            ))}
          </ul>
        </section>
      )}
        </main>

        <aside className="suggest card">
          <h2>Post ideas</h2>
          <div className="tabs">
            {categories.map((c) => (
              <button
                key={c.name}
                className={`tab ${c.name === activeCategory ? "active" : ""}`}
                onClick={() => selectCategory(c)}
              >
                {c.name}
              </button>
            ))}
          </div>

          <form onSubmit={searchIdeas} className="suggest-search">
            <input
              value={searchQuery}
              onChange={(e) => setSearchQuery(e.target.value)}
              placeholder="Specific tech, e.g. Kubernetes"
            />
            <button type="submit" disabled={loadingIdeas || !searchQuery.trim()}>
              {loadingIdeas ? "…" : "Ideas"}
            </button>
          </form>

          <ul className="suggestions">
            {suggestions.map((s, i) => (
              <li key={`${s.title}-${i}`}>
                <button className="suggestion" onClick={() => pickSuggestion(s)}>
                  <span className="suggestion-title">{s.title}</span>
                  {s.description && (
                    <span className="suggestion-desc">{s.description}</span>
                  )}
                </button>
              </li>
            ))}
          </ul>

          <button
            className="link more"
            onClick={() => generateIdeas(searchQuery.trim())}
            disabled={loadingIdeas}
          >
            {loadingIdeas ? "Thinking…" : "✨ Generate more with AI"}
          </button>
        </aside>
      </div>

      {confirmReq && (
        <ConfirmDialog
          request={confirmReq}
          onClose={(ok) => {
            confirmReq.resolve(ok);
            setConfirmReq(null);
          }}
        />
      )}
    </div>
  );
}
