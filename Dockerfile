# syntax=docker/dockerfile:1

FROM --platform=$BUILDPLATFORM golang:1.22-alpine AS build
ARG TARGETOS=linux
ARG TARGETARCH=amd64
WORKDIR /src
COPY go.mod ./
COPY *.go ./
COPY web ./web
RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} \
    go build -trimpath -buildvcs=false \
    -ldflags "-s -w" \
    -o /out/TwitchLiveNotifier .

FROM alpine:3.20
ARG VERSION=dev

LABEL org.opencontainers.image.title="Twitch Live to Discord" \
      org.opencontainers.image.description="Cross-platform Twitch live alerts for labelled Discord webhooks" \
      org.opencontainers.image.source="https://github.com/ausbitbank/twitch-live-notifier" \
      org.opencontainers.image.licenses="MIT" \
      org.opencontainers.image.version="${VERSION}"

RUN apk add --no-cache ca-certificates tzdata \
    && addgroup -S app \
    && adduser -S -G app -h /data app \
    && mkdir -p /data \
    && chown app:app /data

COPY --from=build /out/TwitchLiveNotifier /app/TwitchLiveNotifier

USER app
WORKDIR /data
VOLUME ["/data"]
EXPOSE 8787

HEALTHCHECK --interval=30s --timeout=6s --start-period=20s --retries=3 \
    CMD ["/app/TwitchLiveNotifier", "--healthcheck", "--healthcheck-url", "http://127.0.0.1:8787/healthz"]

ENTRYPOINT ["/app/TwitchLiveNotifier"]
CMD ["--serve", "--listen", "0.0.0.0:8787", "--config", "/data/config.json"]
