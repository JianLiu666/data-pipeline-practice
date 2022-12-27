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

	infra1 := accessor.BuildAccessor()
	defer infra1.Close(ctx)
	infra1.InitRDB(ctx)

	infra2 := accessor.BuildAccessor()
	defer infra2.Close(ctx)
	infra2.InitRDB(ctx)

	// business logic

	// 模擬不可重複讀情境 (Non-repeatable Read)
	// trx1 以 read committed 的隔離等級執行
	//
	// 1. trx1 讀取 wallets table id = 1 的錢包餘額 -> 100k
	// 2. trx2 讀取 wallets table id = 1 的錢包餘額 -> 100k
	// 3. some how, trx2 的執行速度比 trx1 還要快，扣除 60k 後並執行 commit
	// 4. trx1 再次讀取 wallets table id = 1 的錢包餘額變成 40k (發生 non-repeatable read!)
	//
	// 必須將 trx1 的隔離等級調整成 Repeatable Read 等級以上才能避免此問題

	wg := new(sync.WaitGroup)
	wg.Add(2)

	go func(_wg *sync.WaitGroup) {
		defer _wg.Done()

		tx1, err := infra1.RDB.Conn.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
		if err != nil {
			logrus.Panicf("failed to start transaction: %v", err)
		}

		rows, err := tx1.Query("SELECT amount FROM wallets WHERE id = 1;")
		if err != nil {
			logrus.Panicf("failed to qeury: %v", err)
		}

		var amount int = 0
		for rows.Next() {
			if err := rows.Scan(&amount); err != nil {
				logrus.Panicf("failed to scan rows: %v", err)
			}
		}
		logrus.Infof("trx1 read amount = %v", amount)

		time.Sleep(3 * time.Second)

		rows, err = tx1.Query("SELECT amount FROM wallets WHERE id = 1;")
		if err != nil {
			logrus.Panicf("failed to qeury: %v", err)
		}

		for rows.Next() {
			if err := rows.Scan(&amount); err != nil {
				logrus.Panicf("failed to scan rows: %v", err)
			}
		}
		logrus.Infof("trx1 read amount = %v", amount)

		if err := tx1.Commit(); err != nil {
			logrus.Panicf("failed to commit transaction: %v", err)
		}

		logrus.Infoln("trx1 committed.")

	}(wg)

	go func(_wg *sync.WaitGroup) {
		defer _wg.Done()

		time.Sleep(1 * time.Second)

		tx2, err := infra2.RDB.Conn.Begin()
		if err != nil {
			logrus.Panicf("failed to start transaction: %v", err)
		}

		rows, err := tx2.Query("SELECT amount FROM wallets WHERE id = 1;")
		if err != nil {
			logrus.Panicf("failed to qeury: %v", err)
		}

		var amount int = 0
		for rows.Next() {
			if err := rows.Scan(&amount); err != nil {
				logrus.Panicf("failed to scan rows: %v", err)
			}
		}
		logrus.Infof("trx2 read amount = %v", amount)

		if _, err := tx2.Exec("UPDATE wallets SET amount = amount - 60000 WHERE id = 1;"); err != nil {
			logrus.Panicf("failed to execute: %v", err)
		}

		if err := tx2.Commit(); err != nil {
			logrus.Panicf("failed to commit transaction: %v", err)
		}

		logrus.Infoln("trx2 committed.")

	}(wg)

	wg.Wait()

	return nil
}
