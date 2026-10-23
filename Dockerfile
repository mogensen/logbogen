# Where the base images come from. CI sets a mirror, since Docker Hub rate limits anonymous pulls on shared runners.
ARG BASE_REGISTRY=docker.io/library

# mattn/go-sqlite3 needs CGO, so build on a glibc image that ships gcc and link the binary statically
FROM ${BASE_REGISTRY}/golang:1.27-bookworm AS builder
WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=1 GOOS=linux go build -trimpath -tags netgo,osusergo \
    -ldflags '-linkmode external -extldflags "-static"' -o /out/logbogen .

# Distroless has no shell, so the data directory is prepared here and copied over with its owner
RUN mkdir -p /out/data

# The static image ships CA certificates and tzdata but no libc, shell or package manager
FROM gcr.io/distroless/static-debian12
WORKDIR /app

COPY --from=builder /out/logbogen ./logbogen
COPY views ./views
COPY assets ./assets

# The database and sessions live in ./data, mount a volume here
COPY --from=builder --chown=10001:10001 /out/data ./data
VOLUME /app/data

ENV LISTEN_ADDR=0.0.0.0:3000
EXPOSE 3000
USER 10001:10001
ENTRYPOINT ["./logbogen"]
