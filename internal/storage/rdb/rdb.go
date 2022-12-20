package rdb

import (
	"context"
	"database/sql"

	"github.com/sirupsen/logrus"
)

type Rdb struct {
	Conn *sql.DB
}

func (c *Rdb) Shutdown(ctx context.Context) {
	if err := c.Conn.Close(); err != nil {
		logrus.Panicf("failed to close mysql connection: %v", err)
	}
}
