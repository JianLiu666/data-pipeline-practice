package cmd

import (
	"database/sql"
	"fmt"
	"practice/internal/config"
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
	cfg := config.NewFromViper()

	// connect to mysql database
	dsn := fmt.Sprintf("%s:%s@tcp(%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		cfg.MySQL.UserName,
		cfg.MySQL.Password,
		cfg.MySQL.Address,
		cfg.MySQL.DBName,
	)
	db, err := sql.Open("mysql", dsn)
	defer db.Close()
	if err != nil {
		logrus.Panicf("failed to open mysql: %v", err)
	}

	// check connection is still alive
	if err := db.Ping(); err != nil {
		logrus.Panicf("failed to ping mysql: %v", err)
	}

	// bussiness logic
	showTablesQuery, err := db.Query("SHOW TABLES")
	if err != nil {
		logrus.Panicf("failed to execute sql task: %v", err)
	}

	for showTablesQuery.Next() {
		var tbName string

		if err := showTablesQuery.Scan(&tbName); err != nil {
			logrus.Errorf("querying tables failed: %v", err)
			continue
		}

		selectQuery, err := db.Query(fmt.Sprintf("SELECT * FROM %s", tbName))
		defer selectQuery.Close()
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
