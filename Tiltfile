load("ext://helm_resource", "helm_resource")

helm_resource(
    "ldap-server",
    chart="./helm/ldap-server/",
    deps=[
        "tilt-values.yaml",
        "helm/ldap-server/",
    ],
    flags=[
        "--values", "tilt-values.yaml",
    ],
    image_deps=[
        "ghcr.io/y0-l0/ldap-server-helm/ldap-manager:latest",
    ],
    image_keys=[
        ("ldapManager.image.registry", "ldapManager.image.repository", "ldapManager.image.tag"),
    ],
)

docker_build(
    "ghcr.io/y0-l0/ldap-server-helm/ldap-manager:latest",
    "./ldap-manager/",
    dockerfile="./ldap-manager/Dockerfile",
)

if config.tilt_subcommand == "down" and "-v" in sys.argv:
    local("kubectl delete pvc data-ldap-server-0 --wait=false --ignore-not-found")
