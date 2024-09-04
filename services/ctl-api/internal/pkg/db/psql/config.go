package psql

import (
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func (d *database) gormConfig() *gorm.Config {
	return &gorm.Config{
		Logger:         d.Logger,
		TranslateError: true,
	}
}

func (d *database) postgresConfig(pool *pgxpool.Pool) postgres.Config {
	return postgres.Config{
		Conn: stdlib.OpenDBFromPool(pool),
	}
}
