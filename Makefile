langs := go ts cpp
packages := go ts cpp common
soft.generators.version := 1.6.0
soft.generator.common.version := 1.4.1
soft.generator.cpp.version := 1.2.4
soft.generator.go.version := 1.9.1
soft.generator.ts.version := 1.2.0

export DOCKER_BUILDKIT=1

.PHONY : dist dist.clean versions publish

default: generators tests

# creates all docker images for soft generators
generators: generator.go generator.cpp generator.ts

generator.go:
	@echo "[generator.go]"
	@docker build --file Dockerfile-go --tag masagroup/soft.generator.go .

generator.cpp:
	@echo "[generator.cpp]"
	@docker build --file Dockerfile-cpp --tag masagroup/soft.generator.cpp .

generator.ts:
	@echo "[generator.ts]"
	@docker build --file Dockerfile-ts --tag masagroup/soft.generator.ts .

#test all generators
tests: test.go test.cpp test.ts

test.go:
	@echo "[test.go]"
	@test $(shell docker run --rm -i masagroup/soft.generator.go -v 2>&1 | sed -r 's#soft.generator.go version: (.*)#\1#g') = "$(soft.generator.go.version)" || (echo "invalid soft.generator.go version"; exit 1)

test.cpp:
	@echo "[test.cpp]"
	@test $(shell docker run --rm -i masagroup/soft.generator.cpp -v 2>&1 | sed -r 's#soft.generator.cpp version: (.*)#\1#g') = "$(soft.generator.cpp.version)" || (echo "invalid soft.generator.cpp version"; exit 1)

test.ts:
	@echo "[test.ts]"
	@test $(shell docker run --rm -i masagroup/soft.generator.ts -v 2>&1 | sed -r 's#soft.generator.ts version: (.*)#\1#g') = "$(soft.generator.ts.version)" || (echo "invalid soft.generator.ts version"; exit 1)

define export_generator
	$(eval id := $(shell docker create masagroup/soft.generator.$(1) bash))
	docker cp $(id)://usr/share/soft.generator.$(1) - | gzip > dist/soft.generator.$(1)-$(soft.generator.$(1).version).tar.gz
	docker rm $(id) > /dev/null
endef

# dist generators binaries for distribution
dist:
	@mkdir -p dist
	@$(foreach lang,$(langs),$(call export_generator,$(lang));)

# distclean clean generators binaries distribution
distclean:
	@rm -rf dist

# update projects versions in :
#	* maven pom.xml
#	* java MANIFEST.MF
#	* dockerfile entry point
versions:
	@docker run --rm -v $(CURDIR):/pwd -w /pwd klakegg/saxon xslt -s:/pwd/pom.xml -xsl:/pwd/pom.xslt -o:/pwd/pom.xml artifactId=soft.generators version=$(soft.generators.version)
	@$(foreach package,$(packages), docker run --rm -v $(CURDIR):/pwd -w /pwd klakegg/saxon xslt -s:/pwd/pom.xml -xsl:/pwd/pom.xslt -o:/pwd/pom.xml artifactId=soft.generator.$(package) version=$(soft.generator.$(package).version);)
	@$(foreach package,$(packages), sed -i "s#Bundle-Version: .*#Bundle-Version: $(soft.generator.$(package).version)#g" soft.generators/soft.generator.$(package)/META-INF/MANIFEST.MF;)
	@$(foreach lang,$(langs), sed -i "s#[0-9]*\.[0-9]*\.[0-9]*\.jar#$(soft.generator.$(lang).version).jar#g" Dockerfile-$(lang);)

# publish generators images on docker hub
publish:
	@$(foreach lang,$(langs), docker image tag masagroup/soft.generator.$(lang) masagroup/soft.generator.$(lang):v$(soft.generator.$(lang).version);)
	@$(foreach lang,$(langs), docker image push masagroup/soft.generator.$(lang):v$(soft.generator.$(lang).version);)
