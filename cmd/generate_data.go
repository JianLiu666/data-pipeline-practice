package cmd

import (
	"context"
	"practice/internal/accessor"

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

	infra.RDB.GenerateData(ctx)

	return nil
}
