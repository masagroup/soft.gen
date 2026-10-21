# soft.gen: the generators.
#
#   make build       build ./bin/soft-gen
#   make test        go vet + go test
#   make image       build the docker image (every language included)
#   make image.test  check the version of the image
#   make publish     push the image (tags <VERSION> and latest)
#
# VERSION is computed from the Git state by version.sh (a Go pseudo-version
# such as v2.0.0-20261021125246-3166c30bb912, or the v2 tag of the commit).

ifndef VERSION
VERSION := $(shell ./version.sh)
endif
IMAGE ?= masagroup/soft.generator

.PHONY: build
build:
	go build -ldflags "-X main.Version=$(VERSION)" -o bin/soft-gen ./cmd/soft-gen

.PHONY: test
test:
	go vet ./...
	go test ./...

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
