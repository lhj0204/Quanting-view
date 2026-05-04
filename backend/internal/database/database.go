package database

import (
	"context"
	"database/sql"
	"log"
	"os"
	"path/filepath"

	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/sqlitedialect"
	"github.com/uptrace/bun/driver/sqliteshim"
	"quant/internal/models"
)

var DB *bun.DB

func Init(dsn string) error {
	dir := filepath.Dir(dsn)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	sqldb, err := sql.Open(sqliteshim.ShimName, dsn)
	if err != nil {
		return err
	}
	sqldb.SetMaxOpenConns(1)

	DB = bun.NewDB(sqldb, sqlitedialect.New())

	if err := runMigrations(); err != nil {
		return err
	}

	seedDefaultAccount()
	log.Println("Database initialized successfully")
	return nil
}

func runMigrations() error {
	ctx := context.Background()
	models := []interface{}{
		(*models.ExchangeKey)(nil),
		(*models.Strategy)(nil),
		(*models.Backtest)(nil),
		(*models.Order)(nil),
		(*models.Trade)(nil),
		(*models.Position)(nil),
		(*models.RiskRule)(nil),
		(*models.Account)(nil),
		(*models.MarketDataCache)(nil),
	}

	for _, m := range models {
		if _, err := DB.NewCreateTable().IfNotExists().Model(m).Exec(ctx); err != nil {
			return err
		}
	}
	return nil
}

func seedDefaultAccount() {
	ctx := context.Background()
	count, _ := DB.NewSelect().Model((*models.Account)(nil)).Count(ctx)
	if count == 0 {
		acct := &models.Account{
			Balance:        10000,
			Currency:       "USDT",
			InitialBalance: 10000,
		}
		_, err := DB.NewInsert().Model(acct).Exec(ctx)
		if err != nil {
			log.Printf("Warning: failed to seed default account: %v", err)
		}
	}
}
