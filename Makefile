.PHONY: dependency unittest test

dependency:
	@go get -v ./...

test:
	@go test -v ./...

unittest: dependency
	@go test -v -short ./...
