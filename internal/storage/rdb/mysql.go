package rdb

import (
	"context"
	"database/sql"
	"time"

	"github.com/sirupsen/logrus"

	_ "github.com/go-sql-driver/mysql"
)

type mysqlClient struct {
	sqlDB *sql.DB
}

func NewMysqlClient(ctx context.Context, dsn string, connMaxLifetime time.Duration, maxOpenConns, maxIdleConns int) RDB {
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		logrus.Errorf("failed to open mysql database: %v", err)
	}

	if err := db.Ping(); err != nil {
		logrus.Panicf("failed to ping mysql: %v", err)
	}

	db.SetConnMaxIdleTime(connMaxLifetime)
	db.SetMaxOpenConns(maxOpenConns)
	db.SetMaxIdleConns(maxIdleConns)

	return &mysqlClient{
		sqlDB: db,
	}
}

func (c *mysqlClient) Shutdown(ctx context.Context) {
	if err := c.sqlDB.Close(); err != nil {
		logrus.Panicf("failed to close mysql connection: %v", err)
	}
}
