package testhelper

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"acl-challenge/internal/infrastructure/persistence"

	"github.com/joho/godotenv"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func SetupTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	loadRepoEnv(t)

	testDatabaseURL := strings.TrimSpace(os.Getenv("TEST_DATABASE_URL"))
	// Allow DSNs wrapped in quotes (common in .env examples / shells).
	testDatabaseURL = strings.Trim(testDatabaseURL, `"`)
	require.NotEmpty(t, testDatabaseURL, "TEST_DATABASE_URL must be set for integration tests")

	db, err := gorm.Open(postgres.Open(testDatabaseURL), &gorm.Config{})
	require.NoError(t, err)

	sqlDB, err := db.DB()
	require.NoError(t, err)

	err = persistence.RunMigrations(sqlDB, migrationsPath(t))
	require.NoError(t, err)

	t.Cleanup(func() {
		_ = sqlDB.Close()
	})

	t.Cleanup(func() {
		TruncateAll(t, db)
	})

	return db
}

func TruncateAll(t *testing.T, db *gorm.DB) {
	t.Helper()
	err := db.Exec("TRUNCATE TABLE notifications, users RESTART IDENTITY CASCADE").Error
	require.NoError(t, err)
}

func migrationsPath(t *testing.T) string {
	t.Helper()
	return filepath.Join(repoRoot(t), "migrations")
}

func repoRoot(t *testing.T) string {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	require.True(t, ok, "failed to resolve caller info for repo root")

	// tests/testhelper/db.go -> repo root is two levels up.
	return filepath.Clean(filepath.Join(filepath.Dir(file), "..", ".."))
}

// loadRepoEnv loads <repo_root>/.env into the process environment when present.
// Existing env vars take precedence: godotenv.Load does not overwrite them.
func loadRepoEnv(t *testing.T) {
	t.Helper()
	envPath := filepath.Join(repoRoot(t), ".env")
	if _, err := os.Stat(envPath); err != nil {
		return
	}
	if err := godotenv.Load(envPath); err != nil {
		t.Logf("warning: failed to load %s: %v", envPath, err)
	}
}
