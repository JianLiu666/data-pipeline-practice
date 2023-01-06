package cmd

import (
	"context"
	"practice/internal/accessor"

	"github.com/spf13/cobra"
)

var lockFailed1Cmd = &cobra.Command{
	Use:   "lock_failed_1",
	Short: "",
	Long:  ``,
	RunE:  RunLockFailed1Cmd,
}

func init() {
	rootCmd.AddCommand(lockFailed1Cmd)
}

func RunLockFailed1Cmd(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	infra := accessor.BuildAccessor()
	defer infra.Close(ctx)

	infra.InitRDB(ctx)

	infra.RDB.SimulateLockFailed1(ctx)

	return nil
}
