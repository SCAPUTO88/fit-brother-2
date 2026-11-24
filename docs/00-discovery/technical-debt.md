# Inventário de Dívida Técnica

Classificação dos débitos técnicos observados diretamente no código em `develop` (`16a8e69`).

---

## 1. Dívida Crítica de Arquitetura e Contrato

| ID | Área | Descrição | Impacto | Fase Alvo |
| :--- | :--- | :--- | :--- | :--- |
| TD-01 | Domínio / DB | Ausência de vínculo `trainer_id` ou `trainer_student_relationships` em `students` | Impossibilita isolamento multi-treinador (BOLA crítico) | Fase 1 / Fase 3 |
| TD-02 | Domínio / AuthZ | Confusão entre `users.id` e `students.id` nas rotas de `workouts` | Alunos autenticados nunca enxergam ou abrem os próprios treinos | Fase 1 |
| TD-03 | Backend / Video | `application/video/usecases.go` importa `adapters/storage` e depende de `*storage.R2Storage` | Viola Arquitetura Hexagonal e impede testes/execução sem S3/R2 | Fase 2 |
| TD-04 | API ↔ Flutter | Contrato de upload de vídeo diverge totalmente (`request-upload`/`confirm` vs `upload-url`/`videos`) | Upload de vídeo do app Flutter falha sempre com 404/400 | Fase 1 / Fase 2 |
| TD-05 | Config / Boot | `config.Load()` exige `REDIS_URL` (não usado) e 5 variáveis `R2_*` com panic/fatal | Impede subida simples local/homologação sem Redis/R2 | Fase 2 |

---

## 2. Dívida de Confiabilidade e Consistência Transacional

| ID | Área | Descrição | Impacto | Fase Alvo |
| :--- | :--- | :--- | :--- | :--- |
| TD-06 | Auth Refresh | `refresh.go` revoga o refresh token antigo antes de validar se usuário está ativo e sem transação SQL | Race conditions e tokens órfãos em falhas parciais | Fase 1 |
| TD-07 | Create Student | `create_student.go` insere `users` e `students` em queries separadas e ignora erros de `GetByEmail`/`GetByCPFHash` (`_`) | Pode deixar `users` órfão sem registro `students` ou mascarar erro de DB | Fase 1 |
| TD-08 | Anonimização LGPD | `anonymize.go` desativa `users` e deleta `students` sem transação; não anonimiza `users.email`/`full_name` nem revoga tokens | Inconsistência em falhas parciais e retenção indevida de PII | Fase 1 |
| TD-09 | Middleware Auth | `auth_middleware.go` confia apenas na assinatura do JWT até expirar (15m), sem verificar `is_active`/revogação | Conta desativada/anonimizada continua operando até expiração do access token | Fase 1 |
| TD-10 | Error Handling | Handlers retornam `err.Error()` diretamente em respostas JSON (`500`/`400`) e mapeiam `"acesso negado"` como HTTP `400` | Vazamento de detalhes internos e semântica REST incorreta (`403`/`404`) | Fase 1 |

---

## 3. Dívida de Frontend (Flutter)

| ID | Área | Descrição | Impacto | Fase Alvo |
| :--- | :--- | :--- | :--- | :--- |
| TD-11 | Auth Provider | `auth_provider.dart` seta `isLoggedIn = true` sem validar presença de tokens; `logout()` não envia `refresh_token` | Estado inconsistente no app e refresh tokens nunca revogados no logout | Fase 1 |
| TD-12 | Dio Client | `onError` para HTTP `401` é apenas um comentário stub; sem fila de refresh | Sessão cai abruptamente após 15 minutos de uso | Fase 2 |
| TD-13 | Navegação | `app_router.dart` sem `redirect` guards de autenticação ou separação por role (`TRAINER` vs `STUDENT`) | Rotas internas acessíveis sem estado autenticado válido | Fase 1 |
| TD-14 | UI Morta / Incompleta | Drawer "Vídeos" sem `onTap`, clique em aluno no dashboard sem detalhe, zero player de vídeo, zero telas de aluno | Experiência incompleta e atalhos não funcionais | Fase 5–8 |
| TD-15 | Web / PWA | `web/index.html` e `web/manifest.json` com boilerplate Flutter padrão e `flutter_secure_storage` sobre `localStorage` | Marca não configurada e exposição a XSS se tokens permanecerem acessíveis ao JS | Fase 2 / Fase 8 |

---

## 4. Dívida de Pipeline, Testes e Operação

| ID | Área | Descrição | Impacto | Fase Alvo |
| :--- | :--- | :--- | :--- | :--- |
| TD-16 | Testes Integração | `tests/integration/health_test.go` é apenas um `t.Log(...)` sem banco real | Zero validação de SQL, migrations, joins, constraints e FK em CI | Fase 1 |
| TD-17 | CI/CD | `.github/workflows/ci.yml` usa `latest` em ferramentas, não restringe `permissions:`, não roda Flutter nem SAST/secrets/SBOM | Baixa reprodutibilidade e risco de regressão silenciosa no Flutter/segurança | Fase 2 |
| TD-18 | Scripts Locais | `scripts/debug_crypto.go` imprime `PASSWORD_PEPPER` em `stdout`; `seed_trainer.go` possui senha padrão hardcoded | Risco de exposição operacional acidental de segredo | Fase 1 |
| TD-19 | `main` vs `develop` | `origin/main` está 4 meses/sprints atrás de `develop` | Risco de deploy acidental de código antigo ou confusão de onboarding | Fase 1 / 2 |
| TD-20 | `.gitignore` | `.gitignore` ignora `docs/` e `AGENTS.md` | Documentação estratégica não é versionada por padrão | Fase 0 (Decisão) |
