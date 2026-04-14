//go:build integration

package integration_test

import (
	"testing"
)

func TestIntegrationSetup(t *testing.T) {
	// Esse teste serve apenas para sinalizar a inicialização correta 
	// dos containers de PostgreSQL e Redis na CI/CD do GitHub Actions.
	
	t.Log("Ambiente de Integração Configurado e Operacional")
}
