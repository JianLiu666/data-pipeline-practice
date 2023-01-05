GIT_NUM ?= ${shell git rev-parse --short=6 HEAD}
BUILD_TIME ?= ${shell date +'%Y-%m-%d_%T'}

MYSQL_USER ?= root
MYSQL_PASSWORD ?= 0
MYSQL_HOST ?= localhost
MYSQL_PORT ?= 3306
MYSQL_DATABASE ?= development
MYSQL_DSN ?= $(MYSQL_USER):$(MYSQL_PASSWORD)@tcp($(MYSQL_HOST):$(MYSQL_PORT))/$(MYSQL_DATABASE)

POSTGRES_USER ?= user
POSTGRES_PASSWORD ?= password
POSTGRES_HOST ?= localhost
POSTGRES_PORT ?= 5432
POSTGRES_DATABASE ?= development
POSTGRES_DSN ?= $(POSTGRES_USER):$(POSTGRES_PASSWORD)@$(POSTGRES_HOST):$(POSTGRES_PORT)/$(POSTGRES_DATABASE)

.PHONY: help init setup-all shutdown-all lint migrate-up migrate-down show-tables gen-data dirty-read read-skew lost-update write-skew-1 write-skew-2 lock-failed-1

help:
	@echo "Usage make [commands]\n"
	@echo "Commands:"
	@echo "  init  初始化建置環境 (docker volume, build image, etc.)"
	@echo "  setup-all      透過 docker-compose 啟動所有服務 (主要系統, 壓力測試工具, 各項監控工具)"
	@echo "  shutdown-all   關閉 docker-cpmpose 所有服務"
	@echo "  lint           執行 Go Linter (golangci-lint)"
	@echo "  migrate-up     透過 golang-migrate 執行所有 up migrations"
	@echo "  migrate-down   透過 golang-migrate 執行所有 down migrations"
	@echo "  show-tables    "
	@echo "  gen-data       "
	@echo "  dirty-read     模擬 Transaction 中的 Dirty Read 情境與解決辦法"
	@echo "  read-skew      模擬 Transaction 中的 Read Skew 情境與解決辦法"
	@echo "  lost-update    模擬 Transaction 中的 Lost Update 情境與解決辦法"
	@echo "  write-skew-1   模擬 Transaction 中的第一種 Write Skew 情境與解決辦法"
	@echo "  write-skew-2   模擬 Transaction 中的第二種 Write Skew 情境與解決辦法"
	@echo "  lock-failed-1  模擬 Transaction 中因為命中不同索引導致上鎖失敗的情境與解決辦法"

init:
	rm -rf deployments/data
	mkdir -p deployments/data/mysql
	mkdir -p deployments/data/postgres
	mkdir -p deployments/data/pgadmin

	go mod download
	go mod tidy

setup-all:
	docker-compose -f deployments/00.infra.yaml down -v

	docker-compose -f deployments/00.infra.yaml up -d

	docker ps -a

shutdown-all:
	docker-compose -f deployments/00.infra.yaml down -v

lint:
	golangci-lint run

migrate-up:
	migrate -database 'mysql://$(MYSQL_DSN)?multiStatements=true' -source 'file://deployments/mysql/migration' -verbose up
	migrate -database 'postgres://$(POSTGRES_DSN)?sslmode=disable' -source 'file://deployments/postgres/migration' -verbose up

migrate-down:
	echo y | migrate -database 'mysql://$(MYSQL_DSN)?multiStatements=true' -source 'file://deployments/mysql/migration' -verbose down
	echo y | migrate -database 'postgres://$(POSTGRES_DSN)?sslmode=disable' -source 'file://deployments/postgres/migration' -verbose down

show-tables:
	go run main.go show_tables -f ./conf.d/env.yaml

gen-data:
	go run main.go generate_data -f ./conf.d/env.yaml

dirty-read:
	go run main.go dirty_read -f ./conf.d/env.yaml

read-skew:
	go run main.go read_skew -f ./conf.d/env.yaml

lost-update:
	go run main.go lost_update -f ./conf.d/env.yaml

write-skew-1:
	go run main.go write_skew_1 -f ./conf.d/env.yaml

write-skew-2:
	go run main.go write_skew_2 -f ./conf.d/env.yaml

lock-failed-1:
	go run main.go lock_failed_1 -f ./conf.d/env.yaml