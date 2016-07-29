BUILDID=$(shell git rev-parse HEAD)

.PHONY: all
all: jolokiabeat/deb jolokiabeat/rpm jolokiabeat/darwin jolokiabeat/win jolokiabeat/bin \
	build/upload/build_id.txt

.PHONY: jolokiabeat
jolokiabeat: build/upload
	# cross compile on ubuntu
	docker run --rm \
		-v $(abspath build):/build \
		-v $(abspath xgo-scripts):/scripts \
		-v $(abspath ../..):/source \
		-e PACK=$@ \
		-e BEFORE_BUILD=before_build.sh \
		-e SOURCE=/source \
		-e BUILDID=${BUILDID} \
		tudorg/beats-builder \
		github.com/neonmori/jolokiabeat

include ./scripts/Makefile
