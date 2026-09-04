# Security Policy

The Business Central Admin Center provider manages privileged, production
infrastructure: it needs `AdminCenter.ReadWrite.All` and `AdminAgents`
membership, and it can permanently delete Business Central environments. We
take reports about it seriously and we would rather hear about a suspected
problem than not.

## Reporting a Vulnerability

**Please do not open a public issue, discussion, or pull request for a security
problem.**

Report it privately through GitHub's private vulnerability reporting:

**[Report a vulnerability](https://github.com/axiansinfoma/terraform-provider-bcadmincenter/security/advisories/new)**

(Repository → *Security* → *Advisories* → *Report a vulnerability*.) The report
is visible only to you and the maintainers, and the advisory doubles as the
place where we coordinate the fix and the disclosure with you.

Please include as much of the following as you have:

- The provider version (or commit) and how it was obtained — Terraform Registry,
  a GitHub release archive, or a local build.
- The affected resource, data source, or provider setting.
- A minimal Terraform configuration and the commands that reproduce it.
- What you observed and why you believe it is a security problem — the impact
  matters more than the mechanism.
- Any log output, with credentials and tenant identifiers redacted.

Please **do not** include live secrets, access tokens, or state files in a
report. A redacted excerpt is enough; we will ask if we need more.

### What to expect

| Stage | Target |
| --- | --- |
| Acknowledgement of your report | 5 working days |
| Initial assessment (severity, scope, whether it is in scope at all) | 10 working days |
| Fix released, or a public status update if it takes longer | 90 days |

This is a small, volunteer-maintained project — these are the targets we aim
for, not a contractual commitment. If a report goes unanswered past the
acknowledgement window, feel free to ping the advisory thread.

We will keep you updated as we work on a fix, credit you in the advisory and the
changelog unless you ask us not to, and agree a disclosure date with you before
publishing. Please give us a reasonable window to ship a fix before disclosing
publicly. We do not operate a bug bounty and cannot offer payment.

## Scope

**In scope** — the code and artifacts this repository produces:

- The provider source in `internal/`, including credential handling, request
  construction, and anything that could leak a token or send one to the wrong
  host.
- Sensitive values reaching Terraform state, plan files, or logs when the schema
  says they should not.
- Resources acting on a different tenant or environment than the configuration
  names.
- The release pipeline: the GoReleaser configuration, the release workflow, and
  the integrity of published archives, checksums, and signatures.
- Guidance in this repository's documentation that leads operators into an
  insecure configuration.

**Out of scope** — report these to their owners instead:

- The Business Central Admin Center API itself, or Business Central environment
  behaviour → [Microsoft](https://msrc.microsoft.com/create-report).
- Terraform CLI, the plugin framework, or state-backend behaviour →
  [HashiCorp](https://www.hashicorp.com/en/trust/security).
- Vulnerabilities in third-party Go modules that the provider does not reach.
  We track dependency updates through Renovate; if provider code makes the
  vulnerability reachable, that is in scope — say so in your report.
- Configuration mistakes in a user's own Terraform, Azure AD tenant, or CI
  system.
- Builds made with the `bcadmincenter_testing` build tag. That tag deliberately
  enables affordances that released binaries refuse (see below); a finding that
  depends on it is a finding about a build that should never be distributed.

## Supported Versions

The provider is pre-1.0 and only the latest release is supported. Fixes go out
in a new patch or minor release from `main`; we do not backport to older lines.

| Version | Supported |
| --- | --- |
| Latest release (see [releases](https://github.com/axiansinfoma/terraform-provider-bcadmincenter/releases)) | ✅ |
| Everything older | ❌ — upgrade |

If you pin a version, pin with `~>` so you pick up patch releases.

## Security Properties You Can Rely On

These are deliberate design decisions. If you find one of them does not hold,
that is a vulnerability worth reporting.

**Credentials never leave the intended host.** Every request goes to the Admin
Center API over TLS. A `base_url` override must use `https` — release builds
reject an `http://` URL, because the `Authorization` header is attached before
the destination is examined, and a plaintext URL would put a live Azure AD
access token on the wire in the clear. Overriding `base_url` at all raises a
warning.

**Testing affordances are compiled out of releases.** Accepting a static bearer
token in place of Azure AD authentication (`BCADMINCENTER_TEST_TOKEN`) and
talking to a plaintext base URL are both gated behind the
`bcadmincenter_testing` build tag, as a compile-time constant, so the branches
are removed from release binaries rather than left reachable at runtime. Never
build a distributed artifact with that tag, and never set it in a pipeline that
touches a real tenant.

**Secrets are marked sensitive.** `client_secret`, `oidc_token`, and
`oidc_request_token` are `Sensitive` in the provider schema, so Terraform
redacts them in plan and apply output. Terraform state is *not* encrypted by
this provider — treat state and plan files as secret material regardless, and
store them in a backend with encryption and access control.

**A resource talks to the tenant it names.** `aad_tenant_id` selects the tenant
a resource is read from and written to, and changing it forces replacement
rather than an in-place update against the wrong tenant.

## Recommendations for Operators

- **Prefer federated credentials over secrets.** Workload identity / OIDC
  (GitHub Actions, Azure DevOps, Kubernetes) avoids long-lived client secrets
  entirely. See the [workload identity
  guides](../tutorials/workload-identity-github.md). Use a client secret or
  certificate only where federation is not available, and rotate it.
- **Scope the service principal narrowly.** The provider needs a highly
  privileged application; do not reuse that app registration for anything else,
  and keep its Admin Center authorization list current.
- **Protect state.** State contains environment topology and configuration.
  Use a remote backend with encryption at rest, restricted access, and audit
  logging.
- **Review plans before applying.** The provider will destroy and recreate
  environments when an immutable attribute changes. Environment deletion is
  permanent.
- **Verify what you install.** Releases ship a `_SHA256SUMS` file with a
  detached GPG signature made by the maintainers' signing key. The Terraform
  Registry verifies this for you on `terraform init`; if you install a release
  archive by hand, check the checksum and the signature yourself.
- **Keep CI credentials least-privilege.** The provider's own workflows request
  `contents: read` (and `contents: write` only for releases) and pin every
  third-party action to a commit SHA. The same is worth doing in pipelines that
  run the provider.

## Hardening in This Repository

For transparency, the controls currently in place on the repository itself:

- Secret scanning and push protection are enabled.
- Third-party GitHub Actions are pinned to commit SHAs, and workflow
  permissions are scoped per workflow.
- Dependencies are updated by Renovate; the Go toolchain is tracked in `go.mod`
  and CI builds against it.
- Release artifacts are built by GoReleaser with `-trimpath` and `CGO_ENABLED=0`,
  and the checksum file is GPG-signed.
- Linting includes `depguard` rules that keep deprecated SDK packages out of the
  codebase, plus `errcheck` and `forcetypeassert` to stop silently ignored
  failures.
