IMAGE ?= docker.io/openshift/origin-cluster-svcat-apiserver-operator
TAG ?= latest
PROG  := cluster-svcat-apiserver-remover
REPO_PATH:= github.com/openshift/cluster-svcat-apiserver-operator
GO_LD_FLAGS := -ldflags "-X '${REPO_PATH}/pkg/version.SourceGitCommit=$(shell git rev-parse HEAD)'"
SOURCES := $(shell find . -name '*.go' -not -path "*/vendor/*")

all: build build-image verify
.PHONY: all
build:
	GODEBUG=tls13=1 go build ${GO_LD_FLAGS} ./cmd/cluster-svcat-apiserver-remover
.PHONY: build

image:
	docker build -t "$(IMAGE):$(TAG)" .
.PHONY: build-image

test: test-unit test-e2e
.PHONY: test

test-unit:
ifndef JUNITFILE
	go test $(GO_LD_FLAGS) -race ./...
else
ifeq (, $(shell which gotest2junit 2>/dev/null))
$(error gotest2junit not found! Get it by `go get -u github.com/openshift/release/tools/gotest2junit`.)
endif
	go test $(GO_LD_FLAGS) -race -json ./... | gotest2junit > $(JUNITFILE)
endif
.PHONY: test-unit

test-e2e:
	go test -v ./test/e2e/...
.PHONY: test-e2e

verify: verify-govet
	@gofmt -l $(SOURCES) | grep ".*\.go"; if [ "$$?" = "0" ]; then exit 1; fi
.PHONY: verify

verify-govet:
	go vet $(GO_LD_FLAGS) ./...
.PHONY: verify-govet

clean:
	rm -f $(PROG)
.PHONY: clean
