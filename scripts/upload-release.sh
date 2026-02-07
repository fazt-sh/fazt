#!/bin/bash
# Upload a fazt release to fazt-releases.zyt.app
#
# API key setup (one-time):
#   curl -X POST https://fazt-releases.zyt.app/api/setup \
#     -H "Content-Type: application/json" -d '{"key":"your-secret-key"}'
#
# Then set FAZT_RELEASES_API_KEY in your .env or export it.
set -e

VERSION=${1:-}
FILE=${2:-}
DESCRIPTION=${3:-}

if [ -z "$VERSION" ] || [ -z "$FILE" ]; then
  echo "Usage: $0 <version> <file> [description]"
  echo ""
  echo "Example:"
  echo "  $0 v0.28.0 fazt-v0.28.0-linux-amd64.tar.gz 'Add custom URL upgrade support'"
  echo ""
  echo "Requires FAZT_RELEASES_API_KEY env var (or source .env)"
  exit 1
fi

if [ ! -f "$FILE" ]; then
  echo "Error: File '$FILE' not found"
  exit 1
fi

# Load .env if present
if [ -f .env ]; then
  source .env
fi

if [ -z "$FAZT_RELEASES_API_KEY" ]; then
  echo "Error: FAZT_RELEASES_API_KEY not set"
  echo "Set it in .env or export it"
  exit 1
fi

echo "Uploading $FILE as $VERSION to fazt-releases.zyt.app..."
echo ""

# Upload using curl with progress
RESPONSE=$(curl -s -w "\n%{http_code}" -X POST "https://fazt-releases.zyt.app/api/releases" \
  -H "X-API-Key: $FAZT_RELEASES_API_KEY" \
  -F "version=$VERSION" \
  -F "description=$DESCRIPTION" \
  -F "file=@$FILE" \
  --progress-bar)

HTTP_CODE=$(echo "$RESPONSE" | tail -1)
BODY=$(echo "$RESPONSE" | head -n -1)

if [ "$HTTP_CODE" = "201" ]; then
  RELEASE_ID=$(echo "$BODY" | jq -r '.id')
  echo ""
  echo "Upload complete!"
  echo ""
  echo "Download URL:"
  echo "  https://fazt-releases.zyt.app/api/releases/$RELEASE_ID/download"
  echo ""
  echo "Upgrade command:"
  echo "  fazt upgrade fazt-releases.zyt.app"
  echo ""
  echo "View releases:"
  echo "  https://fazt-releases.zyt.app"
else
  echo ""
  echo "Error ($HTTP_CODE): $BODY"
  exit 1
fi
