# ADR-003 — Camada de IA plugável com Ollama como padrão

- **Status:** Aceito
- **Data:** 2026-05-25

## Contexto

A geração de conteúdo é o coração do produto, mas o projeto tem **custo zero** como regra.
Provedores de LLM em nuvem (Claude, OpenAI) exigem uma chave paga. Ao mesmo tempo, quem quiser
mais qualidade deve poder usar a própria chave.

## Decisão

Definir uma **interface `AIProvider`** desacoplada, com implementações intercambiáveis:

- **`OllamaProvider`** — padrão. Modelo executado localmente via [Ollama](https://ollama.com/),
  custo zero.
- **`APIProvider`** — "traga sua própria chave" (BYO) para Claude/OpenAI, configurável.

A escolha do provedor fica em configuração (settings no SQLite / `.env`). O prompt de geração
segue a especificação do agente `agents/linkedin-post-writer.agent.md`.

## Alternativas consideradas

- **Somente BYO key:** mais simples e melhor qualidade de texto, mas exige uma chave paga para
  o app funcionar — viola "custo zero".
- **Somente Ollama:** 100% grátis e privado, mas a qualidade depende do hardware do usuário e
  remove a flexibilidade de quem prefere um modelo de ponta.

## Consequências

**Positivas**
- Honra "custo zero" e "open source" com o default local.
- Flexível: trocar de provedor não toca a lógica de domínio.
- Privacidade máxima na opção local (nada sai da máquina).

**Negativas / trade-offs**
- Manter mais de uma implementação e normalizar diferenças entre provedores.
- Qualidade do default (Ollama) varia conforme o modelo/hardware local.
