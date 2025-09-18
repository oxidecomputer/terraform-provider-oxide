FROM rust

SHELL ["/bin/bash", "-c"]

RUN \
  git clone https://github.com/oxidecomputer/omicron.git --depth 1 --branch jmcarp/prereqs-omit-sudo && \
  cd omicron && \
  source env.sh && \
  ./tools/install_builder_prerequisites.sh -y -s

COPY nexus-config.toml omicron

WORKDIR omicron
