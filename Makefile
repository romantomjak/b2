build:
	go build -o b2 cmd/b2/main.go

test:
	go test -cover ./

.PHONY: build
