package ai

import (
	"context"
	"strings"
)

// Idea is a post suggestion: a topic title plus a short context the user can
// expand on. Both seed the generation form when picked.
type Idea struct {
	Title       string `json:"title"`
	Description string `json:"description"`
}

// Category groups curated post ideas shown instantly in the UI (offline, zero
// cost). The AI can extend a category on demand via Suggest.
type Category struct {
	Name  string `json:"name"`
	Ideas []Idea `json:"ideas"`
}

// curatedCatalog is the built-in seed list. Each idea is a post angle a
// developer can expand on, written in English to match the generated posts.
var curatedCatalog = []Category{
	{Name: "Frontend", Ideas: []Idea{
		{"A rendering pattern that finally made my React app fast", "Walk through the bottleneck, the fix, and the measured before/after."},
		{"Why I stopped reaching for a state library on small apps", "Show what built-in tools covered and when a library would still pay off."},
		{"Accessibility wins that took 10 minutes and mattered", "List small, concrete changes and the impact on real users."},
		{"The CSS feature I wish I'd learned years earlier", "Explain the feature, a practical example, and what it replaced."},
		{"What a design system actually saves a small team", "Contrast life before/after, with honest trade-offs."},
	}},
	{Name: "Backend", Ideas: []Idea{
		{"A bug that taught me how timeouts really propagate", "Tell the incident story and the mental model you kept."},
		{"When I chose a boring database and never regretted it", "Explain the decision, the pressure to over-engineer, and the outcome."},
		{"The API design mistake I keep seeing in code reviews", "Describe the anti-pattern, why it hurts, and the cleaner shape."},
		{"How I think about idempotency in payment flows", "Share the rule of thumb and a failure it prevented."},
		{"Lessons from migrating a monolith one slice at a time", "Outline the strategy, the first slice, and what you'd do differently."},
	}},
	{Name: "Databases", Ideas: []Idea{
		{"The index that turned a 4s query into 40ms", "Show the query, the plan, and the one change that fixed it."},
		{"Why I default to SQLite for local-first tools", "Cover the trade-offs and where it stops being the right call."},
		{"N+1 queries: how I spot and kill them", "Give the symptom, the detection trick, and the fix."},
		{"What I learned the hard way about migrations in production", "Tell the near-miss and the checklist you now follow."},
		{"Normalization vs. denormalization: a real trade-off I faced", "Describe the case and how you decided."},
	}},
	{Name: "DevOps", Ideas: []Idea{
		{"The CI step that catches the most bugs for the least effort", "Name the check and quantify what it saves."},
		{"How a single binary simplified my whole deployment", "Contrast the old pipeline with the new one."},
		{"What I monitor first when something feels slow", "Share your triage order and why."},
		{"The infra cost I cut without anyone noticing", "Explain the waste you found and how you removed it safely."},
		{"Why I keep my Dockerfiles boring on purpose", "List the boring choices and the bugs they avoid."},
	}},
	{Name: "Security", Ideas: []Idea{
		{"How I store secrets so they never hit the repo", "Walk through the setup and the failure mode it prevents."},
		{"Encrypting tokens at rest: a pattern I now use everywhere", "Explain the approach in plain terms and when it matters."},
		{"The OAuth detail that trips up most first integrations", "Pinpoint the gotcha and the fix."},
		{"A threat model I sketch before writing any auth code", "Show the few questions you always ask."},
		{"Why least privilege saved me during an incident", "Tell the story and the principle it reinforced."},
	}},
	{Name: "AI / ML", Ideas: []Idea{
		{"Running models locally: what changed in my workflow", "Compare cost, privacy, and speed before/after."},
		{"How I keep AI features cheap with a pluggable provider", "Explain the abstraction and the free-by-default choice."},
		{"Prompting lessons from shipping an AI feature", "Share 3 concrete prompt changes that improved output."},
		{"When a smaller model was the right call", "Describe the task and why bigger wasn't better."},
		{"Guardrails I add before letting an LLM touch user data", "List the checks and the risk each removes."},
	}},
	{Name: "Career", Ideas: []Idea{
		{"The habit that made me a better code reviewer", "Describe the habit and a review it improved."},
		{"What I'd tell my junior self about scope", "Share the lesson and a project where it mattered."},
		{"How writing in public changed my engineering career", "Tell what you wrote, what happened, and what you'd start today."},
		{"Saying no to work: a skill I had to learn", "Give a concrete example and the framing that helped."},
		{"The side project that taught me more than any course", "Explain what it was and the specific skills it built."},
	}},
}

// Categories returns the curated catalog.
func Categories() []Category { return curatedCatalog }

// completer is the low-level capability each provider implements: a single
// system+user completion. Generate and Suggest are both built on top of it.
type completer interface {
	complete(ctx context.Context, system, user string) (string, error)
}

// ideaSeparator splits the title from the description in each suggestion line.
const ideaSeparator = "::"

const suggestionSystemPrompt = `You generate concise LinkedIn post IDEAS for software developers — not full posts.
Rules:
- Return exactly 6 ideas.
- One idea per line, formatted as: <punchy title> :: <one-sentence angle/context>
- No numbering, no bullets, no Markdown.
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

// suggestVia produces ideas through any completer and parses them.
func suggestVia(ctx context.Context, c completer, category, query string) ([]Idea, error) {
	out, err := c.complete(ctx, suggestionSystemPrompt, buildSuggestionPrompt(category, query))
	if err != nil {
		return nil, err
	}
	return parseSuggestions(out), nil
}

// parseSuggestions turns a raw multi-line completion into ideas, stripping list
// prefixes and splitting each line into title and description.
func parseSuggestions(raw string) []Idea {
	var ideas []Idea
	for _, line := range strings.Split(raw, "\n") {
		s := strings.TrimSpace(line)
		s = strings.TrimLeft(s, "-*•0123456789. )")
		s = strings.TrimSpace(s)
		if s == "" {
			continue
		}
		title, desc, _ := strings.Cut(s, ideaSeparator)
		ideas = append(ideas, Idea{
			Title:       strings.TrimSpace(title),
			Description: strings.TrimSpace(desc),
		})
	}
	return ideas
}
