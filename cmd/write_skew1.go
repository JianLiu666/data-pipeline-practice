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
	//
	//                    Transaction 1                                  Database                                       Transaction 2
	//                         |                                            |                                                |
	//                         |                                            |   wallets                                      |
	//                         |                                            |  +----+--------+-----+                         |
	//                         |                                            |  | id | amount | ... |                         |
	//                         |                                            |  +----+--------+-----+                         |
	//                         |                                            |  | 1  | 100000 | ... |                         |
	//                         |                                            |  +----+--------+-----+                         |
	//                         |                                            |                                                |
	//                         |                                            |                            START TRANSACTION   |
	//                         |                                            | <--------------------------------------------- |   wallets
	//                         |                                            |                   SELECT amount FROM wallets   |  +----+--------+-----+
	//                         |                                            | <--------------------------------------------- |  | id | amount | ... |
	//                         |   START TRANSACTION                        |                                                |  +----+--------+-----+
	//   wallets               | -----------------------------------------> |                                                |  | 1  | 100000 | ... |
	//  +----+--------+-----+  |   INSERT INTO wallets (...) VALUES (...)   |                                                |  +----+--------+-----+
	//  | id | amount | ... |  | -----------------------------------------> |                                                |
	//  +----+--------+-----+  |   COMMIT                                   |                                                |
	//  | 1  | 100000 | ... |  | -----------------------------------------> |                                                |   wallets
	//  +----+--------+-----+  |                                            |                   SELECT amount FROM wallets   |  +----+--------+-----+
	//  | 2  | 100000 | ... |  |                                            | <--------------------------------------------- |  | id | amount | ... |
	//  +----+--------+-----+  |                                            |                                                |  +----+--------+-----+
	//                         |                                            |                                                |  | 1  | 100000 | ... |
	//                         |                                            |                                                |  +----+--------+-----+
	//                         |                                            |                                                |
	//                         |                                            |                                                |
	//                         |                                            |                                                |   wallets
	//                         |                                            |   UPDATE wallets SET amount = amount + 10000   |  +----+--------+-----+
	//                         |                                            | <--------------------------------------------- |  | id | amount | ... |
	//                         |                                            |                                       COMMIT   |  +----+--------+-----+
	//                         |                                            | <--------------------------------------------- |  | 1  | 110000 | ... |
	//                         |                                            |                                                |  +----+--------+-----+
	//                         |                                            |                                                |  | 2  | 110000 | ... |
	//                         |                                            |                                                |  +----+--------+-----+
	//                         |                                            |                                                |
	//                         |                                            |                                                | 發生 write skew 導致更新到未讀取過的資料
	//                         |                                            |                                                | 不一定所有的 repeatable read 等級都能阻止!
	//
	// 上述情境可以直接透過調整 isolation level 至 serializable level 解決 Write Skew 問題

	wg := new(sync.WaitGroup)
	wg.Add(2)

	go func(_wg *sync.WaitGroup) {
		defer _wg.Done()

		tx2, err := infra.RDB.Conn.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelRepeatableRead})
		checkError(err, "failed to start transaction:")
		logrus.Infoln("trnsaction 2 started.")

		var count int
		err = tx2.QueryRow("SELECT COUNT(amount) FROM wallets").Scan(&count)
		checkError(err, "failed to querying row:")
		logrus.Infof("trnsaction 2 selected, count = %v", count)

		time.Sleep(1 * time.Second)

		err = tx2.QueryRow("SELECT COUNT(amount) FROM wallets").Scan(&count)
		checkError(err, "failed to querying row:")
		logrus.Infof("trnsaction 2 selected, count = %v", count)

		_, err = tx2.Exec("UPDATE wallets SET amount = amount + 10000")
		checkError(err, "failed to execute:")
		logrus.Infoln("trnsaction 2 updated")

		// err = tx2.QueryRow("SELECT COUNT(amount) FROM wallets").Scan(&count)
		// checkError(err, "failed to querying row:")
		// logrus.Warnf("trnsaction 2 selected, count = %v", count)

		err = tx2.Commit()
		checkError(err, "failed to commit:")
		logrus.Infoln("trnsaction 2 committed.")

	}(wg)

	time.Sleep(1 * time.Millisecond)

	go func(_wg *sync.WaitGroup) {
		defer _wg.Done()

		tx1, err := infra.RDB.Conn.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelRepeatableRead})
		checkError(err, "failed to start transaction:")
		logrus.Infoln("trnsaction 1 started.")

		timeNow := time.Now().Format("2006-01-02 15:04:05")
		_, err = tx1.Exec("INSERT INTO wallets (user_id, amount, created_at, modified_at) VALUES (?, ?, ?, ?);",
			"2",
			100000,
			timeNow,
			timeNow,
		)
		checkError(err, "failed to execute:")
		logrus.Infoln("trnsaction 1 inserted.")

		err = tx1.Commit()
		checkError(err, "failed to commit:")
		logrus.Infoln("trnsaction 1 committed.")

	}(wg)

	wg.Wait()

	var count int
	err = infra.RDB.Conn.QueryRow("SELECT COUNT(amount) FROM wallets WHERE amount >= 110000").Scan(&count)
	checkError(err, "failed to querying row:")
	logrus.Warnf("SELECT COUNT(amount) FROM wallets WHERE amount >= 110000 is %v", count)

	return nil
}
