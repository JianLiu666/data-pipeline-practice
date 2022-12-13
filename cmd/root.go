package cmd

import (
	"fmt"
	"os"
	"practice/internal/config"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfgFile string
var rootCmd = &cobra.Command{
	Use:   "root",
	Short: "",
	Long:  ``,
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "f", "./conf.d/env.yaml", "config file (default is ./conf.d/env.yaml)")
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	cfg := config.NewFromViper()
	fmt.Println(cfg)
}

func initConfig() {
	viper.SetConfigFile(cfgFile)
	viper.AutomaticEnv()
}
