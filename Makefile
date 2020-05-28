.PHONY: all run build

all: build run

build:
	@go build ./

run:
	@./mqlog
