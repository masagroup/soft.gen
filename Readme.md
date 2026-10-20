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

```shell
$ docker run --rm -v "$PWD":/gen -w /gen masagroup/soft.generator.go -s -m /gen/model.ecore -o /gen
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

## Several implementations of a model (Go) ##

`generateModel` generates the interfaces of a model and one implementation of it. The templates
`generateInterfaces` and `generateImplementations` generate them apart: with an implementation
package path distinct from the interface package path (`packages` property:
`<nsURI>=<interface path>#<implementation path>`), each implementation is a complete package, with its
own package and factory instances (`GetPackage`, `GetFactory`), and several implementations of the same
model, with different options, can be generated beside the same interfaces, one run each:

```shell
# the interfaces, once
$ docker run --rm -v "$PWD":/gen -w /gen masagroup/soft.generator.go -m /gen/model.ecore -o /gen -t generateInterfaces \
    -p packages=<nsURI>=example.com/m/model/model#example.com/m/model/store/store -p module=example.com/m
# an implementation with reflective feature delegation
$ docker run --rm -v "$PWD":/gen -w /gen masagroup/soft.generator.go -m /gen/model.ecore -o /gen -t generateImplementations \
    -p packages=<nsURI>=example.com/m/model/model#example.com/m/model/store/store -p module=example.com/m \
    -p featureDelegation=Reflective
# an implementation with typed fields
$ docker run --rm -v "$PWD":/gen -w /gen masagroup/soft.generator.go -m /gen/model.ecore -o /gen -t generateImplementations \
    -p packages=<nsURI>=example.com/m/model/model#example.com/m/model/fields/fields -p module=example.com/m
```

The implementations have the same nsURI and classifier IDs: a program chooses one by registering its
package (in the package registry of a resource set, for example). With equal interface and
implementation paths, the package and factory instances are generated with the interfaces:
use `generateModel`.

### One package shared by the implementations ###

With the property `packageInstance=interfaces` (and distinct paths), the package instance (`GetPackage`)
is generated with the interfaces, by `generateInterfaces`, with a factory of the data types of the model,
and shared by the implementations: `generateImplementations` generates the classes, whose `EClass` is
the shared one, and a factory (`GetFactory`) whose package is the shared one, but no package. The objects
of an implementation do not mix with the objects of another: the setters of the references check it
(soft.go `ecore.CheckSameImplementation`, when the program is built with the tag `ecore_check`; the
property `checkImplementation=false` leaves these checks out of the generated code).

```shell
$ docker run --rm -v "$PWD":/gen -w /gen masagroup/soft.generator.go -m /gen/model.ecore -o /gen -t generateInterfaces \
    -p packageInstance=interfaces -p packages=<nsURI>=example.com/m/model/model#example.com/m/model/store/store -p module=example.com/m
$ docker run --rm -v "$PWD":/gen -w /gen masagroup/soft.generator.go -m /gen/model.ecore -o /gen -t generateImplementations \
    -p packageInstance=interfaces -p packages=<nsURI>=example.com/m/model/model#example.com/m/model/store/store -p module=example.com/m \
    -p featureDelegation=Reflective
$ docker run --rm -v "$PWD":/gen -w /gen masagroup/soft.generator.go -m /gen/model.ecore -o /gen -t generateImplementations \
    -p packageInstance=interfaces -p packages=<nsURI>=example.com/m/model/model#example.com/m/model/memory/memory -p module=example.com/m
```

A resource set chooses its implementation in its package registry: the generated `Install(registry)` of an
implementation registers the package and the factory of the implementation (`memory.Install(resourceSet.GetPackageRegistry())`,
or `registry.RegisterFactory(memory.GetFactory())`). There is no global default implementation: the factory
instance of the package only converts data types, and its `Create` panics.

## Quick Build ##
If you want to bootstrap generators yourself, you'll need:
- Make
- Docker
- Run Make, specifying dist target with the following command:
    ```
    make all dist
    ```

## IDE ##
If you want to debug generators with an Eclipse IDE, you'll need:
- [Eclipse IDE Modeling Tools](https://www.eclipse.org/downloads/)
- [Eclipse IDE integration for Maven](https://download.eclipse.org/technology/m2e/releases/latest/)
  - maven should be configured with github as available maven repository by adding [settings.xml](settings.m2.xml) in ${user}/.m2 directory
- [Annotation Processor in IDE](https://immutables.github.io/apt.html)
- [Acceleo](https://download.eclipse.org/acceleo/updates/releases)

