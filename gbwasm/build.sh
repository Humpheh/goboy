GOPATH="$1"
GOOS=js GOARCH=wasm $GOPATH/bin/go build -o net/goboy.wasm main.go