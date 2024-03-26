# Build the manager binary
FROM golang:1.19 AS builder

WORKDIR /workspace
# Copy the Go Modules manifests
COPY go.mod go.mod
COPY go.sum go.sum
# cache deps before building and copying source so that we don't need to re-download as much
# and so that source changes don't invalidate our downloaded layer

# It's a proxy for CN developer, please unblock it if you have network issue
#ARG GOPROXY
#ENV GOPROXY=${GOPROXY:-https://goproxy.cn,direct}

RUN go mod download

# Copy the go source
COPY api/ api/
COPY cmd/ cmd/
COPY pkg/ pkg/
COPY version/ version/

# Build
ARG TARGETARCH
ARG VERSION
ARG GITCOMMIT
RUN CGO_ENABLED=0 GOOS=linux GOARCH=${TARGETARCH} \
    go build -a -ldflags "-s -w -X kdp-oam-operator/version.CoreVersion=${VERSION:-undefined} -X kdp-oam-operator/version.GitRevision=${GITCOMMIT:-undefined}" \
        -o apiserver-${TARGETARCH} cmd/apiserver/main.go

FROM ${BASE_IMAGE:-alpine:3.15}
RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g' /etc/apk/repositories && \
    apk add --no-cache ca-certificates tzdata bash expat && \
    rm -rf /var/cache/apk/*

ENV TZ=${TZ:-Asia/Shanghai}
RUN cp /usr/share/zoneinfo/${TZ} /etc/localtime
RUN echo ${TZ} > /etc/timezone
WORKDIR /

ARG TARGETARCH

COPY --from=builder /workspace/apiserver-${TARGETARCH} /usr/local/bin/apiserver

COPY docs/ docs/
USER 5555

CMD ["apiserver"]
