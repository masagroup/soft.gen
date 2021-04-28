.PHONY : dist dist.clean versions

default: generators

# generators cerates all docker images for soft generators
generators: generator.base generator.go generator.cpp generator.ts

generator.base:
	@echo "[generator.base]"
	@docker build --file Dockerfile --tag masagroup/soft.generator.base .

generator.go:
	@echo "[generator.go]"
	@docker build --file Dockerfile-go --tag masagroup/soft.generator.go .

generator.cpp:
	@echo "[generator.cpp]"
	@docker build --file Dockerfile-cpp --tag masagroup/soft.generator.cpp .

generator.ts:
	@echo "[generator.ts]"
	@docker build --file Dockerfile-ts --tag masagroup/soft.generator.ts .

langs := go ts cpp
packages := go ts cpp common
soft.generators.version := 1.4.0
soft.generator.common.version := 1.2.2
soft.generator.cpp.version := 1.2.2
soft.generator.go.version := 1.4.0
soft.generator.ts.version := 1.0.2

# dist generators binaries for distribution
dist:
	@mkdir -p dist
	@$(eval id := $(shell docker create masagroup/soft.generator.base bash))
	@$(foreach lang,$(langs), docker cp $(id):/soft.generator.$(lang) - | gzip > dist/soft.generator.$(lang)-$(soft.generator.$(lang).version).tar.gz;)
	@docker rm $(id) > /dev/null

# distclean clean generators binaries distribution
distclean:
	@rm -rf dist



# update projects versions in :
#	* maven pom.xml
#	* java MANIFEST.MF
#	* dockerfile entry point
versions:
	@docker run --rm -d -v $(pwd):/pwd -w /pwd klakegg/saxon xslt -s:/pwd/build.xml -xsl:/pwd/pom.xslt -o:/pwd/build.xml artifactId=soft.generators version=$(soft.generators.version) > /dev/null
	@$(foreach package,$(packages), docker run --rm -d -v $(CURDIR):/pwd -w /pwd klakegg/saxon xslt -s:/pwd/build.xml -xsl:/pwd/pom.xslt -o:/pwd/build.xml artifactId=soft.generator.$(lang) version=$(soft.generator.$(lang).version) > /dev/null;)
	@$(foreach package,$(packages), sed -i "s#Bundle-Version: .*#Bundle-Version: $(soft.generator.$(package).version)#g" soft.generators/soft.generator.$(package)/META-INF/MANIFEST.MF;)
	@$(foreach lang,$(langs), sed -i "s#[0-9]*\.[0-9]*\.[0-9]*\.jar#$(soft.generator.$(lang).version).jar#g" Dockerfile-$(lang);)
