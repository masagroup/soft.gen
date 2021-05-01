[![](https://img.shields.io/github/license/masagroup/soft.gen.svg)](https://github.com/masagroup/soft.gen/blob/master/LICENSE)
![](https://github.com/masagroup/soft.gen/actions/workflows/build_and_test.yml/badge.svg)
![](https://img.shields.io/github/v/release/masagroup/soft.gen)
# Soft Generators #

Soft Generators is an implementation of EMF generators for C++, Go and TypeScript languages

Soft Generators is part of [Soft](https://github.com/masagroup/soft) project



## Installation ##

### Github Release ###
- Install a Java JRE 8+
- Download archive of a generator from the [project releases page](https://github.com/masagroup/soft.gen/releases)
- Unzip archive

Generator can now be run with the following command:

```shell
$ java -jar sof.generator.<lang>-<version>.jar --version

<lang>    = the generator language : go, ts, cpp
<version> = the generator version
```

Example:
```shell
$ java -jar soft.generator.go-1.5.0.jar --version
soft.generator.go version: 1.5.0
```

### Docker Hub ###
Use the Docker images available from Docker Hub:
- [Cpp generator image](https://hub.docker.com/r/masagroup/soft.generator.cpp)
- [Go generator image](https://hub.docker.com/r/masagroup/soft.generator.go)
- [TypeScript generator image](https://hub.docker.com/r/masagroup/soft.generator.ts)

Example:
```shell
$ docker run --rm masagroup/soft.generator.go  --version
soft.generator.go version: 1.5.0
```

## Usage ##

```
Usage: [-hsv] -m=<model> -o=<folder> [-p=<property=value>]... [-P=<propertyfile>]... [-t=<template>]...
  -h, --help                        print this help and exit
  -m, --model=<model>               the input model
  -o, --output=<folder>             the output folder
  -p, --property=<property=value>   set value for given property
  -P, --properties=<propertyfile>   load properties from a property file
  -s, --silent                      print nothing but failures
  -t, --template=<template>         the template to be executed: <templates list>
  -v, --version                     print version information and exit
```

## Quick Build ##
If you want to bootstrap generators yourself, you'll need:
- Make
- Docker
- Run Make, specifying dist target with the following command:
    ```
    make all dist
    ```



