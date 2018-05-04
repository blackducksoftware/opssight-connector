ifdef IMAGE_PREFIX
PREFIX="$(IMAGE_PREFIX)-"
endif

ifneq (, $(findstring gcr.io,$(REGISTRY)))
PREFIX_CMD="gcloud"
DOCKER_OPTS="--"
endif

OUTDIR=_output

CURRENT_DIR:=$(shell dirname $(realpath $(lastword $(MAKEFILE_LIST))))
OUTDIR=_output

.PHONY: $(SUBDIRS) clean build local_build container push test

all: build

build: $(OUTDIR)
	docker run --rm -e CGO_ENABLED=0 -e GOOS=linux -e GOARCH=amd64 -v "${CURRENT_DIR}/..":/go/src/github.com/blackducksoftware/opssight-connector -w /go/src/github.com/blackducksoftware/opssight-connector/processors golang go build -o ${OUTDIR}/image-processor cmd/image-processor/image-processor.go
	docker run --rm -e CGO_ENABLED=0 -e GOOS=linux -e GOARCH=amd64 -v "${CURRENT_DIR}/..":/go/src/github.com/blackducksoftware/opssight-connector -w /go/src/github.com/blackducksoftware/opssight-connector/processors golang go build -o ${OUTDIR}/pod-processor cmd/pod-processor/pod-processor.go

local_build: $(OUTDIR) cmd/image-processor/image-processor.go cmd/pod-processor/pod-processor.go
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o ${OUTDIR}/image-processor cmd/image-processor/image-processor.go
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o ${OUTDIR}/pod-processor cmd/pod-processor/pod-processor.go

container: registry_check $(OUTDIR)/image-processor ${OUTDIR}/pod-processor
	docker build -t $(REGISTRY)/$(PREFIX)opssight-image-processor -f Dockerfile.image .
	docker build -t $(REGISTRY)/$(PREFIX)opssight-pod-processor -f Dockerfile.pod .

push: registry_check container
	$(PREFIX_CMD) docker $(DOCKER_OPTS) push $(REGISTRY)/$(PREFIX)opssight-image-processor:latest
	$(PREFIX_CMD) docker $(DOCKER_OPTS) push $(REGISTRY)/$(PREFIX)opssight-pod-processor:latest

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
