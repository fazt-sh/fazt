#!/bin/bash
# Fast local release script - builds all platforms and uploads to GitHub
# Usage: ./scripts/release.sh v0.10.2 "Release title" "Release notes"
#
# Requires: GITHUB_PAT_FAZT in .env or environment

set -e

VERSION="$1"
TITLE="${2:-$VERSION}"
NOTES="${3:-See CHANGELOG.md for details.}"

if [ -z "$VERSION" ]; then
  echo "Usage: $0 <version> [title] [notes]"
  echo "Example: $0 v0.10.2 'Bug fixes' 'Fixed XYZ issue'"
  exit 1
fi

# Load PAT from .env if not in environment
if [ -z "$GITHUB_PAT_FAZT" ] && [ -f .env ]; then
  export $(grep GITHUB_PAT_FAZT .env | xargs)
fi

if [ -z "$GITHUB_PAT_FAZT" ]; then
  echo "Error: GITHUB_PAT_FAZT not found in environment or .env"
  exit 1
fi

REPO="fazt-sh/fazt"
BUILD_DIR="/tmp/fazt-release-$$"
mkdir -p "$BUILD_DIR"

echo "==> Building $VERSION"

# Check if admin needs rebuild (optional - skip if recent)
ADMIN_DIST="internal/assets/system/admin"
if [ ! -d "$ADMIN_DIST" ] || [ "$REBUILD_ADMIN" = "1" ]; then
  echo "==> Building admin SPA..."
  (cd admin && npm install --silent && npm run build --silent)
  rm -rf "$ADMIN_DIST"
  cp -r admin/dist "$ADMIN_DIST"
fi

# Build all platforms in parallel
echo "==> Cross-compiling (4 targets)..."
LDFLAGS="-w -s"

build_target() {
  local os=$1 arch=$2
  local out="$BUILD_DIR/fazt"
  CGO_ENABLED=0 GOOS=$os GOARCH=$arch go build -ldflags="$LDFLAGS" -o "$out" ./cmd/server
  tar -czf "$BUILD_DIR/fazt-$VERSION-$os-$arch.tar.gz" -C "$BUILD_DIR" fazt
  rm "$out"
  echo "  ✓ $os/$arch"
}

# Run builds (parallel if possible)
build_target linux amd64 &
build_target linux arm64 &
build_target darwin amd64 &
build_target darwin arm64 &
wait

echo "==> Creating GitHub release..."

# Create release
RELEASE_ID=$(curl -s -X POST \
  -H "Authorization: token $GITHUB_PAT_FAZT" \
  -H "Accept: application/vnd.github+json" \
  "https://api.github.com/repos/$REPO/releases" \
  -d "{
    \"tag_name\": \"$VERSION\",
    \"name\": \"$TITLE\",
    \"body\": \"$NOTES\",
    \"draft\": false,
    \"prerelease\": false
  }" | jq -r '.id')

if [ "$RELEASE_ID" = "null" ] || [ -z "$RELEASE_ID" ]; then
  echo "Error: Failed to create release"
  exit 1
fi

echo "==> Uploading assets to release $RELEASE_ID..."

upload_asset() {
  local file=$1
  local name=$(basename "$file")
  curl -s -X POST \
    -H "Authorization: token $GITHUB_PAT_FAZT" \
    -H "Content-Type: application/gzip" \
    "https://uploads.github.com/repos/$REPO/releases/$RELEASE_ID/assets?name=$name" \
    --data-binary @"$file" | jq -r '.name // .message'
}

for f in "$BUILD_DIR"/*.tar.gz; do
  result=$(upload_asset "$f")
  echo "  ✓ $result"
done

# Cleanup
rm -rf "$BUILD_DIR"

echo ""
echo "==> Release $VERSION complete!"
echo "    https://github.com/$REPO/releases/tag/$VERSION"
echo ""
echo "To upgrade zyt: fazt remote upgrade zyt"
