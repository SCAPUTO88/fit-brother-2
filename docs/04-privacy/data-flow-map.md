# Mapa de Fluxo de Dados Pessoais (Data Flow Map)

```mermaid
flowchart LR
    Titular["Titular (Aluno / Trainer)"]
    PWA["Flutter Web/PWA (Browser)"]
    Edge["Cloudflare DNS/WAF/Pages"]
    API["Fit Brother API (Go)"]
    DB[(Supabase PostgreSQL)]
    R2[(Cloudflare R2 - Vídeos Demo)]
    SMTP["Resend / Brevo (E-mail Transacional)"]

    Titular -- "1. Cadastro/Convite, Treino, Avaliação (HTTPS)" --> Edge --> PWA
    PWA -- "2. API Calls HTTPS (Sessão Cookie HttpOnly / CSRF)" --> Edge --> API
    API -- "3. Grava dados cifrados em trânsito (TLS) + AES em notas" --> DB
    API -- "4. Envia apenas nome/e-mail + link de ativação/reset" --> SMTP
    API -- "5. Gera Signed URL curta (5m) após validar vínculo" --> R2
    PWA -- "6. Stream/Upload direto sem trafegar dados sensíveis de saúde ao R2" --> R2
```

---

## Fronteiras e Salvaguardas por Etapa

1. **Browser (Flutter Web/PWA):** Proibido armazenar avaliações biométricas, tokens ou dados privados em cache persistente de Service Worker ou `localStorage` em claro após migração para cookies `HttpOnly` (Fase 2/8).
2. **Trânsito Internacional de Dados:** Provedores gratuitos em nuvem (Cloudflare, Render, Resend/Brevo, GitHub, e Supabase caso não alocado exclusivamente em `sa-east-1`) podem processar ou rotear dados fora do Brasil. Isso exige cláusulas de transferência internacional (Art. 33 LGPD / Resolução CD/ANPD nº 19/2024), seleção de região São Paulo (`sa-east-1`) sempre que disponível no Supabase sem custo extra e transparência explícita na Política de Privacidade.
3. **E-mail Transacional (SMTP):** Nunca incluir dados de saúde, medidas corporais, senhas ou CPF no corpo de e-mails — enviar somente nome, instrução neutra e link opaco temporário.
4. **Logs de Aplicação (`zerolog`):** Redação estrita; registrar apenas `request_id`, `user_id` (UUID), rota, status HTTP e latência.
