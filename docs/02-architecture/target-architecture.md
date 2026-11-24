# Arquitetura Alvo (TO-BE) — Fit Brother 2.0

## 1. Direção Arquitetural

Preservamos o núcleo tecnológico produtivo do projeto (**Backend Go em Arquitetura Hexagonal + PostgreSQL + Frontend Flutter Web/PWA-first**) — ver [ADR-002](../adr/ADR-002-backend-frontend-db-preservation.md) — corrigindo o acoplamento de adapters, a dependência obrigatória de Redis/R2 no boot e a exposição de infraestrutura GCP.

### Princípios Estruturais TO-BE
1. **Arquitetura Hexagonal Estrita:** `domain` define entidades e portas (`UserRepository`, `RelationshipRepository`, `ExerciseCatalogRepository`, `WorkoutRepository`, `StoragePort`, `MailerPort`, `AuditLogger`); `application` orquestra transações e autorização de posse; `adapters` implementam HTTP/BFF, Postgres, S3/R2 ou Storage Local e SMTP.
2. **Multi-Tenancy e Vínculo Explícito:** Todo acesso a aluno, ficha, avaliação ou mídia passa por validação de `organization_id` / `trainer_student_relationships` no backend.
3. **Sessão Web Segura:** Para o cliente Flutter Web no navegador, a arquitetura alvo prevê sessão via cookies `HttpOnly; Secure; SameSite` sob mesmo site ou camada BFF/proxy reverso (ver [ADR-008](../adr/ADR-008-web-session-security-bff-cookies.md)), eliminando dependência exclusiva de tokens em `localStorage`.
4. **Sem Dependências Obrigatórias Ociosas:** `REDIS_URL` deixa de ser obrigatório; rate limiting e revogação usam memória/PostgreSQL no início (ver [ADR-004](../adr/ADR-004-redis-removal-and-rate-limiting.md)). Storage S3/R2 torna-se plugável atrás de `StoragePort` (ver [ADR-005](../adr/ADR-005-media-storage-and-exercise-library.md)).
5. **Fora do GCP:** API Go em PaaS de contêiner sem GCP (Render Free no dev/homologação, upgrade previsível ou Fly/Railway no piloto), Postgres no Supabase (`sa-east-1` quando disponível no projeto), Flutter Web estático em Cloudflare Pages e mídias no Cloudflare R2 (ver [ADR-007](../adr/ADR-007-target-infrastructure-gcp-exit.md)).

---

## 2. Evolução entre Fases

- **Curto prazo (Fases 1–2):** Blindagem de segurança no monólito Go atual, remoção de Redis/GCP, `StoragePort` opcional e correção do contrato/auth no Flutter.
- **Médio prazo (Fases 3–7):** Multi-tenancy, convites via SMTP, catálogo independente (`free-exercise-db`), versionamento de treinos, fluxo completo do aluno e avaliações físicas.
- **Longo prazo (Fases 8–10):** Design system completo, PWA endurecida, observabilidade, DAST e Go/No-Go comercial.
