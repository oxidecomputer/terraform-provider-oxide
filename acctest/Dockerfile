# Stage 1: Build omicron-dev and download dependencies
FROM rust:bookworm AS builder

ARG OMICRON_BRANCH="main"

# Download omicron source - works with branch names, tags, or commit SHAs
RUN curl -sL "https://github.com/oxidecomputer/omicron/archive/${OMICRON_BRANCH}.tar.gz" | tar xz && \
    mv omicron-* omicron

WORKDIR /omicron

# Install builder prerequisites (downloads cockroach, clickhouse, dpd, mgd)
RUN bash -c "source env.sh && ./tools/install_builder_prerequisites.sh -y -s"

# Build omicron-dev binary (debug build for faster compilation)
RUN bash -c "source env.sh && cargo build -p omicron-dev"

# Stage 2: Minimal runtime image
FROM debian:bookworm-slim

# Install runtime dependencies (from install_runner_prerequisites.sh)
RUN apt-get update && apt-get install -y \
    ca-certificates \
    libpq5 \
    libsqlite3-0 \
    libssl3 \
    libxmlsec1-openssl \
    curl \
    && rm -rf /var/lib/apt/lists/*

# Copy omicron-dev binary
COPY --from=builder /omicron/target/debug/omicron-dev /usr/local/bin/

# Copy required out/ subdirectories (dpd needs relative paths preserved)
# Excludes: downloads/ (source tarballs), console-assets/ (web UI), transceiver-control/
COPY --from=builder /omicron/out/clickhouse /omicron/out/clickhouse
COPY --from=builder /omicron/out/cockroachdb /omicron/out/cockroachdb
COPY --from=builder /omicron/out/dendrite-stub /omicron/out/dendrite-stub
COPY --from=builder /omicron/out/mgd /omicron/out/mgd

# Copy config files to original locations (omicron-dev expects these paths)
COPY --from=builder /omicron/nexus/examples /omicron/nexus/examples
COPY --from=builder /omicron/gateway-test-utils/configs /omicron/gateway-test-utils/configs
COPY --from=builder /omicron/smf /omicron/smf

# Patch config to bind to all interfaces (for Docker networking)
RUN sed -i 's/bind_address = "127.0.0.1:12220"/bind_address = "0.0.0.0:12220"/' \
    /omicron/nexus/examples/config.toml

# Create directories needed for path resolution
RUN mkdir -p /omicron/test-utils /omicron/nexus/test-utils

# Add downloaded binaries to PATH
ENV PATH="/omicron/out/cockroachdb/bin:/omicron/out/clickhouse:/omicron/out/dendrite-stub/bin:/omicron/out/mgd/root/opt/oxide/mgd/bin:${PATH}"

WORKDIR /omicron

CMD ["omicron-dev", "run-all"]
