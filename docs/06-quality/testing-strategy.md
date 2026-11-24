# Estratégia de Testes e Qualidade — Fit Brother 2.0

## 1. Pirâmide de Testes Alvo

| Camada | Ferramentas | Escopo |
| :--- | :--- | :--- |
| **Unitários Backend** | `testing`, `testify`, table-driven tests, `-race` | Regras de domínio, criptografia, JWT, validações, use cases isolados e fuzzing (`Fuzz*`) em parsers/CPF/inputs críticos. |
| **Integração & Contrato DB** | `httptest`, PostgreSQL descartável / Testcontainers | Repositórios reais, migrations `001`–`012`, constraints FK/UNIQUE, transações SQL e testes HTTP completos com router + middleware. |
| **Segurança & Regressão (RED/GREEN)** | Suíte dedicada `tests/security/` ou `*_authz_test.go` | Garante matematicamente os 20+ cenários obrigatórios de segurança e LGPD. |
| **Frontend Flutter** | `flutter_test`, widget tests, provider/GoRouter tests, integration tests Web | Fluxos de login/refresh/logout, guards por perfil, responsividade (desktop vs mobile), acessibilidade e fluxos offline/idempotência. |
| **DevSecOps / DAST** | `golangci-lint`, `govulncheck`, `gitleaks`, `semgrep`, `osv-scanner`, `trivy`, SBOM, OWASP ZAP (staging) | Gates automáticos no GitHub Actions bloqueando merge/release em vulnerabilidades críticas/altas. |

---

## 2. Cenários Obrigatórios de Regressão (Checklist)

Todo cenário abaixo deverá possuir teste automatizado vinculado até a conclusão da respectiva fase:
1. Usuário anônimo não cadastra alunos (`POST /v1/students` retorna `401`).
2. Aluno não lista outros alunos (`GET /v1/students` retorna `403`).
3. Aluno não acessa dados físicos/perfil de terceiros (`403`).
4. Treinador A não acessa/edita/anonimiza alunos do Treinador B (`403`).
5. Treinador A não altera treinos/sessões/exercícios do Treinador B (`403`).
6. Aluno não altera prescrições (`403`).
7. Usuários não elevam privilégios via payload ou claims (`403`/ignorado).
8. IDs manipulados (UUIDs de terceiros ou inexistentes) não permitem acesso indevido (`403`/`404`).
9. Usuário não autorizado não emite URLs assinadas de playback/upload (`403`).
10. Uploads não podem referenciar objetos ou caminhos arbitrários (`400`/`403` + `HeadObject`).
11. Convites expirados são rejeitados (`410`/`400`).
12. Convites já aceitos não podem ser reutilizados (`409`/`410`).
13. Refresh tokens revogados não podem ser reutilizados (revoga sessão).
14. Contas inativas/desativadas não mantêm acesso indevido.
15. Logout limpa dados privados no Flutter e revoga refresh token no servidor.
16. Prescrições publicadas preservam histórico executado (`workout_versions` imutável).
17. Reenvios de execução de treino não duplicam sessões (`client_idempotency_key`).
18. Dados sensíveis/PII não aparecem em logs (`zerolog` + regressão de `login.go`).
19. Credenciais/segredos não aparecem nos bundles Web Flutter nem no repositório (`gitleaks`).
20. Backups podem ser restaurados em PostgreSQL limpo (teste mensal documentado).
