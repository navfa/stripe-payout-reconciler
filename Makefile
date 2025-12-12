.PHONY: build test lint vet fmt tidy verify clean

BINARY := stripe-payout-reconciler
CMD    := ./cmd/$(BINARY)

build:
	go build -trimpath -o $(BINARY) $(CMD)

test:
	go test -race ./...

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

clean:
	rm -f $(BINARY)
	rm -rf dist/
