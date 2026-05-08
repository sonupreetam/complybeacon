#!/bin/bash
# Unified version sync: aligns Go and OTel versions across all modules,
# Containerfiles, CI workflows, and documentation.
#
# Source of truth:
#   Go version  — go.work `go` directive
#   OTel version — truthbeam/go.mod (experimental v0.x and stable v1.x)
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "$0")/../.." && pwd)"
cd "$ROOT_DIR"

GO_WORK="go.work"
TRUTHBEAM_GOMOD="truthbeam/go.mod"
MANIFEST="beacon-distro/manifest.yaml"

# Auto-discover go.mod files and Containerfiles with Go base images
mapfile -t GO_MODS < <(find . -name go.mod -not -path '*/vendor/*' -print | sort || true)
mapfile -t CONTAINERFILES < <(grep -rl '^FROM golang:' . --include='Containerfile*' --include='Dockerfile*' 2>/dev/null | sort || true)

# Workspace modules (from go.work use block) — these carry OTel deps
mapfile -t WORKSPACE_MODULES < <(sed -n '/^use (/,/^)/{ s/^[[:space:]]*\.\///p }' "$GO_WORK" || true)

# ── Extract Go version from go.work ──────────────────────────────
GO_VERSION=$(sed -n 's/^go \([0-9]*\.[0-9]*\.[0-9]*\)/\1/p' "$GO_WORK" | head -1)
if [[ -z "$GO_VERSION" ]]; then
    echo "ERROR: Could not extract Go version from $GO_WORK"
    exit 1
fi
GO_MINOR="${GO_VERSION%.*}"
echo "=== Go version sync (source: $GO_WORK) ==="
echo "  Target: $GO_VERSION (minor: $GO_MINOR)"

# ── Sync go.mod files ────────────────────────────────────────────
for GOMOD in "${GO_MODS[@]}"; do
    perl -i -pe "s/^go \d+\.\d+(\.\d+)?$/go $GO_VERSION/" "$GOMOD"
    perl -i -ne 'print unless /^toolchain go/' "$GOMOD"
    echo "  go.mod: $GOMOD"
done

# ── Sync Containerfile Go image tags ─────────────────────────────
for CF in "${CONTAINERFILES[@]}"; do
    perl -i -pe "s{FROM golang:\d+\.\d+\.\d+}{FROM golang:$GO_VERSION}" "$CF"
    echo "  Containerfile: $CF"
done

# ── Sync CI workflow GO_VERSION ──────────────────────────────────
CI_LOCAL=".github/workflows/ci_local.yml"
if [[ -f "$CI_LOCAL" ]]; then
    perl -i -pe "s{^(\s*GO_VERSION:\s*)\S+}{\${1}$GO_MINOR}" "$CI_LOCAL"
    echo "  CI workflow: $CI_LOCAL (GO_VERSION: $GO_MINOR)"
fi

# ── Sync documentation ──────────────────────────────────────────
DOCS=(
    "docs/DEVELOPMENT.md"
    "README.md"
    "AGENTS.md"
)
for DOC in "${DOCS[@]}"; do
    if [[ ! -f "$DOC" ]]; then
        continue
    fi
    perl -i -pe "s{Go \d+\.\d+\+}{Go ${GO_MINOR}+}g" "$DOC"
    perl -i -pe "s{Go \d+\.\d+\.\d+}{Go $GO_VERSION}g" "$DOC"
    perl -i -pe "s{golang:\d+\.\d+\.\d+}{golang:$GO_VERSION}g" "$DOC"
    echo "  Doc: $DOC"
done

echo ""

# ── Extract OTel versions from truthbeam ─────────────────────────
echo "=== OTel version sync (source: $TRUTHBEAM_GOMOD) ==="

OTEL_EXPERIMENTAL=$(grep -E 'go\.opentelemetry\.io/collector/[^/]+' "$TRUTHBEAM_GOMOD" | \
                    grep -v 'go.opentelemetry.io/contrib' | \
                    grep -oE 'v0\.[0-9]+\.[0-9]+' | \
                    sort -V -u | tail -1)

OTEL_STABLE=$(grep -E 'go\.opentelemetry\.io/collector/[^/]+' "$TRUTHBEAM_GOMOD" | \
              grep -v 'go.opentelemetry.io/contrib' | \
              grep -oE 'v1\.[0-9]+\.[0-9]+' | \
              sort -V -u | tail -1)

if [[ -z "$OTEL_EXPERIMENTAL" ]]; then
    echo "ERROR: Could not extract experimental (v0.x) OTel version from $TRUTHBEAM_GOMOD"
    exit 1
fi
if [[ -z "$OTEL_STABLE" ]]; then
    echo "ERROR: Could not extract stable (v1.x) OTel version from $TRUTHBEAM_GOMOD"
    exit 1
fi

echo "  Experimental: $OTEL_EXPERIMENTAL"
echo "  Stable: $OTEL_STABLE"

# ── Sync OTel versions in workspace module go.mod files ──────────
for MODULE in "${WORKSPACE_MODULES[@]}"; do
    GOMOD="$MODULE/go.mod"
    if [[ ! -f "$GOMOD" ]]; then
        continue
    fi
    perl -i -pe "s{(go\.opentelemetry\.io/collector/[^/\s]+)\s+v0\.\d+\.\d+}{\$1 $OTEL_EXPERIMENTAL}g" "$GOMOD"
    perl -i -pe "s{(go\.opentelemetry\.io/collector/[^/\s]+)\s+v1\.\d+\.\d+}{\$1 $OTEL_STABLE}g" "$GOMOD"
    echo "  go.mod: $GOMOD"
done

# ── Sync manifest.yaml ──────────────────────────────────────────
perl -i -pe "s{(go\.opentelemetry\.io/collector/(exporter|processor|receiver)/\w+) v[\d.]+}{\$1 $OTEL_EXPERIMENTAL}g" "$MANIFEST"
perl -i -pe "s{(github\.com/open-telemetry/opentelemetry-collector-contrib/(exporter|processor|receiver|connector|extension)/\w+) v[\d.]+}{\$1 $OTEL_EXPERIMENTAL}g" "$MANIFEST"
perl -i -pe "s{(go\.opentelemetry\.io/collector/confmap/provider/\w+) v[\d.]+}{\$1 $OTEL_STABLE}g" "$MANIFEST"
echo "  Manifest: $MANIFEST"

# ── Sync Containerfile builder version ───────────────────────────
COLLECTOR_CF="beacon-distro/Containerfile.collector"
if [[ -f "$COLLECTOR_CF" ]]; then
    perl -i -pe "s{builder\@v[\d.]+}{builder\@$OTEL_EXPERIMENTAL}g" "$COLLECTOR_CF"
    echo "  Builder: $COLLECTOR_CF"
fi

echo ""

# ── Run go mod tidy on all discovered modules ────────────────────
echo "=== Running go mod tidy ==="
for GOMOD in "${GO_MODS[@]}"; do
    MODULE_DIR=$(dirname "$GOMOD")
    echo "  Tidying $MODULE_DIR..."
    (cd "$MODULE_DIR" && go mod tidy 2>&1 | sed 's/^/    /')
done

echo ""
echo "=== Version sync complete ==="
echo "  Go: $GO_VERSION | OTel: experimental=$OTEL_EXPERIMENTAL stable=$OTEL_STABLE"
echo ""
echo "NOTE: Containerfile @sha256: digests are NOT updated automatically."
echo "      To update digests, pull the new image and replace the hash:"
echo "      podman pull golang:$GO_VERSION && podman inspect golang:$GO_VERSION --format '{{.Digest}}'"
