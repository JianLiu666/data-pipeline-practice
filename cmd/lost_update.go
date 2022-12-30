package cmd

import (
	"context"
	"database/sql"
	"practice/internal/accessor"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var lostUpdateCmd = &cobra.Command{
	Use:   "lost_update",
	Short: "",
	Long:  ``,
	RunE:  RunLostUpdateCmd,
}

func init() {
	rootCmd.AddCommand(lostUpdateCmd)
}

func RunLostUpdateCmd(cmd *cobra.Command, args []string) error {
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

	// 模擬更新丟失情境 (Lost Update)
	//
	//                    Transaction 1                                          Database                                           Transaction 2
	//                         |                                                    |                                                    |
	//                         |                                                    |   wallets                                          |
	//                         |                                                    |  +----+--------+-----+                             |
	//                         |                                                    |  | id | amount | ... |                             |
	//                         |                                                    |  +----+--------+-----+                             |
	//                         |                                                    |  | 1  | 100000 | ... |                             |
	//                         |                                                    |  +----+--------+-----+                             |
	//                         |                                                    |                                                    |
	//                         |                                                    |                                START TRANSACTION   |
	//                         |                                                    | <------------------------------------------------- |
	//                         |   START TRANSACTION                                |                                                    |
	//                         | -------------------------------------------------> |                                                    |   wallets
	//                         |                                                    |          SELECT amount FROM wallets WHERE id = 1   |  +----+--------+-----+
	//   wallets               |                                                    | <------------------------------------------------- |  | id | amount | ... |
	//  +----+--------+-----+  |   SELECT amount FROM wallets WHERE id = 1          |                                                    |  +----+--------+-----+
	//  | id | amount | ... |  | -------------------------------------------------> |                                                    |  | 1  | 100000 | ... |
	//  +----+--------+-----+  |                                                    |                                                    |  +----+--------+-----+
	//  | 1  | 100000 | ... |  |                                                    |                                                    |
	//  +----+--------+-----+  |                                                    |                                                    |
	//                         |                                                    |                                                    |   wallets
	//                         |                                                    |   UPDATE wallets SET amount = 60000 WHERE id = 1   |  +----+--------+-----+
	//                         |                                                    | <------------------------------------------------- |  | id | amount | ... |
	//                         |                                                    |                                           COMMIT   |  +----+--------+-----+
	//   wallets               |                                                    | <------------------------------------------------- |  | 1  |  60000 | ... |
	//  +----+--------+-----+  |   UPDATE wallets SET amount = 40000 WHERE id = 1   |                                                    |  +----+--------+-----+
	//  | id | amount | ... |  | -------------------------------------------------> |                                                    |
	//  +----+--------+-----+  |   COMMIT                                           |                                                    |  transaction2 的更新結果最後被 transaction1 覆蓋掉
	//  | 1  |  40000 | ... |  | -------------------------------------------------> |                                                    |  造成 lost update
	//  +----+--------+-----+  |                                                    |                                                    |
	//                         |                                                    |                                                    |
	//
	// 兩種解決 Lost Update 的辦法:
	//
	// 1. 交給 Database 的 atomic write
	//     - 改寫 UPDATE wallets SET amount = {value} WHERE id = 1 成 UPDATE wallets SET amount = amount - {value} WHERE id = 1
	//     - 要特別注意 transaction failed 時的重試流程，若未能保證冪等性可能造成重複扣款問題
	//
	// 2. 自行實現樂觀鎖流程 (CAS)
	//     - 改寫 UPDATE walltes SET amount = {value} WHERE id = 1 成 UPDATE wallets SET amount = {new} WHERE id = 1 AND amount = {old}
	//     - 強制 transaction 1 更新失敗, 但要自行驗證 transaction 執行結果是否符合預期

	tx2, err := infra.RDB.Conn.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelRepeatableRead})
	checkError(err, "failed to start transaction:")

	tx1, err := infra.RDB.Conn.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelRepeatableRead})
	checkError(err, "failed to start transaction:")

	var amount_tx1, amount_tx2, amount_result int

	err = tx2.QueryRow("SELECT amount FROM wallets WHERE id = 1").Scan(&amount_tx2)
	checkError(err, "failed to querying row:")

	err = tx1.QueryRow("SELECT amount FROM wallets WHERE id = 1").Scan(&amount_tx1)
	checkError(err, "failed to querying row:")

	_, err = tx2.Exec("UPDATE wallets SET amount = 60000 WHERE id = 1")
	checkError(err, "failed to execute:")

	err = tx2.Commit()
	checkError(err, "failed to commit:")

	_, err = tx1.Exec("UPDATE wallets SET amount = 40000 WHERE id = 1")
	checkError(err, "failed to execute:")

	err = tx1.Commit()
	checkError(err, "failed to commit:")

	err = infra.RDB.Conn.QueryRow("SELECT amount FROM wallets WHERE id = 1").Scan(&amount_result)
	checkError(err, "failed to querying row:")

	logrus.Warnf("Amount = %v", amount_result)

	return nil
}
