package cmd

import (
	"context"
	"practice/internal/accessor"

	"github.com/spf13/cobra"

	_ "github.com/go-sql-driver/mysql"
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

	// // business logic
	// // local 維護一組資料, 併發寫入 DB 後跟 DB 比較兩邊交易結果是否一致

	return nil
}
