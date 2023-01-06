package cmd

import (
	"context"
	"practice/internal/accessor"

	"github.com/spf13/cobra"
)

var readSkewCmd = &cobra.Command{
	Use:   "read_skew",
	Short: "",
	Long:  ``,
	RunE:  RunReadSkewCmd,
}

func init() {
	rootCmd.AddCommand(readSkewCmd)
}

func RunReadSkewCmd(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	infra := accessor.BuildAccessor()
	defer infra.Close(ctx)
	infra.InitRDB(ctx)

	infra.RDB.SimulateReadSkew(ctx)

	return nil
}
