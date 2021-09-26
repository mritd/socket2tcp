BUILD_VERSION   := $(shell cat version)
BUILD_DATE      := $(shell date "+%F %T")
COMMIT_SHA1     := $(shell git rev-parse HEAD)

all: clean
	bash .cross_compile.sh

release: clean all
	ghr -u mritd -t ${GITHUB_TOKEN} -replace -recreate --debug ${BUILD_VERSION} dist

clean:
	rm -rf dist

install:
	go install -ldflags	"-X 'main.Version=${BUILD_VERSION}' \
               			-X 'main.BuildDate=${BUILD_DATE}' \
               			-X 'main.CommitID=${COMMIT_SHA1}'"

docker:
	cat Dockerfile | docker build -t mritd/socket2tcp:${BUILD_VERSION} -f - .

.PHONY : all release clean install docker
