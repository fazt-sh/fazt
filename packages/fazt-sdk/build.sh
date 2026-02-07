#!/bin/bash
# Build fazt-sdk into a single ES module
# Usage: ./build.sh
set -e

DIR="$(cd "$(dirname "$0")" && pwd)"
ESBUILD="${DIR}/../../admin/node_modules/.bin/esbuild"

mkdir -p "$DIR/dist"

$ESBUILD "$DIR/index.js" \
  --bundle \
  --format=esm \
  --outfile="$DIR/dist/fazt-sdk.mjs"

echo "Built: $(wc -c < "$DIR/dist/fazt-sdk.mjs" | tr -d ' ') bytes → dist/fazt-sdk.mjs"
