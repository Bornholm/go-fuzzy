webapp:
	rm -rf dist/webapp
	mkdir -p dist/webapp
	cp -f "$(shell go env GOROOT)/lib/wasm/wasm_exec.js" ./dist/webapp
	cp -f misc/webapp/index.html ./dist/webapp
	GOOS=js GOARCH=wasm go build -o dist/webapp/main.wasm ./misc/webapp

test:
	go test -v -race ./...

watch: tools/modd/bin/modd
	tools/modd/bin/modd

tools/modd/bin/modd:
	mkdir -p tools/modd/bin
	GOBIN=$(PWD)/tools/modd/bin go install github.com/cortesi/modd/cmd/modd@latest
