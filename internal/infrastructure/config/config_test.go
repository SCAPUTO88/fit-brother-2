package config

import "testing"

func TestResolveAppPortUsesRailwayPortWhenAppPortIsUnset(t *testing.T) {
	t.Setenv("APP_PORT", "")
	t.Setenv("PORT", "49152")

	if got := resolveAppPort(); got != "49152" {
		t.Fatalf("resolveAppPort() = %q, want %q", got, "49152")
	}
}
