# Resultados da Baseline de Testes e Builds (Fase 0)

**Data da verificação:** 2026-11-23  
**Ambiente:** Contêiner isolado de análise (Linux, `git`, `python3`, `unzip`/`tar` disponíveis).  
**Ferramentas ausentes neste contêiner:** `go`, `flutter`, `dart`, `docker`, `psql`.

---

## 1. Declaração de Transparência e Bloqueio de Execução

Em conformidade estrita com as regras da Fase 0 (**não simular resultados nem esconder falhas**):
- **Nenhuma suíte de testes Go, build Docker, `flutter analyze`, `flutter test` ou migração PostgreSQL pôde ser executada neste ambiente isolado**, pois os toolchains `go`, `flutter`, `dart`, `docker` e `psql` não estão instalados no contêiner atual.
- Os comandos padrão do projeto foram inspecionados nos arquivos de build/CI, e **todos os 14 arquivos de teste existentes** foram auditados estaticamente linha a linha.

### Comandos Previstos (Não Executados Neste Contêiner)
- Backend unitário: `go test -v -race -count=1 ./internal/...` → **NÃO EXECUTADO (`go: command not found`)**
- Backend integração: `go test -v -race -count=1 -tags integration ./tests/integration/...` → **NÃO EXECUTADO (`go` e `docker/postgres` ausentes)**
- Backend lint / vulnerabilidades: `golangci-lint run`, `gofmt -l .`, `govulncheck ./...` → **NÃO EXECUTADO**
- Backend build: `CGO_ENABLED=0 GOOS=linux go build ./cmd/api` → **NÃO EXECUTADO**
- Frontend Flutter: `cd fit_brother_app && flutter analyze && flutter test && flutter build web` → **NÃO EXECUTADO (`flutter: command not found`)**

---

## 2. Inventário Auditado dos Testes Existentes (Inspeção Estática)

### Backend Go (`12` arquivos)

| Arquivo | O que Testa na Prática | Limitações / Lacunas Encontradas |
| :--- | :--- | :--- |
| `internal/infrastructure/config/config_test.go` | Precedência `APP_PORT` vs `PORT` (teste ainda nomeado `...RailwayPort...`) | Não valida comportamento sem `REDIS_URL`/`R2_*`. |
| `internal/infrastructure/crypto/crypto_test.go` | Criptografia/decriptação AES-256-GCM, hash HMAC determinístico de CPF e bcrypt+pepper | Bom isolamento unitário. |
| `internal/infrastructure/jwt/service_test.go` | Geração/validação RS256 com par de chaves temporário e expiração | Não cobre rotação (`kid`) nem blacklist/revogação. |
| `internal/domain/student/entity_test.go` | Validação de e-mail, senha e algoritmo completo de dígitos verificadores de CPF | Reforça obrigatoriedade atual de CPF (a revisar na Fase 3). |
| `internal/domain/workout/entity_test.go` | Regras de entidade de `WorkoutPlan`, `WorkoutSession`, `Exercise` | Não testa versionamento/templates (inexistentes ainda). |
| `internal/application/auth/login_test.go` | Fluxo de login e **regressão confirmada**: garante que `passwordHash` não vaza nos logs em falha | Não testa login de usuário deletado/bloqueio por brute force. |
| `internal/application/student/create_student_test.go` | Criação de aluno com mocks de repositório e cripto | Mocks escondem ausência de transação SQL e erros ignorados em `GetByEmail`/`GetByCPFHash`. |
| `internal/application/student/get_student_test.go` | Busca e decriptação de aluno via mocks | Não contempla vínculo com `trainer_id`. |
| `internal/application/student/anonymize_test.go` | Desativação e delete via mocks | Não detecta que `users.email`/`full_name` e `refresh_tokens` ficam intactos. |
| `internal/application/workout/usecases_test.go` | Casos de uso de criação e leitura de treinos | Não cobre tentativa de treinador A alterar sessão/exercício de treinador B; valida `student.ID == tokenUserID` mascarando o bug real `users.id != students.id`. |
| `internal/adapters/repository/postgres/student_repository_test.go` | Queries SQL via `go-sqlmock` | Apenas valida SQL mockado; não valida schema real do Postgres. |
| `internal/adapters/http/router_test.go` | Apenas `GET /v1/health` (200) | Não testa autenticação/autorização de nenhuma rota protegida. |
| `internal/adapters/http/handler/student_handler_test.go` | `POST /v1/students` com mock de usecase | Testa a rota **sem middleware de auth**, refletindo a vulnerabilidade atual. |
| `tests/integration/health_test.go` | **Placeholder vazio** (`t.Log("Ambiente de Integração Configurado e Operacional")`) | **Zero testes reais** de conexão SQL, migrations, FKs ou PostgreSQL. |

### Frontend Flutter (`fit_brother_app/test/` — `2` arquivos)

| Arquivo | O que Testa na Prática | Limitações / Lacunas Encontradas |
| :--- | :--- | :--- |
| `test/widget_test.dart` | Smoke test que monta `FitBrotherApp` dentro de `ProviderScope` | Cobre apenas renderização inicial da tela de login. |
| `test/core/api/dio_client_test.dart` | Instanciação básica do `DioClient` | Não testa interceptors, headers, logout nem refresh flow. |

---

## 3. Conclusão da Baseline

- A cobertura de testes atual é **insuficiente como barreira de segurança e contrato**:
  - Não existem testes de contrato entre `workout_handler` e `workout_repository` (onde o bug `users.id != students.id` teria falhado imediatamente contra Postgres real).
  - Não existem testes de isolamento multi-treinador (`Trainer A` vs `Trainer B`), autorização em vídeos (`video_handler`) ou segurança em `POST /students`.
  - A suíte de integração (`tests/integration/`) precisa ser implementada de verdade com PostgreSQL descartável (Testcontainers ou service container no CI) já na **Fase 1**.
