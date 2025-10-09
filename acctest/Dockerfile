FROM rust

SHELL ["/bin/bash", "-c"]

ARG OMICRON_REF=main

RUN \
  git clone https://github.com/oxidecomputer/omicron.git && \
  cd omicron && \
  git checkout $OMICRON_REF && \
  source env.sh && \
  ./tools/install_builder_prerequisites.sh -y -s && \
  cargo xtask omicron-dev help

COPY nexus-config.toml omicron

WORKDIR omicron
