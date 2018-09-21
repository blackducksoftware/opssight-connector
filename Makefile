BINARY = $(shell ls cmd)

ifdef IMAGE_PREFIX
PREFIX="$(IMAGE_PREFIX)-"
endif

TAG="latest"
ifdef IMAGE_TAG
TAG="$(IMAGE_TAG)"
endif


ifneq (, $(findstring gcr.io,$(REGISTRY)))
PREFIX_CMD="gcloud"
DOCKER_OPTS="--"
endif

OUTDIR=_output
BUILDDIR=build
LOCAL_TARGET=local
FEDERATOR:=federator
OPSSIGHT_CORE:=opssight-core

CURRENT_DIR:=$(shell dirname $(realpath $(lastword $(MAKEFILE_LIST))))

BUILD_TIME:=$(shell date)

LAST_COMMIT=$(shell git rev-parse HEAD)

.PHONY: all clean test push test ${BINARY} container local

all: build

build: ${OUTDIR} $(BINARY)

${LOCAL_TARGET}: ${OUTDIR} $(BINARY)

$(BINARY):
ifeq ($(MAKECMDGOALS),${LOCAL_TARGET})
	cd cmd/$@; CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o $@
else
	docker run --rm -e CGO_ENABLED=0 -e GOOS=linux -e GOARCH=amd64 -v "${CURRENT_DIR}":/go/src/github.com/blackducksoftware/opssight-connector -w /go/src/github.com/blackducksoftware/opssight-connector/cmd/$@ golang:1.9 go build -o $@
endif
	mv cmd/$@/$@ ${OUTDIR}

container: registry_check container_prep
	$(foreach p,${BINARY}, \
		if [[ $(p) != $(FEDERATOR) ]]; then \
			cd ${CURRENT_DIR}/${BUILDDIR}/$p; \
			docker build . -t $(REGISTRY)/$(PREFIX)${p}:$(TAG) --build-arg VERSION=$(TAG) --build-arg 'BUILDTIME=$(BUILD_TIME)' --build-arg LASTCOMMIT=$(LAST_COMMIT);\
		fi; \
	 )

container_prep: ${OUTDIR} $(BINARY)
	$(foreach p,${BINARY},\
		if [[ $(p) = $(FEDERATOR) ]]; then \
			mkdir -p ${CURRENT_DIR}/${BUILDDIR}/opssight-core; \
			cp ${CURRENT_DIR}/cmd/$p/* ${OUTDIR}/$p ${CURRENT_DIR}/${BUILDDIR}/opssight-core; \
		else \
			if [[ $(p) != $(OPSSIGHT_CORE) ]]; then \
				mkdir -p ${CURRENT_DIR}/${BUILDDIR}/$p; \
			fi; \
			cp ${CURRENT_DIR}/cmd/$p/* LICENSE ${OUTDIR}/$p ${CURRENT_DIR}/${BUILDDIR}/$p; \
		fi; \
	 )

push: container
	$(foreach p,${BINARY},$(PREFIX_CMD) docker $(DOCKER_OPTS) push $(REGISTRY)/$(PREFIX)${p}:$(TAG);)

test:
	docker run --rm -e CGO_ENABLED=0 -e GOOS=linux -e GOARCH=amd64 -v "${CURRENT_DIR}":/go/src/github.com/blackducksoftware/opssight-connector -w /go/src/github.com/blackducksoftware/opssight-connector golang:1.9 go test ./pkg/...

clean:
	rm -rf ${OUTDIR} ${BUILDDIR}
	$(foreach p,${BINARY},rm -f cmd/$p/$p;)

${OUTDIR}:
	mkdir -p ${OUTDIR}

registry_check:
ifndef REGISTRY
	echo "Must set REGISTRY to create containers"
	exit 1
endif
