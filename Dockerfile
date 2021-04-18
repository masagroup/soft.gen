#
# Build stage
#
FROM maven:3.8.1-jdk-11-slim AS build
COPY . /home/soft.gen/src
RUN mvn -f /home/soft.gen/src/soft.acceleo/pom.xml -Pdocker clean install \
 && mvn -f /home/soft.gen/src/soft.generators/pom.xml -Pdocker clean verify
