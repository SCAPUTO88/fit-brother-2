# ADR-002 — Preservação de Go (Hexagonal) + PostgreSQL + Flutter

**Status:** Proposto (Aguardando aprovação na Fase 0)  
**Data:** 2026-11-23

## Contexto
Avaliou-se se o projeto deveria reescrever o backend ou migrar o frontend Flutter Web (`fit_brother_app`) para React/Next.js ou outro framework.

## Decisão
**Manter Backend Go (Gin + sqlx/pgx em Arquitetura Hexagonal) + PostgreSQL + Frontend Flutter (Riverpod + GoRouter + Dio)** no MVP.

## Justificativa
1. O código Go atual possui boa separação estrutural (`domain`/`application`/`adapters`/`infrastructure`), testes unitários e binário único de baixo consumo de RAM (~512 MB compatível com free-tier).
2. A base Flutter já existe e atende Web/PWA agora preservando reuso futuro para aplicativos móveis nativos Android/iOS sem duplicar código.
3. Não existem impedimentos técnicos estruturais do Flutter Web que justifiquem o custo/atraso de reescrita para React neste momento — os bugs encontrados são de integração/sessão/UI corrigíveis nas Fases 1–8.

## Riscos Monitorados
- Sessão segura Web (`ADR-008`), tamanho inicial do bundle Web (mitigado com CDN Cloudflare Pages + cache adequado de estáticos) e Service Worker customizado (Fase 8).
