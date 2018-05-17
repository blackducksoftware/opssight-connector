BINARY = $(shell ls cmd)

ifdef IMAGE_PREFIX
PREFIX="$(IMAGE_PREFIX)-"
endif

ifneq (, $(findstring gcr.io,$(REGISTRY))) 
PREFIX_CMD="gcloud"
DOCKER_OPTS="--"
endif

OUTDIR=_output
LOCAL_TARGET=local

CURRENT_DIR:=$(shell dirname $(realpath $(lastword $(MAKEFILE_LIST))))

.PHONY: all clean test push test ${BINARY} container local

all: build

build: ${OUTDIR} $(BINARY)

${LOCAL_TARGET}: ${OUTDIR} $(BINARY)

$(BINARY):
ifeq ($(MAKECMDGOALS),${LOCAL_TARGET})
	cd cmd/$@; CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o $@
else
	docker run --rm -e CGO_ENABLED=0 -e GOOS=linux -e GOARCH=amd64 -v "${CURRENT_DIR}":/go/src/github.com/blackducksoftware/opssight-connector/scanner -w /go/src/github.com/blackducksoftware/opssight-connector/scanner/cmd/$@ golang:1.9 go build -o $@
endif
	cp cmd/$@/$@ ${OUTDIR}

container: $(BINARY)
	$(foreach p,${BINARY},cd ${CURRENT_DIR}/cmd/$p; docker build -t $(REGISTRY)/$(PREFIX)${p} .;)

push: container
	$(foreach p,${BINARY},$(PREFIX_CMD) docker $(DOCKER_OPTS) push $(REGISTRY)/$(PREFIX)${p}:latest;)

test:
	docker run --rm -e CGO_ENABLED=0 -e GOOS=linux -e GOARCH=amd64 -v "${CURRENT_DIR}":/go/src/github.com/blackducksoftware/perceivers -w /go/src/github.com/blackducksoftware/ossight-connector/scanner golang:1.9 go test ./pkg/...

clean:
	rm -rf ${OUTDIR}
	$(foreach p,${BINARY},rm -f cmd/$p/$p;)

${OUTDIR}:
	mkdir -p ${OUTDIR}
