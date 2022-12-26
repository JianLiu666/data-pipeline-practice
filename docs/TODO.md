# TODO

- [TODO](#todo)
  - [Application](#application)
  - [Mechanisms](#mechanisms)
  - [Deployment](#deployment)
  - [Relational Database Isolation Level](#relational-database-isolation-level)

---

## Application

- [ ] Database interface
  - [ ] MySQL implementation
  - [ ] PostgreSQL implementation
- [ ] Benchmark
  - [ ] Read committed 
  - [ ] Snapshot isolation

## Mechanisms

- [ ] Research capture data change (CDC) solutions
  - [ ] Apache Flink CDC

## Deployment

- [x] Docker-compose configuration file
- [x] Study relational database migration tool `flyway`
  - [x] Used `golang-migrate` cli tool
- [ ] MySQL master & slave

## Relational Database Isolation Level

- [ ] 理解 MySQL gap lock & next-key lock
- [ ] 理解 不可重複讀(Non-repeatable Read) 與 讀偏差(Read Skew) 的區別