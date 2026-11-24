# Critérios de Aceitação e Go/No-Go

## 1. Critérios de Aceitação por Marco Principal

- **Fase 1 (Security Baseline):** Todos os 10 primeiros cenários de segurança da lista em `testing-strategy.md` implementados em RED/GREEN e passando contra PostgreSQL real; `users.id != students.id` corrigido e coberto por teste de ponta a ponta; `debug_crypto.go` sanitizado.
- **Fase 2 (Simplificação & DevSecOps):** API inicia sem `REDIS_URL` e sem `R2_*` (com `STORAGE_DRIVER=local`); Flutter renova token automaticamente no 401 e limpa tudo no logout; zero menções ativas ao GCP no código/CI; workflow CI verde com Go + Flutter + scanners com versões fixas.
- **Fase 3–5 (Identidade, Catálogo e Builder):** Convites únicos via SMTP funcionando; zero exigência de CPF para novos alunos; catálogo `free-exercise-db` importado em PT-BR com licença documentada; versionamento de treinos preservando histórico; upload de vídeo Flutter ↔ API integrado.
- **Fase 6–8 (Aluno, Avaliações e UI):** Aluno executa treino com cronômetro e idempotência; avaliações parciais e gráficos sem diagnósticos clínicos; Design System Dark/Light WCAG 2.2 AA; PWA sem cachear API privada ou tokens.
- **Fase 9–10 (Produção):** Homologação validada fora do GCP; DAST OWASP ZAP sem achados Críticos/Altos; restauração de backup testada; documentação jurídica/LGPD revisada externamente.

---

## 2. Checklist de Go/No-Go para Produção Comercial

O sistema **NÃO** pode ir para produção comercial se qualquer item abaixo for **NO-GO**:

1. **[Segurança]** Zero vulnerabilidades Críticas ou Altas abertas no código, dependências (`govulncheck`/`osv-scanner`/`trivy`) ou DAST (ZAP).
2. **[Autorização]** 100% dos testes de regressão multi-tenant / anti-BOLA passando no CI.
3. **[Sessão Web]** Cookies `HttpOnly; Secure; SameSite` + CSRF validados em navegador real (Desktop, Android e iOS Safari) sem tokens expostos em `localStorage` ou Service Worker.
4. **[LGPD]** Política de Privacidade e Termos de Uso aprovados por revisão jurídica, canal de privacidade ativo, fluxos de consentimento/exportação/exclusão testados e decisão formal sobre DPO (Resolução CD/ANPD nº 2/2022) documentada.
5. **[Infra & Backup]** Saída completa do GCP concluída; banco em plano com backup diário automático + **teste real de restauração aprovado e registrado**; alertas de uso/cota configurados sem risco de cobrança ilimitada.
6. **[Operação]** Runbooks, plano de resposta a incidentes (incluindo notificação ANPD), monitoramento `/v1/health` e rollback de deploy testados em homologação.
