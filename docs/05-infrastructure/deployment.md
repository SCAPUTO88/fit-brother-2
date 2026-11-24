# Estratégia de Deploy e Promoção de Ambientes (Fase 9–10)

*Nota:* Conforme diretriz da Fase 0, **nenhum deploy foi executado neste momento**. Este documento define o modelo seguro de promoção após conclusão das Fases 1 a 8.

---

## 1. Fluxo de Promoção

1. **Pull Request → `develop`:**
   - Roda pipeline completa (Go lint, unit, integration Postgres, race detector, `govulncheck`, Flutter analyze/test/build web, `gitleaks`, `semgrep`, `osv-scanner`, `trivy`, SBOM).
   - Deploy automático (após Fase 9) apenas para **Homologação** via tag/merge aprovado em `develop`, nunca usando dados reais de alunos.
2. **Release PR `develop` → `main`:**
   - Exige revisão explícita, CHANGELOG, verificação de migrations e aprovação manual de **Environment Protection Rules** no GitHub Actions.
   - Executa backup pré-deploy, aplica migrations (`migrate up`), promove contêiner Go assinado/tagueado e publica assets estáticos no Cloudflare Pages.

---

## 2. Segredos e Parâmetros de Produção

- Gerenciados exclusivamente no cofre de segredos do provedor de destino (Render Environment Groups / Cloudflare Pages Env / GitHub Environments), **nunca** em arquivos versionados.
- Chaves obrigatórias isoladas por ambiente (dev, homologação e produção nunca compartilham `AES_KEY`, `PASSWORD_PEPPER`, chaves JWT ou bucket R2).
- Health check `/v1/health` deve validar conectividade com PostgreSQL antes de receber tráfego (readiness) e retornar payload mínimo sem detalhes de infra.
