# Plano de Migração de Banco de Dados e Sistema

Este documento atende ao requisito de preservação das migrations `001`–`005` com evolução segura para o esquema TO-BE.

---

## 1. Regra de Ouro das Migrations

1. **Preservar `001` a `005` intactas:** Nunca editar arquivos SQL já existentes (`001`–`005`), para evitar divergência de checksum (`dirty database`) em bancos já migrados.
2. **Proibição absoluta de `migrate down` global:** O comando `go run ./scripts/migrate.go down` reverte todas as migrations destrutivamente. Ele deve ser bloqueado contra ambientes não locais (exigir flag `--i-know-this-wipes-data` apenas com `APP_ENV=development`).

---

## 2. Sequência de Migrations Incrementais Planejadas

| Migration | Fase | Alterações de Schema | Transformação / Backfill | Rollback Seguro |
| :--- | :--- | :--- | :--- | :--- |
| **`006_add_trainer_ownership_and_indexes`** | Fase 1 | Adiciona `trainer_id UUID REFERENCES users(id)` em `students`; adiciona índices em `students(trainer_id)` e `workout_plans(trainer_id, student_id)`. | Em bancos dev com 1 trainer: preenche `trainer_id` com o trainer existente; em bancos multi-trainer: exige atribuição explícita antes de `NOT NULL`. | `DROP COLUMN trainer_id` (somente se ainda não em produção ou após backup). |
| **`007_organizations_invites_and_lgpd`** | Fase 3 | Cria `organizations`, `memberships`, `trainer_student_relationships`, `invitations`, `consent_records`, `privacy_requests`, `audit_events`. Torna `students.cpf_encrypted` e `students.cpf_hash` **NULLABLE**. | Para cada `TRAINER` ativo, cria organização `SOLO` + membership `OWNER`; migra `students.trainer_id` para `trainer_student_relationships` (`ACTIVE`). | Migrations forward-only em produção (desativar novas features via feature flag/código compatível). |
| **`008_deprecate_and_drop_cpf`** *(opcional pós-validação)* | Fase 3/9 | Limpa/anula `cpf_encrypted` e `cpf_hash` após período de transição e remove colunas se zero dependência. | Backfill de limpeza (`UPDATE students SET cpf_encrypted=NULL, cpf_hash=NULL`). | Restaurar do snapshot criptografado pré-migração se estritamente necessário. |
| **`009_exercise_catalog_and_media`** | Fase 4 | Cria `exercise_catalog`, lookup de grupos/equipamentos e `exercise_media` (`status: PENDING/CONFIRMED`). | Seed idempotente do catálogo `free-exercise-db` + migração de registros legados de `videos`. | Preservar tabela `videos` em modo leitura durante janela de transição. |
| **`010_workout_templates_and_versions`** | Fase 5 | Cria `workout_templates`, `workout_versions` (`jsonb` imutável) e `workout_assignments`. | Converte `workout_plans` ativos em `workout_templates` + `v1` + `workout_assignments` sem apagar legado imediatamente. | Código lê `workout_assignments` com fallback somente-leitura para legado se necessário. |
| **`011_execution_and_feedback`** | Fase 6 | Cria `executed_workout_sessions` (com `client_idempotency_key` `UNIQUE`), `set_logs` e `workout_feedback`. | Nenhum backfill necessário (dados novos). | Forward-only. |
| **`012_assessments_and_measurements`** | Fase 7 | Cria `physical_assessments`, `body_measurements`, `weight_entries`. | Migra eventuais registros de `biometric_data` para `physical_assessments` + `weight_entries`. | Manter `biometric_data` até validação da migração. |

---

## 3. Protocolo Antes de Toda Migration em Homologação/Piloto

1. Executar backup lógico completo (`pg_dump --format=custom`) antes de rodar `migrate up`.
2. Validar tempo de lock e aplicar `migrate up` em ambiente cópia/homologação primeiro.
3. Rodar smoke query validando contagem de registros antes/depois do backfill.
4. Confirmar compatibilidade da versão anterior da API em janela de transição (expand-and-contract).
