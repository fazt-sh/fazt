#!/bin/bash

# Test script for tracking endpoint
BASE_URL="http://localhost:4698"

echo "Testing Command Center Tracking Endpoint"
echo "========================================"
echo ""

# Test 1: Basic pageview
echo "Test 1: Basic pageview tracking"
curl -X POST $BASE_URL/track \
  -H "Content-Type: application/json" \
  -d '{"h":"test.com","p":"/page1","e":"pageview","t":["app","test"]}' \
  -w "\nStatus: %{http_code}\n\n"

# Test 2: Pageview with explicit domain
echo "Test 2: Pageview with explicit domain"
curl -X POST $BASE_URL/track \
  -H "Content-Type: application/json" \
  -d '{"d":"custom-domain.com","p":"/","e":"click","t":["campaign-123"]}' \
  -w "\nStatus: %{http_code}\n\n"

# Test 3: Pageview with query params
echo "Test 3: Pageview with query params"
curl -X POST $BASE_URL/track \
  -H "Content-Type: application/json" \
  -d '{"h":"blog.com","p":"/post","e":"pageview","t":["blog"],"q":{"ref":"twitter","utm_campaign":"promo"}}' \
  -w "\nStatus: %{http_code}\n\n"

# Test 4: Click event
echo "Test 4: Click event tracking"
curl -X POST $BASE_URL/track \
  -H "Content-Type: application/json" \
  -d '{"h":"shop.com","p":"/products","e":"click","t":["shop","ecommerce"],"ref":"https://google.com"}' \
  -w "\nStatus: %{http_code}\n\n"

# Test 5: Event with no tags
echo "Test 5: Event with no tags"
curl -X POST $BASE_URL/track \
  -H "Content-Type: application/json" \
  -d '{"h":"example.com","p":"/about","e":"pageview"}' \
  -w "\nStatus: %{http_code}\n\n"

# Test 6: Invalid JSON (should fail)
echo "Test 6: Invalid JSON (should fail)"
curl -X POST $BASE_URL/track \
  -H "Content-Type: application/json" \
  -d '{invalid json}' \
  -w "\nStatus: %{http_code}\n\n"

# Test 7: GET method (should fail)
echo "Test 7: GET method (should fail)"
curl -X GET $BASE_URL/track \
  -w "\nStatus: %{http_code}\n\n"

echo ""
echo "Testing complete!"
echo ""
echo "Verify data in database:"
echo "  sqlite3 cc.db 'SELECT COUNT(*) FROM events WHERE source_type=\"web\"'"
