ARG fluent_bit_version=1.8.6
ARG golang_version=1.16.7

FROM golang:${golang_version} as builder
WORKDIR /build
COPY . .
RUN CGO_ENABLED=1 go build -buildmode=c-shared -o /yc-logging.so .

FROM fluent/fluent-bit:${fluent_bit_version} as fluent-bit
COPY --from=builder /yc-logging.so /fluent-bit/bin/
ENTRYPOINT ["/fluent-bit/bin/fluent-bit", "-e", "/fluent-bit/bin/yc-logging.so"]
CMD ["-c", "/fluent-bit/etc/fluent-bit.conf"]
