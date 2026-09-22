# syntax=docker/dockerfile:1

FROM golang:1.26-alpine AS build
WORKDIR /src

RUN apk add --no-cache git ca-certificates

COPY go.mod go.sum ./
RUN go mod download

COPY . .

ARG VERSION=dev
ARG COMMIT=none
ARG DATE=unknown
RUN CGO_ENABLED=0 go build \
    -trimpath \
    -ldflags "-s -w \
      -X github.com/Blizoman/netwatch/internal/version.Version=${VERSION} \
      -X github.com/Blizoman/netwatch/internal/version.Commit=${COMMIT} \
      -X github.com/Blizoman/netwatch/internal/version.Date=${DATE}" \
    -o /out/netwatch ./cmd/netwatch

FROM alpine:3.20
RUN apk add --no-cache ca-certificates
COPY --from=build /out/netwatch /usr/local/bin/netwatch

# ICMP ping/traceroute need CAP_NET_RAW. Grant it at `docker run` time with
# `--cap-add=NET_RAW`, or run the container as root (the default here).
ENTRYPOINT ["netwatch"]
CMD ["--help"]
