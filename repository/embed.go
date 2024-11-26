package repository

import (
	"embed"
)

const (
	MigrationFilesPath = "migration"
)

//go:embed migration
var MigrationFiles embed.FS
