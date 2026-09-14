package postgres

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"time"

	"gorm.io/gorm"
)

var migrationFilenamePattern = regexp.MustCompile(`^(\d+)_.+\.(up|down)\.sql$`)

// MigrationAction identifies a supported schema migration operation.
type MigrationAction string

const (
	// MigrationActionUp applies pending migrations.
	MigrationActionUp MigrationAction = "up"
	// MigrationActionDown rolls back a positive number of migrations.
	MigrationActionDown MigrationAction = "down"
	// MigrationActionVersion reads the current schema version without changing it.
	MigrationActionVersion MigrationAction = "version"
)

type migrationRecord struct {
	Version   uint      `gorm:"column:version;primaryKey"`
	AppliedAt time.Time `gorm:"column:applied_at"`
}

func (migrationRecord) TableName() string { return "schema_migrations" }

type migrationFiles struct {
	version uint
	up      string
	down    string
}

// RunMigration executes a versioned database schema operation through GORM.
func (c *Client) RunMigration(ctx context.Context, directory string, action MigrationAction, steps int) (uint, bool, error) {
	if err := c.ensureMigrationTable(ctx); err != nil {
		return 0, false, err
	}
	if action == MigrationActionVersion {
		version, err := c.migrationVersion(ctx)

		return version, false, err
	}
	if action != MigrationActionUp && action != MigrationActionDown {
		return 0, false, fmt.Errorf("unsupported migration action %q", action)
	}
	if action == MigrationActionDown && steps <= 0 {
		return 0, false, errors.New("migration rollback steps must be positive")
	}

	migrations, err := readMigrationFiles(directory)
	if err != nil {
		return 0, false, err
	}
	applied, err := c.appliedMigrationVersions(ctx)
	if err != nil {
		return 0, false, err
	}
	selected := selectMigrations(migrations, applied, action, steps)
	for _, migration := range selected {
		if err := c.applyMigration(ctx, migration, action); err != nil {
			return 0, false, err
		}
	}

	version, err := c.migrationVersion(ctx)

	return version, false, err
}

func (c *Client) ensureMigrationTable(ctx context.Context) error {
	const statement = `CREATE TABLE IF NOT EXISTS schema_migrations (
		version BIGINT PRIMARY KEY,
		applied_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
	)`
	if err := c.db.WithContext(ctx).Exec(statement).Error; err != nil {
		return fmt.Errorf("creating migration table: %w", err)
	}

	return nil
}

func (c *Client) migrationVersion(ctx context.Context) (uint, error) {
	var record migrationRecord
	err := c.db.WithContext(ctx).Order("version DESC").Take(&record).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return 0, nil
	}
	if err != nil {
		return 0, fmt.Errorf("reading migration version: %w", err)
	}

	return record.Version, nil
}

func (c *Client) appliedMigrationVersions(ctx context.Context) (map[uint]bool, error) {
	var versions []uint
	if err := c.db.WithContext(ctx).Model(&migrationRecord{}).Pluck("version", &versions).Error; err != nil {
		return nil, fmt.Errorf("reading applied migrations: %w", err)
	}
	applied := make(map[uint]bool, len(versions))
	for _, version := range versions {
		applied[version] = true
	}

	return applied, nil
}

func (c *Client) applyMigration(ctx context.Context, migration migrationFiles, action MigrationAction) error {
	filename := migration.up
	if action == MigrationActionDown {
		filename = migration.down
	}
	content, err := os.ReadFile(filename) //nolint:gosec // Paths come only from entries in the operator-selected migrations directory.
	if err != nil {
		return fmt.Errorf("reading migration %d: %w", migration.version, err)
	}

	err = c.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if lockErr := tx.Exec("LOCK TABLE schema_migrations IN EXCLUSIVE MODE").Error; lockErr != nil {
			return fmt.Errorf("locking migration table: %w", lockErr)
		}
		var appliedCount int64
		if countErr := tx.Model(&migrationRecord{}).Where("version = ?", migration.version).Count(&appliedCount).Error; countErr != nil {
			return fmt.Errorf("checking migration %d state: %w", migration.version, countErr)
		}
		if action == MigrationActionUp && appliedCount > 0 {
			return nil
		}
		if action == MigrationActionDown && appliedCount == 0 {
			return nil
		}
		// Migration files are repository-controlled schema source, not request input.
		if execErr := tx.Exec(string(content)).Error; execErr != nil {
			return fmt.Errorf("executing migration %d: %w", migration.version, execErr)
		}
		if action == MigrationActionUp {
			record := migrationRecord{Version: migration.version, AppliedAt: time.Now().UTC()}
			if createErr := tx.Create(&record).Error; createErr != nil {
				return fmt.Errorf("recording migration %d: %w", migration.version, createErr)
			}

			return nil
		}
		if deleteErr := tx.Delete(&migrationRecord{}, "version = ?", migration.version).Error; deleteErr != nil {
			return fmt.Errorf("removing migration %d record: %w", migration.version, deleteErr)
		}

		return nil
	})
	if err != nil {
		return fmt.Errorf("applying %s migration %d: %w", action, migration.version, err)
	}

	return nil
}

func readMigrationFiles(directory string) ([]migrationFiles, error) {
	entries, err := os.ReadDir(directory)
	if err != nil {
		return nil, fmt.Errorf("reading migrations directory: %w", err)
	}
	byVersion := make(map[uint]*migrationFiles)
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		matches := migrationFilenamePattern.FindStringSubmatch(entry.Name())
		if len(matches) == 0 {
			continue
		}
		versionValue, parseErr := strconv.ParseUint(matches[1], 10, 64)
		if parseErr != nil {
			return nil, fmt.Errorf("parsing migration version %q: %w", matches[1], parseErr)
		}
		version := uint(versionValue)
		migration := byVersion[version]
		if migration == nil {
			migration = &migrationFiles{version: version}
			byVersion[version] = migration
		}
		path := filepath.Join(directory, entry.Name())
		if matches[2] == "up" {
			migration.up = path
		} else {
			migration.down = path
		}
	}

	migrations := make([]migrationFiles, 0, len(byVersion))
	for version, migration := range byVersion {
		if migration.up == "" || migration.down == "" {
			return nil, fmt.Errorf("migration %d requires both up and down files", version)
		}
		migrations = append(migrations, *migration)
	}
	sort.Slice(migrations, func(i, j int) bool { return migrations[i].version < migrations[j].version })

	return migrations, nil
}

func selectMigrations(all []migrationFiles, applied map[uint]bool, action MigrationAction, steps int) []migrationFiles {
	selected := make([]migrationFiles, 0, len(all))
	if action == MigrationActionUp {
		for _, migration := range all {
			if !applied[migration.version] {
				selected = append(selected, migration)
				if steps > 0 && len(selected) == steps {
					break
				}
			}
		}

		return selected
	}

	for index := len(all) - 1; index >= 0 && len(selected) < steps; index-- {
		if applied[all[index].version] {
			selected = append(selected, all[index])
		}
	}

	return selected
}
