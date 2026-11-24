# Checklist de Descomissionamento / Saída do GCP

**Estado encontrado na Fase 0:** Não há chaves JSON de service account, Terraform ou scripts de deploy GCP no código-fonte. As referências existentes são **documentais** (`SAD.md`, `README.md`), comentários de commit (`4a28ac9`) e a leitura genérica da variável `PORT` em `config.go` (que também atende Render, Fly.io e Railway).

---

## 1. Ações Técnicas no Repositório (Fase 2)

- [x] Atualizar `README.md` removendo menções a Google Cloud Run e ajustando para a stack alvo fora do GCP.
- [x] Atualizar/arquivar `docs/architecture/SAD.md` referenciando `docs/02-architecture/target-architecture.md` (sem Cloud Run / Upstash obrigatório).
- [x] Renomear teste `TestResolveAppPortUsesRailwayPortWhenAppPortIsUnset` (`config_test.go`) para nome agnóstico de plataforma (`TestResolveAppPortFallsBackToPortEnvVar`).
- [x] Garantir que `docker/Dockerfile` continua agnóstico de nuvem (OCI container padrão rodando `PORT` ou `APP_PORT`).
- [x] Garantir que `.github/workflows/ci.yml` e futuros workflows de deploy não possuam steps de `google-github-actions/*` ou `gcloud`.

---

## 2. Verificações na Conta Externa do Proprietário (Manual pelo Usuário)

Caso exista algum recurso GCP criado manualmente fora do repositório:
- [ ] Verificar no console GCP se existe projeto criado para Fit Brother (Cloud Run, Artifact Registry, Secret Manager, Cloud Build).
- [ ] Excluir serviços Cloud Run, imagens no Artifact Registry e segredos no Secret Manager para eliminar qualquer cobrança residual (ex.: armazenamento de container images no Artifact Registry).
- [ ] Revogar eventuais chaves de Service Account criadas manualmente em máquinas locais.
