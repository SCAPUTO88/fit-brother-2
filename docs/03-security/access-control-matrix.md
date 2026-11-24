# Matriz de Controle de Acesso (RBAC + Ownership)

Este documento compara o estado atual vulnerável (**AS-IS**) com a matriz exigida (**TO-BE**).

---

## 1. Rotas Atuais — Estado AS-IS vs Correção Imediata (Fase 1)

| Endpoint | AS-IS (Real no Código) | TO-BE Imediato (Fase 1) |
| :--- | :--- | :--- |
| `POST /v1/students` | **Público (sem auth!)** | `RequireAuth` + role `TRAINER` (ou via fluxo de convite na Fase 3); grava `trainer_id = claims.UserID`. |
| `GET /v1/students` | Qualquer autenticado (inclusive `STUDENT`) lista **todos** os alunos | Apenas `TRAINER`, filtrado obrigatoriamente por `trainer_id = claims.UserID` (ou `organization_id`). `STUDENT` recebe `403`. |
| `GET /v1/students/:id` | Qualquer `TRAINER` lê **qualquer** aluno (com CPF) | `TRAINER` só se aluno vinculado a ele (`403` caso contrário); `STUDENT` só o próprio `user_id` (`403` para terceiros). |
| `DELETE /v1/students/:id` | Qualquer `TRAINER` anonimiza **qualquer** aluno; `STUDENT` tem falha silenciosa de error handling | `TRAINER` apenas se dono do vínculo; `STUDENT` apenas a si próprio (confirmado via `user_id`), em transação SQL com revogação de tokens. |
| `POST /v1/workouts` | `TRAINER` cria plano para **qualquer** `student_id` | `TRAINER` somente para aluno vinculado a ele. |
| `GET /v1/workouts` | `STUDENT` busca `student_id = claims.UserID` (bug: retorna vazio) | `TRAINER` lista planos criados na sua org/vínculo; `STUDENT` resolve `students.id` pelo `claims.UserID` e lista apenas os próprios. |
| `GET /v1/workouts/:id` | `STUDENT` recebe erro (`users.id != students.id`); retorna HTTP 400 | `TRAINER` dono do plano OU `STUDENT` dono do `student_id` resolvido via `user_id`. Negado = `403`, inexistente = `404`. |
| `POST /v1/workouts/:id/sessions` & `.../exercises` | Qualquer `TRAINER` cria em planos alheios | Somente `TRAINER` dono do plano/org pai (`403` caso contrário). |
| `POST /v1/videos/upload-url` & `POST /v1/videos` | Qualquer `TRAINER` cria para qualquer `exercise_id` sem checagem | Somente `TRAINER` dono do plano/exercício + validação `HeadObject`. |
| `GET /v1/videos/exercises/:exercise_id/signed-url` | **Qualquer usuário autenticado** baixa URL assinada | Somente `TRAINER` dono OU `STUDENT` com plano/atribuição ativa contendo o exercício. |

---

## 2. Matriz Alvo Completa por Perfil (Fases 3–8)

| Recurso / Ação | Anônimo | Aluno (`STUDENT`) | Personal Trainer (`TRAINER`) | Admin/Owner da Org |
| :--- | :---: | :---: | :---: | :---: |
| Login / Refresh / Logout / Ativar Convite / Reset | Sim (rate-limited) | Própria sessão | Própria sessão | Própria sessão |
| Gerir Organização e Membros | Não | Não | Leitura própria org | Sim (própria org) |
| Enviar / Revogar Convite de Aluno | Não | Não | Sim (própria org) | Sim (própria org) |
| Listar / Consultar Alunos Vinculados | Não | Apenas próprio perfil | Apenas alunos com vínculo ativo | Alunos da org |
| Consultar Catálogo Global de Exercícios | Não | Somente leitura (nos treinos) | Leitura global + CRUD custom da org | Leitura global + CRUD custom da org |
| Criar / Editar / Publicar Templates de Treino | Não | **Não** | Sim (própria org) | Sim (própria org) |
| Visualizar Treino Publicado e Registrar Séries | Não | Apenas os próprios | Leitura de alunos vinculados | Leitura na org |
| Lançar Avaliação Física / Medidas | Não | Leitura própria (+ registro peso se habilitado) | Sim (alunos vinculados) | Sim (alunos da org) |
| Solicitar Exportação / Anonimização LGPD | Não | Próprios dados | Próprios dados | Próprios dados |
| Visualizar Eventos de Auditoria Críticos | Não | Não | Não | Sim (própria org) |
