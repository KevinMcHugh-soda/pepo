#!/bin/bash

# Test script to verify UI changes for Actions card and People list
set -e

echo "=== Testing UI Changes ==="
echo "1. Remove timestamps from Actions card"
echo "2. Add last action time to People list"
echo "3. Highlight people with no recent actions"
echo ""

# Configuration
SERVER_PORT=${PORT:-8080}
BASE_URL="http://localhost:${SERVER_PORT}"
LOGFILE="test_ui_changes.log"

# Clean up previous log
> "$LOGFILE"

# Function to log and execute curl commands
curl_test() {
    local description="$1"
    local url="$2"
    local method="${3:-GET}"
    local data="$4"

    echo "Testing: $description"
    echo "URL: $url" >> "$LOGFILE"
    echo "Method: $method" >> "$LOGFILE"

    if [ "$method" = "POST" ]; then
        curl -s -X POST \
             -H "Content-Type: application/x-www-form-urlencoded" \
             -H "HX-Request: true" \
             -d "$data" \
             "$url" >> "$LOGFILE" 2>&1
    else
        curl -s -H "Accept: text/html" \
             -H "HX-Request: true" \
             "$url" >> "$LOGFILE" 2>&1
    fi

    echo "✓ Complete"
    echo "---" >> "$LOGFILE"
}

# Start server in background
echo "Starting server..."
./bin/pepo > server.log 2>&1 &
SERVER_PID=$!

# Wait for server to start
echo "Waiting for server to start..."
sleep 3

# Check if server is running
if ! curl -s "$BASE_URL" > /dev/null 2>&1; then
    echo "❌ Server failed to start"
    kill $SERVER_PID 2>/dev/null || true
    exit 1
fi

echo "✓ Server started on $BASE_URL"
echo ""

# Test 1: Create test people
echo "=== Creating Test Data ==="
curl_test "Create Person 1 (Alice)" "$BASE_URL/api/v1/people" "POST" "name=Alice"
curl_test "Create Person 2 (Bob)" "$BASE_URL/api/v1/people" "POST" "name=Bob"
curl_test "Create Person 3 (Charlie)" "$BASE_URL/api/v1/people" "POST" "name=Charlie"

# Get people to extract IDs for actions
echo "Getting people list to extract IDs..."
PEOPLE_JSON=$(curl -s -H "Accept: application/json" "$BASE_URL/api/v1/people")
echo "People JSON: $PEOPLE_JSON" >> "$LOGFILE"

# Extract person IDs (this is a simple approach, in a real test we'd use jq)
ALICE_ID=$(echo "$PEOPLE_JSON" | grep -o '"id":"[^"]*"' | head -1 | cut -d'"' -f4)
BOB_ID=$(echo "$PEOPLE_JSON" | grep -o '"id":"[^"]*"' | head -2 | tail -1 | cut -d'"' -f4)
CHARLIE_ID=$(echo "$PEOPLE_JSON" | grep -o '"id":"[^"]*"' | head -3 | tail -1 | cut -d'"' -f4)

echo "Alice ID: $ALICE_ID" >> "$LOGFILE"
echo "Bob ID: $BOB_ID" >> "$LOGFILE"
echo "Charlie ID: $CHARLIE_ID" >> "$LOGFILE"

# Test 2: Create actions with different dates
echo ""
echo "=== Creating Actions ==="

# Recent action for Alice (today)
TODAY=$(date -u +"%Y-%m-%dT%H:%M:%S")
curl_test "Recent action for Alice" "$BASE_URL/api/v1/actions" "POST" \
    "person_id=$ALICE_ID&description=Recent positive action&valence=positive&occurred_at=$TODAY"

# Old action for Bob (2 weeks ago)
OLD_DATE=$(date -u -d '14 days ago' +"%Y-%m-%dT%H:%M:%S" 2>/dev/null || date -u -v-14d +"%Y-%m-%dT%H:%M:%S" 2>/dev/null || echo "2024-01-01T12:00:00")
curl_test "Old action for Bob" "$BASE_URL/api/v1/actions" "POST" \
    "person_id=$BOB_ID&description=Old negative action&valence=negative&occurred_at=$OLD_DATE"

# No action for Charlie (he remains without actions)

echo ""
echo "=== Testing UI Changes ==="

# Test 3: Check Actions card (should not have timestamps)
echo ""
echo "Testing Actions card format..."
curl_test "Get Actions HTML" "$BASE_URL/api/v1/actions"

# Check if timestamps are removed from actions
if grep -q "Occurred:" "$LOGFILE" || grep -q "Created:" "$LOGFILE"; then
    echo "❌ FAIL: Timestamps still present in Actions card"
    echo "Found timestamp patterns in actions response"
else
    echo "✓ PASS: Timestamps removed from Actions card"
fi

# Test 4: Check People list (should show last action times)
echo ""
echo "Testing People list format..."
curl_test "Get People HTML with last actions" "$BASE_URL/api/v1/people"

# Check if last action times are shown
if grep -q "Last action:" "$LOGFILE"; then
    echo "✓ PASS: Last action times are displayed"
else
    echo "❌ FAIL: Last action times not found in People list"
fi

# Check if highlighting for old/no actions works
if grep -q "text-red-600.*No actions recorded" "$LOGFILE" || grep -q "bg-red-50" "$LOGFILE"; then
    echo "✓ PASS: Highlighting for people without recent actions detected"
else
    echo "❌ FAIL: No highlighting detected for people without recent actions"
fi

# Check if recent actions are shown differently
if grep -q "text-green-600" "$LOGFILE"; then
    echo "✓ PASS: Recent actions are highlighted in green"
else
    echo "❌ FAIL: Recent actions highlighting not detected"
fi

# Test 5: Test main page rendering
echo ""
echo "Testing main page..."
curl_test "Get Main Page" "$BASE_URL/"

echo ""
echo "=== Test Results Summary ==="
echo "Check the detailed log in: $LOGFILE"
echo ""

# Show relevant parts of the log
echo "=== Sample Actions HTML (should not have timestamps) ==="
grep -A 10 -B 5 "border-b pb-3 mb-3" "$LOGFILE" | head -20 || echo "No action items found"

echo ""
echo "=== Sample People HTML (should show last action times) ==="
grep -A 5 -B 5 "Last action:" "$LOGFILE" | head -15 || echo "No last action info found"

echo ""
echo "=== Red highlighting for old/no actions ==="
grep -A 2 -B 2 "text-red-600\|bg-red-50" "$LOGFILE" || echo "No red highlighting found"

echo ""
echo "=== Green highlighting for recent actions ==="
grep -A 2 -B 2 "text-green-600" "$LOGFILE" || echo "No green highlighting found"

# Cleanup
echo ""
echo "Cleaning up..."
kill $SERVER_PID 2>/dev/null || true
wait $SERVER_PID 2>/dev/null || true

echo "Test completed. Check $LOGFILE for detailed output."
