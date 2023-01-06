package cmd

import (
	"context"
	"practice/internal/accessor"

	"github.com/spf13/cobra"
)

var lostUpdateCmd = &cobra.Command{
	Use:   "lost_update",
	Short: "",
	Long:  ``,
	RunE:  RunLostUpdateCmd,
}

func init() {
	rootCmd.AddCommand(lostUpdateCmd)
}

func RunLostUpdateCmd(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	infra := accessor.BuildAccessor()
	defer infra.Close(ctx)
	infra.InitRDB(ctx)

	infra.RDB.SimulateLostUpdate(ctx)

	return nil
}
