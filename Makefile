.PHONY: help check-go-version build run install-local relink-local test test-race coverage ci lint fmt tidy clean
.DEFAULT_GOAL := help

# ── Change this to your extension name (without the gh- prefix) ──
EXTENSION_NAME ?= contrib

BINARY ?= bin/gh-$(EXTENSION_NAME)
GO ?= go
GO_MIN_MAJOR ?= 1
GO_MIN_MINOR ?= 21
GO_MIN_PATCH ?= 0

check-go-version:
	@current="$$( $(GO) env GOVERSION 2>/dev/null | sed 's/^go//' )"; \
	current_major="$${current%%.*}"; \
	current_minor="$${current#*.}"; \
	current_minor="$${current_minor%%.*}"; \
	current_patch="$${current#*.*.}"; \
	if [ "$$current_patch" = "$$current" ]; then current_patch=0; fi; \
	current_patch="$${current_patch%%.*}"; \
	if [ -z "$$current_major" ] || [ -z "$$current_minor" ] || [ -z "$$current_patch" ]; then \
		echo "Unable to detect Go version via '$(GO) env GOVERSION'."; \
		exit 1; \
	fi; \
	if [ "$$current_major" -lt "$(GO_MIN_MAJOR)" ] || { [ "$$current_major" -eq "$(GO_MIN_MAJOR)" ] && { [ "$$current_minor" -lt "$(GO_MIN_MINOR)" ] || { [ "$$current_minor" -eq "$(GO_MIN_MINOR)" ] && [ "$$current_patch" -lt "$(GO_MIN_PATCH)" ]; }; }; }; then \
		echo "Go $$current detected. gh-$(EXTENSION_NAME) requires Go $(GO_MIN_MAJOR).$(GO_MIN_MINOR).$(GO_MIN_PATCH)+."; \
		echo "Upgrade Go and retry."; \
		exit 1; \
	fi

help:
	@echo "gh-$(EXTENSION_NAME) developer commands"
	@echo ""
	@echo "  make build          Build ./$(BINARY)"
	@echo "  make run            Build and run locally"
	@echo "  make install-local  Install extension from current checkout"
	@echo "  make relink-local   Reinstall local extension link"
	@echo "  make test           Run unit tests"
	@echo "  make test-race      Run tests with race + coverage.out"
	@echo "  make coverage       Print coverage summary (requires coverage.out)"
	@echo "  make ci             Run build + vet + test-race"
	@echo "  make lint           Run golangci-lint if installed"
	@echo "  make fmt            Format all Go packages"
	@echo "  make tidy           Run go mod tidy"
	@echo "  make clean          Remove build artifacts"

build: check-go-version
	@mkdir -p $(dir $(BINARY))
	$(GO) build -o $(BINARY) .

run: build
	./$(BINARY)

install-local:
	gh extension install .

relink-local: build
	@if gh extension list | grep -qE '^gh $(EXTENSION_NAME)[[:space:]]'; then \
		gh extension remove $(EXTENSION_NAME); \
	fi
	$(GO) build -o gh-$(EXTENSION_NAME) .
	gh extension install .

test: check-go-version
	$(GO) test ./...

test-race: check-go-version
	$(GO) test -v -race -coverprofile=coverage.out ./...

coverage:
	$(GO) tool cover -func=coverage.out

ci: check-go-version
	$(GO) build ./... && $(GO) vet ./... && $(GO) test -v -race -coverprofile=coverage.out ./...

lint:
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run; \
	else \
		echo "golangci-lint not installed; skipping"; \
	fi

fmt: check-go-version
	$(GO) fmt ./...

tidy: check-go-version
	$(GO) mod tidy

clean:
	rm -rf bin coverage.out gh-$(EXTENSION_NAME)
