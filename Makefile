BEATNAME=jolokiabeat
BEAT_DIR=github.com/neonmori
SYSTEM_TESTS=false
TEST_ENVIRONMENT=false
# ES_BEATS?=${GOPATH}/src/github.com/elastic/beats
ES_BEATS=./vendor/github.com/elastic/beats
GOPACKAGES=$(shell glide novendor)
PREFIX?=.

# ARCH=amd64 BEAT=jolokiabeat BUILDID=$(shell git rev-parse HEAD) SNAPSHOT=yes ./vendor/github.com/elastic/beats/dev-tools/packer/platforms/centos/build.sh

# Path to the libbeat Makefile
-include $(ES_BEATS)/libbeat/scripts/Makefile

# Initial beat setup
.PHONY: setup
setup:
	make update

.PHONY: git-init
git-init:
	git init
	git add README.md CONTRIBUTING.md
	git commit -m "Initial commit"
	git add LICENSE
	git commit -m "Add the LICENSE"
	git add .gitignore
	git commit -m "Add git settings"
	git add .
	git reset -- .travis.yml
	git commit -m "Add jolokiabeat"
	git add .travis.yml
	git commit -m "Add Travis CI"

# This is called by the beats packer before building starts
.PHONY: before-build
before-build:
