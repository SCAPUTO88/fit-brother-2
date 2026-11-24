# ADR-004 — Remoção de Redis Obrigatório e Estratégia de Rate Limiting

**Status:** Proposto (Aguardando aprovação na Fase 0)  
**Data:** 2026-11-23

## Contexto
Embora `SAD.md`, `.github/workflows/ci.yml`, `docker-compose.yml` e `internal/infrastructure/config/config.go:77` tratem Redis como requisito obrigatório (`requireEnv("REDIS_URL")`), **nenhum pacote do código-fonte conecta ou usa Redis**.

## Decisão
1. **Remover `REDIS_URL` como variável obrigatória no startup** na Fase 2.
2. Remover Redis do `docker-compose.yml` e do CI padrão até haver necessidade real comprovada.
3. Implementar **rate limiting inicial em memória na API Go** (ex.: token bucket por IP/e-mail com limpeza periódica de chaves antigas) para endpoints críticos (`/auth/login`, `/auth/refresh`, `/auth/activate`, `/auth/password-reset`).
4. Controlar estado de revogação crítica (`refresh_tokens`, `is_active`) no PostgreSQL já existente.

## Critério de Reintrodução Futura (Upstash Redis)
Redis voltará a ser considerado **opcionalmente** (atrás de interface em `domain`/`application`) apenas quando houver múltiplas réplicas simultâneas da API em produção comercial exigindo rate limiting distribuído ou cache compartilhado de alta frequência.
