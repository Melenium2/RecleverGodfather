.PHONY: build
build:
	go build -o bin/main

.PHONY: run
run:
	./bin/main

.DEFAULT_GOAL := build