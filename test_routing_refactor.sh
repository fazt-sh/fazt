#!/bin/bash
set -e

# Build
go build -o fazt ./cmd/server

# Clean env
TEMP_HOME=$(mktemp -d)
export HOME="$TEMP_HOME"
echo "Using HOME=$HOME"

mkdir -p $HOME/.config/fazt

# Init
./fazt server init --domain fazt.test --username admin --password secret

# Start Server
# Port 9090
./fazt server start --port 9090 > server.log 2>&1 &
SERVER_PID=$!
echo "Server started with PID $SERVER_PID"

cleanup() {
  kill $SERVER_PID
  wait $SERVER_PID 2>/dev/null || true
  rm -rf "$TEMP_HOME"
}
trap cleanup EXIT

# Wait for start
sleep 3

# Authenticate (Session Cookie)
curl -s -c cookies.txt -d '{"username":"admin","password":"secret"}' -H "Content-Type: application/json" http://localhost:9090/api/login

# Generate API Key
# Request: POST /api/keys, Body: {"name":"test-key","scopes":"[]"} (Guessing structure)
RAW_KEY_RESP=$(curl -s -b cookies.txt -X POST -d '{"name":"test-key"}' -H "Content-Type: application/json" http://localhost:9090/api/keys)
echo "Raw Key Response: $RAW_KEY_RESP"

TOKEN=$(echo "$RAW_KEY_RESP" | grep -o '"token":"[^"]*"' | cut -d'"' -f4)

if [ -z "$TOKEN" ]; then
  echo "Failed to get token"
  cat server.log
  exit 1
fi

echo "Got Token: ${TOKEN:0:10}..."

# Set token for CLI
./fazt client set-auth-token --token "$TOKEN"

# Create dummy sites
mkdir -p site_root site_404 site_test
echo "I AM ROOT" > site_root/index.html
echo "I AM 404" > site_404/index.html
echo "I AM TEST" > site_test/index.html

# Deploy sites
./fazt client deploy --path site_root --domain root --server http://localhost:9090
./fazt client deploy --path site_404 --domain 404 --server http://localhost:9090
./fazt client deploy --path site_test --domain test --server http://localhost:9090

echo "Sites deployed. Testing routing..."

# 1. Test Dashboard (localhost)
RESP=$(curl -s -H "Host: localhost" http://localhost:9090/login)
if [[ "$RESP" == *"Login"* ]]; then
  echo "✓ localhost -> Dashboard"
else
  echo "✗ localhost -> Dashboard FAILED"
  echo "$RESP" | head
  exit 1
fi

# 2. Test Dashboard (admin.fazt.test)
RESP=$(curl -s -H "Host: admin.fazt.test" http://localhost:9090/login)
if [[ "$RESP" == *"Login"* ]]; then
  echo "✓ admin.fazt.test -> Dashboard"
else
  echo "✗ admin.fazt.test -> Dashboard FAILED"
  echo "$RESP" | head
  exit 1
fi

# 3. Test Root Site (root.fazt.test)
RESP=$(curl -s -H "Host: root.fazt.test" http://localhost:9090)
if [[ "$RESP" == *"I AM ROOT"* ]]; then
  echo "✓ root.fazt.test -> Root Site"
else
  echo "✗ root.fazt.test -> Root Site FAILED"
  echo "Got: $RESP"
  exit 1
fi

# 4. Test Root Site (fazt.test)
RESP=$(curl -s -H "Host: fazt.test" http://localhost:9090)
if [[ "$RESP" == *"I AM ROOT"* ]]; then
  echo "✓ fazt.test -> Root Site"
else
  echo "✗ fazt.test -> Root Site FAILED"
  echo "Got: $RESP"
  exit 1
fi

# 5. Test 404 Site (404.fazt.test)
RESP=$(curl -s -H "Host: 404.fazt.test" http://localhost:9090)
if [[ "$RESP" == *"I AM 404"* ]]; then
  echo "✓ 404.fazt.test -> 404 Site"
else
  echo "✗ 404.fazt.test -> 404 Site FAILED"
  echo "Got: $RESP"
  exit 1
fi

# 6. Test Universal 404 (unknown.fazt.test)
STATUS=$(curl -s -o /dev/null -w "%{http_code}" -H "Host: unknown.fazt.test" http://localhost:9090)
RESP=$(curl -s -H "Host: unknown.fazt.test" http://localhost:9090)
if [[ "$STATUS" == "404" ]] && [[ "$RESP" == *"I AM 404"* ]]; then
  echo "✓ unknown.fazt.test -> Universal 404 Site"
else
  echo "✗ unknown.fazt.test -> Universal 404 Site FAILED"
  echo "Status: $STATUS"
  echo "Got: $RESP"
  exit 1
fi

# 7. Test Subdomain (test.fazt.test)
RESP=$(curl -s -H "Host: test.fazt.test" http://localhost:9090)
if [[ "$RESP" == *"I AM TEST"* ]]; then
  echo "✓ test.fazt.test -> Test Site"
else
  echo "✗ test.fazt.test -> Test Site FAILED"
  echo "Got: $RESP"
  exit 1
fi

echo "All routing tests passed!"