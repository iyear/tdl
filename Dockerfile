# https://www.docker.com/blog/faster-multi-platform-builds-dockerfile-cross-compilation-guide/
FROM --platform=$BUILDPLATFORM golang:1.21-alpine AS builder

ARG VERSION="dev"
ARG COMMIT="unknown"
ARG COMMIT_DATE="unknown"

WORKDIR /

COPY . .

ARG TARGETOS
ARG TARGETARCH

RUN --mount=type=cache,target=/root/.cache/go-build \
    --mount=type=cache,target=/go/pkg \
    GOOS=$TARGETOS GOARCH=$TARGETARCH \
    go build -trimpath \
    -ldflags "-s -w  \
    -X github.com/iyear/tdl/pkg/consts.Version=${VERSION}  \
    -X github.com/iyear/tdl/pkg/consts.Commit=${COMMIT}  \
    -X github.com/iyear/tdl/pkg/consts.CommitDate=${COMMIT_DATE}" \
    -o tdl

FROM alpine:latest

RUN apk add --no-cache ca-certificates

COPY --from=builder /tdl /usr/bin/tdl

ENTRYPOINT ["tdl"]
