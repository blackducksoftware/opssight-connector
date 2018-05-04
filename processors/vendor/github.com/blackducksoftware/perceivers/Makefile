SUBDIRS = pod image

.PHONY: $(SUBDIRS) clean build local_build container push test

all: build

build: $(SUBDIRS)

local_build: $(SUBDIRS)

container: $(SUBDIRS)

push: $(SUBDIRS)

test: $(SUBDIRS)
	go test ./pkg/...

$(SUBDIRS):
	$(MAKE) -C $@ $(MAKECMDGOALS)

clean: $(SUBDIRS)
	rm -rf _output
