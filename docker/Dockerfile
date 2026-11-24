# ==============================================================================
# Stage 1: Build
# ==============================================================================
FROM golang:1.26.11-alpine AS builder

# Dependências de build
RUN apk add --no-cache git ca-certificates tzdata

WORKDIR /app

# Copiar dependências primeiro (cache layer)
COPY go.mod go.sum ./
RUN go mod download

# Copiar código-fonte
COPY . .

# Build com otimizações de tamanho
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build \
    -ldflags="-s -w -X main.version=$(git describe --tags --always --dirty 2>/dev/null || echo 'dev')" \
    -o /app/server \
    ./cmd/api

# ==============================================================================
# Stage 2: Runtime — imagem distroless (sem shell, sem package manager)
# Superfície de ataque mínima para produção
# ==============================================================================
FROM gcr.io/distroless/static-debian12

# Copiar certificados e timezone do builder
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo

# Copiar apenas o binário
COPY --from=builder /app/server /server

# Porta da aplicação
EXPOSE 8080

# Executar como usuário não-root
USER nonroot:nonroot

ENTRYPOINT ["/server"]
