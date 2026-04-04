FROM debian:bookworm-slim AS build

# Go 1.26.1
ADD --checksum=sha256:031f088e5d955bab8657ede27ad4e3bc5b7c1ba281f05f245bcc304f327c987a \
    https://go.dev/dl/go1.26.1.linux-amd64.tar.gz \
    /tmp/go.tar.gz

# ko v0.18.1
ADD --checksum=sha256:048ab11818089a43b7b74bc554494a79a3fd0d9822c061142e5cd3cf8b30cb27 \
    https://github.com/ko-build/ko/releases/download/v0.18.1/ko_0.18.1_Linux_x86_64.tar.gz \
    /tmp/ko.tar.gz

# helm v4.1.3
ADD --checksum=sha256:02ce9722d541238f81459938b84cf47df2fdf1187493b4bfb2346754d82a4700 \
    https://get.helm.sh/helm-v4.1.3-linux-amd64.tar.gz \
    /tmp/helm.tar.gz

# yq v4.45.1
ADD --checksum=sha256:654d2943ca1d3be2024089eb4f270f4070f491a0610481d128509b2834870049 \
    https://github.com/mikefarah/yq/releases/download/v4.45.1/yq_linux_amd64 \
    /target/usr/local/bin/yq

# gh v2.89.0
ADD --checksum=sha256:d0422caade520530e76c1c558da47daebaa8e1203d6b7ff10ad7d6faba3490d8 \
    https://github.com/cli/cli/releases/download/v2.89.0/gh_2.89.0_linux_amd64.tar.gz \
    /tmp/gh.tar.gz

RUN mkdir -p /target/usr/local/bin \
    && tar -C /target/usr/local -xzf /tmp/go.tar.gz \
    && tar -xzf /tmp/ko.tar.gz -C /target/usr/local/bin ko \
    && tar -xzf /tmp/helm.tar.gz -C /target/usr/local/bin --strip-components=1 linux-amd64/helm \
    && tar -xzf /tmp/gh.tar.gz -C /target/usr/local/bin --strip-components=2 gh_2.89.0_linux_amd64/bin/gh \
    && chmod +x /target/usr/local/bin/yq

FROM debian:bookworm-slim

RUN apt-get update && apt-get install -y --no-install-recommends \
    git \
    ca-certificates \
    && rm -rf /var/lib/apt/lists/*

COPY --from=build /target/usr/local /usr/local

ENV PATH="/usr/local/go/bin:${PATH}"
