package rdb

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/sirupsen/logrus"

	_ "github.com/lib/pq"
)

// NewPostgresClient New PostgreSQL Client Driver
// @param ctx
// @param host      mysql dsn
// @param port      mysql dsn
// @param user      mysql dsn
// @param password  mysql dsn
// @param dbName    mysql dsn
func NewPostgresClient(ctx context.Context, host string, port int, user, password, dbName string) *Rdb {
	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		host,
		port,
		user,
		password,
		dbName,
	)

	conn, err := sql.Open("postgres", dsn)
	if err != nil {
		logrus.Errorf("failed to open postgres database: %v", err)
	}

	if err := conn.Ping(); err != nil {
		logrus.Panicf("failed to ping postgres: %v", err)
	}

	return &Rdb{
		Conn: conn,
	}
}
