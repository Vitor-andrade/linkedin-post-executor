# ADR-001 — Backend em Go compilado em binário único

- **Status:** Aceito
- **Data:** 2026-05-25

## Contexto

O projeto é **local-first**, **open source** e precisa ser distribuído como um
**produto individualizado** (cada usuário roda a própria instância). A facilidade de
"baixar e rodar" é um requisito de primeira ordem. Há também um agendador que precisa
rodar em segundo plano para publicar posts no horário marcado.

O autor é proficiente em Node/TypeScript e tem interesse em aprofundar Go, sem pressão de prazo.

## Decisão

Usar **Go** no backend, compilando em um **único binário** que embute a UID (React/Vite)
via `go:embed`. O mesmo processo serve a interface, expõe a API HTTP local e executa o
agendador.

## Alternativas consideradas

- **Next.js full-stack:** entrega mais rápida e familiar, porém a execução local exige
  runtime Node, e um agendador persistente local fica deselegante (precisaria de um worker
  `node-cron` separado). A distribuição vira "clone o repo + npm install" ou Docker.
- **Go API + Next.js separado:** mais flexível, mas são dois processos para rodar localmente,
  o que enfraquece a proposta "baixe e rode".

## Consequências

**Positivas**
- Distribuição trivial: um executável sem dependências de runtime; cross-compile simples.
- Goroutines tornam o agendador em segundo plano natural e leve.
- Baixo consumo de recursos rodando localmente.
- Objetivo de aprendizado do autor atendido.

**Negativas / trade-offs**
- Curva de aprendizado e mais _boilerplate_ que Next.js.
- Ecossistema de frontend permanece em JS de qualquer forma (build do Vite embutido).

> ⚠️ **Importante:** a escolha de Go **não** é por desempenho de CPU. A latência percebida é
> dominada pela geração do LLM e pelas chamadas à API do LinkedIn. A motivação real é o
> **modelo de distribuição** e o **processo persistente** do agendador.
