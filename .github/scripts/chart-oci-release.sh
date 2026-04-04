#!/usr/bin/env bash
set -euo pipefail

TAG="${1:?Usage: chart-oci-release.sh <tag>}"
VERSION="${TAG#v}"
REGISTRY="ghcr.io"
IMAGE_REPO="y0-l0/ldap-server-helm/ldap-manager"
CHART_REPO="oci://ghcr.io/y0-l0/ldap-server-helm"

setup_auth() {
    mkdir -p ~/.docker
    printf '{"auths":{"%s":{"auth":"%s"}}}' \
        "$REGISTRY" \
        "$(printf '%s:%s' "$GITHUB_ACTOR" "$GITHUB_TOKEN" | base64 -w0)" \
        > ~/.docker/config.json
}

build_image() {
    export KO_DOCKER_REPO="${REGISTRY}/${IMAGE_REPO}"
    (cd ldap-manager && ko build --bare --tags="${VERSION},latest" --image-refs=/tmp/ko-image-ref ./cmd/ldap-manager >&2)
    IMAGE_REF=$(cat /tmp/ko-image-ref)

    # Parse: ghcr.io/y0-l0/.../ldap-manager@sha256:abc...
    DIGEST="${IMAGE_REF##*@}"
    REPOSITORY="${IMAGE_REF%%@*}"
    REPOSITORY="${REPOSITORY#"${REGISTRY}/"}"
}

update_yaml() {
    yq e ".ldapManager.image.tag = \"${VERSION}\"" -i helm/ldap-server/values.yaml
    yq e ".ldapManager.image.digest = \"${DIGEST}\"" -i helm/ldap-server/values.yaml
    yq e ".version = \"${VERSION}\"" -i helm/ldap-server/Chart.yaml
}

package_chart() {
    helm dep build helm/ldap-server/
    helm package helm/ldap-server/

    CHART_PUSH=$(helm push "ldap-server-${VERSION}.tgz" "$CHART_REPO" 2>&1)
    printf '%s\n' "$CHART_PUSH"
    CHART_DIGEST=$(printf '%s\n' "$CHART_PUSH" | grep '^Digest:' | awk '{print $2}')
}

create_release() {
    gh release create "$TAG" \
        --title "ldap-server ${TAG}" \
        --notes "## ldap-server ${TAG}

### Install

\`\`\`
helm upgrade --install ldap-server oci://ghcr.io/y0-l0/ldap-server-helm/ldap-server --version ${VERSION}
\`\`\`

### Artifacts

| Artifact | Reference |
|----------|-----------|
| Helm chart | \`oci://ghcr.io/y0-l0/ldap-server-helm/ldap-server@${CHART_DIGEST}\` |
| ldap-manager image | \`${IMAGE_REF}\` |"
}

commit_back() {
    git config user.name "github-actions[bot]"
    git config user.email "github-actions[bot]@users.noreply.github.com"
    git add helm/ldap-server/values.yaml helm/ldap-server/Chart.yaml
    git commit -m "chore(release): bump ldap-manager to ${TAG}"
    git push origin HEAD:main
}

setup_auth
build_image
update_yaml
package_chart
# create_release
# commit_back
