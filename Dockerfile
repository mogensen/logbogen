# Where the base images come from. CI sets a mirror, since Docker Hub rate limits anonymous pulls on shared runners.
ARG BASE_REGISTRY=docker.io/library

# mattn/go-sqlite3 needs CGO, so build on a glibc image that ships gcc
FROM ${BASE_REGISTRY}/golang:1.25-bookworm AS builder
WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=1 GOOS=linux go build -trimpath -o /out/logbogen .

FROM ${BASE_REGISTRY}/debian:bookworm-slim
RUN apt-get update \
    && apt-get install -y --no-install-recommends ca-certificates \
    && rm -rf /var/lib/apt/lists/* \
    && useradd --system --uid 10001 --home-dir /app logbogen
WORKDIR /app

COPY --from=builder /out/logbogen ./logbogen
COPY views ./views
COPY assets ./assets

# The database and sessions live in ./data, mount a volume here
RUN mkdir -p /app/data && chown logbogen /app/data
VOLUME /app/data

ENV LISTEN_ADDR=0.0.0.0:3000
EXPOSE 3000
USER logbogen
ENTRYPOINT ["./logbogen"]
