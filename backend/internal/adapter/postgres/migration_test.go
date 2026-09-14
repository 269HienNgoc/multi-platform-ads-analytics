package postgres

import (
	"os"
	"path/filepath"
	"testing"
)

func TestReadMigrationFiles(t *testing.T) {
	t.Parallel()

	directory := t.TempDir()
	for _, name := range []string{"000002_second.up.sql", "000001_first.down.sql", "000001_first.up.sql", "000002_second.down.sql"} {
		if err := os.WriteFile(filepath.Join(directory, name), []byte("SELECT 1;"), 0o600); err != nil {
			t.Fatalf("writing fixture: %v", err)
		}
	}

	migrations, err := readMigrationFiles(directory)
	if err != nil {
		t.Fatalf("readMigrationFiles() error = %v", err)
	}
	if len(migrations) != 2 || migrations[0].version != 1 || migrations[1].version != 2 {
		t.Fatalf("migrations = %#v, expected versions 1 and 2", migrations)
	}
}

func TestSelectMigrations(t *testing.T) {
	t.Parallel()

	all := []migrationFiles{{version: 1}, {version: 2}, {version: 3}}
	tests := []struct {
		name     string
		action   MigrationAction
		steps    int
		applied  map[uint]bool
		expected []uint
	}{
		{name: "all pending up", action: MigrationActionUp, applied: map[uint]bool{1: true}, expected: []uint{2, 3}},
		{name: "one pending up", action: MigrationActionUp, steps: 1, applied: map[uint]bool{1: true}, expected: []uint{2}},
		{name: "one down", action: MigrationActionDown, steps: 1, applied: map[uint]bool{1: true, 2: true}, expected: []uint{2}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			selected := selectMigrations(all, test.applied, test.action, test.steps)
			if len(selected) != len(test.expected) {
				t.Fatalf("selected count = %d, expected %d", len(selected), len(test.expected))
			}
			for index, expected := range test.expected {
				if selected[index].version != expected {
					t.Errorf("selected[%d] = %d, expected %d", index, selected[index].version, expected)
				}
			}
		})
	}
}
