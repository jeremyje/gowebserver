# Periodic dependency refresh

Produced by `scripts/template-go/update-project.sh`, run on the Wednesday,
Friday and Saturday schedule in `.github/workflows/dependency-update.yaml`. The script
updates every area below in one pass, so the pins in this PR move together
rather than one surface lagging behind another.

- Toolchain and tool pins in the `Makefile_*.mk` files, resolved from each
  tool's authoritative source (GitHub releases, the HashiCorp releases API,
  or the Go module proxy).
- Terraform provider pins and `required_version` under `install/terraform`.
- Vendored protos under `third_party/`.
- Go modules, moved to their latest releases by this repo's own `make deps`
  target (`go get -u ./...`, `go mod tidy`, `go mod download`) and then
  checked with `go mod verify`.

The script leaves changes in the working tree without committing, so this
PR is the review point. Merging it is a human decision.
