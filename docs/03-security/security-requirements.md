# Requisitos de Segurança e Hardening (ASVS / OWASP Baseline)

Estes requisitos funcionam como critérios obrigatórios de aceite para todas as entregas a partir da Fase 1.

1. **Autenticação e Sessão:**
   - Endpoints públicos restritos a `/v1/health`, `/v1/auth/login`, `/v1/auth/refresh`, `/v1/auth/logout`, `/v1/auth/activate` e `/v1/auth/password-reset*` (todos com rate limiting).
   - Refresh token armazenado apenas em hash (`SHA-256`), rotacionado dentro de transação SQL com detecção de reuso (revogação da família se token já revogado for apresentado).
   - Logout revoga o refresh token correspondente e limpa cookies/estado local no Flutter.
   - Contas desativadas (`is_active = false` ou `deleted_at IS NOT NULL`) perdem acesso imediatamente.
   - Sessão Web em produção protegida contra XSS (`HttpOnly; Secure; SameSite` via mesmo site/BFF — ver [ADR-008](../adr/ADR-008-web-session-security-bff-cookies.md)) com proteção anti-CSRF por validação de `Origin`/`Referer` + token CSRF custom header.
2. **Autorização (RBAC + Ownership Multi-Tenant):**
   - Checagem obrigatória em camada de aplicação: papel (`OWNER`/`TRAINER`/`STUDENT`) **E** pertencimento à organização/vínculo ativo no banco.
   - Nenhum endpoint recebe `trainer_id` ou `student_id` confiável sem validar contra o token autenticado e o banco.
   - Respostas HTTP corretas: `401` (não autenticado), `403` (sem permissão/posse), `404` (não encontrado — sem vazar existência quando aplicável).
3. **Proteção de Dados e Criptografia:**
   - Zero CPF obrigatório em novos fluxos; campos sensíveis biométricos tratados com controle estrito.
   - Proibido logar tokens, senhas, peppers, chaves AES, CPFs, e-mails ou corpos de requisição sensíveis.
   - Erros HTTP sanitizados com código estável de erro (ex.: `{"error":"forbidden","code":"AUTH_FORBIDDEN"}`) sem vazar mensagens internas de SQL/driver.
4. **Mídias e Uploads (R2 / StoragePort):**
   - Allowlist estrita de MIME (`video/mp4`, `video/webm`) e limite de tamanho (ex.: 50 MB inicialmente).
   - Chave do objeto gerada 100% pelo servidor (`orgs/{org_id}/exercises/{exercise_id}/{uuid}.{ext}`), ignorando paths enviados pelo cliente.
   - Confirmação obrigatória via `HeadObject` verificando existência, `ContentLength` e `ContentType` real.
   - URLs assinadas de leitura com TTL máximo de 5 minutos e verificação de vínculo ativo.
5. **Cabeçalhos HTTP e PWA:**
   - `X-Content-Type-Options: nosniff`, `Referrer-Policy: strict-origin-when-cross-origin`, `X-Frame-Options: DENY` (ou `frame-ancestors 'none'` via CSP), HSTS em produção e CORS explícito sem wildcard `*` quando `AllowCredentials=true`.
   - Service Worker não pode armazenar respostas autenticadas da API nem URLs assinadas de vídeos.
6. **Pipeline e Segredos:**
   - Versões fixas nas GitHub Actions, `permissions` mínimas, `gitleaks`, `semgrep`, `govulncheck`, `osv-scanner`, `trivy` e geração de SBOM.
