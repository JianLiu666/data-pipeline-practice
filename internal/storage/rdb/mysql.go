package rdb

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/sirupsen/logrus"

	_ "github.com/go-sql-driver/mysql"
)

// NewMysqlClient New MySQL Client Driver
// @param ctx
// @param userName         mysql dsn
// @param password         mysql dsn
// @param address          mysql dsn
// @param dbName           mysql dsn
// @param connMaxLifetime  sets the maximum number of connections in the idle connection pool.
// @param maxOpenConns     sets the maximum number of open connections to the database.
// @param maxIdleConns     sets the maximum amount of time a connection may be reused.
func NewMysqlClient(ctx context.Context, userName, password, address, dbName string, connMaxLifetime time.Duration, maxOpenConns, maxIdleConns int) *Rdb {
	dsn := fmt.Sprintf("%s:%s@tcp(%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		userName,
		password,
		address,
		dbName,
	)

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

	return &Rdb{
		Conn: conn,
	}
}
