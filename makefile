GIT_NUM ?= ${shell git rev-parse --short=6 HEAD}
BUILD_TIME ?= ${shell date +'%Y-%m-%d_%T'}

.PHONY: help init demo shutdown-all

help:
	@echo "Usage make [commands]\n"
	@echo "Commands:"
	@echo "  init  初始化建置環境 (docker volume, build image, etc.)"
	@echo "  demo               透過 docker-compose 啟動所有服務 (主要系統, 壓力測試工具, 各項監控工具)"
	@echo "  shutdown-all       關閉 docker-cpmpose 所有服務"

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