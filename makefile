GIT_NUM ?= ${shell git rev-parse --short=6 HEAD}
BUILD_TIME ?= ${shell date +'%Y-%m-%d_%T'}

MYSQL_USER ?= root
MYSQL_PASSWORD ?= 0
MYSQL_HOST ?= localhost
MYSQL_DATABASE ?= development
MYSQL_PORT ?= 3306
MYSQL_DSN ?= $(MYSQL_USER):$(MYSQL_PASSWORD)@tcp($(MYSQL_HOST):$(MYSQL_PORT))/$(MYSQL_DATABASE)

.PHONY: help init setup-all shutdown-all lint migrate-up migrate-down show-tables gen-data concurrent-transfer dirty-read read-skew lost-update

help:
	@echo "Usage make [commands]\n"
	@echo "Commands:"
	@echo "  init  初始化建置環境 (docker volume, build image, etc.)"
	@echo "  setup-all            透過 docker-compose 啟動所有服務 (主要系統, 壓力測試工具, 各項監控工具)"
	@echo "  shutdown-all         關閉 docker-cpmpose 所有服務"
	@echo "  lint                 "
	@echo "  migrate-up           透過 golang-migrate 執行所有 up migrations"
	@echo "  migrate-down         透過 golang-migrate 執行所有 down migrations"
	@echo "  show-tables          "
	@echo "  gen-data             "
	@echo "  concurrent-transfer  "
	@echo "  dirty-read           "
	@echo "  non-repeatable-read  "
	@echo "  read-skew            "

init:
	rm -rf deployments/data
	mkdir -p deployments/data/mysql

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

migrate-down:
	echo y | migrate -database 'mysql://$(MYSQL_DSN)?multiStatements=true' -source 'file://deployments/mysql/migration' -verbose down

show-tables:
	go run main.go show_tables -f ./conf.d/env.yaml

gen-data:
	go run main.go generate_data -f ./conf.d/env.yaml

concurrent-transfer:
	go run main.go concurrent_transfer -f ./conf.d/env.yaml

dirty-read:
	go run main.go dirty_read -f ./conf.d/env.yaml

non-repeatable-read:
	go run main.go non_repeatable_read -f ./conf.d/env.yaml

read-skew:
	go run main.go read_skew -f ./conf.d/env.yaml

lost-update:
	go run main.go lost_update -f ./conf.d/env.yaml