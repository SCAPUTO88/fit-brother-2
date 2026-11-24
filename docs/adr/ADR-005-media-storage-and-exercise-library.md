# ADR-005 — Armazenamento de Mídias (StoragePort) e Biblioteca de Exercícios

**Status:** Proposto (Aguardando aprovação na Fase 0)  
**Data:** 2026-11-23

## Parte A — Storage de Mídias (`StoragePort` Opcional)

### Contexto
Atualmente `internal/application/video/usecases.go` importa diretamente `*storage.R2Storage`, e `config.go:93-97` dá falha fatal sem as 5 credenciais R2.

### Decisão
- Extrair interface `StoragePort` para `internal/domain/` (Upload URL, Playback URL, HeadObject, DeleteObject).
- Disponibilizar implementações intercambiáveis via `STORAGE_DRIVER`:
  - `local`: armazena em disco local com endpoints assinados locais em dev/testes (sem exigir conta R2).
  - `r2`: Cloudflare R2 (10 GB grátis, **$0 egress**) para homologação/produção com cota rígida por arquivo e por organização.
  - `disabled`: permite subir a API mesmo sem módulo de vídeo habilitado.

---

## Parte B — Escolha da Biblioteca Base de Exercícios (Fase 4)

| Opção Pesquisada | Licença | Custo / Limite Comercial | Cobertura | Veredito |
| :--- | :--- | :--- | :--- | :--- |
| **`yuhonas/free-exercise-db`** | **The Unlicense** (Domínio Público) | **$0, sem restrição comercial**, sem API key externa | ~800+ exercícios com imagens, músculos, equipamento e passos | **ESCOLHIDA.** Permite importar no Postgres, traduzir para PT-BR, hospedar mídias no próprio R2/CDN e usar comercialmente sem risco jurídico. |
| **wger** | Código **AGPL-3.0** / Dados **CC-BY-SA 3.0** | Grátis, porém CC-BY-SA exige atribuição e Share-Alike em bases derivadas | Grande catálogo multilíngue | **DESCARTADA como base principal** pelo efeito viral Share-Alike em adaptações da base e complexidade de licença. |
| **ExerciseDB (RapidAPI / AscendAPI)** | **Proprietária** | Planos gratuitos/OSS são **estritamente não-comerciais**; redistribuição de base proibida | ~1.300+ exercícios GIF | **REJEITADA.** Incompatível com uso SaaS/comercial gratuito e cria dependência de API paga externa. |
