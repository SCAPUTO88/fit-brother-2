# ADR-001 — Retomada do Projeto pela Branch `develop`

**Status:** Proposto (Aguardando aprovação na Fase 0)  
**Data:** 2026-11-23

## Contexto
O histórico do repositório possui duas referências principais:
- `origin/main` (`565e121`), parado no PR #1 (Sprint 2 — Auth + CRUD parcial de alunos).
- `develop` (`16a8e69`), com 30 commits incluindo Sprints 3 (Workouts), 4 (Cloudflare R2 Videos), 5 (Flutter Web + CORS), scripts locais e ajustes de ambiente.

## Decisão
Adotar **`develop` (`16a8e69`) como base oficial de retomada**. Abrir branches de correção/entrega (iniciando por `fix/fase-1-security-baseline`) a partir de `develop` e atualizar `main` via PR revisado somente após conclusão da Fase 1/2.

## Consequências
- **Positivas:** Preserva todo o trabalho de treinos, R2 e Flutter já construído.
- **Negativas:** As vulnerabilidades das Sprints 3–5 e divergências de contrato permanecem presentes até a Fase 1 ser executada.
