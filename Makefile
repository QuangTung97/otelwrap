.PHONY: lint test install-tools install

lint:
	go fmt ./...
	go vet ./...
	revive -config revive.toml -formatter friendly ./...

test:
	go test -covermode=count -coverprofile=coverage.out ./...

test-race:
	go test -race -count=1 ./...

install-tools:
	go install github.com/mgechev/revive

install:
	go install