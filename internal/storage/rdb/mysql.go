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

	// bussiness logic
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

	// ???????????????
	statements := `
	TRUNCATE TABLE users;
	TRUNCATE TABLE wallets;
	TRUNCATE TABLE logs;
	`
	if _, err := m.conn.Exec(statements); err != nil {
		logrus.Panicf("failed to execute sql task: %v", err)
	}

	// ????????? users
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

	// ????????? wallets
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

	// ????????????(Dirty Read) ??????
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
	//                                  |                                         |           SELECT count(*) FROM logs   |  ioslation level ??? read uncommitted ????????????
	//                                  |                                         | <------------------------------------ |  transaction 1 ?????? committed ??????????????? dirty read
	//                                  |   ROLLBACK                              |                                       |  ????????? read committed ???????????????????????????
	//                                  | --------------------------------------> |                                       |
	//                                  |                                         |                              COMMIT   |
	//                                  |                                         | <------------------------------------ |
	//                                  |                                         |                                       |

	// ?????? trx1: ???????????? log
	tx1, err := m.conn.Begin()
	checkError(err, "failed to start transaction:")

	_, err = tx1.Exec("INSERT INTO logs (deposit_user_id, withdraw_user_id, amount, created_at) VALUES (1, 2, 1, '2022-12-22 20:57:47');")
	checkError(err, "failed to execute:")

	// ??? trx1 ?????????, ?????? trx2 ???????????? table ?????????????????????
	// ??????????????? transaction isolation level ?????? read-uncommitted ??????
	tx2, err := m.conn.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelReadUncommitted})
	checkError(err, "failed to start transaction:")

	var count int
	err = tx2.QueryRow("SELECT count(*) FROM logs;").Scan(&count)
	checkError(err, "failed to query:")

	logrus.Warnf("Read Uncommitted: %v", count)

	// ?????? trx2
	err = tx2.Commit()
	checkError(err, "failed to commit transaction:")

	// ?????? trx1
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

	// ???????????????(Read Skew) ??????????????????????????????(Non-repeatable Read)
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
	//                         |                                                             |                                             |  ??????????????? transaction ???????????????????????????
	//                         |                                                             |                                             |  ?????? (read skew)
	//                         |                                                             |                                             |  ????????? repeatable read ???????????????????????????
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

	// ??????????????????(Lost Update) ??????
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
	//  +----+--------+-----+  |   COMMIT                                           |                                                    |  transaction2 ???????????????????????? transaction1 ?????????
	//  | 1  |  40000 | ... |  | -------------------------------------------------> |                                                    |  ?????? lost update
	//  +----+--------+-----+  |                                                    |                                                    |
	//                         |                                                    |                                                    |
	//
	// ???????????? Lost Update ?????????:
	//
	// 1. ?????? Database ??? atomic write
	//     - ?????? UPDATE wallets SET amount = {value} WHERE id = 1 ??? UPDATE wallets SET amount = amount - {value} WHERE id = 1
	//     - ??????????????? transaction failed ???????????????????????????????????????????????????????????????????????????
	//
	// 2. ??????????????????????????? (CAS)
	//     - ?????? UPDATE walltes SET amount = {value} WHERE id = 1 ??? UPDATE wallets SET amount = {new} WHERE id = 1 AND amount = {old}
	//     - ?????? transaction 1 ????????????, ?????????????????? transaction ??????????????????????????????

	tx2, err := m.conn.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelRepeatableRead})
	checkError(err, "failed to start transaction:")

	tx1, err := m.conn.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelRepeatableRead})
	checkError(err, "failed to start transaction:")

	var amount_tx1, amount_tx2, amount_result int

	err = tx2.QueryRow("SELECT amount FROM wallets WHERE id = 1").Scan(&amount_tx2)
	checkError(err, "failed to querying row:")

	err = tx1.QueryRow("SELECT amount FROM wallets WHERE id = 1").Scan(&amount_tx1)
	checkError(err, "failed to querying row:")

	// ??????????????????????????????
	amount_tx2 = 60000
	_, err = tx2.Exec("UPDATE wallets SET amount = ? WHERE id = 1", amount_tx2)
	checkError(err, "failed to execute:")

	err = tx2.Commit()
	checkError(err, "failed to commit:")

	// ??????????????????????????????
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

	// ??????????????????(Phantom Read) ???????????????(Write Skew) ??????
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
	//                         |                                            |                                                | ?????? write skew ????????????????????????????????????
	//                         |                                            |                                                | ?????????????????? repeatable read ??????????????????!
	//
	// ???????????????????????????????????? isolation level ??? serializable level ?????? Write Skew ??????

	wg := new(sync.WaitGroup)
	wg.Add(2)

	go func(_wg *sync.WaitGroup) {
		defer _wg.Done()

		tx2, err := m.conn.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelRepeatableRead})
		checkError(err, "failed to start transaction:")
		logrus.Infoln("trnsaction 2 started.")

		var count int
		err = tx2.QueryRow("SELECT COUNT(amount) FROM wallets").Scan(&count)
		checkError(err, "failed to querying row:")
		logrus.Infof("trnsaction 2 selected, count = %v", count)

		time.Sleep(1 * time.Second)

		err = tx2.QueryRow("SELECT COUNT(amount) FROM wallets").Scan(&count)
		checkError(err, "failed to querying row:")
		logrus.Infof("trnsaction 2 selected, count = %v", count)

		_, err = tx2.Exec("UPDATE wallets SET amount = amount + 10000")
		checkError(err, "failed to execute:")
		logrus.Infoln("trnsaction 2 updated")

		// err = tx2.QueryRow("SELECT COUNT(amount) FROM wallets").Scan(&count)
		// checkError(err, "failed to querying row:")
		// logrus.Warnf("trnsaction 2 selected, count = %v", count)

		err = tx2.Commit()
		checkError(err, "failed to commit:")
		logrus.Infoln("trnsaction 2 committed.")

	}(wg)

	time.Sleep(1 * time.Millisecond)

	go func(_wg *sync.WaitGroup) {
		defer _wg.Done()

		tx1, err := m.conn.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelRepeatableRead})
		checkError(err, "failed to start transaction:")
		logrus.Infoln("trnsaction 1 started.")

		timeNow := time.Now().Format("2006-01-02 15:04:05")
		_, err = tx1.Exec("INSERT INTO wallets (user_id, amount, created_at, modified_at) VALUES (?, ?, ?, ?);",
			"2",
			100000,
			timeNow,
			timeNow,
		)
		checkError(err, "failed to execute:")
		logrus.Infoln("trnsaction 1 inserted.")

		err = tx1.Commit()
		checkError(err, "failed to commit:")
		logrus.Infoln("trnsaction 1 committed.")

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

	// ??????????????????(Phantom Read) ???????????????(Write Skew) ??????
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
	//                         |                                                             |                                                             | ???????????????????????? phantom read (???????????????????????????)
	//                         |                                                             |                                                             | ?????????????????? write skew (?????????????????????????????????)
	//
	// ???????????? Write Skew ?????????:
	//
	// 1. ???????????? Explicit Lock
	//     - ?????? SELECT amount FROM wallets WHERE id = 1 ??? SELECT amount FROM wallets WHERE id = 1 FOR UPDATE
	//     - ??????????????????????????????????????????????????? transation ??????????????????
	//     - ?????? Row Lock ????????? Netx-key Lock ???????????????????????????
	//
	// 2. ??? isolation level ????????? serializable level
	//     - ?????????????????????????????????????????? SELECT ?????????????????????????????? Phantom Read ??????

	wg := new(sync.WaitGroup)
	wg.Add(2)

	go func(_wg *sync.WaitGroup) {
		defer _wg.Done()

		tx1, err := m.conn.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelRepeatableRead})
		checkError(err, "failed to start transaction:")
		logrus.Infoln("trnsaction 1 started.")

		var amount int
		err = tx1.QueryRow("SELECT amount FROM wallets WHERE id = 1").Scan(&amount)
		checkError(err, "failed to querying row:")
		logrus.Infoln("trnsaction 1 selected.")

		time.Sleep(1 * time.Second)

		// ??????????????????????????????
		if amount > 60000 {
			_, err = tx1.Exec("UPDATE wallets SET amount = amount - 60000 WHERE id = 1")
			checkError(err, "failed to execute:")
			logrus.Infoln("trnsaction 1 updated.")
		}

		err = tx1.Commit()
		checkError(err, "failed to commit:")
		logrus.Infoln("trnsaction 1 committed.")

	}(wg)

	time.Sleep(1 * time.Millisecond)

	go func(_wg *sync.WaitGroup) {
		defer _wg.Done()

		tx2, err := m.conn.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelRepeatableRead})
		checkError(err, "failed to start transaction:")
		logrus.Infoln("trnsaction 2 started.")

		var amount int
		err = tx2.QueryRow("SELECT amount FROM wallets WHERE id = 1").Scan(&amount)
		checkError(err, "failed to querying row:")
		logrus.Infoln("trnsaction 2 selected.")

		time.Sleep(1 * time.Second)

		// ??????????????????????????????
		if amount > 60000 {
			_, err = tx2.Exec("UPDATE wallets SET amount = amount - 60000 WHERE id = 1")
			checkError(err, "failed to execute:")
			logrus.Infoln("trnsaction 2 updated.")
		}

		err = tx2.Commit()
		checkError(err, "failed to commit:")
		logrus.Infoln("trnsaction 2 committed.")

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

	// ??????????????????????????????(Covering Index) ??????????????????
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
	//       |                                              |   UPDATE wallets SET amount = 0 WHERE id = 1   |  ???????????????, ?????? Transaction 2 ????????????????????? Transaction 1 ?????????
	//       |                                              | <--------------------------------------------- |  ????????????, ??? Transaction 1 ?????????????????????
	//       |                                              |                                       COMMIT   |
	//       |                                              | <--------------------------------------------- |
	//       |   COMMIT                                     |                                                |
	//       | -------------------------------------------> |                                                |
	//       |                                              |                                                |
	//
	// ???????????? transaction 1 ????????????????????????????????? clustered index ????????????????????????????????? secondary index ?????? (??? user_id)
	// ??? transaction 2 ?????????????????? clustered index, ?????? transaction 2 ???????????????????????????????????? transaction 1 ??????
	//
	// ????????????:
	//
	// 1. ??? transaction 1 ??????????????????, ??????????????????????????? clustered index ???????????????, ???????????? transaction 2 ??????????????? transaction 1 ???????????????????????????
	//
	// 2. ?????????????????? LOCK IN SHARE MODE ????????? FOR UPDATE, ??????????????? clustered index ??????

	wg := new(sync.WaitGroup)
	wg.Add(2)

	go func(_wg *sync.WaitGroup) {
		defer _wg.Done()

		tx1, err := m.conn.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelRepeatableRead})
		checkError(err, "failed to start transaction:")
		logrus.Infoln("trnsaction 1 started.")

		var id int
		err = tx1.QueryRow("SELECT id FROM wallets WHERE user_id = 1 LOCK IN SHARE MODE").Scan(&id)
		checkError(err, "failed to querying row:")
		logrus.Infoln("trnsaction 1 selected")

		time.Sleep(1 * time.Second)

		err = tx1.Commit()
		checkError(err, "failed to commit:")
		logrus.Infoln("trnsaction 1 committed.")

	}(wg)

	time.Sleep(1 * time.Millisecond)

	go func(_wg *sync.WaitGroup) {
		defer _wg.Done()

		tx2, err := m.conn.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelRepeatableRead})
		checkError(err, "failed to start transaction:")
		logrus.Infoln("trnsaction 2 started.")

		_, err = tx2.Exec("UPDATE wallets SET amount = amount - 10000 WHERE id = 1")
		checkError(err, "failed to execute:")
		logrus.Infoln("trnsaction 2 updated.")

		err = tx2.Commit()
		checkError(err, "failed to commit:")
		logrus.Infoln("trnsaction 2 committed.")

	}(wg)

	wg.Wait()
}
