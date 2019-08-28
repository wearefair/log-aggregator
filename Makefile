PACKAGE=log-aggregator
GIT_HASH := $(shell git rev-parse HEAD)
LINUX_BINARY := dist/log-aggregator-$(GIT_HASH).linux
DARWIN_BINARY := dist/log-aggregator-$(GIT_HASH).darwin
YOLO_BINARY := dist/log-aggregator-$(GIT_HASH).yolo

GO_VERSION := 1.10.3

default: help

.PHONY: help
help:
	@echo Makefile help
	@echo -------------
	@echo build: build binaries for all architectures
	@echo build-linux: containerized build amd64 linux binaries
	@echo build-darwin: containerized build amd64 darwin binaries
	@echo build-yolo: local build with arbitrary dependency versions
	@echo release: publish linux binary to s3 bucket
	@echo clean: remove local build artifacts

.PHONY: build
build: $(LINUX_BINARY) $(DARWIN_BINARY) $(YOLO_BINARY)

$(LINUX_BINARY): dep dist
	docker run --rm -v $$(pwd):/usr/src/github.com/wearefair/log-aggregator \
	  -w /usr/src/github.com/wearefair/log-aggregator \
		-e GOOS=linux \
		-e GOARCH=amd64 \
		-e GOPATH=/usr \
		golang:$(GO_VERSION) \
	  bash -c "apt-get update && apt-get install libsystemd-dev --assume-yes && \
			echo 'Running linux go build...' && \
			go build -ldflags \"-w -s\" -v -o $(LINUX_BINARY)"

$(DARWIN_BINARY): dep dist
	docker run --rm -v $$(pwd):/usr/src/github.com/wearefair/log-aggregator \
	  -w /usr/src/github.com/wearefair/log-aggregator \
		-e GOOS=darwin \
		-e GOARCH=amd64 \
		-e GOPATH=/usr \
		golang:$(GO_VERSION) \
	  bash -c "apt-get update && apt-get install libsystemd-dev --assume-yes && \
			echo 'Running darwin go build...' && \
			go build -ldflags \"-w -s\" -v -o $(DARWIN_BINARY)"


$(YOLO_BINARY): dep dist
	go build -o dist/log-aggregator-$(GIT_HASH).yolo

.PHONY: dep
dep:
	dep ensure -v

.PHONY: build-linux
build-linux: $(LINUX_BINARY)

.PHONY: build-darwin
build-darwin: $(DARWIN_BINARY)

.PHONY: build-yolo
build-yolo: $(YOLO_BINARY)

.PHONY: release
release: $(LINUX_BINARY)
	aws s3 cp $(LINUX_BINARY) s3://$(RELEASE_BUCKET)/$(PACKAGE)/$(PACKAGE)-$(GIT_HASH)

dist:
	mkdir dist

.PHONY: clean
clean:
	rm -rf dist
