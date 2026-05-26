package ai

import (
	"fmt"
	"strings"
)

// systemPrompt encodes the LinkedIn Post Writer specification
// (see agents/linkedin-post-writer.agent.md): Unicode typography, a strong
// hook, scannable body, CTA and hashtags — no Markdown, no URLs in the body.
const systemPrompt = `Você é um especialista em escrever posts de alto engajamento para o LinkedIn.

Regras de formatação (OBRIGATÓRIAS):
- Use tipografia Unicode para destaque: negrito (𝗮𝘀𝘀𝗶𝗺) e itálico (𝘢𝘴𝘴𝘪𝘮). NUNCA use Markdown (**, ##, etc.).
- As 2 primeiras linhas são o gancho: precisam gerar curiosidade e o clique em "ver mais".
- Corpo escaneável: parágrafos curtos, uma linha em branco entre eles.
- Use ◈ ou ↳ para itens de lista quando fizer sentido.
- Separadores de seção com ━━━━━━━━━━━━━━━━━━ quando ajudar a leitura.
- Sem emojis no corpo (exceto ♻️ no CTA, se apropriado).
- Sem URLs no corpo do post.
- Termine com um CTA claro e, na ÚLTIMA linha, de 5 a 8 hashtags relevantes.
- Tamanho ideal entre 1500 e 2500 caracteres; máximo absoluto de 3000.

Responda APENAS com o texto final do post, pronto para copiar e colar.`

// buildPrompt assembles the full user prompt from the request seed.
func buildPrompt(req GenerateRequest) string {
	var b strings.Builder
	b.WriteString("Crie um post de LinkedIn com base nestas informações.\n\n")
	b.WriteString(fmt.Sprintf("Título/Tema: %s\n", strings.TrimSpace(req.Title)))
	if d := strings.TrimSpace(req.Description); d != "" {
		b.WriteString(fmt.Sprintf("Descrição/Contexto: %s\n", d))
	}
	return b.String()
}
