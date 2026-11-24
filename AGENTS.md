# AGENTS.md — Regras Permanentes de Desenvolvimento do Fit Brother 2.0

Este documento define as regras obrigatórias para qualquer agente de engenharia ou desenvolvedor atuando no repositório **Fit Brother 2.0**.

---

## 1. Idioma, Postura e Evidências

- **Idioma padrão:** Português do Brasil (PT-BR) em documentação, planos, revisões e comunicações de engenharia. Identificadores de código, rotas SQL/HTTP e Conventional Commits seguem inglês técnico padrão do repositório.
- **Honestidade técnica absoluta:** Nunca simule outputs de testes, builds, cobertura, deploys ou inspeções. Quando uma ferramenta não puder ser executada no ambiente atual, registre o comando exato, o motivo do bloqueio e os resultados apenas do que foi realmente verificado.
- **Nada de código sem evidência:** Não declare funcionalidades, correções de segurança ou critérios de aceitação como concluídos sem evidência objetiva (testes automatizados passando, build limpo, revisão de autorização).

---

## 2. Controle de Versão, Branches e Escopo

- **Branch base oficial de retomada:** `develop` (ver `docs/adr/ADR-001-base-branch-resumption.md`). A branch `main` encontra-se desatualizada (parada na Sprint 2 / commit `565e121`).
- **Regra do `.gitignore`:** Atualmente `.gitignore` (linhas 50–51) ignora `AGENTS.md` e `docs/`. Não faça merge nem altere arquivos fora do escopo autorizado; para versionar a documentação da Fase 0+, obtenha aprovação explícita para atualizar `.gitignore` (ver `docs/07-project/decisions.md`).
- **Commits e PRs:** Use **Conventional Commits** (`feat:`, `fix:`, `docs:`, `security:`, `refactor:`, `test:`, `chore:`, `ci:`), trabalhe em branches dedicadas por entrega e **nunca realize `git push` ou `merge` sem aprovação explícita**.
- **Sem mudanças paralelas fora da fase:** Respeite o roadmap por fases (`docs/07-project/roadmap.md`). Não misture grandes refatorações com entregas de segurança ou novas features.

---

## 3. Fluxo Obrigatório por Tarefa

Para cada tarefa após autorização da fase:
1. Analisar o contexto e ler `docs/07-project/current-state.md`.
2. Apresentar proposta enxuta e mapear riscos/impactos em contrato de API e schema.
3. Definir critérios de aceitação testáveis.
4. Escrever testes primeiro (TDD / RED-GREEN, obrigatório em toda correção de segurança).
5. Implementar respeitando Arquitetura Hexagonal e isolamento multi-tenant.
6. Executar validações reais (`gofmt`, `go test -race ./...`, `golangci-lint`, `govulncheck`, `flutter analyze`, `flutter test`).
7. Revisar autorização, privacidade (LGPD) e ausência de segredos/dados sensíveis em logs.
8. Atualizar documentação afetada e `docs/07-project/current-state.md`.

---

## 4. Regras de Segurança e Privacidade (Invioláveis)

- **Default Deny em rotas e objetos:** Toda rota fora `/v1/health`, `/v1/auth/login`, `/v1/auth/refresh`, `/v1/auth/logout` e fluxos públicos explícitos de convite/ativação exige autenticação (`RequireAuth`) e validação explícita de papel (**RBAC**) + posse (**ownership / vínculo ativo**).
- **Zero IDOR / BOLA:** Nunca confie em IDs enviados pelo cliente sem verificar no backend se o usuário autenticado tem relação ativa autorizada com o recurso (ex.: treinador vinculado ao aluno, aluno dono do plano/mídia).
- **Sem `users.id` vs `students.id` ambíguos:** Sempre resolva explícita e atomicamente a identidade do domínio (`users.id` ↔ `students.id` ↔ `trainer_student_relationships`).
- **Zero exposição de dados sensíveis:**
  - Não exiba CPF ou biometria em listagens, dashboards ou logs.
  - Não imprima segredos (`PASSWORD_PEPPER`, `AES_KEY`, chaves RSA, tokens) em `stdout` ou logs.
  - Scripts utilitários (como `scripts/debug_crypto.go`) que imprimam segredos devem ser removidos ou sanitizados na Fase 1.
- **Minimização LGPD:** Não torne CPF obrigatório em novos cadastros de alunos; planeje a remoção gradual do CPF conforme `docs/adr/ADR-006-lgpd-cpf-removal.md`.
- **Uploads de mídia:** Toda URL assinada deve validar posse do exercício, prefixo/chave controlada pelo servidor, MIME/tamanho permitidos e confirmação via `HeadObject` no storage antes de persistir metadados.
- **Segurança Web (PWA/Flutter Web):** Não trate `flutter_secure_storage` no navegador como equivalente a Keychain/Keystore nativo. Consulte `docs/adr/ADR-008-web-session-security-bff-cookies.md` antes de publicar sessão Web em produção.

---

## 5. Banco de Dados e Migrations

- **Nunca altere migrations já aplicadas** (`001` a `005`). Toda evolução deve usar novas migrations incrementais (`006+`).
- **Nunca execute `go run ./scripts/migrate.go down` em bancos persistentes** (homologação ou produção): o script reverte todas as migrations destrutivamente.
- Operações multi-tabela críticas (ex.: criar usuário + aluno, anonimizar conta + revogar sessões, rotacionar refresh token) devem ocorrer dentro de **transações SQL explícitas**.

---

## 6. Arquitetura e Qualidade de Código

- **Backend Go (`internal/`):** Preserve a separação Hexagonal:
  - `domain`: entidades, regras puras e interfaces (ports). Proibido importar `gin`, `sqlx`, `aws-sdk`, `pgx`.
  - `application`: casos de uso e DTOs. Deve depender exclusivamente de interfaces em `domain` (ex.: `StoragePort`, nunca `*storage.R2Storage` concreto).
  - `adapters`: HTTP handlers, repositórios Postgres, adapters de storage/e-mail.
  - `infrastructure`: config, logger, JWT, crypto.
- **Contratos de API:** Não altere rotas, payloads ou códigos HTTP sem documentar o impacto e atualizar o cliente Flutter simultaneamente.
- **Dependências externas:** Não torne serviços opcionais (como Redis ou R2) bloqueadores obrigatórios no startup da API (`config.Load()`) quando não estiverem habilitados.
- **Frontend Flutter (`fit_brother_app/`):** Preserve Riverpod + GoRouter + Dio; mantenha fluxos separados e responsivos para Personal Trainer (desktop/tablet) e Aluno (mobile-first), sem UI morta (botões sem ação).

---

## 7. Handoff Entre Sessões

Ao encerrar qualquer sessão de trabalho:
- Atualize `docs/07-project/current-state.md` com branch atual, último commit, validações executadas, bloqueios e próximos passos exatos.
- Registre decisões arquiteturais relevantes como novos ADRs em `docs/adr/`.
