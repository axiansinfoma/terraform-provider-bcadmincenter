# Contributing

Thanks for your interest in improving the Terraform provider for Business
Central Admin Center. This document covers how to get set up, what the project
expects from a change, and how a pull request gets reviewed.

By participating you agree to abide by our [Code of
Conduct](CODE_OF_CONDUCT.md). Security problems follow a different route —
please read the [Security Policy](SECURITY.md) and report privately rather than
opening an issue.

## Before You Start

| I want to… | Go here |
| --- | --- |
| Ask how to configure or use the provider | [Discussions](https://github.com/axiansinfoma/terraform-provider-bcadmincenter/discussions) — usage questions in the issue tracker get closed |
| Report a bug | [Bug report](https://github.com/axiansinfoma/terraform-provider-bcadmincenter/issues/new?template=Bug_Report.yml) |
| Request a resource, data source, or attribute | [Feature request](https://github.com/axiansinfoma/terraform-provider-bcadmincenter/issues/new?template=Feature_Request.yml) |
| Report a security vulnerability | [Security Policy](SECURITY.md) — privately, not an issue |
| Report a problem in Terraform itself | [hashicorp/terraform](https://github.com/hashicorp/terraform/issues) |

For anything larger than a bug fix or a documentation correction, **open an
issue first** and let's agree on the shape of the change. This provider wraps
an API whose live behaviour frequently differs from its published
documentation, so the design discussion is usually the expensive part — it is
worth having before you write the code.

## Development Environment

### Requirements

- **Go** — the version in [`go.mod`](../go.mod) (currently 1.27.1). CI builds
  with `go-version-file: go.mod`, so an older toolchain will not match.
- **Terraform** >= 1.13. CI tests against 1.13.x and 1.14.x.
- **[golangci-lint](https://golangci-lint.run/welcome/install/)** v2 — the
  configuration in [`.golangci.yml`](../.golangci.yml) uses the v2 schema.
- **[pre-commit](https://pre-commit.com/)** — optional but recommended; the
  hooks catch most of what CI would fail on.
- Optionally, access to a Business Central tenant and an Azure AD app
  registration. You do **not** need one for most contributions: the unit tests
  and the mock-server acceptance tests run entirely offline.

A [dev container](../.devcontainer/devcontainer.json) is included and installs
Go, Terraform, the Azure CLI, the GitHub CLI, and the pre-commit hooks for you.

### Setup

```shell
git clone https://github.com/axiansinfoma/terraform-provider-bcadmincenter
cd terraform-provider-bcadmincenter
go mod download
pre-commit install   # optional
```

### Make targets

```shell
make build            # go build ./...
make install          # go install ./...
make fmt              # gofmt -s -w -e .
make lint             # golangci-lint run
make test             # unit tests
make testmock         # acceptance tests against the local mock HTTP server
make testacc          # acceptance tests against a REAL tenant — see the warning below
make docs             # regenerate docs/ from templates/ and examples/
make docs-check       # fail if docs/ is out of date
make validate-docs    # docs-check + terraform fmt -check on examples/
make generate         # everything under tools/ (docs generation)
```

Bare `make` runs `fmt lint install generate`.

### Trying your build against a real tenant

[`scripts/setup-local-testing.sh`](../scripts/setup-local-testing.sh) builds the
provider and prints (or, with your confirmation, appends) the
`provider_installation { dev_overrides { … } }` block for `~/.terraformrc` that
points `axiansinfoma/bcadmincenter` at your working tree. With a dev override
in place you skip `terraform init` — rebuild the binary and re-run
`terraform plan`.

The [`dev/`](../dev/README.md) directory holds ready-made configurations
covering every resource and data source for exactly this purpose. Credentials
come from the environment (`ARM_CLIENT_ID`, `ARM_CLIENT_SECRET`,
`ARM_TENANT_ID`, or `az login`).

## Testing

**Tests are not optional.** Every new resource, data source, or service method
needs coverage before the change is considered complete.

The suite has four layers:

1. **Unit tests** — `make test`, or `go test ./...`. Service-layer tests use a
   mock HTTP server and must cover success, API errors, network errors, and
   malformed responses. Schema and metadata tests assert the type name, the
   attributes, and `Configure` behaviour.
2. **Mock acceptance tests** — `make testmock`. Real Terraform lifecycle tests
   run against a local mock HTTP server. These need the
   `bcadmincenter_testing` build tag (see below).
3. **Terraform test framework** — `cd tests && terraform init && terraform test`.
   Configuration-level tests using `mock_provider`.
4. **Live acceptance tests** — `make testacc`. ⚠️ **These create, modify, and
   destroy real Business Central environments in a real tenant**, take up to
   two hours, and may incur cost. Run them only against a tenant you own and
   are willing to have rearranged. They are not run in CI.

### The `bcadmincenter_testing` build tag

The mock-server tests need two things a released provider must never do: accept
a static bearer token instead of authenticating against Azure AD, and talk to a
plaintext `http://` base URL. Both are gated behind the `bcadmincenter_testing`
build tag as a compile-time constant, so they are removed from untagged builds
entirely.

Two rules follow, and CI enforces both:

- The untagged unit-test run must stay untagged. It is what proves the
  release-mode guards still compile and hold.
- Never add a new capability behind that tag without a comment in
  [`buildmode_testing.go`](../internal/client/buildmode_testing.go) explaining
  why it cannot exist in a release build.

## Adding a Resource or Data Source

Resources live in per-service packages under `internal/services/<service>/`,
with the API client in `internal/client/`, shared enums and defaults in
`internal/constants/`, and the provider wiring in `internal/provider/`.

A complete change touches all of these:

1. **Service and schema** — a new package under `internal/services/`, using
   the [Terraform Plugin
   Framework](https://developer.hashicorp.com/terraform/plugin/framework).
   Reuse values from `internal/constants` rather than hardcoding API strings.
2. **Registration** — add it to `Resources()` or `DataSources()` in
   [`internal/provider/provider.go`](../internal/provider/provider.go), and
   update the expected count in `provider_test.go`.
3. **Tests** — service tests plus schema/metadata tests, as described above.
4. **Documentation** — a template in `templates/` and a runnable example in
   `examples/resources/<type_name>/` (`resource.tf`, and `import.sh` for
   importable resources). Then `make docs` and commit the regenerated `docs/`.
5. **Changelog** — an entry under the appropriate heading (see below).

Design conventions worth knowing before you start, drawn from problems we have
already hit:

- **Attributes that select a tenant or are immutable in the API must force
  replacement.** An in-place update that talks to the old tenant while
  recording the new one is a correctness bug, not a convenience.
- **If the API has no operation to unset something, reject the removal at plan
  time.** Silently producing no request lets Terraform record an attribute as
  gone while the environment keeps it.
- **A read that fails is not a read that returned nothing.** Only remove a
  resource from state on a genuine 404; a transient 503 or a timeout must
  surface as an error.
- **If a create partially succeeded, record it in state.** Losing a resource
  that exists remotely leaves the next apply failing as "already exists".
- **Compare API status values and identifiers case-insensitively.** The live
  API is inconsistent about casing across endpoints.
- **Honour the `timeouts` block if you advertise it.**

## Documentation

`docs/` is **generated** — do not edit it by hand. Edit the source instead:

- `templates/` — the page templates ([template
  guide](../templates/README.md)); every resource/data-source template must
  contain `{{ .SchemaMarkdown }}`.
- `examples/` — the `.tf` and `.sh` files the templates embed. Format them with
  `terraform fmt -recursive examples/`, and give every `.tf` file the copyright
  header.
- Schema `MarkdownDescription` strings — these become the attribute reference.
- `templates/guides/` — long-form authentication and workflow guides. These
  are plain Markdown with registry frontmatter (`page_title`, `subcategory`,
  `description`) and are copied verbatim into `docs/guides/`.
  [`tutorials/`](../tutorials/README.md) holds the same guides without
  frontmatter, for reading in the repository — keep the two in sync.

Then run `make docs` and commit the result. CI regenerates the docs and fails
the build if the committed `docs/` differs, and the pre-commit hook does the
same locally — it deliberately does not stage the regenerated files for you, so
that generated changes appear in the diff you are about to review.

## Code Style

`make fmt` and `make lint` are the authority. What they enforce, so you are not
surprised:

- **`gofmt -s`** formatting.
- **`godot`** — top-level comments end with a period.
- **`depguard`** — the legacy `terraform-plugin-sdk/v2` is banned. Use the
  Plugin Framework, and `terraform-plugin-testing` for test helpers.
- **`errcheck`**, **`forcetypeassert`**, **`nilerr`** — no silently discarded
  errors, no unchecked type assertions.
- **`misspell`**, **`unparam`**, **`unused`**, **`staticcheck`**, and friends.

Every source file needs the MPL-2.0 copyright header:

```go
// Copyright (c) 2025 Axians Infoma GmbH
// SPDX-License-Identifier: MPL-2.0
```

[`.copywrite.hcl`](../.copywrite.hcl) records which paths are exempt;
[`hashicorp/copywrite`](https://github.com/hashicorp/copywrite) can add headers
for you (`copywrite headers`).

Write comments that explain *why*, especially where the code works around live
API behaviour that contradicts Microsoft's documentation. Those comments are the
reason the workaround survives the next refactor.

## Commits and Pull Requests

### Commit messages

We use [Conventional Commits](https://www.conventionalcommits.org/):
`feat:`, `fix:`, `docs:`, `chore:`, `refactor:`, `test:`, with an optional
scope — `feat(pte): migrate per-tenant extensions to Admin Center API 2.29`.
Write the subject in the imperative mood and explain the *why* in the body.

### Changelog

Every user-visible change needs an entry in [`CHANGELOG.md`](../CHANGELOG.md)
under a heading for the next unreleased version (add one at the top if it does
not exist yet), in HashiCorp's provider style — one of
`BREAKING CHANGES:`, `FEATURES:`, `ENHANCEMENTS:`, `BUG FIXES:`, prefixed with
the affected component and ending with a link to the PR:

```markdown
BUG FIXES:
* resource/bcadmincenter_environment: `azure_region` is no longer discarded on
  refresh ([#103](https://github.com/axiansinfoma/terraform-provider-bcadmincenter/pull/103))
```

Say what changed *and* what went wrong before, so an operator reading the
changelog can tell whether it affected them. Purely internal changes (CI,
refactors with no behavioural difference) do not need an entry.

### Opening the PR

1. Branch off `main`.
2. Run `make fmt lint test testmock` and `make validate-docs` locally.
3. Fill in the [pull request template](pull_request_template.md) — link the
   issue, describe the change, and say how you tested it.
4. Keep the PR focused. A behavioural fix, a dependency bump, and a
   documentation rewrite are three PRs.
5. Push, and check that CI is green: build and lint, docs generation, unit and
   mock acceptance tests on both Terraform versions, and the Terraform test
   framework suite.

Review comes from the [code owners](CODEOWNERS). Expect questions about live API
behaviour, about what happens on refresh after an out-of-band change, and about
the tests. Maintainers may push small fixups to your branch to get a change
over the line; say so in the PR if you would rather they didn't.

### Releases

Maintainers release by pushing a `v*` tag, which runs GoReleaser: it regenerates
the documentation, builds the cross-platform archives, and publishes them with a
GPG-signed `SHA256SUMS` file for the Terraform Registry. Contributors do not
need to do anything for a release beyond the changelog entry.

## License

The provider is licensed under the [Mozilla Public License
2.0](../LICENSE). Contributions are accepted under the same license.
