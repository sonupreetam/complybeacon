# Image Publishing Process

This guide explains how to publish in GHCR and promote in Quay for container images by using the org-infra reusable workflows.

## Process Overview

The publishing process values **security and automation** to provide predictable, low-cost image releases.

```
Main Branch Push  →  Build + Scan + Sign  →  GHCR
                                              ↓
Release Tag (v*)  →  Verify + Promote     →  Quay.io
```

## Publishing Images

Images are automatically built and published when changes are pushed to the `main` branch.

### What Happens Automatically

1. **Build**: Images with SBOM and SLSA provenance
2. **Scan**: OSV Trivy vulnerability scan (blocks on HIGH/CRITICAL)
3. **Sign**: Keyless signing via Sigstore/Cosign
4. **Push**: Published to GHCR with `sha-<commit>` tag

### Manual Trigger

To manually trigger a build (e.g., for base image updates):

1. Go to **Actions** → **Publish Images to GHCR**
2. Click **Run workflow**
3. Optionally check **Force rebuild without cache**

## Promoting to Quay.io

Promotion copies signed images from GHCR to Quay.io for public distribution.

> **Key Point:** Promotion does **not rebuild** the image. It uses `cosign copy` to transfer the exact same bytes (identical `sha256` digest) from GHCR to Quay, preserving all signatures and attestations. This guarantees the image you tested in GHCR is identical to what's released on Quay.

### Creating a Release

```bash
# Ensure your changes are merged to main and images are built
git checkout main
git pull origin main

# Create a signed tag
git tag -s v1.2.3 -m "Release v1.2.3"
git push origin v1.2.3
```

The `promote.yml` workflow automatically:
- Verifies the GHCR image signature
- Copies the image to Quay.io (preserving signatures)
- Creates semver tags (`v1.2.3` → `1.2`, `1`)
- Re-signs on Quay.io

### Release Cadence

Releases are created as needed. Maintainers coordinate releases via issues or discussions.

## Setting Up a New Repository

### 1. Create Caller Workflows

### 2. Configure Secrets

Add these secrets in **Settings** → **Secrets and variables** → **Actions**:

| Secret | Required For | Description |
|--------|--------------|-------------|
| `QUAY_USERNAME` | Promotion | Quay.io robot account username |
| `QUAY_PASSWORD` | Promotion | Quay.io robot account token |

> **Note:** GHCR uses `GITHUB_TOKEN` automatically—no additional secrets needed.

### 3. Enable Branch Protection

In **Settings** → **Branches** → **main**:
- ✅ Require status checks to pass
- ✅ Require branches to be up to date

## Verifying Images

After publishing, verify images are properly signed:

```bash
# Verify GHCR image
cosign verify ghcr.io/complytime/my-image:sha-abc123 \
  --certificate-identity-regexp='https://github.com/complytime/.*' \
  --certificate-oidc-issuer=https://token.actions.githubusercontent.com

# Verify Quay image
cosign verify quay.io/complytime/my-image:v1.2.3 \
  --certificate-identity-regexp='https://github.com/complytime/.*' \
  --certificate-oidc-issuer=https://token.actions.githubusercontent.com
```

## Quick Reference

| Task | Workflow | Trigger |
|------|----------|---------|
| Build & publish to GHCR | `ci_publish_ghcr.yml` | Push to `main` |
| Promote to Quay.io | `ci_promote_quay.yml` | Push tag `v*.*.*` |

## More Information
- [Sigstore Documentation](https://docs.sigstore.dev/) — Keyless signing details
