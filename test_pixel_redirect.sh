#!/bin/bash

# Test script for pixel and redirect handlers
BASE_URL="http://localhost:4698"

echo "Testing Command Center Pixel & Redirect Handlers"
echo "=================================================="
echo ""

echo "=== Pixel Tests ==="
echo ""

# Test 1: Basic pixel
echo "Test 1: Basic pixel tracking"
curl -s "$BASE_URL/pixel.gif?domain=newsletter&tags=dec,email" \
  -o /tmp/pixel.gif \
  -w "Status: %{http_code}\n"
echo "Pixel saved to /tmp/pixel.gif"
file /tmp/pixel.gif
echo ""

# Test 2: Pixel with custom source
echo "Test 2: Pixel with custom source"
curl -s "$BASE_URL/pixel.gif?domain=blog&source=email-campaign" \
  -o /dev/null \
  -w "Status: %{http_code}\n\n"

# Test 3: Pixel with no parameters (should use referer or unknown)
echo "Test 3: Pixel with no parameters"
curl -s "$BASE_URL/pixel.gif" \
  -o /dev/null \
  -w "Status: %{http_code}\n\n"

echo "=== Redirect Tests ==="
echo ""

# Test 4: Valid redirect (should return 302)
echo "Test 4: Valid redirect (using 'github' slug from mock data)"
curl -I -s "$BASE_URL/r/github" \
  -w "Status: %{http_code}\n" | grep -E "(HTTP|Location|Status)"
echo ""

# Test 5: Redirect with extra tags
echo "Test 5: Redirect with extra tags"
curl -I -s "$BASE_URL/r/twitter?tags=reddit,promo" \
  -w "Status: %{http_code}\n" | grep -E "(HTTP|Location|Status)"
echo ""

# Test 6: Invalid redirect slug (should return 404)
echo "Test 6: Invalid redirect slug (should fail)"
curl -s "$BASE_URL/r/nonexistent-slug" \
  -w "Status: %{http_code}\n\n"

# Test 7: Empty redirect slug (should return 400)
echo "Test 7: Empty redirect slug (should fail)"
curl -s "$BASE_URL/r/" \
  -w "Status: %{http_code}\n\n"

echo ""
echo "Testing complete!"
echo ""
echo "Check events in database:"
echo "  Events from pixel: SELECT COUNT(*) FROM events WHERE source_type='pixel'"
echo "  Events from redirect: SELECT COUNT(*) FROM events WHERE source_type='redirect'"
echo "  Redirect click counts: SELECT slug, click_count FROM redirects ORDER BY click_count DESC LIMIT 5"
