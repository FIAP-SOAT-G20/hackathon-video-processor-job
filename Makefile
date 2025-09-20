.DEFAULT_GOAL := help

# Variables
APP_NAME=video-processor-job
MAIN_FILE=main.go
DOCKER_REGISTRY=ghcr.io
DOCKER_REGISTRY_APP=fiap-soat-g20/hackathon-video-processor-job
VERSION=$(shell git describe --tags --always --dirty)
TEST_PATH=./internal/...
TEST_COVERAGE_FILE_NAME=coverage.out

# Go commands
GOCMD=go
GOBUILD=$(GOCMD) build
GORUN=$(GOCMD) run
GOTEST=$(GOCMD) test
GOCLEAN=$(GOCMD) clean
GOVET=$(GOCMD) vet
GOFMT=$(GOCMD) fmt
GOFMT_TOOL=gofmt
RACE_FLAG=$(if $(filter 1,$(shell go env CGO_ENABLED)),-race,)
GOTIDY=$(GOCMD) mod tidy

# AWS variables
AWS_REGION=us-east-1

# Looks at comments using ## on targets and uses them to produce a help output.
.PHONY: help
help: ALIGN=25
help: ## Print this message
	@echo "Usage: make <command>"
	@awk -F '::? .*## ' -- "/^[^':]+::? .*## /"' { printf "  make '$$(tput bold)'%-$(ALIGN)s'$$(tput sgr0)' - %s\n", $$1, $$2 }' $(MAKEFILE_LIST)

.PHONY: deps
deps: ## Download dependencies
	@echo "🟢 Downloading dependencies..."
	$(GOCMD) mod download
	$(GOTIDY)

.PHONY: install
install: ## Install dependencies (alias like example project)
	@echo "🟢 Installing dependencies..."
	go mod download

.PHONY: fmt
fmt: ## Format the code
	@echo "🟢 Formatting the code..."
	$(GOFMT) ./...

.PHONY: lint
lint: ## Run linter (go vet + gofmt diff)
	@echo "🟢 Running linter..."
	@$(GOVET) ./...
	@$(GOFMT_TOOL) -d -e .

.PHONY: lint-ci
lint-ci: ## Run golangci-lint (like example project)
	@echo "🟢 Running golangci-lint..."
	@go run github.com/golangci/golangci-lint/cmd/golangci-lint@v1.64.7 run --out-format colored-line-number

.PHONY: build
build: fmt ## Build the video processor job
	@echo "🟢 Building the video processor job..."
	$(GOBUILD) -ldflags="-s -w" -o video-processor-job ./cmd/video-processor-job

.PHONY: test
test: lint ## Run tests
	@echo "🟢 Running tests..."
	$(GOTEST) $(TEST_PATH) $(RACE_FLAG) -cover

.PHONY: unit-test
unit-test: lint ## Run unit tests with coverage profile
	@echo "🟢 Running unit tests with coverage..."
	$(GOTEST) ./... $(RACE_FLAG) -cover -covermode=atomic -coverprofile=$(TEST_COVERAGE_FILE_NAME)

.PHONY: coverage-check
coverage-check: unit-test ## Fail if coverage < 80%
	@echo "🟢 Checking coverage threshold (80%)..."
	@filtered=$(TEST_COVERAGE_FILE_NAME).filtered; \
	cat $(TEST_COVERAGE_FILE_NAME) \
		| grep -v "/cmd/" \
		| grep -v "/mocks/" \
		| grep -v "/infrastructure/datasource/" \
		| grep -v "/infrastructure/logger/" \
		| grep -v "/core/domain/" \
		| grep -v "_mock.go" | grep -v "_response.go" \
		| grep -v "_gateway.go" | grep -v "_datasource.go" | grep -v "_presenter.go" \
		| grep -v "config" | grep -v "_entity.go" | grep -v "errors.go" | grep -v "_dto.go" \
		| grep -v "_request.go" | grep -v "middleware" | grep -v "route" | grep -v "util" | grep -v "database" | grep -v "server" | grep -v "httpclient" | grep -v "service" \
		| grep -v "tests/bdd" | grep -v "test/bdd" > $$filtered; \
	total_cov=$$(go tool cover -func=$$filtered | grep '^total:' | awk '{print $$3}' | tr -d '%'); \
	if [ -z "$$total_cov" ]; then \
		echo "Could not parse coverage"; exit 1; \
	fi; \
	awk -v p="$$total_cov" 'BEGIN { if (p + 0 < 80) { printf "Coverage is %.2f%% (< 80%%)\n", p; exit 1 } else { printf "Coverage is %.2f%% (>= 80%%)\n", p; exit 0 } }'


.PHONY: test-coverage
test-coverage: ## Run tests with coverage
	@echo "🟢 Running tests with coverage..."
	$(GOTEST) $(TEST_PATH) $(RACE_FLAG) -cover -coverprofile=$(TEST_COVERAGE_FILE_NAME).tmp
	@cat $(TEST_COVERAGE_FILE_NAME).tmp \
	| grep -v "/cmd/" \
	| grep -v "/mocks/" \
	| grep -v "/infrastructure/datasource/" \
	| grep -v "/infrastructure/logger/" \
	| grep -v "/core/domain/" \
	| grep -v "_mock.go" | grep -v "_response.go" \
	| grep -v "_gateway.go" | grep -v "_datasource.go" | grep -v "_presenter.go" \
	| grep -v "config" | grep -v "_entity.go" | grep -v "errors.go" | grep -v "_dto.go" \
	| grep -v "_request.go" | grep -v "middleware" | grep -v "route" | grep -v "util" | grep -v "database" | grep -v "server" | grep -v "httpclient" | grep -v "service" \
	| grep -v "tests/bdd" | grep -v "test/bdd" > $(TEST_COVERAGE_FILE_NAME)
	@rm $(TEST_COVERAGE_FILE_NAME).tmp
	$(GOCMD) tool cover -html=$(TEST_COVERAGE_FILE_NAME)

.PHONY: coverage
coverage: test-coverage ## Alias for test-coverage (like example project)

.PHONY: test-integration
test-integration: ## Run integration tests (requires AWS credentials)
	@echo "🟢 Running integration tests..."
	$(GOTEST) ./... -tags=integration $(RACE_FLAG) -cover

.PHONY: clean
clean: ## Clean up binaries and coverage files
	@echo "🔴 Cleaning up..."
	$(GOCLEAN)
	rm -f video-processor-job
	rm -f $(TEST_COVERAGE_FILE_NAME)

.PHONY: run
run: build ## Run locally (for testing)
	@echo "🟢 Running locally..."
	./video-processor-job

.PHONY: docker-build
docker-build: ## Build Docker image for Kubernetes job
	@echo "🟢 Building Docker image for Kubernetes job..."
	docker build -t video-processor-job .

.PHONY: docker-push
docker-push: ## Push Docker image
	@echo "🟢 Pushing Docker image..."
	docker push $(DOCKER_REGISTRY)/$(DOCKER_REGISTRY_APP):$(VERSION)
	docker push $(DOCKER_REGISTRY)/$(DOCKER_REGISTRY_APP):latest

.PHONY: docker-run
docker-run: docker-build ## Run Docker container locally
	@echo "🟢 Running Docker container locally..."
	docker run --rm \
		-e K8S_JOB_ENV_VIDEO_KEY=videos/test.mp4 \
		-e K8S_JOB_ENV_VIDEO_BUCKET=test-video-bucket \
		-e K8S_JOB_ENV_PROCESSED_BUCKET=test-processed-bucket \
		video-processor-job


.PHONY: mock
mock: ## Generate mocks (uber mockgen)
	@echo "🟢 Generating mocks (uber mockgen)..."
	@mkdir -p internal/core/port/mocks
	@rm -rf internal/core/port/mocks/*
	@for file in internal/core/port/*.go; do \
		go run go.uber.org/mock/mockgen@latest -source=$$file -destination=internal/core/port/mocks/`basename $$file _port.go`_mock.go; \
	done

.PHONY: security-scan
security-scan: ## Run security scan
	@echo "🟢 Running security scan..."
	@go run golang.org/x/vuln/cmd/govulncheck@latest ./...
