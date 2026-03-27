BINARY := spinctl
GOFLAGS := -trimpath
COVERAGE_THRESHOLD := 90

.PHONY: build test coverage lint clean

build:
	go build $(GOFLAGS) -o bin/$(BINARY) ./cmd/spinctl

test:
	go test ./... -v -race

coverage:
	go test ./... -coverprofile=coverage.out -covermode=atomic
	go tool cover -func=coverage.out
	go tool cover -html=coverage.out -o coverage.html
	@COVERAGE=$$(go tool cover -func=coverage.out | grep total | awk '{print substr($$3, 1, length($$3)-1)}'); \
	if [ $$(echo "$$COVERAGE < $(COVERAGE_THRESHOLD)" | bc -l) -eq 1 ]; then \
		echo "Coverage $$COVERAGE% is below threshold $(COVERAGE_THRESHOLD)%"; \
		exit 1; \
	fi; \
	echo "Coverage: $$COVERAGE% (threshold: $(COVERAGE_THRESHOLD)%)"

lint:
	golangci-lint run ./...

clean:
	rm -rf bin/ coverage.out coverage.html
