package rdb

import (
	"context"
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"
	"github.com/sirupsen/logrus"
)

type postgres struct {
	conn *sql.DB
}

// NewPostgresClient New PostgreSQL Client Driver
// @param ctx
// @param host      mysql dsn
// @param port      mysql dsn
// @param user      mysql dsn
// @param password  mysql dsn
// @param dbName    mysql dsn
func NewPostgresClient(ctx context.Context, host string, port int, user, password, dbName string) Rdb {
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

	return &postgres{
		conn: conn,
	}
}

func (p *postgres) Shutdown(ctx context.Context) {
	if err := p.conn.Close(); err != nil {
		logrus.Panicf("failed to close mysql connection: %v", err)
	}
}

func (p *postgres) ShowTables(ctx context.Context) {

}

func (p *postgres) GenerateData(ctx context.Context) {

}

func (p *postgres) SimulateDirtyRead(ctx context.Context) {

}

func (p *postgres) SimulateReadSkew(ctx context.Context) {

}

func (p *postgres) SimulateLostUpdate(ctx context.Context) {

}

func (p *postgres) SimulateWriteSkew1(ctx context.Context) {

}

func (p *postgres) SimulateWriteSkew2(ctx context.Context) {

}

func (p *postgres) SimulateLockFailed1(ctx context.Context) {

}
