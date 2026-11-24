# ADR-006 — Descontinuação de CPF Obrigatório (Minimização LGPD)

**Status:** Proposto (Aguardando aprovação na Fase 0)  
**Data:** 2026-11-23

## Contexto
Atualmente `migrations/003_create_students_and_biometrics.up.sql` define `cpf_encrypted TEXT NOT NULL` e `cpf_hash TEXT NOT NULL UNIQUE`, `CreateStudentInput` exige CPF obrigatório e `GET /v1/students/:id` devolve o CPF descriptografado para qualquer treinador.

## Decisão
1. **Curto prazo (Fase 1):** Bloquear listagem/acesso indevido entre treinadores e remover exibição de CPF em listagens.
2. **Médio prazo (Fase 3):** Migration incremental tornando `cpf_encrypted` e `cpf_hash` **NULLABLE**, removendo a obrigatoriedade no backend e no Flutter (`create_student_screen.dart`), e substituindo o onboarding por **convite via e-mail + identificação única por UUID/e-mail**.
3. **Limpeza:** Planejar descarte seguro das colunas/valores de CPF após validação de não necessidade para emissão fiscal no MVP (pagamentos estão fora do MVP — Seção 30).

## Consequências
- Reduz drasticamente o impacto em caso de incidente (sem exposição de documento civil oficial), cumpre o princípio da necessidade (Art. 6º, III da LGPD) e reduz atrito no cadastro do aluno.
