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

var lockFailed1Cmd = &cobra.Command{
	Use:   "lock_failed_1",
	Short: "",
	Long:  ``,
	RunE:  RunLockFailed1Cmd,
}

func init() {
	rootCmd.AddCommand(lockFailed1Cmd)
}

func RunLockFailed1Cmd(cmd *cobra.Command, args []string) error {
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

	// 模擬因為觸發覆蓋索引(Covering Index) 導致上鎖失敗
	//
	//  Transaction 1                                    Database                                       Transaction 2
	//       |                                              |                                                |
	//       |                                              |   wallets                                      |
	//       |                                              |  +----+--------+-----+                         |
	//       |                                              |  | id | amount | ... |                         |
	//       |                                              |  +----+--------+-----+                         |
	//       |                                              |  | 1  | 100000 | ... |                         |
	//       |                                              |  +----+--------+-----+                         |
	//       |                                              |                                                |
	//       |   START TRANSACTION                          |                                                |
	//       | -------------------------------------------> |                                                |
	//       |                                              |                            START TRANSACTION   |
	//       |                                              | <--------------------------------------------- |
	//       |   SELECT id FROM wallets WHERE user_id = 1   |                                                |
	//       |   LOCK IN SHARE MODE                         |                                                |
	//       | -------------------------------------------> |                                                |
	//       |                                              |   UPDATE wallets SET amount = 0 WHERE id = 1   |  預想情況中, 此時 Transaction 2 應該要阻塞直到 Transaction 1 結束後
	//       |                                              | <--------------------------------------------- |  才能執行, 但 Transaction 1 卻沒有成功鎖上
	//       |                                              |                                       COMMIT   |
	//       |                                              | <--------------------------------------------- |
	//       |   COMMIT                                     |                                                |
	//       | -------------------------------------------> |                                                |
	//       |                                              |                                                |
	//
	// 原因在於 transaction 1 在執行過程中不需要回到 clustered index 查找資料，因此只需要對 secondary index 上鎖 (即 user_id)
	// 而 transaction 2 請求的鎖是在 clustered index, 因此 transaction 2 可以很順利的執行不必等待 transaction 1 結束
	//
	// 解決辦法:
	//
	// 1. 若 transaction 1 修改查詢欄位, 迫使執行時必須回到 clustered index 查找該欄位, 才會使得 transaction 2 一定得等到 transaction 1 結束後才可繼續動作
	//
	// 2. 將上鎖指令從 LOCK IN SHARE MODE 升級成 FOR UPDATE, 也會同時將 clustered index 上鎖

	wg := new(sync.WaitGroup)
	wg.Add(2)

	go func(_wg *sync.WaitGroup) {
		defer _wg.Done()

		tx1, err := infra.RDB.Conn.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelRepeatableRead})
		checkError(err, "failed to start transaction:")
		logrus.Infoln("trnsaction 1 started.")

		var id int
		err = tx1.QueryRow("SELECT id FROM wallets WHERE user_id = 1 LOCK IN SHARE MODE").Scan(&id)
		checkError(err, "failed to querying row:")
		logrus.Infoln("trnsaction 1 selected")

		time.Sleep(1 * time.Second)

		err = tx1.Commit()
		checkError(err, "failed to commit:")
		logrus.Infoln("trnsaction 1 committed.")

	}(wg)

	time.Sleep(1 * time.Millisecond)

	go func(_wg *sync.WaitGroup) {
		defer _wg.Done()

		tx2, err := infra.RDB.Conn.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelRepeatableRead})
		checkError(err, "failed to start transaction:")
		logrus.Infoln("trnsaction 2 started.")

		_, err = tx2.Exec("UPDATE wallets SET amount = amount - 10000 WHERE id = 1")
		checkError(err, "failed to execute:")
		logrus.Infoln("trnsaction 2 updated.")

		err = tx2.Commit()
		checkError(err, "failed to commit:")
		logrus.Infoln("trnsaction 2 committed.")

	}(wg)

	wg.Wait()

	return nil
}
