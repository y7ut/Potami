package migrate

import (
	"database/sql"
	"fmt"
	"io/fs"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database"
	"github.com/golang-migrate/migrate/v4/database/sqlite3"
	"github.com/y7ut/potami/internal/conf"
	"github.com/y7ut/potami/repository"

	"github.com/golang-migrate/migrate/v4/source/iofs"
	_ "github.com/mattn/go-sqlite3"
)

// MigrateUp 执行迁移
func MigrateUp() error {
	m, err := CreateSqliteMigration()
	if err != nil {
		return fmt.Errorf("failed to initialize migrate instance, %w", err)
	}
	err = m.Up()
	if err != nil {
		if err == migrate.ErrNoChange {
			fmt.Println("migrate up no change")
			return nil
		}
		return fmt.Errorf("migrate up failed, %w", err)
	}
	version, _, _ := m.Version()
	fmt.Printf("migrate up success, version: %v ", version)
	return nil
}

// MigrateDown 执行迁移回滚
func MigrateDown() error {
	m, err := CreateSqliteMigration()
	if err != nil {
		return fmt.Errorf("failed to initialize migrate instance, %w", err)
	}
	err = m.Down()
	if err != nil {
		return fmt.Errorf("migrate down failed, %w", err)
	}
	version, _, _ := m.Version()
	fmt.Printf("migrate up success, version: %v ", version)
	return nil
}

// CreateSqliteMigration 创建sqlite迁移
func CreateSqliteMigration() (*migrate.Migrate, error) {
	sqliteDb, err := sql.Open(conf.DB.Type, conf.DB.Path)
	if err != nil {
		return nil, err
	}
	sqliteDb.Driver()
	driver, err := sqlite3.WithInstance(sqliteDb, &sqlite3.Config{
		DatabaseName: conf.DB.Path,
	})
	if err != nil {
		return nil, err
	}
	return createMigrationFromEmbed(repository.MigrationFiles, repository.MigrationFilesPath, driver)
}

// createMigrationFromEmbed 根据文件系统中的迁移文件创建迁移工具
func createMigrationFromEmbed(fs fs.FS, path string, instance database.Driver) (*migrate.Migrate, error) {
	d, err := iofs.New(fs, path)
	if err != nil {
		return nil, err
	}

	m, err := migrate.NewWithInstance("iofs", d, "sqlite", instance)
	if err != nil {
		return nil, err
	}

	return m, nil
}
