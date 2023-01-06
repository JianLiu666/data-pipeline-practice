package rdb

import (
	"context"

	"github.com/sirupsen/logrus"
)

type Rdb interface {
	Shutdown(ctx context.Context)

	ShowTables(ctx context.Context)
	GenerateData(ctx context.Context)

	SimulateDirtyRead(ctx context.Context)
	SimulateReadSkew(ctx context.Context)
	SimulateLostUpdate(ctx context.Context)
	SimulateWriteSkew1(ctx context.Context)
	SimulateWriteSkew2(ctx context.Context)
	SimulateLockFailed1(ctx context.Context)
}

func checkError(err error, msg string) {
	if err != nil {
		logrus.Panicln(msg, err)
	}
}
