# Decisões Arquiteturais e de Governança (Gate Fase 0 → Fase 1)

**Status:** **APROVADAS pelo proprietário em 2026-11-24** (`"Decisoes aprovadas."`), liberando a execução da **Fase 1 — Security Baseline** na branch `fix/fase-1-security-baseline` a partir de `develop` (`16a8e69`).

| ID | Decisão Solicitada | Recomendação Técnica | Referência |
| :--- | :--- | :--- | :--- |
| **DEC-01** | **Branch base de retomada** | Usar `develop` (`16a8e69`) como base oficial e posteriormente alinhar `main` somente após a Fase 1/2 estabilizada via PR. | [ADR-001](../adr/ADR-001-base-branch-resumption.md) |
| **DEC-02** | **Versionar `AGENTS.md` e `docs/` no Git** | Remover `AGENTS.md` e `docs/` das linhas 50–51 do `.gitignore` para preservar governança e handoff no repositório privado. | [repository-audit.md](../00-discovery/repository-audit.md) |
| **DEC-03** | **Preservação da Stack Principal** | Manter **Backend Go + PostgreSQL + Flutter Web/PWA**, sem reescrever backend ou migrar para React. | [ADR-002](../adr/ADR-002-backend-frontend-db-preservation.md) |
| **DEC-04** | **Estratégia de Autenticação** | Manter **Auth própria no backend Go** (evoluída com sessão segura Web + convites + TOTP opcional) em vez de migrar para Supabase Auth, evitando quebra dos hashes `bcrypt+pepper` e mantendo portabilidade total. | [ADR-003](../adr/ADR-003-auth-strategy-own-vs-supabase.md) |
| **DEC-05** | **Remoção de Redis obrigatório e Saída do GCP** | Remover `REDIS_URL` obrigatório e referências GCP; adotar Render (dev/homolog) + Supabase Postgres + Cloudflare Pages/R2 + Resend/Brevo. | [ADR-004](../adr/ADR-004-redis-removal-and-rate-limiting.md), [ADR-007](../adr/ADR-007-target-infrastructure-gcp-exit.md) |
| **DEC-06** | **Remoção gradual do CPF obrigatório** | Tornar CPF opcional na Fase 3 e eliminar coleta em novos cadastros para aderência à minimização da LGPD. | [ADR-006](../adr/ADR-006-lgpd-cpf-removal.md) |
| **DEC-07** | **Catálogo base de exercícios e Storage opcional** | Adotar **`yuhonas/free-exercise-db` (Unlicense)** traduzido + `StoragePort` opcional (`local` em dev, `R2` em nuvem). | [ADR-005](../adr/ADR-005-media-storage-and-exercise-library.md) |
| **DEC-08** | **Sessão Web via mesmo domínio / Cookies `HttpOnly`** | Servir API sob mesmo site (`/v1/*` via proxy/BFF) com cookies `HttpOnly; Secure; SameSite` + CSRF em produção Web, substituindo `localStorage` para tokens. | [ADR-008](../adr/ADR-008-web-session-security-bff-cookies.md) |
| **DEC-09** | **Revisão jurídica externa antes do piloto/comercial** | Validar com advogado/DPO minuta de Privacidade/Termos, transferência internacional e obrigatoriedade de DPO sob Resolução CD/ANPD nº 2/2022 antes de dados reais em escala. | [legal-basis-matrix.md](../04-privacy/legal-basis-matrix.md) |
