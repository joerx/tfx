default: build

build:
	go build -o out/tfx

test: 
	go test ./...
