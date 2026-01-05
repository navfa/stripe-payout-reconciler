.PHONY: build test lint vet fmt tidy verify clean run install check coverage seed

BINARY := stripe-payout-reconciler
CMD    := ./cmd/$(BINARY)

VERSION ?= dev
ARGS    ?=

build:
	go build -trimpath -ldflags "-X main.version=$(VERSION)" -o $(BINARY) $(CMD)

run: build
	./$(BINARY) $(ARGS)

install:
	go install -trimpath -ldflags "-X main.version=$(VERSION)" $(CMD)

test:
	go test -race -count=1 ./...

coverage:
	go test -race -count=1 -coverprofile=coverage.out ./...
	go tool cover -func=coverage.out
	@echo ""
	@echo "To view in browser: go tool cover -html=coverage.out"

lint:
	golangci-lint run ./...

vet:
	go vet ./...

fmt:
	gofmt -w .
	goimports -w .

tidy:
	go mod tidy

verify:
	go mod verify

check: fmt tidy verify vet lint test

seed:
	./scripts/seed-test-data.sh

clean:
	rm -f $(BINARY)
	rm -f coverage.out
	rm -rf dist/
