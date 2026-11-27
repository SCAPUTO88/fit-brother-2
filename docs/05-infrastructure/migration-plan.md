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
| **`008_exercise_catalog_and_media`** *(implementada)* | Fase 4 | Cria `muscle_groups`, `equipments`, `exercise_catalog` (índice parcial único global), `exercise_media` (`VIDEO`/`IMAGE`/`GIF`, `PENDING`/`CONFIRMED`) e FK opcional `exercises.catalog_exercise_id`. | Seed inicial idempotente de 19 músculos, 12 equipamentos e 32 exercícios clássicos `free-exercise-db` (`The Unlicense`) traduzidos para PT-BR. | `008_exercise_catalog_and_media.down.sql` remove tabelas/coluna preservando `exercises` legado. |
| **`009_workout_templates_and_versions`** *(implementada)* | Fase 5 | Cria `workout_templates`, `workout_versions` (`jsonb` imutável protegido pelo trigger `trg_workout_versions_immutable`) e `workout_assignments` (índice único parcial em `legacy_workout_plan_id`). | Converte idempotentemente `workout_plans` existentes em `workout_templates` + `v1` + `workout_assignments` sem apagar tabelas legadas. | `009_workout_templates_and_versions.down.sql` remove tabelas/trigger preservando `workout_plans` legado. |
| **`010_execution_and_feedback`** | Fase 6 | Cria `executed_workout_sessions` (com `client_idempotency_key` `UNIQUE`), `set_logs` e `workout_feedback`. | Nenhum backfill necessário (dados novos). | Forward-only. |
| **`011_assessments_and_measurements`** | Fase 7 | Cria `physical_assessments`, `body_measurements`, `weight_entries`. | Migra eventuais registros de `biometric_data` para `physical_assessments` + `weight_entries`. | Manter `biometric_data` até validação da migração. |
| **`012_deprecate_and_drop_cpf`** *(opcional pós-piloto)* | Fase 9 | Limpa/anula `cpf_encrypted` e `cpf_hash` após período de transição e remove colunas se zero dependência. | Backfill de limpeza (`UPDATE students SET cpf_encrypted=NULL, cpf_hash=NULL`). | Restaurar do snapshot criptografado pré-migração se estritamente necessário. |

---

## 3. Protocolo Antes de Toda Migration em Homologação/Piloto

1. Executar backup lógico completo (`pg_dump --format=custom`) antes de rodar `migrate up`.
2. Validar tempo de lock e aplicar `migrate up` em ambiente cópia/homologação primeiro.
3. Rodar smoke query validando contagem de registros antes/depois do backfill.
4. Confirmar compatibilidade da versão anterior da API em janela de transição (expand-and-contract).
