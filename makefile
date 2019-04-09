all: ./static/main.wasm ./server/server

# Run target
run: ./static/main.wasm ./server/server
	./server/server

# Build targets
./server/server:
	go build -o ./server/server ./server/server.go

./static/main.wasm: ./wasm/main.wasm
	cp ./wasm/main.wasm ./static/main.wasm

./wasm/main.wasm:
	GOOS=js GOARCH=wasm go build -o ./wasm/main.wasm ./wasm/wasm.go

# Clean targets
clean: clean-server clean-static clean-wasm

clean-server:
	find ./server -type f ! -name *.go -delete

clean-static:
	rm -f ./static/*.wasm

clean-wasm:
	rm -f ./wasm/*.wasm