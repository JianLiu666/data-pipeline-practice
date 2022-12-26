package cmd

import (
	"context"
	"database/sql"
	"practice/internal/accessor"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	_ "github.com/go-sql-driver/mysql"
)

var dirtyReadCmd = &cobra.Command{
	Use:   "dirty_read",
	Short: "",
	Long:  ``,
	RunE:  RunDirtyReadCmd,
}

func init() {
	rootCmd.AddCommand(dirtyReadCmd)
}

func RunDirtyReadCmd(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	infra1 := accessor.BuildAccessor()
	defer infra1.Close(ctx)
	infra1.InitRDB(ctx)

	infra2 := accessor.BuildAccessor()
	defer infra2.Close(ctx)
	infra2.InitRDB(ctx)

	// 執行 trx1: 寫入一筆 log
	tx1, err := infra1.RDB.Conn.Begin()
	if err != nil {
		logrus.Panicf("failed to start transaction: %v", err)
	}

	_, err = tx1.Exec("INSERT INTO logs (deposit_user_id, withdraw_user_id, amount, created_at) VALUES (1, 2, 1, '2022-12-22 20:57:47');")
	if err != nil {
		logrus.Panicf("failed to start transaction: %v", err)
	}

	// 在 trx1 結束前, 執行 trx2 取得相同 table 裡面的資料數量
	// 強制本次的 transaction isolation level 使用 read-uncommitted 等級
	tx2, err := infra2.RDB.Conn.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelReadUncommitted})
	if err != nil {
		logrus.Panicf("failed to start transaction: %v", err)
	}
	rows, err := tx2.Query("SELECT count(*) FROM logs;")
	if err != nil {
		logrus.Panicf("failed to start transaction: %v", err)
	}

	for rows.Next() {
		s := 0
		err = rows.Scan(&s)
		if err != nil {
			logrus.Panicf("failed to start transaction: %v", err)
		}
		logrus.Warnf("Read Uncommitted: %v", s)
	}

	// 結束 trx2
	if err := tx2.Commit(); err != nil {
		logrus.Panicf("failed to commit transaction: %v", err)
	}

	// 結束 trx1
	if err := tx1.Rollback(); err != nil {
		logrus.Panicf("failed to rollback transaction: %v", err)
	}

	return nil
}