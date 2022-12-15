GIT_NUM ?= ${shell git rev-parse --short=6 HEAD}
BUILD_TIME ?= ${shell date +'%Y-%m-%d_%T'}

MYSQL_USER ?= root
MYSQL_PASSWORD ?= 0
MYSQL_HOST ?= localhost
MYSQL_DATABASE ?= development
MYSQL_PORT ?= 3306
MYSQL_DSN ?= $(MYSQL_USER):$(MYSQL_PASSWORD)@tcp($(MYSQL_HOST):$(MYSQL_PORT))/$(MYSQL_DATABASE)

.PHONY: help init demo shutdown-all migrate-up migrate-down

help:
	@echo "Usage make [commands]\n"
	@echo "Commands:"
	@echo "  init  初始化建置環境 (docker volume, build image, etc.)"
	@echo "  demo               透過 docker-compose 啟動所有服務 (主要系統, 壓力測試工具, 各項監控工具)"
	@echo "  shutdown-all       關閉 docker-cpmpose 所有服務"
	@echo "  migrate-up-all     透過 golang-migrate 執行所有 up migrations"
	@echo "  migrate-down-all   透過 golang-migrate 執行所有 down migrations"

init:
	rm -rf deployments/data
	mkdir -p deployments/data/mysql

	go mod download
	go mod tidy

demo:
	docker-compose -f deployments/00.infra.yaml down -v

	docker-compose -f deployments/00.infra.yaml up -d

	docker ps -a

shutdown-all:
	docker-compose -f deployments/00.infra.yaml down -v

migrate-up:
	migrate -database 'mysql://$(MYSQL_DSN)?multiStatements=true' -source 'file://deployments/mysql/migration' -verbose up

migrate-down:
	echo y | migrate -database 'mysql://$(MYSQL_DSN)?multiStatements=true' -source 'file://deployments/mysql/migration' -verbose down