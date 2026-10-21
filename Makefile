# soft-gen: the generators.
#
#   make build       build ./bin/soft-gen
#   make test        go vet + go test
#   make lint        golangci-lint (.golangci.yaml), in docker
#   make image       build the docker image (every language included)
#   make image.test  check the version of the image
#   make publish     push the image (tags <VERSION> and latest)
#
# VERSION is computed from the Git state by version.sh (a Go pseudo-version
# such as v2.0.0-20261021125246-3166c30bb912, or the v2 tag of the commit).

ifndef VERSION
VERSION := $(shell ./version.sh)
endif
IMAGE ?= masagroup/soft-gen
# the golangci-lint image of `make lint`: its Go must support the go directive
# of go.mod
GOLANGCI_LINT_IMAGE ?= golangci/golangci-lint:v2.14.0-alpine

.PHONY: build
build:
	go build -ldflags "-X main.Version=$(VERSION)" -o bin/soft-gen ./cmd/soft-gen

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
	DOCKER_BUILDKIT=1 docker build --build-arg VERSION=$(VERSION) --tag $(IMAGE) .

.PHONY: image.test
image.test:
	@version=$$(docker run --rm $(IMAGE) -v); \
	expected="soft-gen version: $(VERSION)"; \
	if [ "$$version" != "$$expected" ]; then echo "image: got '$$version', expected '$$expected'"; exit 1; fi; \
	echo "image: $$version"

.PHONY: publish
publish:
	docker tag $(IMAGE) $(IMAGE):$(VERSION)
	docker push $(IMAGE):$(VERSION)
	docker push $(IMAGE):latest
