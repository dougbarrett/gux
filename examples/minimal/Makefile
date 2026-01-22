.PHONY: server wasm all clean

all: server wasm

server:
	go build -o bin/server ./cmd/server

wasm:
	GOOS=js GOARCH=wasm go build -o bin/app.wasm ./cmd/wasm

clean:
	rm -rf bin/

run: server
	./bin/server
