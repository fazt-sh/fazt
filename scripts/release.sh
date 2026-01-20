#!/bin/bash
# Fast local release script
# Usage: ./scripts/release.sh v0.10.7
#
# Requires: GITHUB_PAT_FAZT environment variable (from .env)

set -e

VERSION=$1
if [ -z "$VERSION" ]; then
  echo "Usage: $0 <version>"
  echo "Example: $0 v0.10.7"
  exit 1
fi

if [ -z "$GITHUB_PAT_FAZT" ]; then
  echo "Error: GITHUB_PAT_FAZT not set. Run: source .env"
  exit 1
fi

REPO="fazt-sh/fazt"
echo "Building fazt $VERSION for all platforms..."

# Build admin if needed
if [ ! -d "internal/assets/system/admin" ]; then
  echo "Building admin SPA..."
  npm run build --prefix admin
  rm -rf internal/assets/system/admin
  cp -r admin/dist internal/assets/system/admin
fi

# Build all platforms
PLATFORMS="linux/amd64 linux/arm64 darwin/amd64 darwin/arm64"
for PLATFORM in $PLATFORMS; do
  OS=${PLATFORM%/*}
  ARCH=${PLATFORM#*/}
  echo "  Building $OS/$ARCH..."
  CGO_ENABLED=0 GOOS=$OS GOARCH=$ARCH go build -ldflags="-w -s" -o fazt ./cmd/server
  tar -czvf "fazt-${VERSION}-${OS}-${ARCH}.tar.gz" fazt > /dev/null
done

# Create release
echo "Creating GitHub release..."
RELEASE_RESPONSE=$(curl -s -X POST \
  -H "Authorization: token $GITHUB_PAT_FAZT" \
  -H "Accept: application/vnd.github.v3+json" \
  "https://api.github.com/repos/$REPO/releases" \
  -d "{\"tag_name\":\"$VERSION\",\"name\":\"$VERSION\",\"generate_release_notes\":true}")

RELEASE_ID=$(echo "$RELEASE_RESPONSE" | python3 -c "import sys,json; print(json.load(sys.stdin).get('id',''))")

if [ -z "$RELEASE_ID" ]; then
  echo "Error creating release:"
  echo "$RELEASE_RESPONSE"
  exit 1
fi

echo "Release created (ID: $RELEASE_ID). Uploading assets..."

# Upload assets
for PLATFORM in $PLATFORMS; do
  OS=${PLATFORM%/*}
  ARCH=${PLATFORM#*/}
  ASSET="fazt-${VERSION}-${OS}-${ARCH}.tar.gz"
  echo "  Uploading $ASSET..."
  curl -s -X POST \
    -H "Authorization: token $GITHUB_PAT_FAZT" \
    -H "Content-Type: application/gzip" \
    "https://uploads.github.com/repos/$REPO/releases/$RELEASE_ID/assets?name=$ASSET" \
    --data-binary "@$ASSET" > /dev/null
done

# Cleanup
rm -f fazt-*.tar.gz fazt

echo "Done! Release $VERSION ready at:"
echo "  https://github.com/$REPO/releases/tag/$VERSION"
