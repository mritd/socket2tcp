FROM golang:1-alpine AS builder

COPY . /go/src/github.com/mritd/socket2tcp

WORKDIR /go/src/github.com/mritd/socket2tcp

RUN set -ex \
    && apk add git \
    && BUILD_VERSION=`cat version` \
    && BUILD_DATE=`date "+%F %T"` \
    && COMMIT_SHA1=`git rev-parse HEAD` \
    && go install -ldflags  "-X 'main.Version=${BUILD_VERSION}' \
                             -X 'main.BuildDate=${BUILD_DATE}' \
                             -X 'main.CommitID=${COMMIT_SHA1}'"

FROM alpine

LABEL maintainer="mritd <mritd@linux.com>"

ENV TZ Asia/Shanghai

RUN apk add tzdata --no-cache \
    && ln -sf /usr/share/zoneinfo/${TZ} /etc/localtime \
    && echo ${TZ} > /etc/timezone \

# set up nsswitch.conf for Go's "netgo" implementation
# - https://github.com/golang/go/blob/go1.9.1/src/net/conf.go#L194-L275
# - docker run --rm debian:stretch grep '^hosts:' /etc/nsswitch.conf
RUN [ ! -e /etc/nsswitch.conf ] && echo 'hosts: files dns' > /etc/nsswitch.conf

RUN set -ex \
    && apk upgrade \
    && apk add bash tzdata ca-certificates \
    && rm -rf /var/cache/apk/*

COPY --from=builder /go/bin/socket2tcp /usr/local/bin/socket2tcp

ENTRYPOINT ["socket2tcp"]
