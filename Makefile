build:
	@cp "$(shell go env GOROOT)/misc/wasm/wasm_exec.js" ./public/assets/wasm_exec.js
	@GOOS=js GOARCH=wasm go build -o ./public/assets/optim.wasm
.PHONY: build

dev: build
	@go run ./server/...
.PHONY: dev
