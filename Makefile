PACKAGE=log-aggregator
GIT_HASH := $(shell git rev-parse HEAD)

dist/log-aggregator-$(GIT_HASH).linux: | dist
	docker run --rm -v $$(pwd):/usr/src/github.com/wearefair/log-aggregator \
	  -w /usr/src/github.com/wearefair/log-aggregator \
		-e GOOS=linux \
		-e GOARCH=amd64 \
		-e GOPATH=/usr \
		golang:1.8.1 \
	  bash -c "apt-get update && apt-get install libsystemd-dev && \
			echo 'Running go build...' && go build -ldflags \"-w -s\" -v -o dist/log-aggregator-$(GIT_HASH).linux"

build-linux: dist/log-aggregator-$(GIT_HASH).linux

build: | dist
	go build -o dist/log-aggregator-$(GIT_HASH).darwin

release: dist/log-aggregator-$(GIT_HASH).linux
	aws s3 cp dist/log-aggregator-$(GIT_HASH).linux s3://fair-releases-stage-secure/$(PACKAGE)/$(PACKAGE)-$(GIT_HASH)

dist:
	mkdir dist

clean:
	rm -rf dist
