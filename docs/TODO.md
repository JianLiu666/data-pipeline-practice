# TODO

- [TODO](#todo)
  - [Application](#application)
  - [Mechanisms](#mechanisms)
  - [Deployment](#deployment)
  - [Relational Database Isolation Level](#relational-database-isolation-level)

---

## Application

- [x] Database interface
  - [x] MySQL implementation
  - [ ] PostgreSQL implementation
  - [ ] MongoDB implementation
- [ ] Benchmark
  - [ ] Read committed 
  - [ ] Snapshot isolation
- [ ] CDC flow
  - [ ] From MySQL to MongoDB
  - [ ] 設計情境

## Mechanisms

- [ ] Research capture data change (CDC) solutions
  - [ ] Apache Flink CDC

## Deployment

- [x] Docker-compose configuration file
- [x] Study relational database migration tool `flyway`
  - [x] Used `golang-migrate` cli tool
- [ ] MySQL master & slave

## Relational Database Isolation Level

- [x] 理解 MySQL gap lock & next-key lock
- [x] 理解 不可重複讀(Non-repeatable Read) 與 讀偏差(Read Skew) 的區別
- [ ] MySQL 版本升級差異 v5.7 -> v8.0
- [ ] Case: 2 trxs, trx1 走 primary key 更新資料, trx2 走 covering index 取回資料, 確認 trx2 是否會被 block