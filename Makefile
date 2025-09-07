GO ?= go
GOBIN ?= $(shell $(GO) env GOBIN)
ifeq ($(GOBIN),)
GOBIN = $(shell $(GO) env GOPATH)/bin
endif

MOCKGEN = $(GOBIN)/mockgen
GOLANGCI_LINT = $(GOBIN)/golangci-lint
STATICCHECK = $(GOBIN)/staticcheck

.PHONY: build

build:
	$(GO) build -v ./...

lint:
	@if ! command -v $(GOLANGCI_LINT) >/dev/null 2>&1; then \
		curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(GOBIN) latest; \
	fi
	@if ! command -v $(STATICCHECK) >/dev/null 2>&1; then \
    	$(GO) install honnef.co/go/tools/cmd/staticcheck@latest; \
    fi
	$(GOLANGCI_LINT) run ./...
	$(STATICCHECK) ./...

tests:
	$(GO) test -race -v ./...

coverage:
	$(GO) test -v -coverprofile=coverage.out ./...
	$(GO) tool cover -html=coverage.out -o coverage.html

mocks:
	@if ! command -v $(MOCKGEN) >/dev/null 2>&1; then \
		$(GO) install github.com/golang/mock/mockgen@latest; \
	fi
	$(MOCKGEN) -source=internal/service/service.go -destination=internal/service/mock/service_mock.go -package=mock
	$(MOCKGEN) -source=internal/infra/uniswap/client.go -destination=internal/infra/uniswap/mock/client_mock.go -package=mock

clean:
	@rm -f coverage.out coverage.html
