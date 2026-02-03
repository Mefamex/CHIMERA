#!/usr/bin/env bash
# Download third-party web assets used by the static site
set -euo pipefail

# target directory inside the repo where static assets are served
OUT_DIR="web/static/third_party"
mkdir -p "$OUT_DIR"
BOOTSTRAP_VERSION="5.3.8"
echo "Downloading Bootstrap $BOOTSTRAP_VERSION to $OUT_DIR"
curl -L "https://cdn.jsdelivr.net/npm/bootstrap@${BOOTSTRAP_VERSION}/dist/css/bootstrap.min.css" -o "$OUT_DIR/bootstrap.min.css"
curl -L "https://cdn.jsdelivr.net/npm/bootstrap@${BOOTSTRAP_VERSION}/dist/js/bootstrap.bundle.min.js" -o "$OUT_DIR/bootstrap.bundle.min.js"
curl -L "https://cdn.jsdelivr.net/npm/bootstrap@${BOOTSTRAP_VERSION}/dist/js/bootstrap.bundle.min.js.map" -o "$OUT_DIR/bootstrap.bundle.min.js.map" || true
curl -L "https://cdn.jsdelivr.net/npm/bootstrap@${BOOTSTRAP_VERSION}/dist/css/bootstrap.min.css.map" -o "$OUT_DIR/bootstrap.min.css.map" || true
echo "BOOTSTRAP $BOOTSTRAP_VERSION downloaded successfully."
echo "Saved Path: $OUT_DIR"
