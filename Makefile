BINARY=music

all: fmt clean build

test:
	go test ./...

fmt:
	go fmt ./...

build:
	go build -o $(BINARY) main.go

clean:
	rm -rf $(BINARY)

