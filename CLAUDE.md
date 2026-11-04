# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## What this is

A Kubebuilder (go.kubebuilder.io/v4) Kubernetes controller that generates configuration files for the [Homer](https://github.com/bastienwirtz/homer) static dashboard from CRDs (`homer.bananaops.io/v1alpha1`: `HomerServices`, `HomerPage`). Module: `github.com/BananaOps/homer-k8s`, Go 1.26 (required by k8s.io v0.37; code relies on `os.Root`).

## Commands

```bash
go build ./...                       # quick compile check
make test                            # generate + manifests + fmt + vet + go test -race with coverage (cover.out)
go test ./internal/controller -run TestMergeGroupWithSameName   # single test
make lint                            # golangci-lint v2 (.golangci.yml; tool downloaded into ./bin)
make generate manifests              # regenerate deepcopy + CRDs (helm/homer-k8s/crds/)
skaffold build / skaffold run        # build image with ko (main: ./cmd/) and deploy the Helm chart
```

No envtest: tests use the controller-runtime fake client, temp dirs and `HOMER_CONFIG_DIR`. The EC2 code is tested with a fake `ec2.DescribeInstancesAPIClient`.

CI (`.github/workflows/go-test.yml`) runs golangci-lint (version pinned to match the Makefile), gosec, `go test -race` with a coverage summary, and a codegen drift check (`make generate manifests` must leave no diff). Gosec findings matter: suppress them only with a justified `// #nosec Gxxx` comment.

## Architecture

The runtime model spans the Go code and the Helm chart, so you need both to understand it:

- **One pod, two containers sharing an `emptyDir`** (`helm/homer-k8s/templates/deployment.yaml`): the controller mounts it at `/assets`, and the Homer dashboard image mounts it at `/www/assets`. The controller's only output is YAML files written to that directory. It never creates Kubernetes objects.
- **`/config/global_config.yml`** comes from a ConfigMap rendered by the `homer-k8s.calculatedConfig` helper (`values.yaml` `homer.config` merged with `homer.structuredConfig`, plus `customCss` injection). The controllers treat it as the base or fallback config. The path is hard-coded.
- `HOMER_CONFIG_DIR` (default `/assets`) overrides the output directory, which is useful for local runs and tests.

Controllers (`internal/controller/`), wired in `cmd/main.go`:
- **HomerServicesReconciler**: on any change, lists *all* `HomerServices`, appends their groups to `global_config.yml` services, merges groups with the same name (`mergeGroupWithSameName`), sorts them by item count, and rewrites `config.yml` only if the services differ.
- **HomerPageReconciler**: writes `<pageName>.yml` (services only, via `HomerPageConfig`) for each `HomerPage`. When `autoLink` is set, it rebuilds the `links` section of `config.yml`. Links whose URL starts with `#` are treated as auto-managed. It then backfills empty top-level fields from the global config (`mergeGlobalIntoConfig`) and sets `status.ready`. All file I/O goes through an `os.Root` handle (path-traversal hardening). Keep that pattern, along with the `pageName` validation.
- Both controllers read and write the same `config.yml`, so changes in one can clobber the other's sections. Check both when you touch config merging.

`pkg/config` holds the Homer config schema (`HomerConfig`). Its `Extra map[string]interface{} \`yaml:",inline"\`` passes through Homer keys that aren't modeled. Service groups and items reuse the CRD types (`api/v1alpha1.Group`/`Item`) directly.

`internal/ec2`: optional manager `Runnable` registered when `HOMER_EC2_ENABLED=true`. Every 30s a `Syncer` lists running EC2 instances (paginated, default AWS credential chain) and writes `<HOMER_EC2_PAGE_NAME|ec2>.yml`. Errors are logged and never stop the loop.

`Group`/`Item` carry both `json` (CRD) and `yaml` (Homer files, yaml.v3) tags. They must stay identical, which `TestItemYAMLTagsMirrorJSONTags` enforces. Without the yaml tags, Homer's camelCase keys are lowercased and silently ignored.

## Making API / RBAC changes

- After editing `api/v1alpha1/*_types.go`: run `make generate manifests`. New root types go into `knownTypes` in their `init()` (the deprecated controller-runtime `scheme.Builder` is not used). CRDs live in `helm/homer-k8s/crds/`, not `config/` (this repo has no kustomize `config/` dir).
- `make manifests` does **not** generate RBAC. `helm/homer-k8s/templates/clusterrole.yaml` is hand-maintained and must be kept in sync with the `+kubebuilder:rbac` markers. Missing `homerpages` there was a past release bug.
- Sample CRs are in `samples/`.

## Releases and commits

- PR titles must be Conventional Commits (`feat|fix|docs|test|ci|refactor|perf|chore|revert`). release-please builds `CHANGELOG.md` and the version from them, using `release-please-config.json` and `.release-please-manifest.json` (tags stay `vX.Y.Z`).
- Dependabot titles follow the same convention: Go modules come in as `fix(deps)` (these trigger a patch release), and Actions as `ci(deps)`. `k8s.io/*` and `sigs.k8s.io/*` are grouped because they must be bumped together.
- On a `v*.*.*` tag, CI rewrites `version`/`appVersion` in `helm/homer-k8s/Chart.yaml` and commits `ci: bump charts to X`. Don't bump those by hand.
