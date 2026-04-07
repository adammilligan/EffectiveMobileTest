FROM mcr.microsoft.com/devcontainers/go:1-1.23-bookworm

RUN apt-get update \
  && apt-get install -y --no-install-recommends \
    make \
    curl \
    ca-certificates \
  && rm -rf /var/lib/apt/lists/*
