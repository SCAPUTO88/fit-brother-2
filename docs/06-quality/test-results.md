# Registro Contínuo de Resultados de Testes

Este arquivo registra os resultados executados ao longo das fases.

---

## Fase 0 — Baseline Inicial (2026-11-23)

- **Ambiente:** Contêiner isolado de auditoria (sem `go`, `flutter`, `dart`, `docker`, `psql`).
- **Status de execução:** Nenhum comando Go/Flutter/Docker pôde ser executado localmente neste contêiner. Detalhes completos em [`../00-discovery/baseline-test-results.md`](../00-discovery/baseline-test-results.md).

---

## Fase 1 — Security Baseline (`fix/fase-1-security-baseline`, 2026-11-24)

- **Ambiente:** Mesmo contêiner isolado (sem `go`, `flutter`, `dart`, `docker`, `psql`).
- **Verificação estática realizada no contêiner:**
  - Conferência de assinaturas, interfaces de domínio/repositório, remoção de referências obsoletas (`ListAll`, `/videos/request-upload`, `/videos/confirm`, senhas fixas em scripts) e aderência às queries SQL parametrizadas.
- **Testes adicionados/atualizados nesta fase (prontos para execução em ambiente com toolchains):**
  1. `internal/domain/student/entity_test.go`: obrigatoriedade de `TrainerID` em `NewStudent`.
  2. `internal/adapters/repository/postgres/student_repository_test.go`: `sqlmock` atualizado para `trainer_id` e filtro `ListByTrainer`.
  3. `internal/application/auth/refresh_test.go`: rotação transacional de refresh token, revogação da família inteira fora da transação em caso de reuso e bloqueio para usuário inativo/excluído.
  4. `internal/application/student/{create_student_test,get_student_test,anonymize_test}.go`: criação transacional vinculada ao treinador, propagação de erros de banco, isolamento cruzado entre treinadores/alunos e anonimização completa de PII + revogação de tokens.
  5. `internal/application/workout/usecases_test.go`: resolução de `students.id` a partir do `users.id` autenticado em `ListPlans.ByStudent` e `GetPlanDetails`, além de bloqueio cruzado em planos, sessões e exercícios.
  6. `internal/application/video/usecases_test.go`: validação de propriedade do exercício, allowlist de MIME, sanitização de chave no R2, fluxo `PENDING` → `HeadObject` → `CONFIRMED` e assinatura condicionada ao vínculo.
  7. `internal/adapters/http/middleware/auth_middleware_test.go`: validação em tempo real de usuário ativo e cabeçalhos HTTP de segurança.
  8. `internal/adapters/http/handler/student_handler_test.go`: bloqueio de `POST /students` sem sessão (`401`) ou como aluno (`403`), resposta `201 Created` para treinador e isolamento cruzado em `GET`/`DELETE`.
  9. `tests/integration/security_baseline_test.go`: suíte ponta a ponta com PostgreSQL aplicando migrations `001`–`006` e validando isolamento de dados, treinos de aluno, upload/playback de vídeo e anonimização LGPD.
  10. `fit_brother_app/test/widget_test.dart`: verificação do redirecionamento automático de rotas protegidas para `/login` quando não autenticado.
- **Comandos obrigatórios no ambiente local/CI para fechar a verificação numérica:**
  ```bash
  go test -v -race -count=1 ./internal/...
  DATABASE_URL=postgres://fitbrother:secret@localhost:5432/fitbrother_test?sslmode=disable go test -v -race -count=1 -tags integration ./tests/integration/...
  cd fit_brother_app && flutter analyze && flutter test
  ```
