package rdb

import (
	"context"

	"github.com/sirupsen/logrus"
)

type Rdb interface {
	Shutdown(ctx context.Context)

	// 顯示目前關連式資料庫中所有的 tables & columns
	ShowTables(ctx context.Context)

	// 建立測試資料
	GenerateData(ctx context.Context)

	// 模擬髒讀(Dirty Read) 情境
	SimulateDirtyRead(ctx context.Context)

	// 模擬讀偏差(Read Skew) 情境
	SimulateReadSkew(ctx context.Context)

	// 模擬更新丟失(Lost Update) 情境
	SimulateLostUpdate(ctx context.Context)

	// 模擬寫偏差(Write Skew) 情境
	// 可以透過 Serializable Isolation 解決的情境
	SimulateWriteSkew1(ctx context.Context)

	// 模擬寫偏差(Write Skew) 情境
	// 無法單靠 Serializable Isolation 解決的情境
	SimulateWriteSkew2(ctx context.Context)

	// 模擬因為聚簇索引(Clustered index) 與覆蓋索引(Covering index) 不同造成上鎖失敗的情境
	SimulateLockFailed1(ctx context.Context)
}

func checkError(err error, msg string) {
	if err != nil {
		logrus.Panicln(msg, err)
	}
}
