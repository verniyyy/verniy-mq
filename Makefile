.PHONY: build
build:
	go build -o bin/vmq main.go

.PHONY: run
run:
	go run main.go