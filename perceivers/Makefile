ifdef IMAGE_PREFIX
PREFIX="$(IMAGE_PREFIX)-"
endif

ifneq (, $(findstring gcr.io,$(REGISTRY)))
PREFIX_CMD="gcloud"
DOCKER_OPTS="--"
endif

OUTDIR=_output
BINARY=image-perceiver

CURRENT_DIR:=$(shell dirname $(realpath $(lastword $(MAKEFILE_LIST))))
OUTDIR=_output

.PHONY: $(SUBDIRS) clean build local_build container push test

all: build

build: $(OUTDIR)
	docker run --rm -e CGO_ENABLED=0 -e GOOS=linux -e GOARCH=amd64 -v "${CURRENT_DIR}/..":/go/src/github.com/blackducksoftware/opssight-connector -w /go/src/github.com/blackducksoftware/opssight-connector/perceivers golang go build -o ${OUTDIR}/image-perceiver cmd/image-perceiver/image-perceiver.go
	docker run --rm -e CGO_ENABLED=0 -e GOOS=linux -e GOARCH=amd64 -v "${CURRENT_DIR}/..":/go/src/github.com/blackducksoftware/opssight-connector -w /go/src/github.com/blackducksoftware/opssight-connector/perceivers golang go build -o ${OUTDIR}/pod-perceiver cmd/pod-perceiver/pod-perceiver.go

local_build: $(OUTDIR) cmd/image-perceiver/image-perceiver.go cmd/pod-perceiver/pod-perceiver.go
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o ${OUTDIR}/image-perceiver cmd/image-perceiver/image-perceiver.go
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o ${OUTDIR}/pod-perceiver cmd/pod-perceiver/pod-perceiver.go

container: registry_check $(OUTDIR)/image-perceiver ${OUTDIR}/pod-perceiver
	docker build -t $(REGISTRY)/$(PREFIX)image-perceiver -f Dockerfile.image .
	docker build -t $(REGISTRY)/$(PREFIX)pod-perceiver -f Dockerfile.pod .

push: registry_check container
	$(PREFIX_CMD) docker $(DOCKER_OPTS) push $(REGISTRY)/$(PREFIX)image-perceiver:latest
	$(PREFIX_CMD) docker $(DOCKER_OPTS) push $(REGISTRY)/$(PREFIX)pod-perceiver:latest

registry_check:
ifndef REGISTRY
	echo "Must set REGISTRY to create containers"
	exit 1
endif

test: $(SUBDIRS)
	go test ./pkg/...

${OUTDIR}:
	mkdir -p ${OUTDIR}

clean: $(SUBDIRS)
	rm -rf _output
