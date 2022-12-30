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

	infra := accessor.BuildAccessor()
	defer infra.Close(ctx)
	infra.InitRDB(ctx)

	// init
	_, err := infra.RDB.Conn.Exec("TRUNCATE TABLE logs")
	checkError(err, "failed to execute:")

	logrus.Info("========== start ==========")
	defer logrus.Info("=========== end ===========")

	// 模擬髒讀情境 (Dirty Read)
	//
	//                            Transaction 1                                Database                              Transaction 2
	//                                  |                                         |                                       |
	//                                  |                                         |  +----------------------------+       |
	//                                  |                                         |  |           logs             |       |
	//                                  |                                         |  +----+-----------------+-----+       |
	//                                  |                                         |  | id | deposit_user_id | ... |       |
	//  +----------------------------+  |                                         |  +----+-----------------+-----+       |
	//  |           logs             |  |   START TRANSACTION                     |                                       |
	//  +----+-----------------+-----+  | --------------------------------------> |                                       |
	//  | id | deposit_user_id | ... |  |   INSERT INTO logs (...) VALUES (...)   |                                       |
	//  +----+-----------------+-----+  | --------------------------------------> |                                       |
	//  | 1  | 1               | ... |  |                                         |                   START TRANSACTION   |
	//  +----+-----------------+-----+  |                                         | <------------------------------------ |
	//                                  |                                         |           SELECT count(*) FROM logs   | -> ioslation level 為 read uncommitted 時會讀到
	//                                  |                                         | <------------------------------------ |    transaction 1 尚未 committed 的資料導致 dirty read
	//                                  |   ROLLBACK                              |                                       |
	//                                  | --------------------------------------> |                                       |    必須是 read committed 以上的等級才可避免
	//                                  |                                         |                              COMMIT   |
	//                                  |                                         | <------------------------------------ |
	//                                  |                                         |                                       |

	// 執行 trx1: 寫入一筆 log
	tx1, err := infra.RDB.Conn.Begin()
	checkError(err, "failed to start transaction:")

	_, err = tx1.Exec("INSERT INTO logs (deposit_user_id, withdraw_user_id, amount, created_at) VALUES (1, 2, 1, '2022-12-22 20:57:47');")
	checkError(err, "failed to execute:")

	// 在 trx1 結束前, 執行 trx2 取得相同 table 裡面的資料數量
	// 強制本次的 transaction isolation level 使用 read-uncommitted 等級
	tx2, err := infra.RDB.Conn.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelReadUncommitted})
	checkError(err, "failed to start transaction:")

	var count int
	err = tx2.QueryRow("SELECT count(*) FROM logs;").Scan(&count)
	checkError(err, "failed to query:")

	logrus.Warnf("Read Uncommitted: %v", count)

	// 結束 trx2
	err = tx2.Commit()
	checkError(err, "failed to commit transaction:")

	// 結束 trx1
	err = tx1.Rollback()
	checkError(err, "failed to rollback transaction:")

	return nil
}
