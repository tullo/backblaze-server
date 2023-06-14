FROM golang:1.17.3-alpine3.13 AS go-builder
ENV CGO_ENABLED 0
WORKDIR /build
COPY . .
WORKDIR /build/app/backblaze-server
RUN go build -mod=vendor -o server

FROM alpine:3.18.2
RUN apk --no-cache add ca-certificates
RUN addgroup -g 3000 -S app && adduser -u 100000 -S app -G app --no-create-home --disabled-password \
    && mkdir -p /app/badger.db && chown app:app /app/badger.db
USER 100000
WORKDIR /app
COPY --from=go-builder --chown=app:app /build/app/backblaze-server/server /app/server
ENTRYPOINT ["./server"]
