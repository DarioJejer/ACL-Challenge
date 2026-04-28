package testhelper

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"acl-challenge/internal/infrastructure/persistence"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func SetupTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	testDatabaseURL := os.Getenv("TEST_DATABASE_URL")
	require.NotEmpty(t, testDatabaseURL, "TEST_DATABASE_URL must be set for integration tests")

	db, err := gorm.Open(postgres.Open(testDatabaseURL), &gorm.Config{})
	require.NoError(t, err)

	sqlDB, err := db.DB()
	require.NoError(t, err)

	err = persistence.RunMigrations(sqlDB, migrationsPath(t))
	require.NoError(t, err)

	t.Cleanup(func() {
		TruncateAll(t, db)
	})

	t.Cleanup(func() {
		_ = sqlDB.Close()
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
	_, file, _, ok := runtime.Caller(0)
	require.True(t, ok, "failed to resolve caller info for migrations path")

	// tests/testhelper/db.go -> repo root is two levels up.
	rootDir := filepath.Clean(filepath.Join(filepath.Dir(file), "..", ".."))
	return filepath.Join(rootDir, "migrations")
}
