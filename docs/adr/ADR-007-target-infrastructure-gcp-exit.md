# ADR-007 — Arquitetura de Hospedagem Alvo Fora do GCP

**Status:** Proposto (Aguardando aprovação na Fase 0)  
**Data:** 2026-11-23

## Contexto
A documentação legada (`SAD.md`, `README.md`) previa **Google Cloud Run** para o backend Go. Há diretriz expressa para **não utilizar nenhum serviço do Google Cloud Platform (GCP)** no projeto.

## Decisão
Adotar arquitetura portátil baseada em contêiner Docker padrão + SQL padrão:
- **Frontend Web/PWA:** **Cloudflare Pages** (free-tier, banda estática ilimitada, integração nativa com DNS/WAF).
- **Backend API Go:**
  - *Dev/Homologação gratuita:* **Render Free** (sem cartão obrigatório; aceita Dockerfile padrão; 512 MB RAM, sleep após 15m). *(Alternativa: Koyeb Free ou Railway $1/mês crédito após trial).*
  - *Piloto real / Produção inicial:* **Render Starter ($7/mês)** ou **Railway Hobby ($5/mês)** ou **Fly.io**, eliminando o cold start de ~50s durante treinos na academia.
- **Banco de Dados:** **Supabase PostgreSQL** (`sa-east-1` quando disponível) conectado exclusivamente via `DATABASE_URL` SQL padrão (`pgx/v5`), usando **Free** em homologação (com `pg_dump` externo) e **Pro ($25/mês)** com backups automáticos no piloto/produção.
- **Mídias:** **Cloudflare R2** (via `StoragePort`) com $0 egress.
- **SMTP:** **Resend** (3.000/mês, 100/dia grátis) ou **Brevo** (300/dia grátis).

## Consequências
- **Zero lock-in:** Como a API fala SQL padrão via `DATABASE_URL` e roda como container OCI simples, trocar de PaaS (ex.: Render → Fly → VPS Hetzner/AWS) leva minutos sem reescrever código.
