package cmd

import (
	"context"
	"fmt"
	"practice/internal/accessor"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var generateDataCmd = &cobra.Command{
	Use:   "generate_data",
	Short: "",
	Long:  ``,
	RunE:  RunGenerateDataCmd,
}

func init() {
	rootCmd.AddCommand(generateDataCmd)
}

func RunGenerateDataCmd(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	infra := accessor.BuildAccessor()
	defer infra.Close(ctx)

	infra.InitRDB(ctx)

	// business logic

	// 清空舊資料
	statements := `
	TRUNCATE TABLE users;
	TRUNCATE TABLE wallets;
	TRUNCATE TABLE logs;
	`
	if _, err := infra.RDB.Conn.Exec(statements); err != nil {
		logrus.Panicf("failed to execute sql task: %v", err)
	}

	// 初始化 users
	seq := 1
	for idx := 0; idx < 100; idx++ {
		sql := "INSERT INTO `users` (`account`, `password`, `nickname`, `email`, `created_at`, `modified_at`) VALUES "
		end := ","

		for i := 0; i < 100; i++ {
			timeNow := time.Now().Format("2006-01-02 15:04:05")

			sql += fmt.Sprintf("('%v', '%v', '%v', '%v', '%v', '%v')%v",
				fmt.Sprintf("user%v", seq),
				"password",
				fmt.Sprintf("user%v", seq),
				"email",
				timeNow,
				timeNow,
				end,
			)

			seq++
			if seq%100 == 0 {
				end = ";"
			}
		}

		if _, err := infra.RDB.Conn.Exec(sql); err != nil {
			logrus.Panicf("failed to execute sql task: %v", err)
		}
	}

	// 初始化 wallets
	seq = 1
	for idx := 0; idx < 100; idx++ {
		sql := "INSERT INTO `wallets` (`user_id`, `amount`, `created_at`, `modified_at`) VALUES "
		end := ","

		for i := 0; i < 100; i++ {
			timeNow := time.Now().Format("2006-01-02 15:04:05")

			sql += fmt.Sprintf("(%v, %v, '%v', '%v')%v",
				seq,
				100000,
				timeNow,
				timeNow,
				end,
			)

			seq++
			if seq%100 == 0 {
				end = ";"
			}
		}

		if _, err := infra.RDB.Conn.Exec(sql); err != nil {
			logrus.Panicf("failed to execute sql task: %v", err)
		}
	}

	return nil
}
