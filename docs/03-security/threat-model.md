# Modelagem de Ameaças (Threat Model — STRIDE)

Análise baseada na arquitetura TO-BE do Fit Brother 2.0 cobrindo API Go, Flutter Web/PWA, PostgreSQL, Storage R2 e fluxo de e-mail/convites.

---

## 1. Ativos Críticos a Proteger

1. **Dados Pessoais Sensíveis (LGPD Art. 11):** Peso, medidas, avaliações físicas, histórico de treinos e observações do aluno.
2. **Dados de Identificação (em fase de remoção):** CPF (`cpf_encrypted`/`cpf_hash`), e-mail, telefone, contato de emergência.
3. **Credenciais e Chaves:** Chaves privadas JWT (`RS256`), `AES_KEY`, `CPF_HMAC_SECRET`, `PASSWORD_PEPPER`, credenciais R2 e `DATABASE_URL`.
4. **Integridade de Prescrições e Execuções:** Fichas publicadas e séries registradas pelos alunos.
5. **Mídias Privadas de Exercícios:** Vídeos enviados pelos treinadores no bucket R2.

---

## 2. Matriz STRIDE e Mitigações Obrigatórias

| Categoria STRIDE | Ameaça Concreta no Fit Brother | Vetor / Superfície | Contramedida Arquitetural Obrigatória | Fase |
| :--- | :--- | :--- | :--- | :--- |
| **S — Spoofing** | Força bruta em `/auth/login`; uso de convite/reset interceptado; roubo de JWT/refresh no navegador via XSS | Auth endpoints, PWA WebCrypto/`localStorage`, links de e-mail | Rate limit por IP+e-mail; cookies `HttpOnly; Secure; SameSite` (mesmo site/BFF); tokens de convite/reset armazenados como SHA-256 com uso único e TTL curto; MFA TOTP opcional para trainer. | F1–F3 |
| **T — Tampering** | Trainer A altera ficha/exercício do Trainer B; aluno altera prescrição publicada; cliente forja `size_bytes`/`r2_key` em `ConfirmUpload`; alteração de payload em trânsito | `workout_handler`, `video_handler`, CORS/CSRF | Validação estrita de posse no DB; snapshots imutáveis `workout_versions`; `HeadObject` obrigatório antes de confirmar upload; CSRF token/Origin check quando usar cookies; TLS obrigatório. | F1, F3, F5 |
| **R — Repudiation** | Exclusão/edição sem rastro de quem alterou ficha ou solicitou remoção LGPD; duplicação de execuções por reenvio de rede | Prescrições, anonimização, logs de execução | Tabela `audit_events` para ações críticas (convite, vínculo, publicação, anonimização/exportação); `client_idempotency_key` único nas execuções de treino. | F3, F6 |
| **I — Information Disclosure** | `GET /students/:id` vaza CPF entre treinadores; `GET /students` lista todos para qualquer aluno; `GetSignedURL` vaza vídeos privados; `debug_crypto.go` imprime pepper; erros SQL crus no HTTP | Handlers Student/Video, scripts, logs, PWA Service Worker | Isolamento por organização + relacionamento ativo; remoção do CPF obrigatório; checagem de vínculo em `GetSignedURL`; zero PII/segredos em logs; Service Worker proibido de cachear `/v1/*`; remover `debug_crypto.go`. | F1–F3, F8 |
| **D — Denial of Service** | `POST /students` aberto permitindo flood de hashes `bcrypt`; upload ilimitado de vídeos gigantes enchendo R2/banda; reenvios massivos | Endpoints públicos, presigned PUT R2, Render Free CPU/RAM | Exigir auth em `POST /students` e migrar para fluxo de convite; `ContentLengthRange` / limite máximo (ex.: 50 MB) e cota por org/trainer no upload; rate limiting por IP e usuário. | F1–F2, F4 |
| **E — Elevation of Privilege** | Manipulação de role no signup/payload; usuário desativado mantendo JWT ativo por 15m; trainer de uma org acessando outra org | Claims JWT, `users.role`, middleware | Role definida exclusivamente no servidor via `memberships`/convite; TTL curto + checagem de status `is_active`/revogação; separação explícita entre `users.id` e `students.id`. | F1, F3 |
