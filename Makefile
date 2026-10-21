# soft-gen: the generators.
#
#   make build       build ./bin/soft-gen
#   make test        go vet + go test
#   make lint        golangci-lint (.golangci.yaml), in docker
#   make image       build the docker image (every language included)
#   make image.test  check the versions of the image
#   make publish     push the image (tags <VERSION> and latest)
#
# VERSION, the version of soft-gen ("soft-gen -v"), is computed from the Git
# state by version.sh (a Go pseudo-version such as
# v2.0.0-20261021125246-3166c30bb912, or the v2 tag of the commit).
#
# The version of a language ("soft-gen <language> -v"), the version of the
# runtime its templates target, is a constant of its package:
# internal/lang/golang.Version follows soft.go, internal/lang/typescript.Version
# soft.ts. Every build has it, "go run .../cmd/soft-gen@v2.x" included.

ifndef VERSION
VERSION := $(shell ./version.sh)
endif
LDFLAGS := -X main.Version=$(VERSION)
IMAGE ?= masagroup/soft-gen
# the languages whose version image.test checks
LANGUAGES := go ts
# the golangci-lint image of `make lint`: its Go must support the go directive
# of go.mod
GOLANGCI_LINT_IMAGE ?= golangci/golangci-lint:v2.14.0-alpine

.PHONY: build
build:
	go build -ldflags "$(LDFLAGS)" -o bin/soft-gen ./cmd/soft-gen

.PHONY: test
test:
	go vet ./...
	go test ./...

.PHONY: lint
# The sources are mounted read-only; the Go modules and the golangci-lint cache
# are docker volumes, kept between runs. MSYS_NO_PATHCONV keeps Git Bash on
# Windows from rewriting the container paths.
lint:
	MSYS_NO_PATHCONV=1 docker run --rm \
		-v "$(CURDIR)":/src:ro -w /src \
		-v soft-gen-gomod:/go/pkg/mod \
		-v soft-gen-golangci:/root/.cache \
		$(GOLANGCI_LINT_IMAGE) golangci-lint run ./...

.PHONY: image
image:
	docker build --build-arg VERSION=$(VERSION) --tag $(IMAGE) .

.PHONY: image.test
# The version of soft-gen, then the version of each language, which must be
# the one of the sources ("soft-gen go version: v1.13.1").
image.test:
	@version=$$(docker run --rm $(IMAGE) -v); \
	expected="soft-gen version: $(VERSION)"; \
	if [ "$$version" != "$$expected" ]; then echo "image: got '$$version', expected '$$expected'"; exit 1; fi; \
	echo "image: $$version"
	@for lang in $(LANGUAGES); do \
		expected=$$(go run ./cmd/soft-gen $$lang -v) || exit 1; \
		version=$$(docker run --rm $(IMAGE) $$lang -v) || exit 1; \
		if [ "$$version" != "$$expected" ]; then echo "image: got '$$version', expected '$$expected'"; exit 1; fi; \
		echo "image: $$version"; \
	done

.PHONY: publish
publish:
	docker tag $(IMAGE) $(IMAGE):$(VERSION)
	docker push $(IMAGE):$(VERSION)
	docker push $(IMAGE):latest
