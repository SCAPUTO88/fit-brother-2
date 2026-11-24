# Inventário de Dados Pessoais e Sensíveis (LGPD)

**Aviso:** Mapeamento técnico preliminar de engenharia; não substitui revisão jurídica formal (ROPA/RIPD).

---

## 1. Dados Mapeados no Sistema Atual (AS-IS) vs Proposto (TO-BE)

| Categoria | Dado / Campo | Onde Existe (AS-IS) | Classificação LGPD | Estado no TO-BE |
| :--- | :--- | :--- | :--- | :--- |
| **Identificação e Conta** | `full_name`, `email`, `role`, `is_active` | `users` (`migrations/001`) | Dado pessoal comum (Art. 5º, I) | **Manter** (anonimizar `email` e `full_name` no descarte!) |
| **Credenciais** | `password_hash`, `token_hash` | `users`, `refresh_tokens` | Credencial de segurança | **Manter** (apenas hashes fortes + revogação no logout/exclusão) |
| **Documento Oficial** | `cpf_encrypted`, `cpf_hash` | `students` (`migrations/003`), `create_student_screen.dart` | Dado pessoal altamente sensível ao risco de fraude | **REMOVER / DEPRECAR (ADR-006)** — violação de minimização |
| **Contato e Emergência** | `phone`, `emergency_contact`, `birth_date` | `students` (`migrations/003`) | Dado pessoal comum (data nasc. pode identificar faixa etária) | **Tornar opcionais** e justificados apenas quando realmente utilizados |
| **Dados Biométricos e Saúde** | `weight_kg`, `height_cm`, `body_fat_pct`, `muscle_mass_kg`, `notes_encrypted`, circunferências | `biometric_data` (`migrations/003` — hoje sem endpoint) | **Dado Pessoal Sensível — Saúde (Art. 5º, II e Art. 11)** | **Manter apenas no módulo explícito de Avaliações (Fase 7)** com consentimento específico e controle estrito |
| **Histórico de Atividade Física** | Treinos prescritos, séries executadas, cargas, RPE, feedbacks | `workout_plans` / `executed_workout_sessions` / `set_logs` | Dado comportamental ligado à saúde/condicionamento | **Tratar com proteção de dado ligado à saúde (Art. 11)** |
| **Mídia de Exercício** | Vídeos de demonstração enviados pelo trainer | `videos` (`migrations/005`) / R2 | Dado corporativo/imagem do instrutor (evitar imagens identificáveis de alunos no MVP) | **Manter apenas para demonstrações do catálogo/trainer** (sem fotos/vídeos de evolução de alunos nesta fase) |
| **Registros de Conformidade** | `consent_records`, `privacy_requests`, `audit_events` | **Ausente no AS-IS** | Evidência de conformidade e segurança | **Adicionar na Fase 3** |

---

## 2. Observação Crítica Sobre Menores e Idosos (Resolução CD/ANPD nº 2/2022)

- O sistema atual coleta `birth_date` sem bloquear ou exigir consentimento parental específico de menores (Art. 14 da LGPD).
- Na Fase 3, o produto deve exigir declaração de maioridade (18+) nos Termos de Uso no MVP (ou fluxo específico caso passe a atender menores no futuro) e garantir designação/avaliação de **Encarregado (DPO)** dado que o tratamento envolve dados de saúde em potencial alto risco (ver [`legal-basis-matrix.md`](legal-basis-matrix.md)).
