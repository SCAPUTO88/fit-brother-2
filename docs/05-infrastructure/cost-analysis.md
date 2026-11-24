# Análise de Custos e Estratégia Financeira

Valores estimados em **USD/mês** com base nas tabelas públicas vigentes dos provedores avaliados (sujeitos a alteração pelos fornecedores).

---

## 1. Premissas por Cenário de Escala

| Cenário | Treinadores | Alunos | Storage Vídeos Est. | Egress Vídeo/API Est. | E-mails/Mês (Convites/Reset) |
| :--- | :---: | :---: | :---: | :---: | :---: |
| **C1 — Inicial** | 1 | 20 | 1–2 GB | 5–10 GB | < 100 |
| **C2 — Autônomo Cheio** | 1 | 50 | 3–5 GB | 15–25 GB | < 200 |
| **C3 — Pequeno Estúdio** | 5 | 250 | 8–15 GB | 60–100 GB | ~500 |
| **C4 — Rede / Multi-Estúdios** | 20 | 1.000 | 30–50 GB | 250–400 GB | ~2.000 |
| **C5 — Escala Inicial** | 100 | 5.000 | 120–200 GB | 1–1,5 TB | ~8.000–10.000 |

---

## 2. Custos por Estágio Operacional vs Cenários

### A. Desenvolvimento e Homologação (Sem Dados Reais Sensíveis em Produção)
- **Custo estimado:** **$0/mês**
- Stack: Local Docker Compose (Postgres 16) + Render Free (750h/mês, 512 MB RAM, sleep 15m, 5 GB egress) + Supabase Free (500 MB DB, pausa 7 dias sem uso) + Cloudflare Pages Free (banda estática ilimitada) + Resend Free (3.000/mês, 100/dia) + GitHub Actions Free (2.000 min/mês em repo privado) + `LocalStorage` ou R2 Free (10 GB, $0 egress).
- **Por que NÃO serve para produção/piloto real contínuo:** Render Free dorme (cold start ~50s ruim na academia), Supabase Free pausa após 7 dias sem uso e **não tem backups automáticos nem PITR**.

### B. Piloto com Dados Reais (C1: 1/20 e C2: 1/50 — Recomendado Mínimo Seguro)
| Componente | C1 (1 / 20) — Piloto Seguro | C2 (1 / 50) — Piloto Seguro | Observação |
| :--- | :---: | :---: | :--- |
| **Backend (Go API)** | $0 (Render Free*) a **$7** (Render Starter 512MB, sem sleep) | **$5–$7** (Railway Hobby / Render Starter) | *Free só aceitável se cold start de 50s for tolerado em teste controlado; para dados reais contínuos recomenda-se Starter ($7) ou Railway ($5). |
| **Banco PostgreSQL** | **$0** (c/ `pg_dump` diário externo obrigatório) ou **$25** (Supabase Pro) | **$25** (Supabase Pro — backups diários, sem pausa) | **Dados de saúde (LGPD)** em piloto real exigem backup confiável; Supabase Pro ($25) resolve pausa e backup. |
| **Storage / Vídeos (R2)** | **$0** (dentro dos 10 GB grátis, $0 egress) | **$0** (<= 10 GB, $0 egress) | Usar vídeos curtos comprimidos (720p H.264 <= 15–30 MB) + catálogo `free-exercise-db` leve. |
| **Frontend + DNS/CDN (Cloudflare Pages)** | **$0** | **$0** | Banda estática ilimitada. |
| **Auth / SMTP (Resend/Brevo)** | **$0** | **$0** | Própria API Go (Auth) + Resend Free (3k/mês). |
| **Logs / Observabilidade** | **$0** (Better Stack / Grafana Cloud Free / stdout estruturado) | **$0** | Retenção curta sem PII. |
| **CI/CD (GitHub Actions)** | **$0** (<= 2.000 min/mês) | **$0** | Builds otimizados com cache. |
| **Domínio (`.com.br` / `.com`)** | ~$1–$1,50/mês (~R$ 40–70/ano) | ~$1–$1,50/mês | Registro.br / Cloudflare Registrar. |
| **Total Estimado Piloto** | **$1–$8,50** (c/ DB Free+backup próprio) ou **$33** (c/ Supabase Pro) | **~$32–$34/mês** | Custo real previsível sem lock-in. |

### C. Produção Comercial (C3 a C5)
| Item de Custo | C3 (5 / 250) | C4 (20 / 1.000) | C5 (100 / 5.000) |
| :--- | :---: | :---: | :---: |
| **Backend Go** | $7–$15 (1–2 instâncias 512MB–1GB) | $25–$50 (2 instâncias 1–2GB + LB) | $85–$160 (cluster stateless / autoscale) |
| **Banco PostgreSQL (Supabase Pro+)** | $25 (8 GB incluídos) | $25–$45 (upgrade disco/compute micro) | $75–$150 (Compute Small/Medium + PITR $100 opc.) |
| **Storage + Vídeos (Cloudflare R2)** | ~$0,10–$0,30 (15 GB; $0 egress) | ~$0,60–$1,00 (50 GB; $0 egress) | ~$3–$4 (200 GB; $0 egress!) |
| **Tráfego API (Egress)** | $0 (incluído 100GB Render/Fly) | $0–$10 | $15–$40 (depende de plataforma/proxy) |
| **Auth** | $0 (próprio) ou $0 (Supabase <= 50k MAU) | $0 | $0 (até 50k–100k MAU) |
| **SMTP (Resend / Brevo)** | $0 (3k/mês) | $0–$20 (Resend Pro $20 p/ 50k) | $20 |
| **Logs / Observabilidade / Backups** | $0–$15 | $20–$40 | $50–$100 |
| **CI/CD + Domínio** | ~$1,50 | ~$1,50–$10 | ~$20 (minutos extra Actions) |
| **Total Estimado/Mês** | **~$34–$57/mês** | **~$72–$177/mês** | **~$268–$494/mês** |

---

## 3. Governança de Gastos e Riscos de Cobrança Automática

1. **Provedores que NÃO garantem teto rígido sem cobrança se houver cartão:**
   - **Cloudflare R2 / Pages paid add-ons**, **Fly.io** (transição automática para Pay-As-You-Go após crédito de $5), **Render** (overage de bandwidth a $0,15/GB se houver cartão) e **Supabase** (spend cap precisa estar explicitamente **ON** no plano Pro).
2. **Controles Compensatórios Obrigatórios:**
   - **Nunca cadastrar cartão** nos ambientes de desenvolvimento/homologação (`Render Free`, `Supabase Free`, `Cloudflare Pages Free`, `Resend Free`).
   - Quando ativar **R2**, implementar no backend **limite rígido por arquivo (ex.: 50 MB)**, **cota total por organização** e **alerta/bloqueio automático antes de 8 GB** (margem de segurança dos 10 GB gratuitos).
   - No **Supabase Pro**, manter **Spend Cap ativado** para impedir cobrança de overage ilimitada.
   - Monitorar consumo de egress da API Render (limite 5 GB no Free / 100 GB no Starter) mantendo vídeos via presigned URL direta no R2 ($0 egress!).
