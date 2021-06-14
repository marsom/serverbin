[![Build](https://github.com/marsom/serverbin/workflows/build/badge.svg)](https://github.com/marsom/serverbin/actions)
[![License](https://img.shields.io/github/license/marsom/serverbin)](/LICENSE)
[![Go Report Card](https://goreportcard.com/badge/github.com/marsom/serverbin)](https://goreportcard.com/report/github.com/marsom/serverbin)
[![Go Reference](https://pkg.go.dev/badge/github.com/marsom/serverbin.svg)](https://pkg.go.dev/github.com/marsom/serverbin)
[![Powered By: GoReleaser](https://img.shields.io/badge/powered%20by-goreleaser-green.svg)](https://github.com/goreleaser)

# serverbin

A simple request and response service with support for

- HTTP 
- TCP


## Usage

### docker

Run the http test server:
```
docker run -p 8080:8080 -p 8081:8081 -ti --rm marsom/serverbin http
```


Run the tcp test server:
```
docker run -p 8080:8080 -p 8081:8081 -ti --rm marsom/serverbin tcp
```

### go install

Install binary with go
```
go install github.com/marsom/serverbin/cmd/serverbin
```

Run the http test server:
```
serverbin http
```


Run the tcp test server:
```
serverbin tcp
```

### manually

Download the pre-compiled binaries from the [releases](https://github.com/marsom/serverbin/releases) page and copy to 
the desired location.

## Build

Download and install the build requirements:
* [go](https://golang.org/) 1.16+
* [goreleaser](https://goreleaser.com/)

Optional development tools:
* [golangci-lint](https://golangci-lint.run/)

Build binaries
```
goreleaser build --snapshot --rm-dist
```

Build binaries and docker images
```
goreleaser release --snapshot --rm-dist
```

## Release

* Create and push a release candidate tag, i.e., v0.0.1-rc1
* If all look good then create the tagged version, i.e., v0.0.1
* Update release notes in github
  * goreleaser generated docker images section contains architecture based images. this is not correct because we 
    build a docker manifest in a separate step.
  * remove prerelease flag
* Manually trigger the "Release latest images" github action with the 