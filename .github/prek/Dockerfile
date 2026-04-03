FROM debian:bookworm-slim AS build

# Go 1.26.1
ADD --checksum=sha256:031f088e5d955bab8657ede27ad4e3bc5b7c1ba281f05f245bcc304f327c987a \
    https://go.dev/dl/go1.26.1.linux-amd64.tar.gz \
    /tmp/go.tar.gz

# golangci-lint v2.11.4
ADD --checksum=sha256:200c5b7503f67b59a6743ccf32133026c174e272b930ee79aa2aa6f37aca7ef1 \
    https://github.com/golangci/golangci-lint/releases/download/v2.11.4/golangci-lint-2.11.4-linux-amd64.tar.gz \
    /tmp/golangci-lint.tar.gz

# helm v4.1.3
ADD --checksum=sha256:02ce9722d541238f81459938b84cf47df2fdf1187493b4bfb2346754d82a4700 \
    https://get.helm.sh/helm-v4.1.3-linux-amd64.tar.gz \
    /tmp/helm.tar.gz

# helm-unittest v1.0.3
ADD --checksum=sha256:9761f23d9509c98770c026e019e743b524b57010f4bc29175f78d2582ace0633 \
    https://github.com/helm-unittest/helm-unittest/releases/download/v1.0.3/helm-unittest-linux-amd64-1.0.3.tgz \
    /tmp/helm-unittest.tgz

# kubeconform v0.7.0
ADD --checksum=sha256:c31518ddd122663b3f3aa874cfe8178cb0988de944f29c74a0b9260920d115d3 \
    https://github.com/yannh/kubeconform/releases/download/v0.7.0/kubeconform-linux-amd64.tar.gz \
    /tmp/kubeconform.tar.gz

# prek 0.3.8
ADD --checksum=sha256:80ec6adb9f1883344de52cb943d371ecfd25340c4a6b5b81e2600d27e246cfa1 \
    https://github.com/j178/prek/releases/download/v0.3.8/prek-x86_64-unknown-linux-gnu.tar.gz \
    /tmp/prek.tar.gz

# helm-snoop v0.7.0
ADD --checksum=sha256:b79aa3869588bf066d5534b75319489dd821f3a75230a7526dcf457562e8942b \
    https://github.com/Y0-L0/helm-snoop/releases/download/v0.7.0/helm-snoop_0.7.0_linux_amd64.tar.gz \
    /tmp/helm-snoop.tar.gz

RUN mkdir -p /target/usr/local/bin /target/helm-plugins/unittest \
    && tar -C /target/usr/local -xzf /tmp/go.tar.gz \
    && tar -xzf /tmp/golangci-lint.tar.gz -C /target/usr/local/bin --strip-components=1 golangci-lint-2.11.4-linux-amd64/golangci-lint \
    && tar -xzf /tmp/helm.tar.gz -C /target/usr/local/bin --strip-components=1 linux-amd64/helm \
    && tar -xzf /tmp/helm-unittest.tgz -C /target/helm-plugins/unittest \
    && tar -xzf /tmp/kubeconform.tar.gz -C /target/usr/local/bin kubeconform \
    && tar -xzf /tmp/prek.tar.gz -C /target/usr/local/bin --strip-components=1 prek-x86_64-unknown-linux-gnu/prek \
    && tar -xzf /tmp/helm-snoop.tar.gz -C /target/usr/local/bin helm-snoop

FROM debian:bookworm-slim

RUN apt-get update && apt-get install -y --no-install-recommends \
    git \
    python3 \
    python3-pip \
    ca-certificates \
    && rm -rf /var/lib/apt/lists/* \
    && pip install --no-cache-dir --break-system-packages yamllint==1.38.0

COPY --from=build /target/usr/local /usr/local
COPY --from=build /target/helm-plugins/unittest /opt/helm-plugins/unittest

ENV PATH="/usr/local/go/bin:${PATH}" \
    HELM_PLUGINS="/opt/helm-plugins"
