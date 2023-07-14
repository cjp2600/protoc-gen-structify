GO               = go
VERSION          := $(shell git describe --tags || true)
DATE             := $(shell date +%FT%T%z)
GIT_COMMIT       := $(shell git rev-parse HEAD)
GIT_BRANCH       := $(shell git rev-parse --abbrev-ref HEAD)
GOBIN			 = $(CURDIR)/bin

PROTOC_VER := 3.15.8
PROTOC_ZIP := protoc-$(PROTOC_VER)-osx-x86_64.zip
PROTOC := $(GOBIN)/bin/protoc

$(GOBIN):
	mkdir -p $@

.PHONY: build-example
build-example: build ## Build example: make build-example f=example/blog.proto
ifndef f
f = example/blog.proto
endif
	@$(PROTOC) -I/usr/local/include -I.  \
	-I$(GOPATH)/src   \
	--plugin=protoc-gen-structify=$(GOBIN)/structify \
	--structify_out=. --structify_opt=paths=source_relative \
	$(f)

.PHONY: install-protoc
install-protoc: $(GOBIN) ## Install protocol buffer compiler
	@if [ ! -f $(PROTOC) ]; then \
		curl -L https://github.com/protocolbuffers/protobuf/releases/download/v$(PROTOC_VER)/$(PROTOC_ZIP) -o $(GOBIN)/$(PROTOC_ZIP); \
		unzip -o $(GOBIN)/$(PROTOC_ZIP) -d $(GOBIN); \
		rm $(GOBIN)/$(PROTOC_ZIP); \
	else \
		echo "protoc already exists"; \
	fi

.PHONY: install-protoc-gen-go
install-protoc-gen-go: $(GOBIN) ## Install protoc-gen-go plugin
	@GOBIN=$(GOBIN) $(GO) install github.com/golang/protobuf/protoc-gen-go@v1.3.2

.PHONY: build-options
build-options: install-tools ## Build options plugin
	@$(PROTOC) -I/usr/local/include -I. \
	-I$(GOPATH)/src \
	--plugin=protoc-gen-go=$(GOBIN)/protoc-gen-go \
	--go_out=. --go_opt=paths=source_relative \
	plugin/options/structify.proto

.PHONY: fmt
fmt: ## Format code
	$(info $(M) running gofmt...)
	@ret=0 && for d in $$($(GO) list -f '{{.Dir}}' ./... | grep -v /vendor/); do \
		$(GO) fmt $$d/*.go || ret=$$? ; \
		done ; exit $$ret

.PHONY: install-tools
install-tools: install-protoc install-protoc-gen-go ## Install tools for development

.PHONY: test
test: ## Run all tests
	$(info $(M) running tests...)
	@$(GO) test ./... -v -cover

.PHONY: clean
clean: ## Clean up
	rm -rf $(GOBIN)

.PHONY: build
build: ## Build the binary file
	@$(GO) build -o bin/structify

help:                   ##Show this help.
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

%:
	@: