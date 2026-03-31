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
        "gitregistry.knut.univention.de/univention/customers/dataport/upx/container-ldap/ldap-manager:latest",
    ],
    image_keys=[
        ("ldapManager.image.registry", "ldapManager.image.repository", "ldapManager.image.tag"),
    ],
)

docker_build(
    "gitregistry.knut.univention.de/univention/customers/dataport/upx/container-ldap/ldap-manager:latest",
    "./ldap-manager/",
    dockerfile="./ldap-manager/Dockerfile",
)
