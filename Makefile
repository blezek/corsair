.PHONY: default server client deps fmt clean all release-all assets client-assets server-assets contributors
export GOPATH:=$(shell pwd)

BUILDTAGS=debug
default: all

deps:
	go get -tags '$(BUILDTAGS)' -d -v corsair/...

fmt:
	go fmt corsair/...

corsair: assets deps
	go install -tags '$(BUILDTAGS)' corsair

assets: bin/go-bindata
	bin/go-bindata -nomemcopy -tags=$(BUILDTAGS) \
		-debug=$(if $(findstring debug,$(BUILDTAGS)),true,false) \
		-o=src/corsair/assets.go \
		Readme.md assets/...

release: BUILDTAGS=release
release: corsair

all: fmt corsair

bin/go-bindata:
	GOOS="" GOARCH="" go get -u github.com/jteeuwen/go-bindata/...

clean:
	go clean -r corsair/...
