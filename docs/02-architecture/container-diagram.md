# Diagrama de Contêineres (C4 Nível 2 — TO-BE)

```mermaid
C4Container
    title C4 Nível 2 — Contêineres (Arquitetura Alvo Fora do GCP)
    Person(trainer, "Personal Trainer", "Desktop / Tablet Browser")
    Person(student, "Aluno", "Mobile Browser / PWA")

    System_Boundary(edge, "Edge / Proxy Mesmo Domínio (Cloudflare)") {
        Container(cdn, "Cloudflare DNS / WAF / Proxy", "Edge", "Terminação TLS, roteamento app.<domínio> e proxy /api para mesma origem de cookies")
        Container(pwa, "Fit Brother Web/PWA", "Flutter Web + Service Worker Seguro", "SPA responsiva (Trainer + Aluno); zero tokens expostos ou cache de dados privados")
    }

    System_Boundary(app_host, "Runtime API (Render / Fly.io / Railway)") {
        Container(api, "Fit Brother API", "Go 1.25+/1.26 + Gin (Hexagonal)", "Auth/Sessão HttpOnly + CSRF, RBAC/ABAC, Prescrição, Execução, Avaliações, LGPD, Signed URLs")
    }

    System_Boundary(data, "Persistência e Serviços Gerenciados") {
        ContainerDb(pg, "PostgreSQL Gerenciado", "Supabase PostgreSQL", "Dados relacionais, versões de treino, logs de séries, auditoria e consentimentos")
        Container(r2, "Object Storage", "Cloudflare R2 (Opcional / StoragePort)", "Vídeos privados e catálogo de mídias")
        Container(smtp, "SMTP Transacional", "Resend / Brevo", "Convites, ativação de conta e recuperação de senha")
    }

    Rel(trainer, cdn, "HTTPS")
    Rel(student, cdn, "HTTPS")
    Rel(cdn, pwa, "Serve estáticos (Cloudflare Pages)")
    Rel(cdn, api, "Encaminha /v1/* (Same-Site Cookie / CORS estrito)")
    Rel(api, pg, "SQL transacional via DATABASE_URL", "pgx/v5 TLS")
    Rel(api, r2, "Presigned URLs + HeadObject (via StoragePort)", "HTTPS S3")
    Rel(pwa, r2, "Upload/Stream direto com URL curta autorizada", "HTTPS")
    Rel(api, smtp, "Envio de convites e reset", "HTTPS/TLS")
```

### Notas Arquiteturais Importantes
1. **Roteamento Same-Site (`/v1/*` ou subdomínio sob mesmo eTLD+1):** Essencial para usar cookies `HttpOnly; Secure; SameSite=Lax` sem sofrer bloqueios de cookies third-party (especialmente Safari/iOS ITP) e sem obrigar um segundo servidor Node separado apenas para BFF — o proxy/CDN encaminha `/v1/*` para a API Go ou o próprio subdomínio compartilhado usa `SameSite=Lax` + CSRF token.
2. **Service Worker:** Configurado para fazer cache apenas de shell estático/fontes/assets públicos da UI, **nunca** interceptando/armazenando respostas de `/v1/*` ou URLs assinadas de vídeos.
3. **Sem Redis por padrão:** Rate limit e revogação de refresh tokens operam em memória (por instância) + PostgreSQL, mantendo apenas 1 serviço backend stateless para rodar dentro do free-tier.
