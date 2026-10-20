.DEFAULT_GOAL := dev

.PHONY: dev
dev: ## dev build
dev: clean install generate-prim-glb generate fmt fix spell vet lint test mod-tidy

.PHONY: ci
ci: ## CI build
ci: dev diff

.PHONY: clean
clean: ## remove files created during build pipeline
	$(call print-target)
	rm -rf dist
	rm -f coverage.*

.PHONY: install
install: ## go install tool
	$(call print-target)
	go install tool
	npm --prefix tools/gltf-checker install
	uv sync

.PHONY: generate-prim-glb
generate-prim-glb: ## go generate
	$(call print-target)
	./tools/examples/process-all.sh ./examples list-prim-step uv run poe convert-batch glb

.PHONY: generate
generate: ## generate outside CI
generate: generate-prim-glb generate-go

.PHONY: generate-go
generate-go: ## go generate
	$(call print-target)
	go generate ./...

.PHONY: generate-unreproducible
generate-unreproducible: ## generate outside CI
generate-unreproducible: generate-obj-step

.PHONY: generate-obj-step
generate-obj-step: ## generate STEP files outputs with nondeterministically-ordered contents
	$(call print-target)
	./tools/examples/process-all.sh ./examples list-assm-report uv run poe assemble-batch _assembly _objects

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
test: test-go test-gltf

.PHONY: test-go
test-go: ## go test with race detector and code coverage
	$(call print-target)
	go test -race -covermode=atomic -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

.PHONY: test-gltf
test-gltf: ## gltf file validator
	$(call print-target)
	tools/gltf-checker/check-all.sh

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
