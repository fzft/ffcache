build:
	go build -o bin/ffcache

build-linux:
	GOOS=linux GOARCH=arm64 go build -o bin/ffcache-64

run: build
	./bin/ffcache

test:
	go test -v ./...