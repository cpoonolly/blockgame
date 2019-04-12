all: ./static/bundle.wasm ./server/server

rerun: clean run

# Run target
run: ./static/bundle.wasm ./server/server
	./server/server

# Build targets
./server/server:
	go build -o ./server/server ./server/server.go

./static/bundle.wasm: ./wasm/bundle.wasm
	cp ./wasm/bundle.wasm ./static/bundle.wasm

./wasm/bundle.wasm: ./core/core
	GOOS=js GOARCH=wasm go build -o ./wasm/bundle.wasm ./wasm/

./core/core:
	go build -o ./core/core ./core/core.go

# Clean targets
clean: clean-server clean-static clean-wasm clean-core

clean-server:
	find ./server -type f ! -name *.go -delete

clean-static:
	rm -f ./static/*.wasm

clean-wasm:
	rm -f ./wasm/*.wasm

clean-core:
	find ./core -type f ! -name *.go -delete