# Auditoria de Segurança — Fit Brother (AS-IS)

**Escopo:** Código completo da branch `develop` (`16a8e69`) e histórico Git.  
**Nota de método:** Auditoria baseada em inspeção estática linha a linha dos arquivos de backend, frontend, migrations, scripts e workflows. Nenhum achado foi fabricado; todos os pontos abaixo possuem evidência verificada no código.

---

## 1. Resumo Executivo de Severidade

| Severidade | Quantidade | Principais Riscos |
| :--- | :---: | :--- |
| **Crítica** | 4 | Cadastro público sem auth (`POST /students`), acesso universal a CPF descriptografado, bug `users.id != students.id` travando alunos, URL assinada de vídeo sem checagem de posse. |
| **Alta** | 7 | Listagem global de alunos para qualquer role, delete sem vínculo de trainer/sem transação/sem anonimizar `users`, manipulação de treinos alheios, upload R2 sem validação/HeadObject, logout/refresh quebrados, segredo impresso em script dev. |
| **Média** | 5 | Tokens em `localStorage` na Web (`flutter_secure_storage`), ausência de checagem `is_active` durante TTL do JWT (15m), erros internos crus no HTTP, CI com `latest` sem SAST/secrets/Flutter, ausência de rate limiting e headers de segurança. |

---

## 2. Validação dos Achados Preliminares + Novos Achados Confirmados

| ID | Sev. | Achado e Evidência Exata | Risco / Exploração |
| :--- | :--- | :--- | :--- |
| **SEC-01** | **CRÍTICA** | **`POST /v1/students` público sem autenticação:** `internal/adapters/http/router.go:99` e `student_handler.go:37` registram `group.POST("", h.Create)` fora do grupo `protected`. | Qualquer atacante anônimo na internet cria contas e alunos, consome CPU com bcrypt e grava dados no Postgres. |
| **SEC-02** | **CRÍTICA** | **IDOR/BOLA em `GET /v1/students/:id` com exposição de CPF:** `student_handler.go:65-91` bloqueia `STUDENT` lendo outro aluno, mas permite que **qualquer `TRAINER` leia CPF descriptografado de qualquer aluno** (não há `trainer_id` na tabela `students`). | Um personal trainer malicioso ou comprometido enumera UUIDs e extrai CPFs e dados de alunos de todos os outros treinadores. |
| **SEC-03** | **CRÍTICA** | **Bug lógico/AuthZ quebrando treinos de alunos:** `migrations/004:4` referencia `students(id)`, mas `workout_handler.go:110` e `workout/usecases.go:196` comparam `student_id` com `claims.UserID` (`users.id`). | Como `users.id != students.id`, alunos autenticados recebem lista vazia ou `"acesso negado"` (HTTP `400`) nos próprios treinos. |
| **SEC-04** | **CRÍTICA** | **URL assinada de playback sem validação de posse:** `video_handler.go:92-113` e `video/usecases.go:109-129` (`GetSignedURLUseCase.Execute`) não recebem `claims.UserID`/`Role` nem verificam se o usuário é dono ou aluno daquele exercício. | Qualquer usuário autenticado (inclusive um aluno sem vínculo) obtém URL assinada do R2 para qualquer `exercise_id`. |
| **SEC-05** | **ALTA** | **Listagem global sem RBAC/filtro (`GET /v1/students`):** `student_handler.go:124-138` apenas verifica existência de claims (`_`) sem checar se `Role == TRAINER` nem filtrar por trainer. | Um usuário `STUDENT` lista todos os alunos cadastrados (nomes, e-mails, IDs, telefone). |
| **SEC-06** | **ALTA** | **Exclusão/Anonimização insegura e incompleta (`DELETE /v1/students/:id`):** `student_handler.go:93-122` e `anonymize.go:25-44`: (a) trainer deleta qualquer aluno; (b) na branch `STUDENT`, `out, _ := h.getUC.Execute(...)` ignora erro; (c) sem transação SQL; (d) mantém `users.email` e `users.full_name` em claro, não preenche `deleted_at` e não revoga `refresh_tokens`. | Violação de isolamento + retenção indevida de PII após pedido LGPD + estado inconsistente em erro parcial. |
| **SEC-07** | **ALTA** | **IDOR/BOLA em Sessões e Exercícios:** `CreateSessionUseCase` (`workout/usecases.go:59-88`) e `CreateExerciseUseCase` (`98-131`) não recebem `trainerID` nem checam posse do plano pai; `CreatePlan` não checa posse do `student_id`. | Treinador A injeta sessões ou exercícios nos planos do Treinador B. |
| **SEC-08** | **ALTA** | **Upload URL sem posse + path injection + metadata spoofing:** `video/usecases.go:41-97`: (a) não valida posse do `exercise_id`; (b) concatena `input.Filename` cru na `r2_key`; (c) `ConfirmUpload` aceita `r2_key`, `size_bytes` e `mime_type` do cliente sem `HeadObject` no R2 nem intenção prévia persistida. | Escrita/associação de chaves arbitrárias no bucket e registro de metadados falsos (ex.: HTML/executável disfarçado). |
| **SEC-09** | **ALTA** | **Fluxos Auth sem transação e falhas no Flutter:** `refresh.go:34-65` revoga o token antes de checar `usr.IsActive` e fora de transação; `create_student.go` cria `user` e `student` sem transação e ignora erros `_` nas consultas de duplicidade; `auth_provider.dart:54` marca `isLoggedIn=true` sem validar tokens; `logout` no app não envia `refresh_token`. | Contas órfãs, revogação quebrada no logout e sessões inconsistentes. |
| **SEC-10** | **ALTA** | **Exposição de segredo em script local:** `scripts/debug_crypto.go:25-26` imprime `PASSWORD_PEPPER` em hex e string pura no `stdout`; `seed_trainer.go:40` usa senha padrão `"Senha123!"`. | Vazamento acidental de pepper global se executado em logs/terminais compartilhados. |
| **SEC-11** | **ALTA (RESOLVIDO NO CÓDIGO)** | **Log histórico de `passwordHash` no login:** `login.go:40-54` **já foi corrigido** e possui teste de regressão em `login_test.go` (`TestLoginUseCaseDoesNotLogPasswordHashOnInvalidPassword`). | Manter regressão ativa no CI; verificar se não há logs sensíveis em outros fluxos. |
| **SEC-12** | **MÉDIA** | **Tokens Web via `flutter_secure_storage` (`localStorage` + WebCrypto):** `dio_client.dart`, `auth_provider.dart`. | Em caso de XSS no Flutter Web/dependência JS, `access_token` e `refresh_token` podem ser exfiltrados. |
| **SEC-13** | **MÉDIA** | **Janela de token após desativação + falta de `kid`:** `auth_middleware.go` não verifica `is_active` nem revogação em janela ativa; `jwt/service.go` não emite `kid` ou `aud`. | Conta desativada mantém acesso por até 15 minutos. |
| **SEC-14** | **MÉDIA** | **Exposição de `err.Error()` e semântica HTTP:** `workout_handler.go:87` retorna HTTP 400 para `"acesso negado"`/`"não encontrado"`; vários handlers retornam `err.Error()` cru no HTTP 500/400. | Auxilia enumeração de estados internos/SQL. |
| **SEC-15** | **MÉDIA** | **Ausência de Rate Limiting, Security Headers e Scanners CI:** Sem rate limit em `/auth/login`, `/auth/refresh`; sem CSP/HSTS/X-Content-Type-Options; `.github/workflows/ci.yml` sem `gitleaks`, `semgrep`, `osv-scanner`, `trivy` e Flutter. | Vulnerabilidade a força bruta/credential stuffing e regressões não detectadas. |
