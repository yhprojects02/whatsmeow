package upgrades

import "testing"

func TestEmbeddedUpgradeTableKeepsAllRegisteredMigrations(t *testing.T) {
	if got, want := len(Table), 13; got != want {
		t.Fatalf("registered embedded migrations = %d, want %d", got, want)
	}
}
