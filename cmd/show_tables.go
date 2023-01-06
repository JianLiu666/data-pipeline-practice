package cmd

import (
	"context"
	"practice/internal/accessor"

	"github.com/spf13/cobra"
)

var showTablesCmd = &cobra.Command{
	Use:   "show_tables",
	Short: "",
	Long:  ``,
	RunE:  RunShowTablesCmd,
}

func init() {
	rootCmd.AddCommand(showTablesCmd)
}

func RunShowTablesCmd(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	infra := accessor.BuildAccessor()
	defer infra.Close(ctx)

	infra.InitRDB(ctx)

	infra.RDB.ShowTables(ctx)

	return nil
}
