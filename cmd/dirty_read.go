package cmd

import (
	"context"
	"practice/internal/accessor"

	"github.com/spf13/cobra"

	_ "github.com/go-sql-driver/mysql"
)

var dirtyReadCmd = &cobra.Command{
	Use:   "dirty_read",
	Short: "",
	Long:  ``,
	RunE:  RunDirtyReadCmd,
}

func init() {
	rootCmd.AddCommand(dirtyReadCmd)
}

func RunDirtyReadCmd(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	infra := accessor.BuildAccessor()
	defer infra.Close(ctx)
	infra.InitRDB(ctx)

	infra.RDB.SimulateDirtyRead(ctx)

	return nil
}
