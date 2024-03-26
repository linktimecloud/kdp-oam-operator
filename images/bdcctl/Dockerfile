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
COPY pkg/ pkg/
COPY version/ version/
COPY reference/ reference/

# Build
RUN CGO_ENABLED=0 GOOS=${TARGETOS:-linux} GOARCH=${TARGETARCH} go build -a -ldflags "-s -w -X kdp-oam-operator/version.CoreVersion=${VERSION:-undefined} -X kdp-oam-operator/version.GitRevision=${GITVERSION:-undefined}" -o bdcctl ./reference/cmd/cli/main.go


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

COPY --from=builder /workspace/bdcctl .

RUN chmod 755 ./bdcctl && \
    mv ./bdcctl /usr/bin/bdcctl
