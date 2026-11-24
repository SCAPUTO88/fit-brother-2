# Guia de Desenvolvimento Local (Estado Atual vs Alvo Fase 2)

## 1. Requisitos Locais (Para o Desenvolvedor)

- **Go:** `1.26.6` (alinhado ao CI/Dockerfile; `go.mod` atualmente declara `1.25.0`)
- **Flutter / Dart:** Flutter `3.41.x` / Dart `3.11.4`
- **Docker & Docker Compose** (PostgreSQL 16)
- **OpenSSL** (para geração local de par de chaves RS256 e segredos)

---

## 2. Como Rodar Hoje (AS-IS na `develop`)

*Importante:* Na branch `develop` atual, `config.Load()` exige `REDIS_URL` e todas as variáveis `R2_*` no boot, mesmo que Redis não seja utilizado.
1. Copiar `.env.example` para `.env`.
2. Gerar chaves JWT em `secrets/jwt_rsa.pem` e `secrets/jwt_rsa.pub.pem` + segredos aleatórios `AES_KEY`, `CPF_HMAC_SECRET`, `PASSWORD_PEPPER`.
3. Subir banco via Docker Compose (`docker compose -f docker/docker-compose.yml up -d`).
4. Rodar migrations (`go run ./scripts/migrate.go up`) e seed (`go run ./scripts/seed_trainer.go`).
5. Subir API (`go run ./cmd/api`) e Flutter Web (`cd fit_brother_app && flutter run -d chrome`).

---

## 3. Como Ficará Após a Fase 2 (TO-BE Simplificado)

- **Zero Redis no Compose e no `.env` obrigatório:** apenas `postgres:16-alpine` (e opcionalmente `mailpit` para testar e-mails de convite/reset na Fase 3).
- **`STORAGE_DRIVER=local` por padrão em dev:** vídeos de teste salvos em diretório local ignorado pelo Git, sem exigir conta R2 para rodar ou testar a API.
- **Script único `make dev-setup`:** gera segredos locais de dev automaticamente e sobe tudo de forma reproduzível com um comando.
