# Escopo do MVP vs Backlog Futuro

## 1. Dentro do MVP (Fases 1 a 10)

1. **Baseline de Segurança (Fase 1):** Fechamento de `POST /students`, isolamento Trainer/Aluno, correção do bug `users.id != students.id` nos treinos, proteção de URLs assinadas e validação `HeadObject` em uploads, sanitização de scripts/logs, transações em Auth/Student/Anonimização.
2. **Simplificação e DevSecOps (Fase 2):** Remoção de `REDIS_URL` obrigatório no boot, `StoragePort` opcional (Local + S3/R2), saída do GCP no código/CI/docs, refresh token automático no app, pipeline CI com Flutter + segurança.
3. **Identidade, Convites e Multi-Tenancy (Fase 3):** `organizations`, `memberships`, `trainer_student_relationships`, convites por link/e-mail com expiração e uso único, ativação de conta, reset de senha, MFA TOTP opcional para trainer, depreciação do CPF obrigatório e fluxos LGPD iniciais (consentimentos e solicitações).
4. **Biblioteca de Exercícios (Fase 4):** Catálogo base importado (`free-exercise-db` — Unlicense) + tradução PT-BR + exercícios customizados do trainer com suporte a vídeo privado.
5. **Workout Builder e Versionamento (Fase 5):** Templates de ficha, clonagem, edição com snapshot de versão publicada, atribuição a aluno e visualização do aluno.
6. **Experiência do Aluno (Fase 6):** App mobile-first com treino do dia, cronômetro de descanso, registro de séries/cargas/reps (idempotente), histórico e feedback ao treinador.
7. **Avaliações Físicas e Gráficos (Fase 7):** Cadastro flexível de avaliações e evolução visual (peso, medidas, cargas, volume, frequência, PRs) + métricas reais no Dashboard do Trainer.
8. **Design System e PWA Seguro (Fase 8):** Identidade grafite/preto/branco/verde-lima, Light/Dark mode, WCAG 2.2 AA, fluxo Web/PWA seguro sem tokens expostos em caches offline.
9. **Infraestrutura, Homologação e Go-Live (Fases 9–10):** Ambiente de homologação e piloto com SMTP, backups verificados, runbooks, auditoria DAST e revisão jurídica/LGPD.

---

## 2. Explicitamente Fora do MVP (Backlog Futuro — `DEFER`)

Mantidos documentados sem implementação até conclusão do MVP:
- Aplicativos nativos publicados na Google Play e Apple App Store (embalagem futura aproveitando a base Flutter);
- Fotografias de evolução corporal (exigem fluxo e retenção de mídia sensível específicos);
- Chat em tempo real e notificações push nativas;
- Agenda / agendamento de aulas presenciais;
- Pagamentos, cobrança recorrente e assinaturas SaaS comerciais;
- White-label para academias;
- Gamificação avançada;
- Integração com smartwatches, Apple Health e Health Connect;
- Inteligência artificial assistiva (geração/ajuste automático de treinos).
