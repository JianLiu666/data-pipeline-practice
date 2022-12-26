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

	// bussiness logic
	showTablesQuery, err := infra.RDB.Conn.Query("SHOW TABLES")
	if err != nil {
		logrus.Panicf("failed to execute sql task: %v", err)
	}

	for showTablesQuery.Next() {
		var tbName string

		if err := showTablesQuery.Scan(&tbName); err != nil {
			logrus.Errorf("querying tables failed: %v", err)
			continue
		}

		selectQuery, err := infra.RDB.Conn.Query(fmt.Sprintf("SELECT * FROM %s", tbName))
		defer func() {
			if err := selectQuery.Close(); err != nil {
				logrus.Errorf("failed to close cursor: %v", err)
			}
		}()
		if err != nil {
			logrus.Errorf("executing qeury failed: %v", err)
			continue
		}

		columns, err := selectQuery.Columns()
		if err != nil {
			logrus.Errorf("failed to get columns from table %v: %v", tbName, err)
			continue
		}

		logrus.Infof("table name: %s -- columns: %v", tbName, strings.Join(columns, ", "))
	}

	logrus.Info("done")
	return nil
}
