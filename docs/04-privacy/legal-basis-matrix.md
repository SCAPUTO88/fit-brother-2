# Matriz Preliminar de Bases Legais e Papéis LGPD

**AVISO LEGAL:** Esta matriz é um insumo técnico preliminar sujeito à validação por advogado/DPO especializado antes do lançamento comercial.

---

## 1. Papéis dos Agentes de Tratamento (Cenário MVP B2B2C)

- **Personal Trainer / Estúdio:** Atua tipicamente como **Controlador** no relacionamento profissional com seus alunos quanto às prescrições e avaliações físicas conduzidas por ele.
- **Plataforma Fit Brother:** Atua como **Operadora** em relação aos dados tratados sob instrução do personal trainer e como **Controladora** quanto aos dados de conta, autenticação, segurança, logs de acesso, suporte e conformidade da própria plataforma. *(Em modelo futuro B2C direto ou estúdio único do próprio criador, o criador acumula o papel de Controlador principal).*

---

## 2. Bases Legais Preliminares por Finalidade

| Operação de Tratamento | Dados Envolvidos | Base Legal Sugerida (LGPD) | Justificativa / Salvaguarda |
| :--- | :--- | :--- | :--- |
| Criação de conta, autenticação, convite e recuperação de senha | Nome, e-mail, hash de senha, tokens de sessão | **Execução de contrato** (Art. 7º, V) | Estritamente necessário para acesso seguro à plataforma. |
| Prescrição de treinos e acompanhamento operacional pelo trainer | Vínculo, fichas, séries, cargas, frequência | **Execução de contrato** (Art. 7º, V) e, quando ligado à condição física/saúde, **Consentimento específico e destacado** (Art. 11, I) | Separar aceite dos Termos de Uso do consentimento granular para métricas de saúde. |
| Avaliações físicas, peso, medidas corporais e % gordura | Peso, altura, circunferências, % gordura, massa magra | **Consentimento específico e destacado do titular** (Art. 11, I) | Todas as medidas opcionais conforme necessário; revogável sem impedir consulta aos dados já anonimizáveis. |
| Logs de segurança, prevenção a fraude e auditoria (`audit_events`) | UUID usuário, IP/User-Agent em segurança, timestamp, evento | **Legítimo interesse / Segurança / Cumprimento de obrigação legal** (Art. 7º, IX e II; Marco Civil da Internet quando aplicável) | Retenção limitada e segregada dos dados de negócio. |
| **CPF (legado atual em `students`)** | `cpf_encrypted`, `cpf_hash` | **Sem base legal sólida de necessidade para MVP de treino** | **Descontinuar/remover na Fase 3** (princípio da necessidade — Art. 6º, III). |

---

## 3. Nota Sobre Pequeno Porte e Encarregado (Resolução CD/ANPD nº 2/2022)

- A Resolução CD/ANPD nº 2/2022 flexibiliza obrigações para agentes de tratamento de pequeno porte, incluindo a dispensa de DPO em regra geral.
- **Entretanto**, a própria norma **afasta a dispensa quando há tratamento de alto risco**, o que inclui o uso em larga escala ou combinação com **dados sensíveis de saúde (Art. 11)** ou grupos vulneráveis.
- **Recomendação:** Mesmo em fase inicial, estruturar um **canal formal de atendimento aos titulares** (e-mail dedicado de privacidade), manter ROPA simplificado atualizado e **submeter a obrigatoriedade formal de nomeação de DPO à análise jurídica antes da abertura comercial**.
