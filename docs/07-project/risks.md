# Registro de Riscos e Bloqueios

| ID | Risco / Bloqueio | Prob. | Impacto | Mitigação / Ação Requerida |
| :--- | :--- | :---: | :---: | :--- |
| **RK-01** | **Ambiente de execução atual sem Go/Flutter/Docker/Postgres** *(Bloqueio atual de validação)* | Confirmado | Alto | Validar suítes reais logo na abertura da Fase 1 em máquina/CI com Go 1.26, Flutter 3.41 e Docker/Postgres antes de alterar código. |
| **RK-02** | **Exposição de dados caso código atual de `develop` vá para rede pública** | Alta | Crítico | Não realizar nenhum deploy antes da conclusão integral da Fase 1 (Security Baseline). |
| **RK-03** | **`.gitignore` atual bloqueia `docs/` e `AGENTS.md`** | Confirmado | Médio | Aprovar na revisão da Fase 0 a remoção de `AGENTS.md` e `docs/` do `.gitignore` para versionamento no Git. |
| **RK-04** | **Flutter Web com `flutter_secure_storage` vulnerável a XSS e Service Worker cacheando sessão** | Média/Alta | Alto | Migrar sessão Web para cookies `HttpOnly; Secure; SameSite` via mesmo domínio/BFF (ADR-008) e personalizar Service Worker na Fase 8. |
| **RK-05** | **Supabase Free pausa após 7 dias sem uso e não possui backup automático** | Alta | Crítico | Usar Free apenas em dev/homologação com `pg_dump` agendado; migrar para Supabase Pro ($25/mês) ou backup externo automatizado no piloto real. |
| **RK-06** | **Render Free dorme após 15 min (~50s cold start)** | Alta | Médio | Aceitável em dev/homologação; para uso real na academia, adotar Render Starter ($7/mês) ou alternativa sem sleep no piloto. |
| **RK-07** | **Conformidade LGPD (Art. 11 Dados de Saúde + Resolução CD/ANPD nº 2/2022 sobre DPO)** | Média | Crítico | Eliminar CPF obrigatório (ADR-006), implantar consentimento/exclusão e obter revisão jurídica formal antes do Go-Live comercial. |
| **RK-08** | **Licença de catálogos de exercícios terceiros (ex.: ExerciseDB / wger)** | Média | Alto | Adotar `free-exercise-db` (Unlicense / domínio público) conforme ADR-005 e rejeitar datasets sem licença comercial clara. |
| **RK-09** | **Cobrança automática acidental em R2/Cloudflare/Fly/Render** | Média | Alto | Não cadastrar cartão em homologação; impor limites rígidos de upload/cota por org no backend e Spend Cap onde houver plano pago. |
