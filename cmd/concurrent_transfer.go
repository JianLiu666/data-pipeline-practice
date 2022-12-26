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

var concurrentTransferCmd = &cobra.Command{
	Use:   "concurrent_transfer",
	Short: "",
	Long:  ``,
	RunE:  RunConcurrentTransferCmd,
}

func init() {
	rootCmd.AddCommand(concurrentTransferCmd)
}

func RunConcurrentTransferCmd(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	infra1 := accessor.BuildAccessor()
	defer infra1.Close(ctx)
	infra1.InitRDB(ctx)

	infra2 := accessor.BuildAccessor()
	defer infra2.Close(ctx)
	infra2.InitRDB(ctx)

	// business logic

	// UPDATE ... WHERE ... 命令屬於排他鎖
	// desc: sets an exclusive next-key lock on every record the search encounters.
	statements := `	
	-- transaction
	UPDATE wallets INNER JOIN users ON wallets.user_id = users.id SET amount = amount + 1 WHERE users.id = 1;
	UPDATE wallets INNER JOIN users ON wallets.user_id = users.id SET amount = amount - 1 WHERE users.id = %v;
	
	-- logging
	INSERT INTO logs (deposit_user_id, withdraw_user_id, amount, created_at) VALUES (1, 2, 1, '%v');
	`

	// 模擬併發交易情境

	wg := new(sync.WaitGroup)
	wg.Add(2)

	go func(_wg *sync.WaitGroup) {
		defer _wg.Done()

		tx1, err := infra1.RDB.Conn.Begin()
		if err != nil {
			logrus.Panicf("failed to start transaction: %v", err)
		}

		if _, err := tx1.Exec(fmt.Sprintf(statements, 2, time.Now().Format("2006-01-02 15:04:05"))); err != nil {
			logrus.Panicf("failed to execute sql task: %v", err)
		}

		time.Sleep(3 * time.Second)

		if err := tx1.Rollback(); err != nil {
			logrus.Panicf("failed to rollback transaction: %v", err)
		}
	}(wg)

	go func(_wg *sync.WaitGroup) {
		defer _wg.Done()

		tx2, err := infra2.RDB.Conn.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelReadUncommitted})
		if err != nil {
			logrus.Panicf("failed to start transaction: %v", err)
		}

		if _, err := tx2.Exec(fmt.Sprintf(statements, 2, time.Now().Format("2006-01-02 15:04:05"))); err != nil {
			logrus.Panicf("failed to execute sql task: %v", err)
		}

		if err := tx2.Commit(); err != nil {
			logrus.Panicf("failed to commit transaction: %v", err)
		}
	}(wg)

	wg.Wait()

	return nil
}
