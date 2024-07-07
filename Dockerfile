FROM golang:1.22.3 as build-env

ADD . /app
WORKDIR /app

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -mod vendor

# -----------------------------------------------------------------------------

FROM alpine:3.19.0

RUN addgroup -S -g 10000 server \
 && adduser -S -D -u 10000 -s /sbin/nologin -G server server

RUN mkdir /app
RUN chown -R 10000:10000 /app

USER 10000

COPY --from=build-env /app/server /app/server

ENTRYPOINT ["/app/server"]
CMD ["server", "--config", "/app/config/config.yaml"]
