ifndef REGISTRY
REGISTRY=gcr.io/gke-verification
endif

ifdef IMAGE_PREFIX
PREFIX="$(IMAGE_PREFIX)-"
endif

ifneq (, $(findstring gcr.io,$(REGISTRY))) 
PREFIX_CMD="gcloud"
DOCKER_OPTS="--"
endif

OUTDIR=_output
BINARY=pod-perceiver

CURRENT_DIR:=$(shell dirname $(realpath $(lastword $(MAKEFILE_LIST))))

.PHONY: clean test

all: build

build: ${OUTDIR} cmd/pod-perceiver.go
	docker run --rm -e CGO_ENABLED=0 -e GOOS=linux -e GOARCH=amd64 -v "${CURRENT_DIR}/..":/go/src/github.com/blackducksoftware/perceivers -w /go/src/github.com/blackducksoftware/perceivers/pod/cmd golang go build -o ../${OUTDIR}/${BINARY} pod-perceiver.go	

local_build: ${OUTDIR} cmd/pod-perceiver.go
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o ${OUTDIR}/${BINARY} cmd/pod-perceiver.go

container: ${OUTDIR}/${BINARY}
	docker build -t $(REGISTRY)/$(PREFIX)${BINARY} .

push: container
	$(PREFIX_CMD) docker $(DOCKER_OPTS) push $(REGISTRY)/$(PREFIX)${BINARY}:latest

test:
	go test ./pkg/...

clean:
	rm -rf ${OUTDIR}

${OUTDIR}:
	mkdir -p ${OUTDIR}
