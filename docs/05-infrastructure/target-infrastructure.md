# Infraestrutura Alvo (TO-BE — Fora do GCP)

Arquitetura de hospedagem otimizada para custo mínimo, portabilidade por contêiner padrão e **zero dependência do Google Cloud Platform**.

---

## 1. Mapa de Ambientes

| Ambiente | Backend API (Go) | Banco PostgreSQL | Frontend Web/PWA | Mídias / Vídeos | SMTP |
| :--- | :--- | :--- | :--- | :--- | :--- |
| **Local (Dev)** | Go nativo ou Docker (`cmd/api`) | PostgreSQL 16 (Docker Compose — **sem Redis**) | `flutter run -d chrome` / web-server | `LocalStorageAdapter` em disco (`/tmp` / volume) ou R2 dev | Console / Mailpit local |
| **Homologação (Staging)** | **Render Free Web Service** (Docker) | **Supabase Free** (projeto separado, idealmente `sa-east-1`) | **Cloudflare Pages** (preview branch `develop`) | **Cloudflare R2** (prefixo `staging/` ou bucket separado) | **Resend Free** (ou Brevo Free) |
| **Piloto Real (Dados Reais)** | **Render Starter ($7/mês)** ou Fly.io/Railway (sem sleep nem cold start de 50s) | **Supabase Pro ($25/mês)** (backups diários automáticos, sem pausa de 7 dias!) | **Cloudflare Pages** + domínio próprio | **Cloudflare R2** (cota monitorada) | **Resend Free / Pro** |
| **Produção Comercial** | Contêiner Go escalável horizontalmente (Render / Fly / Railway ou VPS gerenciada) | **Supabase Pro+** com PITR e réplicas conforme carga | **Cloudflare Pages + WAF/Proxy** mesmo domínio (`/v1/*`) | **Cloudflare R2** + CDN/Signed URLs | **Resend / Postmark / SES** |

---

## 2. Por que Render + Supabase + Cloudflare Pages/R2 + Resend?

1. **Zero GCP:** Remove integralmente Cloud Run, Artifact Registry, Secret Manager e Cloud Build.
2. **Sem cartão obrigatório em Dev/Homologação:** Render Free, Supabase Free, Cloudflare Pages Free e Resend Free permitem validar o sistema em homologação sem risco de fatura surpresa. *(Observação: Cloudflare R2 exige verificação de método de pagamento na ativação, portanto para dev local usamos `LocalStorageAdapter` e ativamos R2 somente quando necessário com travas no backend).*
3. **Portabilidade total:** A API continua sendo uma imagem Docker limpa conectada por `DATABASE_URL`, trocável entre Render, Railway, Fly.io ou VPS em minutos.

Ver detalhamento financeiro por escala em [`cost-analysis.md`](cost-analysis.md).
