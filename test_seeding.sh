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
./fazt server start --port 9091 > server.log 2>&1 &
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

# Check Logs for Seeding
if grep -q "Seeded system site: root" server.log; then
  echo "✓ Log confirms root site seeded"
else
  echo "✗ Log missing seeding confirmation"
  cat server.log
  exit 1
fi

if grep -q "Seeded system site: 404" server.log; then
  echo "✓ Log confirms 404 site seeded"
else
  echo "✗ Log missing 404 seeding confirmation"
  cat server.log
  exit 1
fi

# Verify Content via HTTP
# 1. Root Site
RESP=$(curl -s -H "Host: root.fazt.test" http://localhost:9091)
if [[ "$RESP" == *"Welcome to Fazt"* ]]; then
  echo "✓ root.fazt.test serves Welcome Page"
else
  echo "✗ root.fazt.test FAILED"
  echo "Got: ${RESP:0:100}..."
  exit 1
fi

# 2. 404 Site
RESP=$(curl -s -H "Host: 404.fazt.test" http://localhost:9091)
if [[ "$RESP" == *"Lost in Space"* ]]; then
  echo "✓ 404.fazt.test serves Space Page"
else
  echo "✗ 404.fazt.test FAILED"
  echo "Got: ${RESP:0:100}..."
  exit 1
fi

echo "Seeding tests passed!"
