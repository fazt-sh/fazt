#!/bin/bash
# API Migration Verification Script
# Checks progress of API standardization migration

set -e

echo "=== API Migration Status ==="
echo ""

# Count handlers (excluding test files)
TOTAL_HANDLERS=$(find internal/handlers -name "*.go" -type f ! -name "*_test.go" | wc -l)

# Count handlers using new api.Success or api.Error
MIGRATED_HANDLERS=$(grep -l "api\.Success\|api\.Error" internal/handlers/*.go 2>/dev/null | grep -v "_test.go" | wc -l)

echo "📊 Handlers migrated: $MIGRATED_HANDLERS / $TOTAL_HANDLERS"
echo ""

# Count legacy patterns
LEGACY_HTTP_ERROR=$(grep -n "http\.Error(" internal/handlers/*.go 2>/dev/null | grep -v "_test.go" | wc -l)
LEGACY_JSON_ENCODE=$(grep -n "json\.NewEncoder(w)\.Encode" internal/handlers/*.go 2>/dev/null | grep -v "_test.go" | wc -l)
LEGACY_JSON_ERROR=$(grep -n "jsonError(w," internal/handlers/*.go 2>/dev/null | grep -v "_test.go" | wc -l)

echo "🔍 Legacy Patterns Remaining:"
echo "   - http.Error() calls: $LEGACY_HTTP_ERROR"
echo "   - json.NewEncoder().Encode() calls: $LEGACY_JSON_ENCODE"
echo "   - jsonError() calls: $LEGACY_JSON_ERROR"
echo ""

# List files with legacy patterns
if [ $LEGACY_HTTP_ERROR -gt 0 ] || [ $LEGACY_JSON_ENCODE -gt 0 ] || [ $LEGACY_JSON_ERROR -gt 0 ]; then
    echo "📁 Files needing migration:"
    grep -l "http\.Error\|jsonError(w,\|json\.NewEncoder(w)\.Encode" internal/handlers/*.go 2>/dev/null | grep -v "_test.go" | sed 's/internal\/handlers\//   - /' || echo "   (none)"
    echo ""
fi

# Run tests
echo "=== Running Handler Tests ==="
echo ""

if go test ./internal/handlers/... -v --count=1 2>&1; then
    TEST_STATUS="✅ PASS"
else
    TEST_STATUS="❌ FAIL"
fi

echo ""
echo "=== Summary ==="
echo ""

if [ $LEGACY_HTTP_ERROR -eq 0 ] && [ $LEGACY_JSON_ENCODE -eq 0 ] && [ $LEGACY_JSON_ERROR -eq 0 ]; then
    echo "✅ Migration complete! All handlers use standardized API."
    echo "   Tests: $TEST_STATUS"
elif [ $MIGRATED_HANDLERS -ge 3 ]; then
    echo "🟡 Migration in progress... ($MIGRATED_HANDLERS/$TOTAL_HANDLERS handlers done)"
    echo "   Tests: $TEST_STATUS"
else
    echo "🔴 Migration not started yet."
fi

echo ""
