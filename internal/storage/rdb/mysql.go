package rdb

import (
	"context"
	"database/sql"
	"time"

	"github.com/sirupsen/logrus"

	_ "github.com/go-sql-driver/mysql"
)

type MysqlClient struct {
	Conn *sql.DB
}

func NewMysqlClient(ctx context.Context, dsn string, connMaxLifetime time.Duration, maxOpenConns, maxIdleConns int) *MysqlClient {
	conn, err := sql.Open("mysql", dsn)
	if err != nil {
		logrus.Errorf("failed to open mysql database: %v", err)
	}

	if err := conn.Ping(); err != nil {
		logrus.Panicf("failed to ping mysql: %v", err)
	}

	conn.SetConnMaxIdleTime(connMaxLifetime)
	conn.SetMaxOpenConns(maxOpenConns)
	conn.SetMaxIdleConns(maxIdleConns)

	return &MysqlClient{
		Conn: conn,
	}
}

func (c *MysqlClient) Shutdown(ctx context.Context) {
	if err := c.Conn.Close(); err != nil {
		logrus.Panicf("failed to close mysql connection: %v", err)
	}
}
