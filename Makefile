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
# <language>.version, the version of a language ("soft-gen <language> -v"), is
# the version of the runtime its templates target: go.version follows soft.go.
# Change it here when the templates follow a new runtime version.

ifndef VERSION
VERSION := $(shell ./version.sh)
endif
go.version := v1.13.1
# the versions of the languages, as "soft-gen" reads them: language=version
# pairs separated by commas (no spaces: they pass unquoted to -ldflags)
LANGUAGE_VERSIONS := go=$(go.version)
LDFLAGS := -X main.Version=$(VERSION) -X main.LanguageVersions=$(LANGUAGE_VERSIONS)
comma := ,
IMAGE ?= masagroup/soft-gen
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
	docker build --build-arg VERSION=$(VERSION) --build-arg LANGUAGE_VERSIONS=$(LANGUAGE_VERSIONS) --tag $(IMAGE) .

.PHONY: image.test
# The version of soft-gen, then the version of each language of
# LANGUAGE_VERSIONS ("soft-gen go version: v1.13.1").
image.test:
	@version=$$(docker run --rm $(IMAGE) -v); \
	expected="soft-gen version: $(VERSION)"; \
	if [ "$$version" != "$$expected" ]; then echo "image: got '$$version', expected '$$expected'"; exit 1; fi; \
	echo "image: $$version"
	@for pair in $(subst $(comma), ,$(LANGUAGE_VERSIONS)); do \
		lang=$${pair%%=*}; expected=$${pair#*=}; \
		version=$$(docker run --rm $(IMAGE) $$lang -v) || exit 1; \
		case "$$version" in *" version: $$expected") echo "image: $$version";; \
		*) echo "image: got '$$version', expected the $$lang version $$expected"; exit 1;; esac; \
	done

.PHONY: publish
publish:
	docker tag $(IMAGE) $(IMAGE):$(VERSION)
	docker push $(IMAGE):$(VERSION)
	docker push $(IMAGE):latest
