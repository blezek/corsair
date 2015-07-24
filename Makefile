.PHONY: default server client deps fmt clean all release-all assets client-assets server-assets contributors
export GOPATH:=$(shell pwd)

define help

Makefile for corsair
  corsair - build the corsair app
  test    - run the tests
  run     - run the whitelist server

whitelist can be accessud using curl:

curl -x http://localhost:47011 -v http://ge.com

endef
export help

help:
	@echo "$$help"



# BUILDTAGS flags assets to be bundled (release) or read
# from the filesystems (debug).  Default is debug.
BUILDTAGS=debug

deps:
	go get -tags '$(BUILDTAGS)' -d -v corsair/...

fmt:
	go fmt corsair/...

doc:
	godoc -http=:6060 -goroot=../go

test:
	go test -v -tags '$(BUILDTAGS)' corsair

run: corsair
	bin/corsair -verbose proxy whitelist.db password.txt

whitelist: corsair
	bin/corsair proxy whitelist.txt

corsair: assets deps
	go install -tags '$(BUILDTAGS)' corsair

assets: bin/go-bindata
	bin/go-bindata -nomemcopy -tags=$(BUILDTAGS) \
		-debug=$(if $(findstring debug,$(BUILDTAGS)),true,false) \
		-o=src/corsair/assets.go \
		-prefix=assets \
		Readme.md assets/...

release: BUILDTAGS=release
release: corsair

all: fmt corsair

bin/go-bindata:
	GOOS="" GOARCH="" go get -u github.com/jteeuwen/go-bindata/...

clean:
	go clean -r corsair/...
	rm -rf pkg
