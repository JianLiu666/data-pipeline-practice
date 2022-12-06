package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var cfgFile string
var rootCmd = &cobra.Command{
	Use:   "root",
	Short: "",
	Long:  ``,
}

func init() {

}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
