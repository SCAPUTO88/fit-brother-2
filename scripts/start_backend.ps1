# start_backend.ps1
Set-Location -Path $PSScriptRoot\..

# Força a limpeza das chaves na sessão atual para que o .env prevaleça rigorosamente
Remove-Item Env:PASSWORD_PEPPER -ErrorAction SilentlyContinue
Remove-Item Env:AES_KEY -ErrorAction SilentlyContinue

# Faz carga forçada do arquivo .env nas variáveis de ambiente desta sessão
Write-Host "Carregando o .env mais recente..."
if (Test-Path ".env") {
    Get-Content .env | Where-Object { $_ -match "^[^#]" -and $_ -match "=" } | ForEach-Object {
        $parts = $_ -split "=", 2
        [System.Environment]::SetEnvironmentVariable($parts[0].Trim(), $parts[1].Trim(), "Process")
    }
} else {
    Write-Host "Arquivo .env não encontrado na raiz!"
}

Write-Host "Realizando o Seed com o mesmo ambiente..."
go run scripts/seed_trainer.go

Write-Host "Iniciando o servidor (porta 8080)..."
go run ./cmd/api/main.go
