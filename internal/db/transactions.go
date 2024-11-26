package db

import (
	"context"
	"sync"

	"github.com/y7ut/potami/internal/conf"
)

var (
	queries         *Queries
	onceInitQueries sync.Once
	mutex           sync.Mutex
)

func GetQueries() *Queries {
	onceInitQueries.Do(func() {
		queries = New(conf.SqliteDB())
	})
	return queries
}

func Transaction(ctx context.Context, handle func(ctx context.Context, qtx *Queries) error) error {
	mutex.Lock()
	tx, err := conf.SqliteDB().Begin()
	if err != nil {
		return err
	}
	sqliteQuery := GetQueries()

	defer func() {
		tx.Rollback()
		queries = sqliteQuery
		mutex.Unlock()
	}()

	qtx := sqliteQuery.WithTx(tx)
	queries = qtx
	if err = handle(ctx, qtx); err != nil {
		return err
	}
	return tx.Commit()
}
