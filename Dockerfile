FROM golang:1.13.8-alpine3.11 AS builder

COPY . /go/src/github.com/mritd/socket2tcp

WORKDIR /go/src/github.com/mritd/socket2tcp

ENV GO111MODULE on
ENV GOPROXY https://goproxy.io
ENV GOSUMDB sum.golang.google.cn

RUN set -ex \
    && apk add git \
    && BUILD_VERSION=`cat version` \
    && BUILD_DATE=`date "+%F %T"` \
    && COMMIT_SHA1=`git rev-parse HEAD` \
    && go install -ldflags  "-X 'main.Version=${BUILD_VERSION}' \
                             -X 'main.BuildDate=${BUILD_DATE}' \
                             -X 'main.CommitID=${COMMIT_SHA1}'"

FROM alpine:3.11

LABEL maintainer="mritd <mritd@linux.com>"

RUN set -ex \
    && apk upgrade \
    && apk add bash tzdata ca-certificates \
    && rm -rf /var/cache/apk/*

COPY --from=builder /go/bin/socket2tcp /usr/local/bin/socket2tcp

ENTRYPOINT ["socket2tcp"]
