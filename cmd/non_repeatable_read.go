package cmd

import (
	"context"
	"database/sql"
	"practice/internal/accessor"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var nonRepeatableReadCmd = &cobra.Command{
	Use:   "non_repeatable_read",
	Short: "",
	Long:  ``,
	RunE:  RunNonRepeatableReadCmd,
}

func init() {
	rootCmd.AddCommand(nonRepeatableReadCmd)
}

func RunNonRepeatableReadCmd(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	infra := accessor.BuildAccessor()
	defer infra.Close(ctx)
	infra.InitRDB(ctx)

	// init
	_, err := infra.RDB.Conn.Exec("TRUNCATE TABLE wallets")
	checkError(err, "failed to execute:")

	timeNow := time.Now().Format("2006-01-02 15:04:05")
	_, err = infra.RDB.Conn.Exec("INSERT INTO wallets (user_id, amount, created_at, modified_at) VALUES (?, ?, ?, ?);",
		"1",
		100000,
		timeNow,
		timeNow,
	)
	checkError(err, "failed to execute:")

	logrus.Info("========== start ==========")
	defer logrus.Info("=========== end ===========")

	// 模擬不可重複讀情境 (Non-repeatable Read)
	//
	//                    Transaction 1                                                   Database                                    Transaction 2
	//                         |                                                             |                                             |
	//                         |                                                             |   wallets                                   |
	//                         |                                                             |  +----+--------+-----+                      |
	//                         |                                                             |  | id | amount | ... |                      |
	//                         |                                                             |  +----+--------+-----+                      |
	//                         |                                                             |  | 1  | 100000 | ... |                      |
	//                         |                                                             |  +----+--------+-----+                      |
	//                         |   START TRANSACTION                                         |                                             |
	//                         | ----------------------------------------------------------> |                                             |
	//                         |                                                             |   START TRANSACTION                         |
	//   wallets               |                                                             | <------------------------------------------ |
	//  +----+--------+-----+  |   UPDATE wallets SET amount = amount - 60000 WHERE id = 1   |                                             |
	//  | id | amount | ... |  | ----------------------------------------------------------> |                                             |   wallets
	//  +----+--------+-----+  |                                                             |   SELECT amount FROM wallets WHERE id = 1   |  +----+--------+-----+
	//  | 1  |  40000 | ... |  |                                                             | <------------------------------------------ |  | id | amount | ... |
	//  +----+--------+-----+  |   COMMIT                                                    |                                             |  +----+--------+-----+
	//                         | ----------------------------------------------------------> |                                             |  | 1  | 100000 | ... |
	//                         |                                                             |                                             |  +----+--------+-----+
	//                         |                                                             |                                             |
	//                         |                                                             |                                             |   wallets
	//                         |                                                             |   SELECT amount FROM wallets WHERE id = 1   |  +----+--------+-----+
	//                         |                                                             | <------------------------------------------ |  | id | amount | ... |
	//                         |                                                             |   COMMIT                                    |  +----+--------+-----+
	//                         |                                                             | <------------------------------------------ |  | 1  |  40000 | ... |
	//                         |                                                             |                                             |  +----+--------+-----+
	//                         |                                                             |                                             |
	//                         |                                                             |                                             |  發生同一個 transaction 內讀取到兩次不同的
	//                         |                                                             |                                             |  結果 (non-repeatable read)
	//                         |                                                             |                                             |  必須是 repeatable read 以上的等級才可避免
	//                         |                                                             |                                             |

	tx1, err := infra.RDB.Conn.Begin()
	checkError(err, "failed to start transaction:")

	tx2, err := infra.RDB.Conn.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
	checkError(err, "failed to start transaction:")

	_, err = tx1.Exec("UPDATE wallets SET amount = amount - 60000 WHERE id = 1;")
	checkError(err, "failed to execute:")

	var amount int
	err = tx2.QueryRow("SELECT amount FROM wallets WHERE id = 1;").Scan(&amount)
	checkError(err, "failed to querying row:")

	logrus.Infof("amount = %v", amount)

	err = tx1.Commit()
	checkError(err, "failed to commit transaction:")

	err = tx2.QueryRow("SELECT amount FROM wallets WHERE id = 1;").Scan(&amount)
	checkError(err, "failed to querying row:")

	logrus.Warnf("amount = %v", amount)

	err = tx2.Commit()
	checkError(err, "failed to commit transaction:")

	return nil
}
