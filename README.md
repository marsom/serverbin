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
* go 1.16+
* goreleaser
* golangci-lint


Build the binaries
```
goreleaser build --snapshot --rm-dist
```

Build the binaries and docker images
```
goreleaser release --snapshot --rm-dist
```
