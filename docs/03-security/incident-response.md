# Plano Preliminar de Resposta a Incidentes

**Status:** Documento operacional preliminar (a ser validado antes do piloto comercial na Fase 10).

---

## 1. Classificação de Severidade

| Nível | Critério | Exemplos no Fit Brother | Prazo Inicial de Contenção |
| :--- | :--- | :--- | :--- |
| **SEV-1 (Crítico)** | Exposição ou comprometimento de dados pessoais/sensíveis (Art. 11 LGPD), segredos criptográficos (`AES_KEY`, `JWT` privada, `DATABASE_URL`) ou acesso não autorizado multi-tenant | Vazamento de banco/CPF/avaliações, exploração de IDOR em produção, chave privada vazada em commit/log | Imediato (< 1 hora) |
| **SEV-2 (Alto)** | Falha de autenticação/autorização sem evidência de extração em massa, indisponibilidade total da API | Queda prolongada do backend/banco, bypass parcial de papel em homologação/piloto | < 4 horas |
| **SEV-3 (Médio/Baixo)** | Vulnerabilidade em dependência sem exploração ativa, bug funcional não sensível | CVE moderado em pacote Go/Flutter sem vetor exposto | Próximo ciclo/sprint |

---

## 2. Procedimento Padrão de Resposta (6 Etapas)

1. **Detecção e Triagem:** Registrar horário UTC, componentes afetados, escopo estimado de titulares e preservar evidências/logs antes de reiniciar instâncias.
2. **Contenção Imediata:**
   - Revogar segredos suspeitos (`JWT` keys gerando novo `kid`, `AES_KEY` exigindo plano de re-encrypt se aplicável, `DATABASE_URL`, `R2_*`, tokens SMTP).
   - Invalidar sessões ativas em massa (`UPDATE refresh_tokens SET revoked_at = NOW()`).
   - Bloquear rotas/deploy comprometido no provedor de edge (Cloudflare) ou colocar API em manutenção controlada.
3. **Investigação:** Correlacionar `audit_events`, logs estruturados HTTP (sem PII) e métricas de acesso do provedor (Supabase / R2 / Render / Cloudflare).
4. **Erradicação e Recuperação:** Corrigir vulnerabilidade com teste de regressão, aplicar patch, validar restauração íntegra e monitorar anomalias pós-recuperação.
5. **Comunicação LGPD (Art. 48):** Quando houver risco ou dano relevante aos titulares:
   - Avaliar com Encarregado (DPO) e assessoria jurídica a comunicação formal à **ANPD** e aos **titulares afetados** dentro do prazo regulamentar da ANPD (atualmente 3 dias úteis segundo a Resolução CD/ANPD nº 15/2024, salvo atualização normativa — validar com jurídico).
6. **Lições Aprendidas (Post-mortem):** Atualizar `threat-model.md`, testes automatizados e runbooks em até 5 dias úteis.
