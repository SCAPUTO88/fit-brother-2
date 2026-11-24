# Requisitos de Negócio e Não-Funcionais

Este documento consolida os requisitos do Fit Brother 2.0 mantendo rastreabilidade com as fases do roadmap.

---

## 1. Requisitos Funcionais (RF)

| ID | Domínio | Requisito | Fase |
| :--- | :--- | :--- | :--- |
| **RF-01** | Segurança / Auth | Login, refresh transacional, logout revogando sessão e proteção default-deny em todas as rotas privadas. | Fase 1 |
| **RF-02** | Isolamento | Isolamento estrito de alunos por vínculo ativo com treinador/organização e acesso exclusivo do aluno aos seus próprios dados. | Fase 1 / 3 |
| **RF-03** | Identidade / Convites | Cadastro sem CPF obrigatório; sistema de convites com token único/expirável e ativação de conta pelo aluno; recuperação de senha via e-mail. | Fase 3 |
| **RF-04** | Multi-Tenancy | Suporte a `organizations` e `memberships` (personal autônomo como org de 1 trainer; estúdio com múltiplos trainers). | Fase 3 |
| **RF-05** | Biblioteca de Exercícios | Catálogo independente com busca, grupos musculares, equipamentos, categorias, traduções PT-BR e mídias licenciadas. | Fase 4 |
| **RF-06** | Workout Builder | Criação de templates reutilizáveis, duplicação, edição, versionamento (`workout_versions`) e publicação por período para alunos. | Fase 5 |
| **RF-07** | Execução pelo Aluno | Visualização do treino do dia, demonstração visual do exercício, cronômetro de descanso, registro idempotente de séries/cargas/reps e feedback pós-treino. | Fase 6 |
| **RF-08** | Avaliações e Evolução | Registro flexível de peso, altura, % gordura, massa magra e circunferências (sem obrigatoriedade total); gráficos de peso, medidas, cargas, volume, frequência e PRs. | Fase 7 |
| **RF-09** | Dashboard Operacional | Painel do trainer com alunos ativos/inativos/sem treino, adesão, avaliações pendentes, treinos próximos do vencimento e feedbacks recentes. | Fase 7 / 8 |
| **RF-10** | Privacidade / LGPD | Fluxo de consentimento granular, exportação dos dados do titular, solicitação de exclusão/anonimização completa e trilha de auditoria. | Fase 1 / 3 / 10 |

---

## 2. Requisitos Não-Funcionais (RNF)

| ID | Categoria | Critério Mensurável |
| :--- | :--- | :--- |
| **RNF-01** | Segurança | Zero rotas administrativas/de dados sem AuthN/AuthZ; validação de posse em 100% dos endpoints de objeto (anti-BOLA); regressão automatizada para toda correção. |
| **RNF-02** | Sessão Web | Tokens não devem ficar vulneráveis a XSS ou em caches de PWA; limpeza integral do estado privado no logout. |
| **RNF-03** | Dados Sensíveis | Dados biométricos/saúde tratados sob Art. 11 da LGPD; criptografia em trânsito (TLS 1.2+) e em repouso; proibição de PII/segredos em logs. |
| **RNF-04** | Confiabilidade | Registro de execução de séries tolerante à instabilidade de rede da academia com chave de idempotência (reenvios não duplicam sessões). |
| **RNF-05** | Performance e Portabilidade | API Go em contêiner único stateless sem dependência obrigatória de Redis ou GCP; tempo de resposta p95 < 300ms em operações CRUD com banco aquecido. |
| **RNF-06** | Acessibilidade e UI | Design System próprio (grafite/preto/branco + verde-lima, Dark/Light mode), conformidade WCAG 2.2 AA e alvos de toque >= 48dp. |
| **RNF-07** | Orçamento | Operação viável em free-tier para desenvolvimento/homologação e piloto controlado, sem cobrança automática oculta. |
