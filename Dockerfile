# syntax=docker/dockerfile:1
ARG TARGETOS
ARG TARGETARCH
ARG TARGETVARIANT

FROM golang:1.25.1-bookworm AS builder
WORKDIR /src

ARG VERSION=""
ARG COMMIT=""
ARG BUILD_DATE=""
ARG TARGETOS
ARG TARGETARCH
ARG TARGETVARIANT

ENV CGO_ENABLED=0

COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Cache Go build and module downloads between builds
RUN --mount=type=cache,target=/root/.cache/go-build \
    --mount=type=cache,target=/go/pkg/mod \
    CGO_ENABLED=0 GOOS=${TARGETOS:-linux} GOARCH=${TARGETARCH:-amd64} \
    go build -trimpath -ldflags "-s -w -X main.Version=${VERSION} -X main.Commit=${COMMIT} -X main.BuildDate=${BUILD_DATE}" -o /app ./

FROM scratch

# re-declare build args in final stage so they're available for LABELs
ARG VERSION=""
ARG COMMIT=""
ARG BUILD_DATE=""

LABEL org.opencontainers.image.source="https://github.com/ericogr/ads1115-to-mqtt"
LABEL org.opencontainers.image.title="ads1115-to-mqtt"
LABEL org.opencontainers.image.version="${VERSION}"

COPY --from=builder /app /app

ENTRYPOINT ["/app"]
