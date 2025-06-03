test:
	@go test -v ./... -count=1 -cover

install:
	@go mod tidy