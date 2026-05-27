package ai

import (
	"context"
	"strings"
)

// Category groups curated post-topic ideas shown instantly in the UI (offline,
// zero cost). The AI can extend a category on demand via Suggest.
type Category struct {
	Name   string   `json:"name"`
	Topics []string `json:"topics"`
}

// curatedCatalog is the built-in seed list. Topics are post *angles* a
// developer can expand on, written in English to match the generated posts.
var curatedCatalog = []Category{
	{Name: "Frontend", Topics: []string{
		"A rendering pattern that finally made my React app fast",
		"Why I stopped reaching for a state library on small apps",
		"Accessibility wins that took 10 minutes and mattered",
		"The CSS feature I wish I'd learned years earlier",
		"What a design system actually saves a small team",
	}},
	{Name: "Backend", Topics: []string{
		"A bug that taught me how timeouts really propagate",
		"When I chose a boring database and never regretted it",
		"The API design mistake I keep seeing in code reviews",
		"How I think about idempotency in payment flows",
		"Lessons from migrating a monolith one slice at a time",
	}},
	{Name: "Databases", Topics: []string{
		"The index that turned a 4s query into 40ms",
		"Why I default to SQLite for local-first tools",
		"N+1 queries: how I spot and kill them",
		"What I learned the hard way about migrations in production",
		"Normalization vs. denormalization: a real trade-off I faced",
	}},
	{Name: "DevOps", Topics: []string{
		"The CI step that catches the most bugs for the least effort",
		"How a single binary simplified my whole deployment",
		"What I monitor first when something feels slow",
		"The infra cost I cut without anyone noticing",
		"Why I keep my Dockerfiles boring on purpose",
	}},
	{Name: "Security", Topics: []string{
		"How I store secrets so they never hit the repo",
		"Encrypting tokens at rest: a pattern I now use everywhere",
		"The OAuth detail that trips up most first integrations",
		"A threat model I sketch before writing any auth code",
		"Why least privilege saved me during an incident",
	}},
	{Name: "AI / ML", Topics: []string{
		"Running models locally: what changed in my workflow",
		"How I keep AI features cheap with a pluggable provider",
		"Prompting lessons from shipping an AI feature",
		"When a smaller model was the right call",
		"Guardrails I add before letting an LLM touch user data",
	}},
	{Name: "Career", Topics: []string{
		"The habit that made me a better code reviewer",
		"What I'd tell my junior self about scope",
		"How writing in public changed my engineering career",
		"Saying no to work: a skill I had to learn",
		"The side project that taught me more than any course",
	}},
}

// Categories returns the curated catalog.
func Categories() []Category { return curatedCatalog }

// completer is the low-level capability each provider implements: a single
// system+user completion. Generate and Suggest are both built on top of it.
type completer interface {
	complete(ctx context.Context, system, user string) (string, error)
}

const suggestionSystemPrompt = `You generate concise LinkedIn post TOPIC IDEAS for software developers — not full posts.
Rules:
- Return exactly 6 ideas.
- One idea per line, no numbering, no bullets, no Markdown.
- Each idea is a punchy, specific angle the author can expand into a post.
- Write in English.
Respond ONLY with the 6 lines.`

func buildSuggestionPrompt(category, query string) string {
	var b strings.Builder
	b.WriteString("Suggest fresh LinkedIn post ideas for a developer.\n")
	if c := strings.TrimSpace(category); c != "" {
		b.WriteString("Area: " + c + "\n")
	}
	if q := strings.TrimSpace(query); q != "" {
		b.WriteString("Specific focus / technology: " + q + "\n")
	}
	return b.String()
}

// suggestVia produces topic ideas through any completer and parses them into a
// clean list.
func suggestVia(ctx context.Context, c completer, category, query string) ([]string, error) {
	out, err := c.complete(ctx, suggestionSystemPrompt, buildSuggestionPrompt(category, query))
	if err != nil {
		return nil, err
	}
	return parseSuggestions(out), nil
}

// parseSuggestions turns a raw multi-line completion into trimmed idea strings,
// stripping common list prefixes the model might add despite instructions.
func parseSuggestions(raw string) []string {
	var ideas []string
	for _, line := range strings.Split(raw, "\n") {
		s := strings.TrimSpace(line)
		s = strings.TrimLeft(s, "-*•0123456789. )")
		s = strings.TrimSpace(s)
		if s != "" {
			ideas = append(ideas, s)
		}
	}
	return ideas
}
