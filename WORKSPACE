http_archive(
    name = "io_bazel_rules_go",
    url = "https://github.com/bazelbuild/rules_go/releases/download/0.6.0/rules_go-0.6.0.tar.gz",
    sha256 = "ba6feabc94a5d205013e70792accb6cce989169476668fbaf98ea9b342e13b59",
)
load("@io_bazel_rules_go//go:def.bzl", "go_rules_dependencies", "go_register_toolchains", "go_repository")
go_rules_dependencies()
go_register_toolchains()

go_repository(
    name = "com_github_rs_cors",
    importpath = "github.com/rs/cors",
    tag = "v1.2",
)

go_repository(
    name = "com_github_stretchr_testify",
    importpath = "github.com/stretchr/testify",
    tag = "v1.1.4",
)

go_repository(
    name = "com_github_prometheus_client_golang",
    importpath = "github.com/prometheus/client_golang",
    tag = "v0.8.0",
)

go_repository(
    name = "com_github_prometheus_procfs",
    importpath = "github.com/prometheus/procfs",
    commit = "a6e9df898b1336106c743392c48ee0b71f5c4efa",
)

go_repository(
    name = "com_github_prometheus_client_model",
    importpath = "github.com/prometheus/client_model",
    commit = "6f3806018612930941127f2a7c6c453ba2c527d2",
)

go_repository(
    name = "com_github_pmezard_go_difflib",
    importpath = "github.com/pmezard/go-difflib",
    tag = "v1.0.0",
)

go_repository(
    name = "com_github_davecgh_go_spew",
    importpath = "github.com/davecgh/go-spew",
    tag = "v1.1.0",
)
go_repository(
    name = "com_github_matttproud_golang_protobuf_extensions",
    importpath = "github.com/matttproud/golang_protobuf_extensions",
    tag = "v1.0.0",
)



go_repository(
    name = "com_github_prometheus_common",
    importpath = "github.com/prometheus/common",
    commit = "1bab55dd05dbff384524a6a1c99006d9eb5f139b",
)
go_repository(
    name = "com_github_beorn7_perks",
    importpath = "github.com/beorn7/perks",
    commit = "4c0e84591b9aa9e6dcfdf3e020114cd81f89d5f9",
)

go_repository(
    name = "com_github_elazarl_go_bindata_assetfs",
    importpath = "github.com/elazarl/go-bindata-assetfs",
    commit = "30f82fa23fd844bd5bb1e5f216db87fd77b5eb43",
)

go_repository(
    name = "in_gopkg_yaml_v2",
    importpath = "gopkg.in/yaml.v2",
    commit = "eb3733d160e74a9c7e442f435eb3bea458e1d19f",
)

