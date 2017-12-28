workspace(name = "com_github_jeremyje_gowebserver")

http_archive(
    name = "io_bazel_rules_go",
    url = "https://github.com/bazelbuild/rules_go/releases/download/0.8.1/rules_go-0.8.1.tar.gz",
    sha256 = "90bb270d0a92ed5c83558b2797346917c46547f6f7103e648941ecdb6b9d0e72",
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

go_repository(
    name = "in_gopkg_src_d_go_billy_v3",
    importpath = "gopkg.in/src-d/go-billy.v3",
    commit = "c329b7bc7b9d24905d2bc1b85bfa29f7ae266314",
)

go_repository(
    name = "com_github_sergi_go_diff",
    importpath = "github.com/sergi/go-diff",
    commit = "1744e2970ca51c86172c8190fadad617561ed6e7",
)

go_repository(
    name = "com_github_src_d_go_git_fixtures",
    importpath = "github.com/src-d/go-git-fixtures",
    commit = "a29d269c3be65e4d1b20c29133c74e0551e1aa5d",
)

go_repository(
    name = "in_gopkg_src_d_go_git_v4",
    importpath = "gopkg.in/src-d/go-git.v4",
    commit = "f9879dd043f84936a1f8acb8a53b74332a7ae135",
)

go_repository(
    name = "com_github_src_d_gcfg",
    importpath = "github.com/src-d/gcfg",
    commit = "f187355171c936ac84a82793659ebb4936bc1c23",
)

go_repository(
    name = "in_gopkg_check_v1",
    importpath = "gopkg.in/check.v1",
    commit = "20d25e2804050c1cd24a7eea1e7a6447dd0e74ec",
)

go_repository(
    name = "com_github_jbenet_go_context",
    importpath = "github.com/jbenet/go-context",
    commit = "d14ea06fba99483203c19d92cfcd13ebe73135f4",
)

go_repository(
    name = "in_gopkg_warnings_v0",
    importpath = "gopkg.in/warnings.v0",
    commit = "ec4a0fea49c7b46c2aeb0b51aac55779c607e52b",
)

go_repository(
    name = "com_github_pkg_errors",
    importpath = "github.com/pkg/errors",
    commit = "e881fd58d78e04cf6d0de1217f8707c8cc2249bc",
)

go_repository(
    name = "com_github_xanzy_ssh_agent",
    importpath = "github.com/xanzy/ssh-agent",
    commit = "ba9c9e33906f58169366275e3450db66139a31a9",
)

go_repository(
    name = "org_golang_x_crypto",
    importpath = "golang.org/x/crypto",
    commit = "d585fd2cc9195196078f516b69daff6744ef5e84",
)

go_repository(
    name = "com_github_mitchellh_go_homedir",
    importpath = "github.com/mitchellh/go-homedir",
    commit = "b8bc1bf767474819792c23f32d8286a45736f1c6",
)
