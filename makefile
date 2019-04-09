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

./wasm/bundle.wasm:
	GOOS=js GOARCH=wasm go build -o ./wasm/bundle.wasm ./wasm/

# Clean targets
clean: clean-server clean-static clean-wasm

clean-server:
	find ./server -type f ! -name *.go -delete

clean-static:
	rm -f ./static/*.wasm

clean-wasm:
	rm -f ./wasm/*.wasm