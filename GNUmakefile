default: fmt lint install generate

build:
	go build -v ./...

install: build
	go install -v ./...

lint:
	golangci-lint run

generate:
	cd tools; go generate ./...

fmt:
	gofmt -s -w -e .

test:
	go test -v -cover -timeout=120s -parallel=10 ./...

testacc:
	TF_ACC=1 go test -v -cover -timeout 120m ./...

docs: generate
	@echo "Generating documentation..."
	@cd tools && go generate
	@echo "Documentation generated successfully"

docs-check: docs
	@echo "Checking if documentation is up to date..."
	@git diff --exit-code docs/ || (echo "Documentation is out of date. Run 'make docs' to update." && exit 1)
	@echo "Documentation is up to date"

validate-examples:
	@echo "Validating example files..."
	@terraform fmt -check -recursive examples/ || (echo "Examples need formatting. Run 'terraform fmt -recursive examples/'" && exit 1)
	@echo "Examples are properly formatted"

validate-docs: docs-check validate-examples
	@echo "All documentation validation passed"

.PHONY: fmt lint test testacc build install generate docs docs-check validate-examples validate-docs
