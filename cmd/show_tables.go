package cmd

import (
	"context"
	"fmt"
	"practice/internal/accessor"
	"strings"

	"github.com/sirupsen/logrus"
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

	logrus.Info("========== start ==========")
	defer logrus.Info("=========== end ===========")

	// bussiness logic
	showTablesQuery, err := infra.RDB.Conn.Query("SHOW TABLES")
	checkError(err, "failed to query:")

	for showTablesQuery.Next() {
		var tbName string

		err = showTablesQuery.Scan(&tbName)
		checkError(err, "querying table failed:")

		selectQuery, err := infra.RDB.Conn.Query(fmt.Sprintf("SELECT * FROM %s", tbName))
		defer func() {
			err = selectQuery.Close()
			checkError(err, "failed to close cursor:")
		}()
		checkError(err, "executing query failed:")

		columns, err := selectQuery.Columns()
		checkError(err, fmt.Sprintf("failed to get columns from table %v", tbName))

		logrus.Infof("table name: %s -- columns: %v", tbName, strings.Join(columns, ", "))
	}

	return nil
}
