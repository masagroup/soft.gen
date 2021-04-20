default: generators

# generators cerates all docker images for soft generators
generators: generator.base generator.go generator.cpp generator.ts

generator.base:
	@docker build --file Dockerfile --tag soft.generator.base .

generator.go:
	@docker build --file Dockerfile-go --tag soft.generator.go .

generator.cpp:
	@docker build --file Dockerfile-cpp --tag soft.generator.cpp .

generator.ts:
	@docker build --file Dockerfile-ts --tag soft.generator.ts .

LANGS := go ts cpp
PACKAGES := go ts cpp common

# dist generators binaries for distribution
dist:
	@mkdir -p dist
	@$(eval id := $(shell docker create soft.generator.base bash))
	@$(foreach lang,${LANGS}, docker cp $(id):/soft.generator.$(lang) - | gzip > dist/soft.generator.$(lang)-$(soft.generator.$(lang).version).tar.gz;)
	@docker rm $(id) > /dev/null

.PHONY : dist

# distclean clean generators binaries distribution
distclean:
	@rm -rf dist

.PHONY : distclean

soft.generators.version := 1.4.1
soft.generator.common.version := 1.2.2
soft.generator.cpp.version := 1.2.2
soft.generator.go.version := 1.4.0
soft.generator.ts.version := 1.0.2
pwd := $(shell pwd)

# update projects versions in :
#	* maven pom.xml
#	* java MANIFEST.MF
#	* dockerfile entry point
versions:
	@docker run --rm -d -v $(pwd):/pwd -w /pwd klakegg/saxon xslt -s:/pwd/build.xml -xsl:/pwd/pom.xslt -o:/pwd/build.xml artifactId=soft.generators version=$(soft.generators.version) > /dev/null
	@$(foreach package,${PACKAGES}, docker run --rm -d -v $(pwd):/pwd -w /pwd klakegg/saxon xslt -s:/pwd/build.xml -xsl:/pwd/pom.xslt -o:/pwd/build.xml artifactId=soft.generator.$(lang) version=$(soft.generator.$(lang).version) > /dev/null;)
	@$(foreach package,${PACKAGES}, sed -i "s#Bundle-Version: .*#Bundle-Version: $(soft.generator.$(package).version)#g" soft.generators/soft.generator.$(package)/META-INF/MANIFEST.MF;)
	@$(foreach lang,${LANGS}, sed -i "s#[0-9]*\.[0-9]*\.[0-9]*\.jar#$(soft.generator.$(lang).version).jar#g" Dockerfile-$(lang);)

.PHONY : versions
