CURRENT_DIR:=$(shell dirname $(realpath $(lastword $(MAKEFILE_LIST))))
OUTDIR=_output

.PHONY: test ${OUTDIR}

all: compile

compile: ${OUTDIR}
	docker run -t -i --rm -v "${CURRENT_DIR}/..":/go/src/github.com/blackducksoftware/opssight-connector -w /go/src/github.com/blackducksoftware/opssight-connector/deployer -e CGO_ENABLED=0 -e GOOS=linux -e GOARCH=amd64 golang:1.9 go build -o ${OUTDIR}/opssight-deployer ./cmd/opssight-deployer.go

test:
	go test ./...

clean:
	rm -rf ${OUTDIR}

${OUTDIR}:
	mkdir -p ${OUTDIR}
