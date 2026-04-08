# ldap-server

[![CI](https://github.com/y0-l0/ldap-server-helm/actions/workflows/prek.yml/badge.svg)](https://github.com/y0-l0/ldap-server-helm/actions/workflows/prek.yml)
[![Release](https://img.shields.io/github/v/release/y0-l0/ldap-server-helm)](https://github.com/y0-l0/ldap-server-helm/releases/latest)
Helm chart to deploy OpenLDAP on Kubernetes. Without a single line of bash.

The ldap-server pod deployed by this Helm chart runs exactly two binaries:
`slapd` and `ldap-manager`. No bash, no Python, no glue scripts.

Scalability and high-availability are planned.
Built with long-term maintenance in mind.

## Getting Started

```bash
helm upgrade --install ldap-server oci://ghcr.io/y0-l0/ldap-server-helm/ldap-server
```

Connect to verify:

```bash
kubectl exec -it ldap-server-0 -- ldapsearch \
  -x -H ldap://localhost \
  -D "cn=admin,dc=example,dc=org" \
  -w changeme \
  -b "dc=example,dc=org"
```

## Features

- **Static configuration:**  complete `slapd.conf` rendered by Helm, no runtime assembly
- **Flexible slapd.conf sourcing:** Helm-templated default, existing ConfigMap, inline content, or local file
- **Seed data:** idempotent LDIF seeding on first startup, same flexible sourcing
- **Health probes:** `/healthz` and `/readyz` backed by live LDAP queries
- **Persistence:** single PVC for `/var/lib/ldap`; `/var/run/slapd` is `emptyDir`
- **Existing secret support:** bring your own admin password secret
- **Security hardening:** non-root, read-only root filesystem, all capabilities dropped, seccomp RuntimeDefault
- **nubus-common:** standard Univention labels, names, and secret handling via library chart

## Roadmap

| Phase | Scope | Status |
|---|---|---|
| 1 | Single primary, seed data, health probes | Done |
| 2 | Secondary replicas, syncrepl replication | Planned |
| 3 | Leader election, hot-standby primary | Planned |
| 4 | Proxy topology | Planned |

## Configuration

### slapd.conf as a Helm template

The `slapd.conf` is a Helm-templated ConfigMap.
It is rendered at deploy time from `values.yaml`.
You can also supply your own `slapd.conf` via `slapdConf.existingConfigMap`, inline
`slapdConf.content`, or a local file path (`slapdConf.file`).
If you are using the `slapdConf.content` or `slapdConf.file` options,
you can embed template strings that will be rendered by the Helm chart:
Example: `suffix    {{ .Values.ldapServer.config.ldapBaseDn | quote }}`
see `files/slapd.conf` for a full example configuration.

### Seed data

On first startup, the sidecar loads seed LDIFs into the running slapd instance via `ldapadd`.
A sentinel file on the PVC (`/var/lib/ldap/.initialized`) prevents re-seeding on subsequent
pod restarts. Because seeding goes through slapd's normal write path, overlays and ACLs are
active during seeding.

Supply seed data via `seedData.existingConfigMap`, inline `seedData.content`, or a local file
(`seedData.file`).

## Architecture

Three containers per pod, two of which use the `ldap-manager` image:

```
┌───────────────────────────────────────────┐
│  Pod                                      │
│                                           │
│  ┌───────────────┐  (init, exits)         │
│  │ ldap-manager  │   creates dirs,        │
│  │ setup         │   writes rootpw.conf   │
│  └───────────────┘                        │
│                                           │
│  ┌───────────────┐  (main container)      │
│  │ slapd         │   reads slapd.conf     │
│  │               │   from ConfigMap       │
│  └───────────────┘                        │
│                                           │
│  ┌───────────────┐  (sidecar)             │
│  │ ldap-manager  │   seeds data,          │
│  │ sidecar       │   serves health probes │
│  └───────────────┘                        │
└───────────────────────────────────────────┘
```

**`ldap-manager`** is a statically compiled Go binary on a scratch image — no bash, no Python,
no coreutils. It handles everything that would otherwise require a custom entrypoint script.

**`slapd`** runs unmodified from its upstream image with no custom entrypoint. It reads a
`slapd.conf` mounted from a ConfigMap.

### kYAML templates

All Helm templates are written in [kYAML](https://github.com/kubernetes/enhancements/tree/master/keps/sig-api-machinery/5295-kyaml)
flow style — structure is encoded in `{}` and `[]`, not indentation.
Wrong indentation cannot change the meaning of the document.
A pre-commit hook enforces this by banning `nindent`, `indent`, and `{{-` from template files.

### Health probes

The `ldap-manager sidecar` serves the health endpoints that Kubernetes uses to probe the
`slapd` container:

| Endpoint | Probe | Check |
|---|---|---|
| `GET /healthz` | Liveness + Startup | LDAP root DSE query succeeds |
| `GET /readyz` | Readiness | LDAP root DSE query succeeds |

## Contributing

### Local development

[Tilt](https://tilt.dev) is used for local development against a running cluster:

```bash
tilt up
```

This builds the `ldap-manager` image locally and deploys the chart from source.
On `tilt down -v`, the data PVC is also deleted.

### Pre-commit hooks

Install [prek](https://github.com/y0-l0/prek) and run:

```bash
prek run --all-files
```

### Tests

Unit tests are automatically executed with `prek`.

End-to-end tests spin up a real `slapd` process.
Run them via docker-compose:

```bash
docker compose run e2e
```
