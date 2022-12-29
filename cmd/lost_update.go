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

	infra1 := accessor.BuildAccessor()
	defer infra1.Close(ctx)
	infra1.InitRDB(ctx)

	infra2 := accessor.BuildAccessor()
	defer infra2.Close(ctx)
	infra2.InitRDB(ctx)

	// 模擬更新丟失情境 (Lost Update)
	//
	// 1. trx1 更新 users table id = 1 的 nickname 成 trx1
	// 2. trx2 更新 users table id = 1 的 nickname 成 trx2, 此時會先阻塞直到 trx1 commit 為止
	// 3. trx1 commit
	// 4. trx2 commit
	// 5. 讀取 users table id = 1 的資料發現 nickname 為 trx2 (發生 lost update!)
	//
	// 該情境的解決方式只能透過樂觀鎖的方式解決
	// e.g.
	//   加入新的 column 'version' 來當作查詢條件
	//   SELET version FROM users WHERE id = 1 (version = 0)
	//   UPDATE users SET nickname = 'trx1', version = version + 1 WHERE id = 1 AND version = 0
	//
	// 如果是 number 類型的欄位可以透過 DB 的原子性操作保護
	// e.g.
	//   UPDATE wallets SET amount = amount + 1 WHERE user_id = 1

	wg := new(sync.WaitGroup)
	wg.Add(2)

	go func(_wg *sync.WaitGroup) {
		defer _wg.Done()

		logrus.Infoln("trx1 start.")

		tx1, err := infra1.RDB.Conn.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelRepeatableRead})
		if err != nil {
			logrus.Panicf("failed to start transaction: %v", err)
		}

		_, err = tx1.Exec("UPDATE users SET nickname = 'trx1' WHERE id = 1")
		if err != nil {
			logrus.Panicf("failed to execute: %v", err)
		}

		logrus.Infoln("trx1 update nickname to 'trx1'")
		time.Sleep(6 * time.Second)

		err = tx1.Commit()
		if err != nil {
			logrus.Panicf("failed to commit transaction: %v", err)
		}

		logrus.Infoln("trx1 committed.")

	}(wg)

	go func(_wg *sync.WaitGroup) {
		defer _wg.Done()

		time.Sleep(1 * time.Second)
		logrus.Infoln("trx2 start.")

		tx2, err := infra2.RDB.Conn.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelRepeatableRead})
		if err != nil {
			logrus.Panicf("failed to start transaction: %v", err)
		}

		_, err = tx2.Exec("UPDATE users SET nickname = 'trx2' WHERE id = 1")
		if err != nil {
			logrus.Panicf("failed to execute: %v", err)
		}

		logrus.Infoln("trx2 update nickname to 'trx2'")

		err = tx2.Commit()
		if err != nil {
			logrus.Panicf("failed to commit transaction: %v", err)
		}

		logrus.Infoln("trx2 committed.")

	}(wg)

	wg.Wait()

	rows, err := infra1.RDB.Conn.Query("SELECT nickname FROM users WHERE id = 1;")
	if err != nil {
		logrus.Panicf("failed to qeury: %v", err)
	}

	var nickname string
	for rows.Next() {
		err = rows.Scan(&nickname)
		if err != nil {
			logrus.Panicf("failed to scan rows: %v", err)
		}
	}

	logrus.Infof("nickname = %v", nickname)

	return nil
}
