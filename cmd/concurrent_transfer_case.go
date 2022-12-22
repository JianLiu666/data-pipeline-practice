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

	infra := accessor.BuildAccessor()
	defer infra.Close(ctx)

	infra.InitRDB(ctx)

	// business logic

	// 模擬併發交易情境
	// local 維護一組資料, 併發寫入 DB 後跟 DB 比較兩邊交易結果是否一致

	statements := `
	-- BEGIN;
	
	-- transaction
	UPDATE wallets INNER JOIN users ON wallets.user_id = users.id SET amount = amount + 1 WHERE users.id = 1;
	UPDATE wallets INNER JOIN users ON wallets.user_id = users.id SET amount = amount - 1 WHERE users.id = %v;
	
	-- logging
	INSERT INTO logs (deposit_user_id, withdraw_user_id, amount, created_at) VALUES (1, 2, 1, '%v');
	
	-- COMMIT;
	`

	wg := new(sync.WaitGroup)
	limit := 10000
	wg.Add(limit - 1)

	for i := 2; i <= limit; i++ {
		go func(_wg *sync.WaitGroup, idx int) {
			defer _wg.Done()

			tx, err := infra.RDB.Conn.BeginTx(context.TODO(), &sql.TxOptions{Isolation: sql.LevelReadUncommitted})
			if err != nil {
				logrus.Panicf("failed to start transaction: %v", err)
			}

			if _, err := tx.Exec(fmt.Sprintf(statements, idx, time.Now().Format("2006-01-02 15:04:05"))); err != nil {
				logrus.Panicf("failed to execute sql task: %v", err)
			}

			if err := tx.Commit(); err != nil {
				logrus.Panicf("failed to commit transaction: %v", err)
			}

		}(wg, i)
	}

	wg.Wait()

	return nil
}
