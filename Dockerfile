ARG plugin_version=dev
ARG fluent_bit_version=1.9.0
ARG golang_version=1.17.8

FROM golang:${golang_version} as builder
WORKDIR /build
COPY . .
RUN CGO_ENABLED=1 go build \
    -buildmode=c-shared \
    -o /yc-logging.so \
    -ldflags "-X main.PluginVersion=${plugin_version}" \
    -ldflags "-X main.FluentBitVersion=${fluent_bit_version}" \
    .

FROM fluent/fluent-bit:${fluent_bit_version} as fluent-bit
COPY --from=builder /yc-logging.so /fluent-bit/bin/
ENTRYPOINT ["/fluent-bit/bin/fluent-bit", "-e", "/fluent-bit/bin/yc-logging.so"]
CMD ["-c", "/fluent-bit/etc/fluent-bit.conf"]
