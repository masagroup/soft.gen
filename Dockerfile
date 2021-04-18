#
# Build stage
#
FROM maven:3.8.1-jdk-8-slim AS build
COPY . /home/soft.gen/src
RUN --mount=type=cache,target=/root/.m2 mvn -f /home/soft.gen/src/soft.acceleo/pom.xml -Pdocker clean install \
 && mvn -f /home/soft.gen/src/soft.generators/pom.xml -Pdocker clean verify

FROM scratch
COPY --from=build /home/soft.gen/out/soft.generator.go/*.jar /soft.generator.go/
COPY --from=build /home/soft.gen/out/soft.generator.ts/*.jar /soft.generator.ts/
COPY --from=build /home/soft.gen/out/soft.generator.cpp/*.jar /soft.generator.cpp/
