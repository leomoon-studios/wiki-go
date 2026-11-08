# Build stage
FROM docker.io/library/golang:1.27.3-alpine3.24 AS builder

RUN apk add --no-cache build-base git gcc musl-dev && rm -rf /var/cache/apk/*

WORKDIR /app

COPY go.mod go.sum vendor/ ./

COPY . .

ARG VERSION=dev
RUN CGO_ENABLED=0 GOOS=linux go build -mod=vendor -a -installsuffix cgo -ldflags="-s -w -X 'wiki-go/internal/version.Version=${VERSION}'" -o wiki-go .

# Final stage
FROM docker.io/library/alpine:3.24

RUN apk add --no-cache bash ca-certificates curl linux-pam tzdata && rm -rf /var/cache/apk/*

ARG PUID=1000
ARG PGID=1000

RUN addgroup -g ${PGID} appgroup && adduser -u ${PUID} -G appgroup -s /bin/bash -D appuser

WORKDIR /wiki

RUN mkdir -p /wiki/data && chown -R appuser:appgroup /wiki && chmod -R 755 /wiki

COPY --chown=appuser:appgroup --chmod=755 --from=builder /app/wiki-go .

USER appuser

EXPOSE 8080 443

VOLUME ["/wiki/data"]

ENTRYPOINT ["./wiki-go"]

CMD []
