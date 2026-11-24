# Visão do Produto — Fit Brother 2.0

## 1. Propósito e Posicionamento

O **Fit Brother 2.0** é uma plataforma de treinamento físico orientada a **personal trainers independentes, pequenos estúdios e seus alunos**, projetada para substituir planilhas, PDFs e mensagens soltas no WhatsApp por um fluxo estruturado de prescrição, execução guiada, registro de cargas e acompanhamento de evolução.

---

## 2. Personas e Princípios de Experiência

| Persona | Contexto de Uso Principal | Necessidades Centrais |
| :--- | :--- | :--- |
| **Personal Trainer (Operador / Admin do Espaço)** | Desktop / Tablet (consultoria, montagem de fichas, revisão de adesão) + uso rápido em celular na academia | Cadastrar/convidar alunos com mínimo atrito, reutilizar biblioteca de exercícios e templates, acompanhar quem treinou/venceu e registrar avaliações. |
| **Aluno (Titular dos Dados)** | Celular (PWA / Browser durante o treino na academia, com conexão instável) | Ver o treino do dia em 1 toque, assistir demonstração rápida sem confusão, cronômetro de descanso, marcar séries/cargas/reps e visualizar progresso pessoal. |

---

## 3. Pilares do Fit Brother 2.0

1. **Segurança e Privacidade desde a Concepção (LGPD-by-design):** Coleta mínima de dados pessoais (eliminação do CPF obrigatório), isolamento rigoroso entre alunos e entre treinadores, e controle real de titularidade sobre dados de saúde/biometria.
2. **Experiência Dual Especializada:** Interface operacional e analítica para o Trainer (sidebar, tabelas, workout builder) e interface tátil mobile-first para o Aluno (bottom navigation, botões grandes, registro rápido).
3. **Prescrição Imutável para Histórico Confiável:** Editar ou publicar novas versões de uma ficha nunca destrói ou distorce os registros passados de treinos executados pelo aluno.
4. **Custo Enxuto e Independência de Fornecedor:** Arquitetura desacoplada de serviços proprietários (sem dependência de GCP ou Redis obrigatório no início), pronta para operar em free-tier durante validação e escalar previsivelmente.

---

## 4. Limites Claros de Produto (Não-Objetivos do MVP)

O Fit Brother 2.0 **não** é um sistema médico, não realiza diagnósticos clínicos automatizados e não entrega nesta fase: chat em tempo real, pagamentos/assinaturas, notificações push nativas, integração com wearables (Apple Health/Health Connect), IA generativa ou apps nativos publicados nas lojas.
