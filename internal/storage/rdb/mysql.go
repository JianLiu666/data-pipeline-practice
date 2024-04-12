package rdb

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/sirupsen/logrus"

	_ "github.com/go-sql-driver/mysql"
)

type mysql struct {
	conn *sql.DB
}

// NewMysqlClient New MySQL Client Driver
// @param ctx
// @param userName         mysql dsn
// @param password         mysql dsn
// @param address          mysql dsn
// @param dbName           mysql dsn
// @param connMaxLifetime  sets the maximum number of connections in the idle connection pool.
// @param maxOpenConns     sets the maximum number of open connections to the database.
// @param maxIdleConns     sets the maximum amount of time a connection may be reused.
func NewMysqlClient(ctx context.Context, userName, password, address, dbName string, connMaxLifetime time.Duration, maxOpenConns, maxIdleConns int) Rdb {
	dsn := fmt.Sprintf("%s:%s@tcp(%s)/%s?charset=utf8mb4&parseTime=True&loc=Local&multiStatements=true",
		userName,
		password,
		address,
		dbName,
	)

	conn, err := sql.Open("mysql", dsn)
	if err != nil {
		logrus.Errorf("failed to open mysql database: %v", err)
	}

	if err := conn.Ping(); err != nil {
		logrus.Panicf("failed to ping mysql: %v", err)
	}

	conn.SetConnMaxIdleTime(connMaxLifetime)
	conn.SetMaxOpenConns(maxOpenConns)
	conn.SetMaxIdleConns(maxIdleConns)

	return &mysql{
		conn: conn,
	}
}

func (m *mysql) Shutdown(ctx context.Context) {
	if err := m.conn.Close(); err != nil {
		logrus.Panicf("failed to close mysql connection: %v", err)
	}
}

func (m *mysql) ShowTables(ctx context.Context) {
	logrus.Info("========== start ==========")
	defer logrus.Info("=========== end ===========")

	// business logic
	showTablesQuery, err := m.conn.Query("SHOW TABLES")
	checkError(err, "failed to query:")

	for showTablesQuery.Next() {
		var tbName string

		err = showTablesQuery.Scan(&tbName)
		checkError(err, "querying table failed:")

		selectQuery, err := m.conn.Query(fmt.Sprintf("SELECT * FROM %s", tbName))
		defer func() {
			err = selectQuery.Close()
			checkError(err, "failed to close cursor:")
		}()
		checkError(err, "executing query failed:")

		columns, err := selectQuery.Columns()
		checkError(err, fmt.Sprintf("failed to get columns from table %v", tbName))

		logrus.Infof("table name: %s -- columns: %v", tbName, strings.Join(columns, ", "))
	}
}

func (m *mysql) GenerateData(ctx context.Context) {
	logrus.Info("========== start ==========")
	defer logrus.Info("=========== end ===========")

	// 清空舊資料
	statements := `
	TRUNCATE TABLE users;
	TRUNCATE TABLE wallets;
	TRUNCATE TABLE logs;
	`
	if _, err := m.conn.Exec(statements); err != nil {
		logrus.Panicf("failed to execute sql task: %v", err)
	}

	// 初始化 users
	seq := 1
	for idx := 0; idx < 100; idx++ {
		sql := "INSERT INTO `users` (`account`, `password`, `nickname`, `email`, `created_at`, `modified_at`) VALUES "
		end := ","

		for i := 0; i < 100; i++ {
			timeNow := time.Now().Format("2006-01-02 15:04:05")

			sql += fmt.Sprintf("('%v', '%v', '%v', '%v', '%v', '%v')%v",
				fmt.Sprintf("user%v", seq),
				"password",
				fmt.Sprintf("user%v", seq),
				"email",
				timeNow,
				timeNow,
				end,
			)

			seq++
			if seq%100 == 0 {
				end = ";"
			}
		}

		if _, err := m.conn.Exec(sql); err != nil {
			logrus.Panicf("failed to execute sql task: %v", err)
		}
	}

	// 初始化 wallets
	seq = 1
	for idx := 0; idx < 100; idx++ {
		sql := "INSERT INTO `wallets` (`user_id`, `amount`, `created_at`, `modified_at`) VALUES "
		end := ","

		for i := 0; i < 100; i++ {
			timeNow := time.Now().Format("2006-01-02 15:04:05")

			sql += fmt.Sprintf("(%v, %v, '%v', '%v')%v",
				seq,
				100000,
				timeNow,
				timeNow,
				end,
			)

			seq++
			if seq%100 == 0 {
				end = ";"
			}
		}

		if _, err := m.conn.Exec(sql); err != nil {
			logrus.Panicf("failed to execute sql task: %v", err)
		}
	}
}

func (m *mysql) SimulateDirtyRead(ctx context.Context) {
	// init
	_, err := m.conn.Exec("TRUNCATE TABLE logs")
	checkError(err, "failed to execute:")

	logrus.Info("========== start ==========")
	defer logrus.Info("=========== end ===========")

	// 模擬髒讀(Dirty Read) 情境
	//
	//                            Transaction 1                                Database                              Transaction 2
	//                                  |                                         |                                       |
	//                                  |                                         |   logs                                |
	//                                  |                                         |  +----+-----------------+-----+       |
	//                                  |                                         |  | id | deposit_user_id | ... |       |
	//                                  |                                         |  +----+-----------------+-----+       |
	//   logs                           |   START TRANSACTION                     |                                       |
	//  +----+-----------------+-----+  | --------------------------------------> |                                       |
	//  | id | deposit_user_id | ... |  |   INSERT INTO logs (...) VALUES (...)   |                                       |
	//  +----+-----------------+-----+  | --------------------------------------> |                                       |
	//  | 1  | 1               | ... |  |                                         |                   START TRANSACTION   |
	//  +----+-----------------+-----+  |                                         | <------------------------------------ |
	//                                  |                                         |           SELECT count(*) FROM logs   |  isolation level 為 read uncommitted 時會讀到
	//                                  |                                         | <------------------------------------ |  transaction 1 尚未 committed 的資料導致 dirty read
	//                                  |   ROLLBACK                              |                                       |  必須是 read committed 以上的等級才可避免
	//                                  | --------------------------------------> |                                       |
	//                                  |                                         |                              COMMIT   |
	//                                  |                                         | <------------------------------------ |
	//                                  |                                         |                                       |

	// 執行 trx1: 寫入一筆 log
	tx1, err := m.conn.Begin()
	checkError(err, "failed to start transaction:")

	_, err = tx1.Exec("INSERT INTO logs (deposit_user_id, withdraw_user_id, amount, created_at) VALUES (1, 2, 1, '2022-12-22 20:57:47');")
	checkError(err, "failed to execute:")

	// 在 trx1 結束前, 執行 trx2 取得相同 table 裡面的資料數量
	// 強制本次的 transaction isolation level 使用 read-uncommitted 等級
	tx2, err := m.conn.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelReadUncommitted})
	checkError(err, "failed to start transaction:")

	var count int
	err = tx2.QueryRow("SELECT count(*) FROM logs;").Scan(&count)
	checkError(err, "failed to query:")

	logrus.Warnf("Read Uncommitted: %v", count)

	// 結束 trx2
	err = tx2.Commit()
	checkError(err, "failed to commit transaction:")

	// 結束 trx1
	err = tx1.Rollback()
	checkError(err, "failed to rollback transaction:")
}

func (m *mysql) SimulateReadSkew(ctx context.Context) {
	// init
	_, err := m.conn.Exec("TRUNCATE TABLE wallets")
	checkError(err, "failed to execute:")

	timeNow := time.Now().Format("2006-01-02 15:04:05")
	_, err = m.conn.Exec("INSERT INTO wallets (user_id, amount, created_at, modified_at) VALUES (?, ?, ?, ?);",
		"1",
		100000,
		timeNow,
		timeNow,
	)
	checkError(err, "failed to execute:")

	logrus.Info("========== start ==========")
	defer logrus.Info("=========== end ===========")

	// 模擬讀偏差(Read Skew) 情境，又稱不可重複讀(Non-repeatable Read)
	//
	//                    Transaction 1                                                   Database                                    Transaction 2
	//                         |                                                             |                                             |
	//                         |                                                             |   wallets                                   |
	//                         |                                                             |  +----+--------+-----+                      |
	//                         |                                                             |  | id | amount | ... |                      |
	//                         |                                                             |  +----+--------+-----+                      |
	//                         |                                                             |  | 1  | 100000 | ... |                      |
	//                         |                                                             |  +----+--------+-----+                      |
	//                         |   START TRANSACTION                                         |                                             |
	//                         | ----------------------------------------------------------> |                                             |
	//                         |                                                             |   START TRANSACTION                         |
	//   wallets               |                                                             | <------------------------------------------ |
	//  +----+--------+-----+  |   UPDATE wallets SET amount = amount - 60000 WHERE id = 1   |                                             |
	//  | id | amount | ... |  | ----------------------------------------------------------> |                                             |   wallets
	//  +----+--------+-----+  |                                                             |   SELECT amount FROM wallets WHERE id = 1   |  +----+--------+-----+
	//  | 1  |  40000 | ... |  |                                                             | <------------------------------------------ |  | id | amount | ... |
	//  +----+--------+-----+  |   COMMIT                                                    |                                             |  +----+--------+-----+
	//                         | ----------------------------------------------------------> |                                             |  | 1  | 100000 | ... |
	//                         |                                                             |                                             |  +----+--------+-----+
	//                         |                                                             |                                             |
	//                         |                                                             |                                             |   wallets
	//                         |                                                             |   SELECT amount FROM wallets WHERE id = 1   |  +----+--------+-----+
	//                         |                                                             | <------------------------------------------ |  | id | amount | ... |
	//                         |                                                             |   COMMIT                                    |  +----+--------+-----+
	//                         |                                                             | <------------------------------------------ |  | 1  |  40000 | ... |
	//                         |                                                             |                                             |  +----+--------+-----+
	//                         |                                                             |                                             |
	//                         |                                                             |                                             |  發生同一個 transaction 內讀取到兩次不同的
	//                         |                                                             |                                             |  結果 (read skew)
	//                         |                                                             |                                             |  必須是 repeatable read 以上的等級才可避免
	//                         |                                                             |                                             |

	tx1, err := m.conn.Begin()
	checkError(err, "failed to start transaction:")

	tx2, err := m.conn.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
	checkError(err, "failed to start transaction:")

	_, err = tx1.Exec("UPDATE wallets SET amount = amount - 60000 WHERE id = 1;")
	checkError(err, "failed to execute:")

	var amount int
	err = tx2.QueryRow("SELECT amount FROM wallets WHERE id = 1;").Scan(&amount)
	checkError(err, "failed to querying row:")

	logrus.Infof("amount = %v", amount)

	err = tx1.Commit()
	checkError(err, "failed to commit transaction:")

	err = tx2.QueryRow("SELECT amount FROM wallets WHERE id = 1;").Scan(&amount)
	checkError(err, "failed to querying row:")

	logrus.Warnf("amount = %v", amount)

	err = tx2.Commit()
	checkError(err, "failed to commit transaction:")
}

func (m *mysql) SimulateLostUpdate(ctx context.Context) {
	// init
	_, err := m.conn.Exec("TRUNCATE TABLE wallets")
	checkError(err, "failed to execute:")

	timeNow := time.Now().Format("2006-01-02 15:04:05")
	_, err = m.conn.Exec("INSERT INTO wallets (user_id, amount, created_at, modified_at) VALUES (?, ?, ?, ?);",
		"1",
		100000,
		timeNow,
		timeNow,
	)
	checkError(err, "failed to execute:")

	logrus.Info("========== start ==========")
	defer logrus.Info("=========== end ===========")

	// 模擬更新丟失(Lost Update) 情境
	//
	//                    Transaction 1                                          Database                                           Transaction 2
	//                         |                                                    |                                                    |
	//                         |                                                    |   wallets                                          |
	//                         |                                                    |  +----+--------+-----+                             |
	//                         |                                                    |  | id | amount | ... |                             |
	//                         |                                                    |  +----+--------+-----+                             |
	//                         |                                                    |  | 1  | 100000 | ... |                             |
	//                         |                                                    |  +----+--------+-----+                             |
	//                         |                                                    |                                                    |
	//                         |                                                    |                                START TRANSACTION   |
	//                         |                                                    | <------------------------------------------------- |
	//                         |   START TRANSACTION                                |                                                    |
	//                         | -------------------------------------------------> |                                                    |   wallets
	//                         |                                                    |          SELECT amount FROM wallets WHERE id = 1   |  +----+--------+-----+
	//   wallets               |                                                    | <------------------------------------------------- |  | id | amount | ... |
	//  +----+--------+-----+  |   SELECT amount FROM wallets WHERE id = 1          |                                                    |  +----+--------+-----+
	//  | id | amount | ... |  | -------------------------------------------------> |                                                    |  | 1  | 100000 | ... |
	//  +----+--------+-----+  |                                                    |                                                    |  +----+--------+-----+
	//  | 1  | 100000 | ... |  |                                                    |                                                    |
	//  +----+--------+-----+  |                                                    |                                                    |
	//                         |                                                    |                                                    |   wallets
	//                         |                                                    |   UPDATE wallets SET amount = 60000 WHERE id = 1   |  +----+--------+-----+
	//                         |                                                    | <------------------------------------------------- |  | id | amount | ... |
	//                         |                                                    |                                           COMMIT   |  +----+--------+-----+
	//   wallets               |                                                    | <------------------------------------------------- |  | 1  |  60000 | ... |
	//  +----+--------+-----+  |   UPDATE wallets SET amount = 40000 WHERE id = 1   |                                                    |  +----+--------+-----+
	//  | id | amount | ... |  | -------------------------------------------------> |                                                    |
	//  +----+--------+-----+  |   COMMIT                                           |                                                    |  transaction2 的更新結果最後被 transaction1 覆蓋掉
	//  | 1  |  40000 | ... |  | -------------------------------------------------> |                                                    |  造成 lost update
	//  +----+--------+-----+  |                                                    |                                                    |
	//                         |                                                    |                                                    |
	//
	// 兩種解決 Lost Update 的辦法:
	//
	// 1. 交給 Database 的 atomic write
	//     - 改寫 UPDATE wallets SET amount = {value} WHERE id = 1 成 UPDATE wallets SET amount = amount - {value} WHERE id = 1
	//     - 要特別注意 transaction failed 時的重試流程，若未能保證冪等性可能造成重複扣款問題
	//
	// 2. 自行實現樂觀鎖流程 (CAS)
	//     - 改寫 UPDATE wallets SET amount = {value} WHERE id = 1 成 UPDATE wallets SET amount = {new} WHERE id = 1 AND amount = {old}
	//     - 強制 transaction 1 更新失敗, 但要自行驗證 transaction 執行結果是否符合預期

	tx2, err := m.conn.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelRepeatableRead})
	checkError(err, "failed to start transaction:")

	tx1, err := m.conn.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelRepeatableRead})
	checkError(err, "failed to start transaction:")

	var amount_tx1, amount_tx2, amount_result int

	err = tx2.QueryRow("SELECT amount FROM wallets WHERE id = 1").Scan(&amount_tx2)
	checkError(err, "failed to querying row:")

	err = tx1.QueryRow("SELECT amount FROM wallets WHERE id = 1").Scan(&amount_tx1)
	checkError(err, "failed to querying row:")

	// 表示業務邏輯處理結果
	amount_tx2 = 60000
	_, err = tx2.Exec("UPDATE wallets SET amount = ? WHERE id = 1", amount_tx2)
	checkError(err, "failed to execute:")

	err = tx2.Commit()
	checkError(err, "failed to commit:")

	// 表示業務邏輯處理結果
	amount_tx1 = 40000
	_, err = tx1.Exec("UPDATE wallets SET amount = ? WHERE id = 1", amount_tx1)
	checkError(err, "failed to execute:")

	err = tx1.Commit()
	checkError(err, "failed to commit:")

	err = m.conn.QueryRow("SELECT amount FROM wallets WHERE id = 1").Scan(&amount_result)
	checkError(err, "failed to querying row:")

	logrus.Warnf("Amount = %v", amount_result)
}

func (m *mysql) SimulateWriteSkew1(ctx context.Context) {
	// init
	_, err := m.conn.Exec("TRUNCATE TABLE wallets")
	checkError(err, "failed to execute:")

	timeNow := time.Now().Format("2006-01-02 15:04:05")
	_, err = m.conn.Exec("INSERT INTO wallets (user_id, amount, created_at, modified_at) VALUES (?, ?, ?, ?);",
		"1",
		100000,
		timeNow,
		timeNow,
	)
	checkError(err, "failed to execute:")

	logrus.Info("========== start ==========")
	defer logrus.Info("=========== end ===========")

	// 模擬因為幻讀(Phantom Read) 造成寫偏差(Write Skew) 情境
	//
	//                    Transaction 1                                  Database                                       Transaction 2
	//                         |                                            |                                                |
	//                         |                                            |   wallets                                      |
	//                         |                                            |  +----+--------+-----+                         |
	//                         |                                            |  | id | amount | ... |                         |
	//                         |                                            |  +----+--------+-----+                         |
	//                         |                                            |  | 1  | 100000 | ... |                         |
	//                         |                                            |  +----+--------+-----+                         |
	//                         |                                            |                                                |
	//                         |                                            |                            START TRANSACTION   |
	//                         |                                            | <--------------------------------------------- |   wallets
	//                         |                                            |                   SELECT amount FROM wallets   |  +----+--------+-----+
	//                         |                                            | <--------------------------------------------- |  | id | amount | ... |
	//                         |   START TRANSACTION                        |                                                |  +----+--------+-----+
	//   wallets               | -----------------------------------------> |                                                |  | 1  | 100000 | ... |
	//  +----+--------+-----+  |   INSERT INTO wallets (...) VALUES (...)   |                                                |  +----+--------+-----+
	//  | id | amount | ... |  | -----------------------------------------> |                                                |
	//  +----+--------+-----+  |   COMMIT                                   |                                                |
	//  | 1  | 100000 | ... |  | -----------------------------------------> |                                                |   wallets
	//  +----+--------+-----+  |                                            |                   SELECT amount FROM wallets   |  +----+--------+-----+
	//  | 2  | 100000 | ... |  |                                            | <--------------------------------------------- |  | id | amount | ... |
	//  +----+--------+-----+  |                                            |                                                |  +----+--------+-----+
	//                         |                                            |                                                |  | 1  | 100000 | ... |
	//                         |                                            |                                                |  +----+--------+-----+
	//                         |                                            |                                                |
	//                         |                                            |                                                |
	//                         |                                            |                                                |   wallets
	//                         |                                            |   UPDATE wallets SET amount = amount + 10000   |  +----+--------+-----+
	//                         |                                            | <--------------------------------------------- |  | id | amount | ... |
	//                         |                                            |                                       COMMIT   |  +----+--------+-----+
	//                         |                                            | <--------------------------------------------- |  | 1  | 110000 | ... |
	//                         |                                            |                                                |  +----+--------+-----+
	//                         |                                            |                                                |  | 2  | 110000 | ... |
	//                         |                                            |                                                |  +----+--------+-----+
	//                         |                                            |                                                |
	//                         |                                            |                                                | 發生 write skew 導致更新到未讀取過的資料
	//                         |                                            |                                                | 不一定所有的 repeatable read 等級都能阻止!
	//
	// 上述情境可以直接透過調整 isolation level 至 serializable level 解決 Write Skew 問題

	wg := new(sync.WaitGroup)
	wg.Add(2)

	go func(_wg *sync.WaitGroup) {
		defer _wg.Done()

		tx2, err := m.conn.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelRepeatableRead})
		checkError(err, "failed to start transaction:")
		logrus.Infoln("transaction 2 started.")

		var count int
		err = tx2.QueryRow("SELECT COUNT(amount) FROM wallets").Scan(&count)
		checkError(err, "failed to querying row:")
		logrus.Infof("transaction 2 selected, count = %v", count)

		time.Sleep(1 * time.Second)

		err = tx2.QueryRow("SELECT COUNT(amount) FROM wallets").Scan(&count)
		checkError(err, "failed to querying row:")
		logrus.Infof("transaction 2 selected, count = %v", count)

		_, err = tx2.Exec("UPDATE wallets SET amount = amount + 10000")
		checkError(err, "failed to execute:")
		logrus.Infoln("transaction 2 updated")

		// err = tx2.QueryRow("SELECT COUNT(amount) FROM wallets").Scan(&count)
		// checkError(err, "failed to querying row:")
		// logrus.Warnf("transaction 2 selected, count = %v", count)

		err = tx2.Commit()
		checkError(err, "failed to commit:")
		logrus.Infoln("transaction 2 committed.")

	}(wg)

	time.Sleep(1 * time.Millisecond)

	go func(_wg *sync.WaitGroup) {
		defer _wg.Done()

		tx1, err := m.conn.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelRepeatableRead})
		checkError(err, "failed to start transaction:")
		logrus.Infoln("transaction 1 started.")

		timeNow := time.Now().Format("2006-01-02 15:04:05")
		_, err = tx1.Exec("INSERT INTO wallets (user_id, amount, created_at, modified_at) VALUES (?, ?, ?, ?);",
			"2",
			100000,
			timeNow,
			timeNow,
		)
		checkError(err, "failed to execute:")
		logrus.Infoln("transaction 1 inserted.")

		err = tx1.Commit()
		checkError(err, "failed to commit:")
		logrus.Infoln("transaction 1 committed.")

	}(wg)

	wg.Wait()

	var count int
	err = m.conn.QueryRow("SELECT COUNT(amount) FROM wallets WHERE amount >= 110000").Scan(&count)
	checkError(err, "failed to querying row:")
	logrus.Warnf("SELECT COUNT(amount) FROM wallets WHERE amount >= 110000 is %v", count)
}

func (m *mysql) SimulateWriteSkew2(ctx context.Context) {
	// init
	_, err := m.conn.Exec("TRUNCATE TABLE wallets")
	checkError(err, "failed to execute:")

	timeNow := time.Now().Format("2006-01-02 15:04:05")
	_, err = m.conn.Exec("INSERT INTO wallets (user_id, amount, created_at, modified_at) VALUES (?, ?, ?, ?);",
		"1",
		100000,
		timeNow,
		timeNow,
	)
	checkError(err, "failed to execute:")

	logrus.Info("========== start ==========")
	defer logrus.Info("=========== end ===========")

	// 模擬因為幻讀(Phantom Read) 造成寫偏差(Write Skew) 情境
	//
	//                    Transaction 1                                                   Database                                                    Transaction 2
	//                         |                                                             |                                                             |
	//                         |                                                             |   wallets                                                   |
	//                         |                                                             |  +----+--------+-----+                                      |
	//                         |                                                             |  | id | amount | ... |                                      |
	//                         |                                                             |  +----+--------+-----+                                      |
	//                         |                                                             |  | 1  | 100000 | ... |                                      |
	//                         |                                                             |  +----+--------+-----+                                      |
	//                         |                                                             |                                                             |
	//                         |   START TRANSACTION                                         |                                                             |
	//                         | ----------------------------------------------------------> |                                                             |
	//                         |                                                             |                                         START TRANSACTION   |
	//   wallets               |                                                             | <---------------------------------------------------------- |
	//  +----+--------+-----+  |   SELECT amount FROM wallets WHERE id = 1                   |                                                             |
	//  | id | amount | ... |  | ----------------------------------------------------------> |                                                             |   wallets
	//  +----+--------+-----+  |                                                             |                   SELECT amount FROM wallets WHERE id = 1   |  +----+--------+-----+
	//  | 1  | 100000 | ... |  |                                                             | <---------------------------------------------------------- |  | id | amount | ... |
	//  +----+--------+-----+  |       does the amount more than 60000? Yes!                 |                                                             |  +----+--------+-----+
	//                         |                                                             |                                                             |  | 1  | 100000 | ... |
	//   wallets               |                                                             |       does the amount more than 60000? Yes!                 |  +----+--------+-----+
	//  +----+--------+-----+  |   UPDATE wallets SET amount = amount - 60000 WHERE id = 1   |                                                             |
	//  | id | amount | ... |  | ----------------------------------------------------------> |                                                             |
	//  +----+--------+-----+  |   COMMIT                                                    |                                                             |
	//  | 1  |  40000 | ... |  | ----------------------------------------------------------> |                                                             |   wallets
	//  +----+--------+-----+  |                                                             |   UPDATE wallets SET amount = amount - 60000 WHERE id = 1   |  +----+--------+-----+
	//                         |                                                             | <---------------------------------------------------------- |  | id | amount | ... |
	//                         |                                                             |                                           COMMIT            |  +----+--------+-----+
	//                         |                                                             | <---------------------------------------------------------- |  | 1  | -20000 | ... |
	//                         |                                                             |                                                             |  +----+--------+-----+
	//                         |                                                             |                                                             |
	//                         |                                                             |                                                             | 因為業務邏輯造成 phantom read (讀到錯誤的錢包餘額)
	//                         |                                                             |                                                             | 導致後續發生 write skew (額度不足導致餘額為負數)
	//
	// 兩種解決 Write Skew 的辦法:
	//
	// 1. 自行加上 Explicit Lock
	//     - 改寫 SELECT amount FROM wallets WHERE id = 1 成 SELECT amount FROM wallets WHERE id = 1 FOR UPDATE
	//     - 加上排他鎖明確限制同時間只允許一個 transaction 進行後續流程
	//     - 注意 Row Lock 升級成 Next-key Lock 可能造成的衍生問題
	//
	// 2. 將 isolation level 升級成 serializable level
	//     - 在上述情境中還是無法避免同時 SELECT 後因為業務邏輯產生的 Phantom Read 問題

	wg := new(sync.WaitGroup)
	wg.Add(2)

	go func(_wg *sync.WaitGroup) {
		defer _wg.Done()

		tx1, err := m.conn.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelRepeatableRead})
		checkError(err, "failed to start transaction:")
		logrus.Infoln("transaction 1 started.")

		var amount int
		err = tx1.QueryRow("SELECT amount FROM wallets WHERE id = 1").Scan(&amount)
		checkError(err, "failed to querying row:")
		logrus.Infoln("transaction 1 selected.")

		time.Sleep(1 * time.Second)

		// 表示業務邏輯處理結果
		if amount > 60000 {
			_, err = tx1.Exec("UPDATE wallets SET amount = amount - 60000 WHERE id = 1")
			checkError(err, "failed to execute:")
			logrus.Infoln("transaction 1 updated.")
		}

		err = tx1.Commit()
		checkError(err, "failed to commit:")
		logrus.Infoln("transaction 1 committed.")

	}(wg)

	time.Sleep(1 * time.Millisecond)

	go func(_wg *sync.WaitGroup) {
		defer _wg.Done()

		tx2, err := m.conn.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelRepeatableRead})
		checkError(err, "failed to start transaction:")
		logrus.Infoln("transaction 2 started.")

		var amount int
		err = tx2.QueryRow("SELECT amount FROM wallets WHERE id = 1").Scan(&amount)
		checkError(err, "failed to querying row:")
		logrus.Infoln("transaction 2 selected.")

		time.Sleep(1 * time.Second)

		// 表示業務邏輯處理結果
		if amount > 60000 {
			_, err = tx2.Exec("UPDATE wallets SET amount = amount - 60000 WHERE id = 1")
			checkError(err, "failed to execute:")
			logrus.Infoln("transaction 2 updated.")
		}

		err = tx2.Commit()
		checkError(err, "failed to commit:")
		logrus.Infoln("transaction 2 committed.")

	}(wg)

	wg.Wait()

	var amount int
	err = m.conn.QueryRow("SELECT amount FROM wallets WHERE id = 1").Scan(&amount)
	checkError(err, "failed to querying row:")

	logrus.Warnf("Amount = %v", amount)
}

func (m *mysql) SimulateLockFailed1(ctx context.Context) {
	// init
	_, err := m.conn.Exec("TRUNCATE TABLE wallets")
	checkError(err, "failed to execute:")

	timeNow := time.Now().Format("2006-01-02 15:04:05")
	_, err = m.conn.Exec("INSERT INTO wallets (user_id, amount, created_at, modified_at) VALUES (?, ?, ?, ?);",
		"1",
		100000,
		timeNow,
		timeNow,
	)
	checkError(err, "failed to execute:")

	logrus.Info("========== start ==========")
	defer logrus.Info("=========== end ===========")

	// 模擬因為觸發覆蓋索引(Covering Index) 導致上鎖失敗
	//
	//  Transaction 1                                    Database                                       Transaction 2
	//       |                                              |                                                |
	//       |                                              |   wallets                                      |
	//       |                                              |  +----+--------+-----+                         |
	//       |                                              |  | id | amount | ... |                         |
	//       |                                              |  +----+--------+-----+                         |
	//       |                                              |  | 1  | 100000 | ... |                         |
	//       |                                              |  +----+--------+-----+                         |
	//       |                                              |                                                |
	//       |   START TRANSACTION                          |                                                |
	//       | -------------------------------------------> |                                                |
	//       |                                              |                            START TRANSACTION   |
	//       |                                              | <--------------------------------------------- |
	//       |   SELECT id FROM wallets WHERE user_id = 1   |                                                |
	//       |   LOCK IN SHARE MODE                         |                                                |
	//       | -------------------------------------------> |                                                |
	//       |                                              |   UPDATE wallets SET amount = 0 WHERE id = 1   |  預想情況中, 此時 Transaction 2 應該要阻塞直到 Transaction 1 結束後
	//       |                                              | <--------------------------------------------- |  才能執行, 但 Transaction 1 卻沒有成功鎖上
	//       |                                              |                                       COMMIT   |
	//       |                                              | <--------------------------------------------- |
	//       |   COMMIT                                     |                                                |
	//       | -------------------------------------------> |                                                |
	//       |                                              |                                                |
	//
	// 原因在於 transaction 1 在執行過程中不需要回到 clustered index 查找資料，因此只需要對 secondary index 上鎖 (即 user_id)
	// 而 transaction 2 請求的鎖是在 clustered index, 因此 transaction 2 可以很順利的執行不必等待 transaction 1 結束
	//
	// 解決辦法:
	//
	// 1. 若 transaction 1 修改查詢欄位, 迫使執行時必須回到 clustered index 查找該欄位, 才會使得 transaction 2 一定得等到 transaction 1 結束後才可繼續動作
	//
	// 2. 將上鎖指令從 LOCK IN SHARE MODE 升級成 FOR UPDATE, 也會同時將 clustered index 上鎖

	wg := new(sync.WaitGroup)
	wg.Add(2)

	go func(_wg *sync.WaitGroup) {
		defer _wg.Done()

		tx1, err := m.conn.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelRepeatableRead})
		checkError(err, "failed to start transaction:")
		logrus.Infoln("transaction 1 started.")

		var id int
		err = tx1.QueryRow("SELECT id FROM wallets WHERE user_id = 1 LOCK IN SHARE MODE").Scan(&id)
		checkError(err, "failed to querying row:")
		logrus.Infoln("transaction 1 selected")

		time.Sleep(1 * time.Second)

		err = tx1.Commit()
		checkError(err, "failed to commit:")
		logrus.Infoln("transaction 1 committed.")

	}(wg)

	time.Sleep(1 * time.Millisecond)

	go func(_wg *sync.WaitGroup) {
		defer _wg.Done()

		tx2, err := m.conn.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelRepeatableRead})
		checkError(err, "failed to start transaction:")
		logrus.Infoln("transaction 2 started.")

		_, err = tx2.Exec("UPDATE wallets SET amount = amount - 10000 WHERE id = 1")
		checkError(err, "failed to execute:")
		logrus.Infoln("transaction 2 updated.")

		err = tx2.Commit()
		checkError(err, "failed to commit:")
		logrus.Infoln("transaction 2 committed.")

	}(wg)

	wg.Wait()
}
