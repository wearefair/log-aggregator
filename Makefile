PACKAGE=log-aggregator
GIT_HASH := $(shell git rev-parse HEAD)

default: build

dist/log-aggregator-$(GIT_HASH).linux: dep | dist
	docker run --rm -v $$(pwd):/usr/src/github.com/wearefair/log-aggregator \
	  -w /usr/src/github.com/wearefair/log-aggregator \
		-e GOOS=linux \
		-e GOARCH=amd64 \
		-e GOPATH=/usr \
		golang:1.10.3 \
	  bash -c "apt-get update && apt-get install libsystemd-dev --assume-yes && \
			echo 'Running go build...' && go build -ldflags \"-w -s\" -v -o dist/log-aggregator-$(GIT_HASH).linux"

dep:
	dep ensure -v

build: dep | dist
	go build -o dist/log-aggregator-$(GIT_HASH).darwin

build-linux: dist/log-aggregator-$(GIT_HASH).linux

release: dist/log-aggregator-$(GIT_HASH).linux
	aws s3 cp dist/log-aggregator-$(GIT_HASH).linux s3://$(RELEASE_BUCKET)/$(PACKAGE)/$(PACKAGE)-$(GIT_HASH)

dist:
	mkdir dist

clean:
	rm -rf dist
