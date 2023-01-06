# Data Pipeline Practice

- [Data Pipeline Practice](#data-pipeline-practice)
  - [Goal](#goal)
  - [Getting Start](#getting-start)
    - [Prerequisites](#prerequisites)
    - [Quick install instructions](#quick-install-instructions)
    - [Usage](#usage)
  - [Project Layout](#project-layout)
  - [References](#references)
    - [Golang-Migrate](#golang-migrate)
    - [RDB Isolation Level](#rdb-isolation-level)
    - [PostgreSQL](#postgresql)

---

## Goal

- 熟悉 Golang 與 RDB 、 NoSQL 操作, CDC 、 ETL 流程
- POC 練習

--- 

## Getting Start

### Prerequisites

- Go
- Docker

### Quick install instructions

```shell
make init
```

### Usage

建置並啟動相關基礎服務，例如：

- MySQL
- PostgreSQL
- pgAdmin

```shell
make setup-all
```

執行資料庫 migration 命令

```shell
make migration-up
```

更多指令請查閱：

```
shell help
```

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
 - [[Blog] MySQL-两类更新丢失及解决办法](https://blog.csdn.net/weixin_44793021/article/details/125107154)
 - [[Blog] MySQL🐬 InnoDB 教我的事：想鎖的沒鎖 ？不該鎖的被鎖了！](https://medium.com/%E7%A8%8B%E5%BC%8F%E7%8C%BF%E5%90%83%E9%A6%99%E8%95%89/mysql-innodb-%E6%95%99%E6%88%91%E7%9A%84%E4%BA%8B-%E6%83%B3%E9%8E%96%E7%9A%84%E6%B2%92%E9%8E%96-%E4%B8%8D%E8%A9%B2%E9%8E%96%E7%9A%84%E8%A2%AB%E9%8E%96%E4%BA%86-ac723fe167fe)
 - [[Blog] 在数据库中不可重复读和幻读到底应该怎么分？](https://www.zhihu.com/question/392569386)

### PostgreSQL

 - [[Blog] Docker-compose創建PostgreSQL](https://cde566.medium.com/docker-compose%E5%89%B5%E5%BB%BApostgresql-7f3f9519fa20)