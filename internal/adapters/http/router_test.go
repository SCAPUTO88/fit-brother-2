package http

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCORSConfigUsesConfiguredOrigins(t *testing.T) {
	corsConfig := newCORSConfig("http://localhost:3000, http://localhost:5173")

	require.Equal(t, []string{
		"http://localhost:3000",
		"http://localhost:5173",
	}, corsConfig.AllowOrigins)
}
