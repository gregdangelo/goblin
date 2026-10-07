.Phony: build start listen sender run init run-init

build:
	go build -o ./bin/goblin ./src/cmd/goblin

init:
	./goblin --init

start:
	./goblin

run:
	go run ./src/cmd/goblin/main.go

run-init:
	go run ./src/cmd/goblin/main.go --init

sender:
	go run ./src/cmd/sender/main.go

listen:
	go run ./src/cmd/receiver/main.go