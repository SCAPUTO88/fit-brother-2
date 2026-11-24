# Inventário de Funcionalidades (AS-IS)

Legenda de status: **Implementado**, **Parcial**, **Quebrado**, **Ausente**.

---

## 1. Autenticação, Identidade e Acesso

| Funcionalidade | Backend | Frontend Flutter | Status Global | Evidência / Observação |
| :--- | :--- | :--- | :--- | :--- |
| Login e-mail + senha | Implementado (`POST /v1/auth/login`) | Parcial | **Parcial** | `auth_provider.dart:54` marca `isLoggedIn=true` mesmo se tokens vierem `null`. |
| Refresh Token | Implementado (`POST /v1/auth/refresh`) | Ausente (stub) | **Parcial / Quebrado no App** | Backend rotaciona sem transação; `dio_client.dart:32-34` tem apenas stub. |
| Logout | Implementado (`POST /v1/auth/logout`) | Quebrado | **Quebrado** | Flutter envia corpo vazio; backend exige `refresh_token` (`LogoutInput`) e rota não exige auth. |
| Registro de Treinador / Organização | Ausente (apenas CLI `seed_trainer.go`) | Ausente | **Ausente** | Não há auto-cadastro, onboarding ou multi-tenancy por organização. |
| Convite e Ativação de Aluno | Ausente | Ausente | **Ausente** | Hoje aluno é criado diretamente com senha no endpoint aberto `POST /v1/students`. |
| Recuperação de Senha / MFA | Ausente | Ausente | **Ausente** | Não há integração SMTP nem fluxo de reset/MFA. |

---

## 2. Alunos e Dados Biométricos

| Funcionalidade | Backend | Frontend Flutter | Status Global | Evidência / Observação |
| :--- | :--- | :--- | :--- | :--- |
| Cadastro de aluno | Implementado sem auth (`POST /v1/students`) | Implementado (`create_student_screen.dart`) | **Crítico / Inseguro** | Endpoint público; exige CPF; sem transação SQL entre `users` e `students`. |
| Listagem de alunos | Implementado sem filtro de trainer/role (`GET /v1/students`) | Implementado (`dashboard_screen.dart`) | **Inseguro** | Qualquer usuário autenticado (incluindo `STUDENT`) vê todos os alunos; clique no card não faz nada. |
| Detalhe de aluno (c/ CPF) | Implementado (`GET /v1/students/:id`) | Ausente | **Inseguro** | Qualquer `TRAINER` acessa CPF descriptografado de qualquer aluno. |
| Anonimização / Exclusão LGPD | Parcial (`DELETE /v1/students/:id`) | Ausente | **Parcial / Incompleto** | Sem transação, sem checagem de trainer dono, não anonimiza `users.email`/`full_name` nem revoga tokens. |
| Vínculo Trainer ↔ Student | Ausente | Ausente | **Ausente** | Sem `trainer_id` ou tabela `trainer_student_relationships`. |
| Avaliações Físicas / Biometria | Apenas tabela DB (`migrations/003`) | Ausente | **Ausente** | Zero endpoints HTTP, zero casos de uso; entidade `BiometricData` omite colunas da migration. |

---

## 3. Biblioteca de Exercícios e Mídias (Vídeos)

| Funcionalidade | Backend | Frontend Flutter | Status Global | Evidência / Observação |
| :--- | :--- | :--- | :--- | :--- |
| Catálogo independente de exercícios | Ausente | Ausente | **Ausente** | Exercícios são strings livres dentro de `workout_sessions`. |
| URL assinada de upload (R2) | Implementado (`POST /v1/videos/upload-url`) | Quebrado (`POST /videos/request-upload`) | **Quebrado / Inseguro** | Contrato incompatível; backend não valida ownership nem sanitiza filename. |
| Confirmação de upload (R2) | Implementado (`POST /v1/videos`) | Quebrado (`POST /videos/confirm`) | **Quebrado / Inseguro** | Contrato incompatível; backend não faz `HeadObject` nem valida metadados reais. |
| URL assinada de playback (R2) | Implementado (`GET /v1/videos/exercises/:exercise_id/signed-url`) | Ausente | **Inseguro / Sem UI** | Backend não verifica se usuário pode acessar o exercício; Flutter não toca vídeo. |

---

## 4. Planos de Treino, Prescrição e Execução

| Funcionalidade | Backend | Frontend Flutter | Status Global | Evidência / Observação |
| :--- | :--- | :--- | :--- | :--- |
| Criar plano de treino | Implementado (`POST /v1/workouts`) | Implementado (`create_plan_screen.dart`) | **Parcial** | Não valida se `student_id` pertence ao trainer. |
| Listar planos de treino | Quebrado p/ aluno (`GET /v1/workouts`) | Implementado (`workout_plans_screen.dart`) | **Quebrado p/ Aluno** | Passa `users.id` como `students.id` para `ByStudent`. |
| Detalhes do plano | Quebrado p/ aluno (`GET /v1/workouts/:id`) | Implementado (`plan_detail_screen.dart`) | **Quebrado p/ Aluno** | Mesmo bug `users.id != students.id`; retorna HTTP 400 para acesso negado/não encontrado. |
| Criar sessão e exercício | Sem ownership check (`POST .../sessions`, `.../exercises`) | Implementado | **Inseguro** | Qualquer treinador pode adicionar sessões/exercícios em planos de outro treinador. |
| Edição, duplicação, templates e versionamento | Ausente | Ausente | **Ausente** | Não há API `PUT/PATCH/DELETE` para treinos, templates ou snapshots imutáveis. |
| Execução de treino pelo aluno (check-in, cargas, timer, histórico, feedback) | Ausente | Ausente | **Ausente** | Nenhuma tela de aluno, nenhum log de séries/cargas ou cronômetro. |
| Dashboard operacional e gráficos de evolução | Ausente | Placeholder simples | **Ausente** | Dashboard do trainer lista apenas alunos e atalhos básicos; item "Vídeos" do menu não faz nada. |
