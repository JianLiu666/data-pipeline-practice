# Data Pipeline Practice

- [Data Pipeline Practice](#data-pipeline-practice)
  - [Goal](#goal)
  - [Project Layout](#project-layout)
  - [References](#references)
    - [Golang-Migrate](#golang-migrate)
    - [RDB Isolation Level](#rdb-isolation-level)

---

## Goal

- 熟悉 Golang 與 RDB 、 NoSQL 操作, CDC 、 ETL 流程
- POC 練習

--- 

## Project Layout

- 參考 [Standard Go Project Layout](https://github.com/golang-standards/project-layout)

```
DATA-PIPELINE-PRACTICE
 ├─ cmd/          # 本專案的主要應用程式
 ├─ conf.d/       # 組態設定的檔案範本及預設設定
 ├─ deployments/  # 系統和容器編配部署的組態設定腳本
 │   ├─ data/        # 保存 docker volume
 │   └─ mysql/       # MySQL 組態設定與動態連結函式庫 (dll)
 ├─ docs/         # 設計和使用者文件 (sequence, db schema, etc.)
 ├─ internal/     # 私有應用程式和函示庫的程式碼
 │   ├─ accessor/    # 基礎建設模組
 │   ├─ config/      # 組態設定模組 (viper)
 │   └─ storage/     # 資料庫模組
 ├─ .gitignore    
 ├─ go.mod        
 ├─ go.sum        
 ├─ LICENSE       
 ├─ main.go       # 主程式進入點
 ├─ makefile      
 └─ README.md     
```

---

## References

### Golang-Migrate

- [[Github] migrate CLI](https://github.com/golang-migrate/migrate/tree/master/cmd/migrate)
- [[Blog] Go: database migrations made easy - an example using MySQL](https://www.linkedin.com/pulse/go-database-migrations-made-easy-example-using-mysql-tiago-melo/)

### RDB Isolation Level

 - [[Blog] MySQL中如何为单个事务设置隔离级别](https://www.jianshu.com/p/b324b368a210)
 - [[Blog] MySQL 交易功能 Transaction 整理](https://xyz.cinc.biz/2013/05/mysql-transaction.html)
 - [[StackOverflow] "Read skew" vs "Non-repeatable read" (Transaction)](https://stackoverflow.com/questions/73917534/read-skew-vs-non-repeatable-read-transaction)
 - [[Blog] 複習資料庫的 Isolation Level 與圖解五個常見的 Race Conditions](https://medium.com/@chester.yw.chu/%E8%A4%87%E7%BF%92%E8%B3%87%E6%96%99%E5%BA%AB%E7%9A%84-isolation-level-%E8%88%87%E5%B8%B8%E8%A6%8B%E7%9A%84%E4%BA%94%E5%80%8B-race-conditions-%E5%9C%96%E8%A7%A3-16e8d472a25c)
 - [[Blog] 對於 MySQL Repeatable Read Isolation 常見的三個誤解](https://medium.com/@chester.yw.chu/%E5%B0%8D%E6%96%BC-mysql-repeatable-read-isolation-%E5%B8%B8%E8%A6%8B%E7%9A%84%E4%B8%89%E5%80%8B%E8%AA%A4%E8%A7%A3-7a9afbac65af)