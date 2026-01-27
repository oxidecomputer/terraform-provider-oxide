# Builder image
FROM rust:bookworm AS builder

ARG OMICRON_SHA

RUN curl -sL "https://github.com/oxidecomputer/omicron/archive/${OMICRON_SHA}.tar.gz" | tar xz && \
    mv omicron-* omicron

WORKDIR /omicron

RUN bash -c "source env.sh && ./tools/install_builder_prerequisites.sh -y -s"

RUN bash -c "source env.sh && cargo build -p omicron-dev"

# Runtime image
FROM debian:bookworm-slim

COPY --from=builder /omicron/tools/install_runner_prerequisites.sh /tmp/
RUN apt-get update && \
    /tmp/install_runner_prerequisites.sh -y -s && \
    apt-get install -y curl && \
    rm -rf /var/lib/apt/lists/* /tmp/install_runner_prerequisites.sh

# Copy binaries and config files from the builder image.
COPY --from=builder /omicron/target/debug/omicron-dev /usr/local/bin/
COPY --from=builder /omicron/out/clickhouse /omicron/out/clickhouse
COPY --from=builder /omicron/out/cockroachdb /omicron/out/cockroachdb
COPY --from=builder /omicron/out/dendrite-stub /omicron/out/dendrite-stub
COPY --from=builder /omicron/out/mgd /omicron/out/mgd

COPY --from=builder /omicron/nexus/examples /omicron/nexus/examples
COPY --from=builder /omicron/gateway-test-utils/configs /omicron/gateway-test-utils/configs
COPY --from=builder /omicron/smf /omicron/smf

RUN sed -i 's/bind_address = "127.0.0.1:12220"/bind_address = "0.0.0.0:12220"/' \
    /omicron/nexus/examples/config.toml

# Add test paths required by omicron-dev.
RUN mkdir -p /omicron/test-utils /omicron/nexus/test-utils

ENV PATH="/omicron/out/cockroachdb/bin:/omicron/out/clickhouse:/omicron/out/dendrite-stub/bin:/omicron/out/mgd/root/opt/oxide/mgd/bin:${PATH}"

WORKDIR /omicron

CMD ["omicron-dev", "run-all"]
