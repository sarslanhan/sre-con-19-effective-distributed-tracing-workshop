FROM golang:alpine as builder

ENV GIN_MODE=release

RUN apk add --no-cache \
            build-base \
            git \
            ca-certificates
ADD . /app
WORKDIR /app

RUN make build

FROM alpine:latest
LABEL maintaniner="sre@zalando.de"

ENV GIN_MODE=release

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /app/ot101 /
ADD ./templates /templates

EXPOSE 8085 8087

ENTRYPOINT ["/ot101"]
