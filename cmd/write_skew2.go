package cmd

import (
	"context"
	"practice/internal/accessor"

	"github.com/spf13/cobra"
)

var writeSkew2Cmd = &cobra.Command{
	Use:   "write_skew_2",
	Short: "",
	Long:  ``,
	RunE:  RunWriteSkew2Cmd,
}

func init() {
	rootCmd.AddCommand(writeSkew2Cmd)
}

func RunWriteSkew2Cmd(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	infra := accessor.BuildAccessor()
	defer infra.Close(ctx)
	infra.InitRDB(ctx)

	infra.RDB.SimulateWriteSkew2(ctx)

	return nil
}
