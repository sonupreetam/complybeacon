#!/bin/bash
# Read-only version validation: checks that Go and OTel versions are
# consistent across all modules, Containerfiles, and manifest.yaml.
# Exit 0 = all aligned, exit 1 = drift detected.
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "$0")/../.." && pwd)"
cd "$ROOT_DIR"

FAILED=0

GO_WORK="go.work"
TRUTHBEAM_GOMOD="truthbeam/go.mod"
MANIFEST="beacon-distro/manifest.yaml"

# Auto-discover go.mod files and Containerfiles with Go base images
mapfile -t GO_MODS < <(find . -name go.mod -not -path '*/vendor/*' -print | sort || true)
mapfile -t CONTAINERFILES < <(grep -rl '^FROM golang:' . --include='Containerfile*' --include='Dockerfile*' 2>/dev/null | sort || true)

# Workspace modules (from go.work use block) — these carry OTel deps
mapfile -t WORKSPACE_MODULES < <(sed -n '/^use (/,/^)/{ s/^[[:space:]]*\.\///p }' "$GO_WORK" || true)

# ── Go version alignment ────────────────────────────────────────
echo "=== Go version check ==="

GO_VERSION=$(sed -n 's/^go \([0-9]*\.[0-9]*\.[0-9]*\)/\1/p' "$GO_WORK" | head -1)
if [[ -z "$GO_VERSION" ]]; then
    echo "ERROR: Could not extract Go version from $GO_WORK"
    exit 1
fi
echo "  Source of truth ($GO_WORK): $GO_VERSION"

for GOMOD in "${GO_MODS[@]}"; do
    MOD_VERSION=$(sed -n 's/^go \([0-9]*\.[0-9]*\.[0-9]*\)/\1/p' "$GOMOD" | head -1)
    if [[ -z "$MOD_VERSION" ]]; then
        MOD_VERSION=$(sed -n 's/^go \([0-9]*\.[0-9]*\)/\1/p' "$GOMOD" | head -1)
    fi
    if [[ "$MOD_VERSION" != "$GO_VERSION" ]]; then
        echo "  FAIL: $GOMOD has go $MOD_VERSION (expected $GO_VERSION)"
        FAILED=1
    else
        echo "  OK: $GOMOD"
    fi
done

for CF in "${CONTAINERFILES[@]}"; do
    CF_VERSION=$(sed -n 's/^FROM golang:\([0-9]*\.[0-9]*\.[0-9]*\).*/\1/p' "$CF" | head -1)
    if [[ -z "$CF_VERSION" ]]; then
        echo "  WARNING: Could not extract Go version from $CF"
        continue
    fi
    if [[ "$CF_VERSION" != "$GO_VERSION" ]]; then
        echo "  FAIL: $CF uses golang:$CF_VERSION (expected $GO_VERSION)"
        FAILED=1
    else
        echo "  OK: $CF"
    fi
done

for GOMOD in "${GO_MODS[@]}"; do
    if grep -q '^toolchain go' "$GOMOD"; then
        echo "  FAIL: $GOMOD has stale toolchain directive"
        FAILED=1
    fi
done

echo ""

# ── OTel version consistency ────────────────────────────────────
echo "=== OTel version check ==="

OTEL_EXPERIMENTAL=$(grep -E 'go\.opentelemetry\.io/collector/[^/]+' "$TRUTHBEAM_GOMOD" | \
                    grep -v 'go.opentelemetry.io/contrib' | \
                    grep -oE 'v0\.[0-9]+\.[0-9]+' | \
                    sort -V -u | tail -1)

OTEL_STABLE=$(grep -E 'go\.opentelemetry\.io/collector/[^/]+' "$TRUTHBEAM_GOMOD" | \
              grep -v 'go.opentelemetry.io/contrib' | \
              grep -oE 'v1\.[0-9]+\.[0-9]+' | \
              sort -V -u | tail -1)

echo "  Source of truth ($TRUTHBEAM_GOMOD): experimental=$OTEL_EXPERIMENTAL stable=$OTEL_STABLE"

for MODULE in "${WORKSPACE_MODULES[@]}"; do
    GOMOD="$MODULE/go.mod"
    if [[ ! -f "$GOMOD" ]]; then
        continue
    fi

    EXP_VERSIONS=$(grep -E 'go\.opentelemetry\.io/collector/[^/]+' "$GOMOD" | \
                   grep -v 'go.opentelemetry.io/contrib' | \
                   grep -oE 'v0\.[0-9]+\.[0-9]+' | \
                   sort -u || true)

    if [[ -n "$EXP_VERSIONS" ]]; then
        EXP_COUNT=$(echo "$EXP_VERSIONS" | wc -l)
        if [[ "$EXP_COUNT" -gt 1 ]]; then
            echo "  FAIL: $GOMOD has mixed experimental OTel versions:"
            echo "${EXP_VERSIONS//$'\n'/$'\n'    }"
            FAILED=1
        else
            FIRST_EXP=$(echo "$EXP_VERSIONS" | head -1 || true)
            echo "  OK: $GOMOD experimental at $FIRST_EXP"
        fi
    fi

    STABLE_VERSIONS=$(grep -E 'go\.opentelemetry\.io/collector/[^/]+' "$GOMOD" | \
                     grep -v 'go.opentelemetry.io/contrib' | \
                     grep -oE 'v1\.[0-9]+\.[0-9]+' | \
                     sort -u || true)

    if [[ -n "$STABLE_VERSIONS" ]]; then
        STABLE_COUNT=$(echo "$STABLE_VERSIONS" | wc -l)
        if [[ "$STABLE_COUNT" -gt 1 ]]; then
            echo "  FAIL: $GOMOD has mixed stable OTel versions:"
            echo "${STABLE_VERSIONS//$'\n'/$'\n'    }"
            FAILED=1
        else
            FIRST_STABLE=$(echo "$STABLE_VERSIONS" | head -1 || true)
            echo "  OK: $GOMOD stable at $FIRST_STABLE"
        fi
    fi
done

MANIFEST_VERSIONS=$(grep -E 'go\.opentelemetry\.io/collector/(exporter|processor|receiver)' "$MANIFEST" | \
                    grep -oE 'v[0-9]+\.[0-9]+\.[0-9]+' | sort -u || true)

for V in $MANIFEST_VERSIONS; do
    if [[ "$V" != "$OTEL_EXPERIMENTAL" ]]; then
        echo "  FAIL: $MANIFEST has component at $V (expected $OTEL_EXPERIMENTAL)"
        FAILED=1
    fi
done

PROVIDER_VERSIONS=$(grep -E 'go\.opentelemetry\.io/collector/confmap/provider' "$MANIFEST" | \
                    grep -oE 'v[0-9]+\.[0-9]+\.[0-9]+' | sort -u || true)

for V in $PROVIDER_VERSIONS; do
    if [[ "$V" != "$OTEL_STABLE" ]]; then
        echo "  FAIL: $MANIFEST has provider at $V (expected $OTEL_STABLE)"
        FAILED=1
    fi
done

CONTRIB_VERSIONS=$(grep -E 'github\.com/open-telemetry/opentelemetry-collector-contrib' "$MANIFEST" | \
                   grep -oE 'v[0-9]+\.[0-9]+\.[0-9]+' | sort -u || true)

for V in $CONTRIB_VERSIONS; do
    if [[ "$V" != "$OTEL_EXPERIMENTAL" ]]; then
        echo "  FAIL: $MANIFEST has contrib at $V (expected $OTEL_EXPERIMENTAL)"
        FAILED=1
    fi
done

COLLECTOR_CF="beacon-distro/Containerfile.collector"
if [[ -f "$COLLECTOR_CF" ]]; then
    BUILDER_VERSION=$(grep 'go.opentelemetry.io/collector/cmd/builder@' "$COLLECTOR_CF" | grep -oE 'v[0-9]+\.[0-9]+\.[0-9]+' || true)
    if [[ -n "$BUILDER_VERSION" && "$BUILDER_VERSION" != "$OTEL_EXPERIMENTAL" ]]; then
        echo "  FAIL: Builder at $BUILDER_VERSION (expected $OTEL_EXPERIMENTAL)"
        FAILED=1
    elif [[ -n "$BUILDER_VERSION" ]]; then
        echo "  OK: Builder at $BUILDER_VERSION"
    fi
fi

if [[ -z "$MANIFEST_VERSIONS" ]] || echo "$MANIFEST_VERSIONS" | grep -q "^${OTEL_EXPERIMENTAL}$"; then
    echo "  OK: $MANIFEST components aligned"
fi

echo ""

if [[ "$FAILED" -ne 0 ]]; then
    echo "FAILED: Version drift detected. Run 'task version:sync' to fix."
    exit 1
fi

echo "All version checks passed."
