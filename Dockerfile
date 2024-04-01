# Build the manager binary
FROM golang:1.19 as builder
ARG TARGETOS
ARG TARGETARCH
ARG VERSION
ARG GITVERSION

WORKDIR /workspace
# Copy the Go Modules manifests
COPY go.mod go.mod
COPY go.sum go.sum
# cache deps before building and copying source so that we don't need to re-download as much
# and so that source changes don't invalidate our downloaded layer

# It's a proxy for CN developer, please unblock it if you have network issue
#ARG GOPROXY
#ENV GOPROXY=${GOPROXY:-https://goproxy.cn}

RUN go mod download

# Copy the go source
COPY api/ api/
COPY cmd/ cmd/
COPY pkg/ pkg/
COPY version/ version/

# Build
# the GOARCH has not a default value to allow the binary be built according to the host where the command
# was called. For example, if we call make docker-build in a local env which has the Apple Silicon M1 SO
# the docker BUILDPLATFORM arg will be linux/arm64 when for Apple x86 it will be linux/amd64. Therefore,
# by leaving it empty we can ensure that the container and binary shipped on it will have the same platform.
RUN CGO_ENABLED=0 GOOS=${TARGETOS:-linux} GOARCH=${TARGETARCH} go build  -a -ldflags "-s -w -X kdp-oam-operator/version.CoreVersion=${VERSION:-undefined} -X kdp-oam-operator/version.GitRevision=${GITVERSION:-undefined}" -o manager cmd/bdc/main.go

# Use distroless as minimal base image to package the manager binary
# Refer to https://github.com/GoogleContainerTools/distroless for more details
FROM ${BASE_IMAGE:-alpine:3.15}
RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g' /etc/apk/repositories && \
    apk add --no-cache ca-certificates tzdata bash expat && \
    rm -rf /var/cache/apk/*

ENV TZ=${TZ:-Asia/Shanghai}
RUN cp /usr/share/zoneinfo/${TZ} /etc/localtime
RUN echo ${TZ} > /etc/timezone
WORKDIR /

COPY --from=builder /workspace/manager .
USER 65532:65532

ENTRYPOINT ["/manager"]
