FROM rust

SHELL ["/bin/bash", "-c"]

RUN \
  git clone https://github.com/oxidecomputer/omicron.git --branch main --depth 1 && \
  cd omicron && \
  source env.sh && \
  ./tools/install_builder_prerequisites.sh -y -s

COPY nexus-config.toml omicron

WORKDIR omicron
