package cmd

import (
	"context"
	"practice/internal/accessor"

	"github.com/spf13/cobra"
)

var writeSkew1Cmd = &cobra.Command{
	Use:   "write_skew_1",
	Short: "",
	Long:  ``,
	RunE:  RunWriteSkew1Cmd,
}

func init() {
	rootCmd.AddCommand(writeSkew1Cmd)
}

func RunWriteSkew1Cmd(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	infra := accessor.BuildAccessor()
	defer infra.Close(ctx)
	infra.InitRDB(ctx)

	infra.RDB.SimulateWriteSkew1(ctx)

	return nil
}
