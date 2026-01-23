#!/bin/bash
# Storage layer stress test

HOST="${1:-http://192.168.64.3:8080}"
APP="cashflow"
SESSION="stress-test-$$"

echo "=== Fazt Storage Stress Test ==="
echo "Host: $HOST"
echo "App: $APP"
echo "Session: $SESSION"
echo ""

# Function to make a request and report result
request() {
    local method=$1
    local endpoint=$2
    local data=$3
    local start=$(date +%s%N)

    if [ "$method" = "GET" ]; then
        response=$(curl -s -w "\n%{http_code}" -H "Host: $APP.192.168.64.3" "$HOST$endpoint" 2>&1)
    else
        response=$(curl -s -w "\n%{http_code}" -X POST -H "Host: $APP.192.168.64.3" -H "Content-Type: application/json" -d "$data" "$HOST$endpoint" 2>&1)
    fi

    local end=$(date +%s%N)
    local duration=$(( (end - start) / 1000000 ))
    local status=$(echo "$response" | tail -n1)

    if [ "$status" = "200" ]; then
        echo "OK ${duration}ms"
        return 0
    else
        echo "FAIL($status) ${duration}ms"
        return 1
    fi
}

# Test 1: Sequential writes
echo "--- Test 1: Sequential Writes (20 requests) ---"
success=0
total=20
for i in $(seq 1 $total); do
    result=$(request POST "/api/test-write" "{\"session\":\"$SESSION\",\"index\":$i}")
    if [[ $result == OK* ]]; then
        ((success++))
    fi
    echo "Write $i: $result"
done
echo "Result: $success/$total ($(( success * 100 / total ))%)"
echo ""

# Test 2: Concurrent writes (10 parallel)
echo "--- Test 2: Concurrent Writes (10 parallel) ---"
pids=()
results_file="/tmp/stress_concurrent_$$"
rm -f "$results_file"

for i in $(seq 1 10); do
    (
        result=$(request POST "/api/test-write" "{\"session\":\"$SESSION-concurrent\",\"index\":$i}")
        if [[ $result == OK* ]]; then
            echo "1" >> "$results_file"
        else
            echo "0" >> "$results_file"
        fi
        echo "Concurrent $i: $result"
    ) &
    pids+=($!)
done

# Wait for all
for pid in "${pids[@]}"; do
    wait $pid
done

success=$(grep -c "1" "$results_file" 2>/dev/null || echo "0")
echo "Result: $success/10 ($(( success * 10 ))%)"
rm -f "$results_file"
echo ""

# Test 3: Sequential reads
echo "--- Test 3: Sequential Reads (50 requests) ---"
success=0
total=50
for i in $(seq 1 $total); do
    result=$(request GET "/api/categories?session=$SESSION")
    if [[ $result == OK* ]]; then
        ((success++))
    fi
    # Only print every 10th to reduce output
    if (( i % 10 == 0 )); then
        echo "Read $i: $result (running: $success/$i)"
    fi
done
echo "Result: $success/$total ($(( success * 100 / total ))%)"
echo ""

# Test 4: Mixed workload
echo "--- Test 4: Mixed Read/Write (20 requests) ---"
success=0
total=20
for i in $(seq 1 $total); do
    if (( i % 2 == 0 )); then
        result=$(request GET "/api/categories?session=$SESSION")
    else
        result=$(request POST "/api/test-write" "{\"session\":\"$SESSION-mixed\",\"index\":$i}")
    fi
    if [[ $result == OK* ]]; then
        ((success++))
    fi
    echo "Mixed $i: $result"
done
echo "Result: $success/$total ($(( success * 100 / total ))%)"
echo ""

echo "=== Stress Test Complete ==="
