BUILD_VERSION   := $(shell cat version)
BUILD_DATE      := $(shell date "+%F %T")
COMMIT_SHA1     := $(shell git rev-parse HEAD)

all: clean
	gox -osarch="darwin/amd64 linux/386 linux/amd64 linux/arm windows/amd64 windows/386" \
		-output="dist/{{.Dir}}_{{.OS}}_{{.Arch}}" \
		-ldflags	"-X 'main.Version=${BUILD_VERSION}' \
					-X 'main.BuildDate=${BUILD_DATE}' \
					-X 'main.CommitID=${COMMIT_SHA1}'"

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

.EXPORT_ALL_VARIABLES:

GO111MODULE = on
GOPROXY = https://goproxy.io
GOSUMDB = sum.golang.google.cn
