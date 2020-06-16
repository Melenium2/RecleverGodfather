-include .env
export tg_chat_id := $(TG_CHAT_ID)
export tg_token := $(TG_TOKEN)
export http_port := $(HTTP_PORT)
export grpc_port := $(GRPC_PORT)
export logger_db := $(LOGGER_DB)
export consul_addr := $(CONSUL_ADDR)

GOBIN=./cmd/main

## build: Build go binary
build:
	go build -o $(GOBIN)

## run: Run go server
.PHONY: run
run:
	$(GOBIN)

## get: Run go get missing dependencies
get:
	go get ./...

## deploy: Run commands to deploy app to container
deploy:
	make get
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
          -ldflags='-w -s -extldflags "-static"' -a \
          -o /go/bin/main .

.PHONY: help
all: help
help: Makefile
	@echo
	@echo " Choose a command"
	@echo
	@sed -n 's/^##//p' $< | column -t -s ':' |  sed -e 's/^/ /'
	@echo

.DEFAULT_GOAL := help