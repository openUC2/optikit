.DEFAULT_GOAL := dev

.PHONY: dev
dev: ## dev build
dev: clean install generate fmt fix spell vet lint test mod-tidy

.PHONY: ci
ci: ## CI build
ci: dev diff

.PHONY: clean
clean: ## remove files created during build pipeline
	$(call print-target)
	find internal/clients/build123d/data/*-* ! -name '.gitattributes' -type f -exec rm -f {} +
	find internal/clients/build123d/data/*-*/* ! -name . -prune -type d -exec rm -R {} +
	rm -rf dist
	rm -f coverage.*

.PHONY: install
install: ## go install tool
	$(call print-target)
	go install tool

.PHONY: install-pip
install-pip: ## embedpip
	$(call print-target)
	cd internal/clients/build123d; go run ./embedpip

.PHONY: install-pip-linux-amd64
install-pip-linux-amd64: ## embedpip linux-amd64
	$(call print-target)
	cd internal/clients/build123d; go run ./embedpip linux-amd64

.PHONY: install-pip-darwin-amd64
install-pip-darwin-amd64: ## embedpip linux-amd64
	$(call print-target)
	cd internal/clients/build123d; go run ./embedpip darwin-amd64

.PHONY: install-pip-darwin-arm64
install-pip-darwin-arm64: ## embedpip darwin-arm64
	$(call print-target)
	cd internal/clients/build123d; go run ./embedpip darwin-arm64

.PHONY: generate
generate: ## go generate
	$(call print-target)
	go generate ./...

.PHONY: vet
vet: ## go vet
	$(call print-target)
	go vet ./...

.PHONY: fix
fix: ## go fix
	$(call print-target)
	go fix ./...

.PHONY: fmt
fmt: ## go fmt
	$(call print-target)
	go fmt ./...

.PHONY: spell
spell: ## misspell
	$(call print-target)
	go tool misspell -error -locale=US -w **.md

.PHONY: lint
lint: ## golangci-lint
	$(call print-target)
	go tool golangci-lint run

.PHONY: test
test: ## go test with race detector and code coverage
	$(call print-target)
	go test -race -covermode=atomic -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

.PHONY: mod-tidy
mod-tidy: ## go mod tidy
	$(call print-target)
	go mod tidy

.PHONY: diff
diff: ## git diff
	$(call print-target)
	git diff --exit-code
	RES=$$(git status --porcelain) ; if [ -n "$$RES" ]; then echo $$RES && exit 1 ; fi

.PHONY: build
build: ## goreleaser --snapshot --skip=publish --clean
build: install
	$(call print-target)
	go tool goreleaser --snapshot --skip=publish --clean

.PHONY: release
release: ## goreleaser --clean
release: install
	$(call print-target)
	go tool goreleaser --clean

.PHONY: run
run: ## go run
	@go run -race .

.PHONY: go-clean
go-clean: ## go clean build, test and modules caches
	$(call print-target)
	go clean -r -i -cache -testcache -modcache

.PHONY: help
help:
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

define print-target
    @printf "Executing target: \033[36m$@\033[0m\n"
endef
