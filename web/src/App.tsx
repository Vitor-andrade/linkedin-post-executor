import { useEffect, useState } from "react";

interface Health {
  status: string;
  aiProvider: string;
}

export default function App() {
  const [health, setHealth] = useState<Health | null>(null);
  const [title, setTitle] = useState("");
  const [description, setDescription] = useState("");
  const [content, setContent] = useState("");
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState("");

  useEffect(() => {
    fetch("/api/health")
      .then((r) => r.json())
      .then(setHealth)
      .catch(() => setHealth(null));
  }, []);

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
      if (!res.ok) throw new Error(data.error ?? "Falha ao gerar");
      setContent(data.content);
    } catch (err) {
      setError(err instanceof Error ? err.message : String(err));
    } finally {
      setLoading(false);
    }
  }

  return (
    <main className="container">
      <header>
        <h1>LinkedIn Post Executor</h1>
        <span className={`badge ${health ? "ok" : "down"}`}>
          {health ? `online · IA: ${health.aiProvider}` : "backend offline"}
        </span>
      </header>

      <form onSubmit={generate} className="card">
        <label>
          Título / Tema
          <input
            value={title}
            onChange={(e) => setTitle(e.target.value)}
            placeholder="Ex.: 5 lições que aprendi escalando uma API em Go"
            required
          />
        </label>
        <label>
          Descrição / Contexto (opcional)
          <textarea
            value={description}
            onChange={(e) => setDescription(e.target.value)}
            rows={4}
            placeholder="Detalhes, pontos-chave, tom desejado..."
          />
        </label>
        <button type="submit" disabled={loading}>
          {loading ? "Gerando..." : "Gerar post"}
        </button>
        {error && <p className="error">{error}</p>}
      </form>

      {content && (
        <section className="card">
          <h2>Rascunho gerado</h2>
          <textarea
            className="output"
            value={content}
            onChange={(e) => setContent(e.target.value)}
            rows={16}
          />
        </section>
      )}
    </main>
  );
}
