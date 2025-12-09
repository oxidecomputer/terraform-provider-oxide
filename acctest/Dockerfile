FROM rust

ARG OMICRON_BRANCH="main"

RUN /bin/bash -c "git clone https://github.com/oxidecomputer/omicron.git --branch '${OMICRON_BRANCH}' --depth 1 && \
  cd omicron && \
  source env.sh && \
  ./tools/install_builder_prerequisites.sh -y -s"

RUN sed -i 's/bind_address = "127.0.0.1:12220"/bind_address = "0.0.0.0:12220"/' omicron/nexus/examples/config.toml

WORKDIR omicron
