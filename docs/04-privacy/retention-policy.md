# Política Preliminar de Retenção, Descarte e Anonimização

**Status:** Diretriz técnica preliminar sujeita à aprovação jurídica antes da operação comercial.

---

## 1. Prazos Propostos por Categoria de Dados

| Categoria | Prazo de Retenção | Ação ao Término / Solicitação do Titular |
| :--- | :--- | :--- |
| **Conta de usuário e perfil do aluno** | Enquanto a conta estiver ativa | Na solicitação de exclusão (Art. 18, VI): anonimizar `email` (`deleted_<uuid>@anon.local`), limpar `full_name`, `phone`, `emergency_contact`, `birth_date`, `cpf_*`, definir `is_active=false`, preencher `deleted_at` e revogar todos os `refresh_tokens`. |
| **Dados de avaliações físicas e histórico de execução** | Enquanto vínculo ativo + prazo definido para exportação prévia | Ao excluir conta: apagar ou anonimizar irrevogavelmente em cascata na mesma transação (`executed_workout_sessions`, `set_logs`, `physical_assessments`, `body_measurements`, `weight_entries`). |
| **Tokens de refresh, convites e reset de senha** | `refresh_tokens`: TTL 7 dias; `invitations`: 72h; `reset`: 1h | Job periódico remove tokens expirados/revogados há mais de 30 dias. |
| **Eventos de auditoria e consentimentos (`audit_events`, `consent_records`)** | 5 anos ou prazo legal aplicável (sem dados de saúde em claro) | Manter apenas UUID anonimizado/pseudonimizado e registro do evento de exclusão como prova de cumprimento de obrigação legal (Art. 16, I). |
| **Logs operacionais HTTP (`zerolog`)** | Máximo 30–90 dias | Rotação automática sem PII no payload. |
| **Backups de banco de dados** | 7 a 30 dias conforme plano do provedor | Documentar na Política de Privacidade que cópias em backups rotativos expiram por sobrescrita dentro da janela de retenção do backup. |

---

## 2. Falhas Atuais de Anonimização a Corrigir na Fase 1

No código atual (`internal/application/student/anonymize.go:38-43`):
- Apenas atualiza `users.is_active = false` e executa `DELETE FROM students`.
- **Deixa `users.email` e `users.full_name` intactos em claro**, não seta `users.deleted_at`, não revoga `refresh_tokens` e não usa transação SQL.
- A correção da Fase 1 tornará o descarte **atômico, completo e verificável por teste de integração**.
