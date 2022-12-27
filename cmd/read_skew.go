package cmd

import (
	"context"
	"database/sql"
	"fmt"
	"practice/internal/accessor"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var readSkewCmd = &cobra.Command{
	Use:   "read_skew",
	Short: "",
	Long:  ``,
	RunE:  RunReadSkewCmd,
}

func init() {
	rootCmd.AddCommand(readSkewCmd)
}

func RunReadSkewCmd(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	infra1 := accessor.BuildAccessor()
	defer infra1.Close(ctx)
	infra1.InitRDB(ctx)

	infra2 := accessor.BuildAccessor()
	defer infra2.Close(ctx)
	infra2.InitRDB(ctx)

	// bussiness logic

	// 模擬讀偏差情境 (Read Skew)
	//
	// 0. 對 wallets table id = 1 的錢包餘額扣款 40k 並新增紀錄至 logs table
	// 1. trx1 讀取 wallets table id = 1 的錢包餘額 -> 60k
	// 2. trx2 對 wallets table id = 1 的錢包再次扣款 40k，也新增紀錄至 logs table
	// 2. trx2 commit
	// 1. trx1 取得 logs table 中 id = 1 的所有交易紀錄發現交易總額為 80k, 與錢包餘額不符 (發生 read skew!)
	//
	// 必須將 trx1 的隔離等級調整成 Repeatable Read 等級以上才能避免此問題

	// 前置作業，產生一筆交易紀錄
	statement := `
	UPDATE wallets SET amount = amount - 40000 WHERE id = 1;
	INSERT INTO logs (deposit_user_id, withdraw_user_id, amount, created_at) VALUES (0, 1, 40000, '%v');
	`

	if _, err := infra1.RDB.Conn.Exec(fmt.Sprintf(statement, time.Now().Format("2006-01-02 15:04:05"))); err != nil {
		logrus.Panicf("failed to execute sql task: %v", err)
	}

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

		var walletAmount int = 0
		for rows.Next() {
			if err := rows.Scan(&walletAmount); err != nil {
				logrus.Panicf("failed to scan rows: %v", err)
			}
		}
		logrus.Infof("trx1 read wallet amount = %v", walletAmount)

		time.Sleep(3 * time.Second)

		rows, err = tx1.Query("SELECT amount FROM logs WHERE withdraw_user_id = 1;")
		if err != nil {
			logrus.Panicf("failed to qeury: %v", err)
		}

		var withdrawAmount int = 0
		for rows.Next() {
			var amount int = 0
			if err := rows.Scan(&amount); err != nil {
				logrus.Panicf("failed to scan rows: %v", err)
			}
			withdrawAmount += amount
		}

		logrus.Infof("trx1 read withdraw amount = %v", withdrawAmount)

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

		if _, err := tx2.Exec(fmt.Sprintf(statement, time.Now().Format("2006-01-02 15:04:05"))); err != nil {
			logrus.Panicf("failed to execute sql task: %v", err)
		}

		if err := tx2.Commit(); err != nil {
			logrus.Panicf("failed to commit transaction: %v", err)
		}

		logrus.Infoln("trx2 committed.")

	}(wg)

	wg.Wait()

	return nil
}
