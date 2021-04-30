[![License](https://img.shields.io/github/license/masagroup/soft.gen.svg)](https://github.com/masagroup/soft.gen/blob/master/LICENSE)

# Soft Generators #

Soft Generators is the implementation of the EMF generators in C++, Go and TypeScript langages

Soft Generators is part of [Soft](https://github.com/masagroup/soft) project



## Installation ##

### Github Release ###
- Install a Java JRE 8+
- Download archive of a generator from the [project releases page](https://github.com/masagroup/soft.gen/releases)
- Unzip archive

Generator can now be run with the following command:
```shell
$ java -jar <generator>-<version>.jar --version

with <generator> = 
soft.generator.go   the Go generator
soft.generator.ts   the TypeScript generator
soft.generator.cpp  the C++ generator

with <version> = the generator version
```

Example:
```shell
$ java -jar soft.generator.go-1.5.0.jar --version
soft.generator.go version: 1.5.0
```

### Docker Hub ###
Use the Docker images available from Docker Hub:
- [Cpp generator image](https://hub.docker.com/repository/docker/masagroup/soft.generator.cpp)
- [Go generator image](https://hub.docker.com/repository/docker/masagroup/soft.generator.go)
- [TypeScript generator image](https://hub.docker.com/repository/docker/masagroup/soft.generator.ts)

Example:
```shell
$ docker run --rm -i masagroup/soft.generator.go  --version
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

## Examples ##
![](https://raw.githubusercontent.com/masagroup/soft.gen/master/docs/example-docker.gif)

## Quick Build ##
If you want to bootstrap generators yourself, you'll need:
- Make
- Docker
- Run Make, specifying dist target with the following command:
    ```
    make all dist
    ```



