.DEFAULT_GOAL := help

# Variables
APP_NAME=video-processor-job
MAIN_FILE=main.go
DOCKER_REGISTRY=ghcr.io
DOCKER_REGISTRY_APP=fiap-soat-g20/hackathon-video-processor-job
VERSION=$(shell git describe --tags --always --dirty)
TEST_PATH=./internal/...
TEST_COVERAGE_FILE_NAME=coverage.out
LAMBDA_ZIP_FILE=function.zip

# Go commands
GOCMD=go
GOBUILD=$(GOCMD) build
GORUN=$(GOCMD) run
GOTEST=$(GOCMD) test
GOCLEAN=$(GOCMD) clean
GOVET=$(GOCMD) vet
GOFMT=$(GOCMD) fmt
GOTIDY=$(GOCMD) mod tidy

# AWS variables
AWS_REGION=us-east-1
FUNCTION_NAME=video-processor-job

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

.PHONY: fmt
fmt: ## Format the code
	@echo "🟢 Formatting the code..."
	$(GOFMT) ./...

.PHONY: lint
lint: ## Run linter
	@echo "🟢 Running linter..."
	@$(GOVET) ./...
	@$(GOFMT) -d -e ./...

.PHONY: build
build: fmt ## Build the Lambda function
	@echo "🟢 Building the Lambda function..."
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 $(GOBUILD) -ldflags="-s -w" -o bootstrap $(MAIN_FILE)

.PHONY: build-lambda
build-lambda: build ## Build Lambda deployment package
	@echo "🟢 Creating Lambda deployment package..."
	@zip -r $(LAMBDA_ZIP_FILE) bootstrap
	@echo "Lambda package created: $(LAMBDA_ZIP_FILE)"

.PHONY: test
test: lint ## Run tests
	@echo "🟢 Running tests..."
	$(GOTEST) $(TEST_PATH) -race -cover

.PHONY: test-coverage
test-coverage: ## Run tests with coverage
	@echo "🟢 Running tests with coverage..."
	$(GOTEST) $(TEST_PATH) -race -cover -coverprofile=$(TEST_COVERAGE_FILE_NAME).tmp
	@cat $(TEST_COVERAGE_FILE_NAME).tmp | grep -v "_mock.go" | grep -v "_response.go" \
	| grep -v "_gateway.go" | grep -v "_datasource.go" | grep -v "_presenter.go" \
	| grep -v "config" | grep -v "logger" | grep -v "_entity.go" | grep -v "errors.go" | grep -v "_dto.go" > $(TEST_COVERAGE_FILE_NAME)
	@rm $(TEST_COVERAGE_FILE_NAME).tmp
	$(GOCMD) tool cover -html=$(TEST_COVERAGE_FILE_NAME)

.PHONY: test-integration
test-integration: ## Run integration tests (requires AWS credentials)
	@echo "🟢 Running integration tests..."
	$(GOTEST) ./... -tags=integration -race -cover

.PHONY: clean
clean: ## Clean up binaries and coverage files
	@echo "🔴 Cleaning up..."
	$(GOCLEAN)
	rm -f bootstrap
	rm -f $(LAMBDA_ZIP_FILE)
	rm -f $(TEST_COVERAGE_FILE_NAME)

.PHONY: run
run: build ## Run locally (for testing)
	@echo "🟢 Running locally..."
	./bootstrap

.PHONY: docker-build
docker-build: ## Build Docker image for Lambda
	@echo "🟢 Building Docker image for Lambda..."
	docker build --platform linux/amd64 -t $(DOCKER_REGISTRY)/$(DOCKER_REGISTRY_APP):$(VERSION) .
	docker tag $(DOCKER_REGISTRY)/$(DOCKER_REGISTRY_APP):$(VERSION) $(DOCKER_REGISTRY)/$(DOCKER_REGISTRY_APP):latest

.PHONY: docker-push
docker-push: ## Push Docker image
	@echo "🟢 Pushing Docker image..."
	docker push $(DOCKER_REGISTRY)/$(DOCKER_REGISTRY_APP):$(VERSION)
	docker push $(DOCKER_REGISTRY)/$(DOCKER_REGISTRY_APP):latest

.PHONY: docker-run
docker-run: docker-build ## Run Docker container locally
	@echo "🟢 Running Docker container locally..."
	docker run --rm -p 9000:8080 \
		-e VIDEO_BUCKET=test-video-bucket \
		-e PROCESSED_BUCKET=test-processed-bucket \
		$(DOCKER_REGISTRY)/$(DOCKER_REGISTRY_APP):latest

.PHONY: lambda-deploy
lambda-deploy: build-lambda ## Deploy to AWS Lambda (requires AWS CLI)
	@echo "🟢 Deploying to AWS Lambda..."
	aws lambda update-function-code \
		--function-name $(FUNCTION_NAME) \
		--zip-file fileb://$(LAMBDA_ZIP_FILE) \
		--region $(AWS_REGION)

.PHONY: lambda-create
lambda-create: build-lambda ## Create AWS Lambda function
	@echo "🟢 Creating AWS Lambda function..."
	aws lambda create-function \
		--function-name $(FUNCTION_NAME) \
		--runtime provided.al2 \
		--role arn:aws:iam::YOUR_ACCOUNT:role/lambda-execution-role \
		--handler bootstrap \
		--zip-file fileb://$(LAMBDA_ZIP_FILE) \
		--timeout 900 \
		--memory-size 1024 \
		--region $(AWS_REGION)

.PHONY: lambda-invoke
lambda-invoke: ## Test Lambda function
	@echo "🟢 Testing Lambda function..."
	aws lambda invoke \
		--function-name $(FUNCTION_NAME) \
		--payload '{"video_key": "test-video.mp4"}' \
		--region $(AWS_REGION) \
		response.json
	@cat response.json
	@rm response.json

.PHONY: lambda-logs
lambda-logs: ## View Lambda function logs
	@echo "🟢 Viewing Lambda function logs..."
	aws logs tail /aws/lambda/$(FUNCTION_NAME) --follow

.PHONY: mock
mock: ## Generate mocks
	@echo "🟢 Generating mocks..."
	@rm -rf internal/core/port/mocks/*
	@for file in internal/core/port/*.go; do \
		go tool mockgen -source=$$file -destination=internal/core/port/mocks/`basename $$file _port.go`_mock.go; \
	done

.PHONY: security-scan
security-scan: ## Run security scan
	@echo "🟢 Running security scan..."
	@go tool govulncheck -show verbose ./...