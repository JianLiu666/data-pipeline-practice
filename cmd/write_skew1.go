package cmd

import (
	"context"
	"database/sql"
	"practice/internal/accessor"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var writeSkew1Cmd = &cobra.Command{
	Use:   "write_skew_1",
	Short: "",
	Long:  ``,
	RunE:  RunWriteSkew1Cmd,
}

func init() {
	rootCmd.AddCommand(writeSkew1Cmd)
}

func RunWriteSkew1Cmd(cmd *cobra.Command, args []string) error {
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

	// 模擬因為幻讀(Phantom Read) 造成寫偏差(Write Skew) 情境
	//
	//                    Transaction 1                                                   Database                                                    Transaction 2
	//                         |                                                             |                                                             |
	//                         |                                                             |   wallets                                                   |
	//                         |                                                             |  +----+--------+-----+                                      |
	//                         |                                                             |  | id | amount | ... |                                      |
	//                         |                                                             |  +----+--------+-----+                                      |
	//                         |                                                             |  | 1  | 100000 | ... |                                      |
	//                         |                                                             |  +----+--------+-----+                                      |
	//                         |                                                             |                                                             |
	//                         |   START TRANSACTION                                         |                                                             |
	//                         | ----------------------------------------------------------> |                                                             |
	//                         |                                                             |                                         START TRANSACTION   |
	//   wallets               |                                                             | <---------------------------------------------------------- |
	//  +----+--------+-----+  |   SELECT amount FROM wallets WHERE id = 1                   |                                                             |
	//  | id | amount | ... |  | ----------------------------------------------------------> |                                                             |   wallets
	//  +----+--------+-----+  |                                                             |                   SELECT amount FROM wallets WHERE id = 1   |  +----+--------+-----+
	//  | 1  | 100000 | ... |  |                                                             | <---------------------------------------------------------- |  | id | amount | ... |
	//  +----+--------+-----+  |       does the amount more than 60000? Yes!                 |                                                             |  +----+--------+-----+
	//                         |                                                             |                                                             |  | 1  | 100000 | ... |
	//   wallets               |                                                             |       does the amount more than 60000? Yes!                 |  +----+--------+-----+
	//  +----+--------+-----+  |   UPDATE wallets SET amount = amount - 60000 WHERE id = 1   |                                                             |
	//  | id | amount | ... |  | ----------------------------------------------------------> |                                                             |
	//  +----+--------+-----+  |   COMMIT                                                    |                                                             |
	//  | 1  |  40000 | ... |  | ----------------------------------------------------------> |                                                             |   wallets
	//  +----+--------+-----+  |                                                             |   UPDATE wallets SET amount = amount - 60000 WHERE id = 1   |  +----+--------+-----+
	//                         |                                                             | <---------------------------------------------------------- |  | id | amount | ... |
	//                         |                                                             |                                           COMMIT            |  +----+--------+-----+
	//                         |                                                             | <---------------------------------------------------------- |  | 1  | -20000 | ... |
	//                         |                                                             |                                                             |  +----+--------+-----+
	//                         |                                                             |                                                             |
	//                         |                                                             |                                                             | 因為業務邏輯造成 phantom read (讀到錯誤的錢包餘額)
	//                         |                                                             |                                                             | 導致後續發生 write skew (額度不足導致餘額為負數)
	//
	// 兩種解決 Write Skew 的辦法:
	//
	// 1. 自行加上 Explicit Lock
	//     - 改寫 SELECT amount FROM wallets WHERE id = 1 成 SELECT amount FROM wallets WHERE id = 1 FOR UPDATE
	//     - 加上排他鎖明確限制同時間只允許一個 transation 進行後續流程
	//     - 注意 Row Lock 升級成 Netx-key Lock 可能造成的衍生問題
	//
	// 2. 將 isolation level 升級成 serializable level
	//     - 在上述情境中還是無法避免同時 SELECT 後因為業務邏輯產生的 Phantom Read 問題

	wg := new(sync.WaitGroup)
	wg.Add(2)

	go func(_wg *sync.WaitGroup) {
		defer _wg.Done()

		tx1, err := infra.RDB.Conn.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
		checkError(err, "failed to start transaction:")
		logrus.Infoln("trnsaction 1 started.")

		var amount int
		err = tx1.QueryRow("SELECT amount FROM wallets WHERE id = 1").Scan(&amount)
		checkError(err, "failed to querying row:")
		logrus.Infoln("trnsaction 1 selected.")

		time.Sleep(1 * time.Second)

		// 表示業務邏輯處理結果
		if amount > 60000 {
			_, err = tx1.Exec("UPDATE wallets SET amount = amount - 60000 WHERE id = 1")
			checkError(err, "failed to execute:")
			logrus.Infoln("trnsaction 1 updated.")
		}

		err = tx1.Commit()
		checkError(err, "failed to commit:")
		logrus.Infoln("trnsaction 1 committed.")

	}(wg)

	time.Sleep(1 * time.Millisecond)

	go func(_wg *sync.WaitGroup) {
		defer _wg.Done()

		tx2, err := infra.RDB.Conn.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
		checkError(err, "failed to start transaction:")
		logrus.Infoln("trnsaction 2 started.")

		var amount int
		err = tx2.QueryRow("SELECT amount FROM wallets WHERE id = 1").Scan(&amount)
		checkError(err, "failed to querying row:")
		logrus.Infoln("trnsaction 2 selected.")

		time.Sleep(1 * time.Second)

		// 表示業務邏輯處理結果
		if amount > 60000 {
			_, err = tx2.Exec("UPDATE wallets SET amount = amount - 60000 WHERE id = 1")
			checkError(err, "failed to execute:")
			logrus.Infoln("trnsaction 2 updated.")
		}

		err = tx2.Commit()
		checkError(err, "failed to commit:")
		logrus.Infoln("trnsaction 2 committed.")

	}(wg)

	wg.Wait()

	var amount int
	err = infra.RDB.Conn.QueryRow("SELECT amount FROM wallets WHERE id = 1").Scan(&amount)
	checkError(err, "failed to querying row:")

	logrus.Warnf("Amount = %v", amount)

	return nil
}
