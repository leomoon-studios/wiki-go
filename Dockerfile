# Build stage
FROM docker.io/library/golang:1.24-alpine3.21 AS builder

RUN apk add --no-cache build-base git gcc musl-dev && rm -rf /var/cache/apk/*

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags="-s -w" -o wiki-go .

# Final stage
FROM docker.io/library/alpine:3.21

RUN apk add --no-cache bash ca-certificates curl linux-pam tzdata && rm -rf /var/cache/apk/*

ARG PUID=1000
ARG PGID=1000

RUN addgroup -g ${PGID} appgroup && adduser -u ${PUID} -G appgroup -s /bin/bash -D appuser

WORKDIR /wiki

COPY --from=builder /app/wiki-go .

RUN chown -R appuser:appgroup /wiki && chmod -R 755 /wiki

USER appuser

EXPOSE 8080

VOLUME ["/wiki/data"]

ENTRYPOINT ["./wiki-go"]

CMD []
