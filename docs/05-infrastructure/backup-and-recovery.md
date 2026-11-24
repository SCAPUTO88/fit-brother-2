# Estratégia de Backup, Restauração e Continuidade

## 1. Realidade dos Provedores Free-Tier

- **Supabase Free:** **NÃO inclui backups automáticos diários nem Point-in-Time Recovery (PITR)** e suspende bancos após 7 dias de inatividade.
- **Supabase Pro ($25/mês):** Inclui backups diários (retenção 7 dias) e evita pausa automática; PITR exige add-on adicional.
- **Cloudflare R2:** Armazenamento durável de objetos; em cenários críticos, requer política de versionamento ou sincronização de bucket se houver mídias insubstituíveis.

---

## 2. Política Obrigatória de Backup por Ambiente

| Ambiente | RPO Alvo | RTO Alvo | Mecanismo |
| :--- | :--- | :--- | :--- |
| **Homologação (Free)** | 24h | < 4h | Dump lógico diário (`pg_dump --format=custom`) via workflow agendado ou script criptografado com idade/GPG. |
| **Piloto Real** | 24h | < 2h | **Supabase Pro** (backups diários gerenciados) **+** exportação lógica diária criptografada armazenada fora do Supabase (ex.: bucket privado R2 separado com retenção 14–30 dias). |
| **Produção Comercial** | <= 1h (ou PITR) | < 1h | Supabase Pro + PITR + snapshots pré-migration obrigatórios. |

---

## 3. Procedimento Mensal de Teste de Restauração (Obrigatório para Go-Live)

Backup não testado equivale a backup inexistente. Antes da Fase 10 e mensalmente em produção:
1. Baixar o último snapshot (`pg_dump` cifrado).
2. Subir instância PostgreSQL descartável limpa (`postgres:16-alpine`).
3. Descriptografar e executar `pg_restore --clean --if-exists`.
4. Rodar query de integridade referencial (contagem de `users`, `students`, `workout_assignments`, `executed_workout_sessions`, `set_logs`, `physical_assessments`).
5. Registrar data, responsável, tempo gasto (RTO real) e status em `docs/06-quality/test-results.md`.
