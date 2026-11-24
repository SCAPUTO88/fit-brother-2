# Backlog de Implementação — Fit Brother 2.0

Backlog priorizado por fase. Nenhuma tarefa abaixo está executada nesta Fase 0.

| ID | Fase | Épico | Tarefa / Entrega | Critério Resumido de Saída |
| :--- | :--- | :--- | :--- | :--- |
| BL-01 | Fase 1 | Segurança | Proteger `POST /v1/students` com `RequireAuth` + role `TRAINER` + vínculo com trainer | Anônimo e `STUDENT` recebem 401/403; teste de regressão |
| BL-02 | Fase 1 | Segurança | Adicionar migration `006` com `trainer_id` (ou relação transitória) e filtrar `GET/DELETE /students` | Trainer A nunca acessa/deleta aluno do Trainer B; `STUDENT` recebe 403 em `GET /students` |
| BL-03 | Fase 1 | Correção Crítica | Resolver `users.id -> students.id` em `WorkoutHandler` e checar ownership em sessões/exercícios | Aluno lista e lê apenas o próprio plano; erro mapeado para 403/404 |
| BL-04 | Fase 1 | Segurança Vídeo | Validar posse do exercício em `UploadURL`, `ConfirmUpload` (`HeadObject` + MIME/size) e `GetSignedURL` | Usuário sem vínculo recebe 403; objeto inexistente ou fora do prefixo é rejeitado |
| BL-05 | Fase 1 | Auth & Transações | Transações SQL em `CreateStudent`, `Anonymize` (anonimizando `users` + revogando tokens) e `Refresh`; validar `is_active` | Refresh revogado e conta desativada falham imediatamente; zero dados parciais |
| BL-06 | Fase 1 | Higiene & Flutter | Sanitizar `debug_crypto.go`/`seed_trainer.go`; corrigir `logout` com `refresh_token` e `auth_provider` no Flutter; testes integração DB reais | Zero segredos impressos em scripts; suíte `./tests/integration/...` roda contra Postgres real |
| BL-07 | Fase 2 | Simplificação | Tornar `REDIS_URL` e `R2_*` opcionais no boot; criar `StoragePort` em `domain` + `LocalStorageAdapter` | API sobe apenas com `DATABASE_URL` + chaves Auth |
| BL-08 | Fase 2 | Auth Web & Refresh | Implementar refresh interceptor no `DioClient` e estratégia de sessão Web (`HttpOnly` cookies / BFF — ADR-008) | Sessão renovada sem queda em 15m; tokens protegidos de XSS na Web |
| BL-09 | Fase 2 | DevSecOps | Modernizar CI (versões fixas, Flutter jobs, Gitleaks, Semgrep, OSV-Scanner, Trivy, SBOM, `permissions` mínimas); remover refs GCP | Pipeline completa verde sem `latest` solto |
| BL-10 | Fase 3 | Identidade & Org | Migrations `organizations`, `memberships`, `trainer_student_relationships`, `invitations` | Multi-tenancy operacional com papéis explícitos |
| BL-11 | Fase 3 | Convite & Senha | Convite por token hash único/expirável, ativação pelo aluno, reset de senha via SMTP, MFA TOTP para trainer | Convite expirado/reutilizado rejeitado |
| BL-12 | Fase 3 | LGPD & CPF | Tornar CPF opcional/deprecado (ADR-006); tabelas `consent_records`, `privacy_requests`, `audit_events` | Fluxo completo de consentimento, exportação e solicitação de exclusão |
| BL-13 | Fase 4 | Catálogo | Modelagem e importador idempotente `free-exercise-db` (Unlicense) + grupos musculares/equipamentos + tradução PT-BR | Catálogo pesquisável sem violar licenças |
| BL-14 | Fase 5 | Workout Builder | Templates, duplicação, edição, `workout_versions` imutáveis na publicação + alinhamento completo de contrato de vídeo no Flutter | Edição de ficha não corrompe histórico passado |
| BL-15 | Fase 6 | App Aluno | Navegação inferior (5 abas), treino do dia, player/mídia, cronômetro, registro de séries com idempotência e feedback | Reenvio offline/rede lenta não duplica séries |
| BL-16 | Fase 7 | Evolução & Painel | Avaliações físicas flexíveis + gráficos de evolução + KPIs reais do dashboard do trainer | Sem diagnósticos clínicos; apenas dados autorizados |
| BL-17 | Fase 8 | Design & PWA | Design System (grafite/preto/branco + verde-lima, Dark/Light, WCAG 2.2 AA), `index.html`/`manifest.json` e Service Worker seguro | Sem cache de dados privados na PWA |
| BL-18 | Fase 9–10 | Produção | Homologação fora do GCP, backups + teste de restauração, observabilidade, DAST ZAP, revisão LGPD e checklist Go/No-Go | Critérios de Go/No-Go 100% assinados |
