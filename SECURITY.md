# Security policy

## Supported versions

Security fixes are applied to the **default branch** (`main`) and may be released as new **patch tags** (for example `v0.1.1`) when applicable.

There is no long-term support promise for older major/minor lines until the project publishes an explicit support policy.

## Code of conduct enforcement

For reports covered by [CODE_OF_CONDUCT.md](CODE_OF_CONDUCT.md), use the **same private channels** as security reports (for example GitHub private reporting if enabled, or a direct message / email to maintainers listed on their GitHub profiles). Do not use public issues for sensitive harassment reports.

## Reporting a vulnerability

**Please do not** file a public GitHub issue that contains exploit details for unfixed vulnerabilities.

Preferred options:

1. **GitHub private vulnerability reporting** (if enabled for `triadmoko/office`): use **Security → Report a vulnerability** on the repository.
2. Otherwise, contact the **repository maintainers** through a private channel (for example email listed on the maintainer’s GitHub profile).

Include:

- A short description of the issue and its impact
- Steps to reproduce (code, file, or command line)
- Affected versions or commit SHA if known

Maintainers will aim to acknowledge receipt in a reasonable timeframe; exact SLAs depend on volunteer availability.

## Scope

This policy covers the **office** module and the **`cmd/office`** demo binary as shipped in this repository.

Out of scope examples:

- Vulnerabilities in Go itself (report to the Go security team)
- Issues in third-party dependencies **not** used by this repository’s current **stdlib-only** policy (there should be none in `go.mod`)

## Safe harbor

We support **good-faith** security research that follows this reporting process and does not violate law or disrupt users.
