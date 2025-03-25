FROM golang:1.24 AS builder

RUN mkdir /src
WORKDIR /src

ADD . .
ARG VERSION
RUN \
	--mount=type=cache,target=/go/pkg/mod \
	--mount=type=cache,target=/root/.cache/go-build \
CGO_ENABLED=0 go build -ldflags="-X main.Version=$VERSION" -v -o /registry ./cmd/registry/

RUN rm -rf /tmp/*

FROM scratch

COPY --from=builder /registry /registry
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo
COPY --from=builder /tmp /tmp
COPY --from=builder /src/ext/db /migrations
ENV MIGRATIONS_DIR=/migrations

CMD ["serve"]
ENTRYPOINT ["/registry"]
