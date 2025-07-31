#!/bin/bash

# Simple test script to verify UI changes
set -e

echo "=== Simple UI Changes Test ==="

# Configuration
SERVER_PORT=8080
BASE_URL="http://localhost:${SERVER_PORT}"

# Function to start server
start_server() {
    echo "Starting server..."
    DATABASE_URL=postgres://postgres:password@localhost:5433/pepo_dev?sslmode=disable PORT=${SERVER_PORT} ./bin/pepo &
    SERVER_PID=$!
    sleep 3

    # Check if server is running
    if ! curl -s "${BASE_URL}/health" > /dev/null 2>&1; then
        echo "❌ Server failed to start"
        kill $SERVER_PID 2>/dev/null || true
        exit 1
    fi
    echo "✅ Server started successfully"
}

# Function to stop server
stop_server() {
    if [ ! -z "$SERVER_PID" ]; then
        echo "Stopping server..."
        kill $SERVER_PID 2>/dev/null || true
        wait $SERVER_PID 2>/dev/null || true
    fi
}

# Trap to ensure server is stopped on exit
trap stop_server EXIT

# Start the server
start_server

echo ""
echo "=== Testing Actions Card (should NOT have timestamps) ==="

# Get actions HTML
ACTIONS_HTML=$(curl -s -H "Accept: text/html" -H "HX-Request: true" "${BASE_URL}/api/v1/actions")

if echo "$ACTIONS_HTML" | grep -q "Occurred:.*Created:"; then
    echo "❌ FAIL: Actions still contain timestamps"
    echo "Found: $(echo "$ACTIONS_HTML" | grep -o "Occurred:.*Created:[^<]*" | head -1)"
else
    echo "✅ PASS: No timestamps found in Actions"
fi

echo ""
echo "=== Testing People List (should show last action times) ==="

# Get people HTML
PEOPLE_HTML=$(curl -s -H "Accept: text/html" -H "HX-Request: true" "${BASE_URL}/api/v1/people")

if echo "$PEOPLE_HTML" | grep -q "Last action:"; then
    echo "✅ PASS: People list shows last action times"
    echo "Found: $(echo "$PEOPLE_HTML" | grep -o "Last action:[^<]*" | head -1)"
else
    echo "❌ FAIL: People list does not show last action times"
fi

# Check for highlighting of people without recent actions
if echo "$PEOPLE_HTML" | grep -q "No actions recorded\|text-red-600.*bg-red-50"; then
    echo "✅ PASS: Found highlighting for people without recent actions"
else
    echo "⚠️  WARNING: No highlighting found for people without recent actions (might not have test data)"
fi

# Check for green highlighting of recent actions
if echo "$PEOPLE_HTML" | grep -q "text-green-600"; then
    echo "✅ PASS: Found green highlighting for recent actions"
else
    echo "⚠️  WARNING: No green highlighting found (might not have recent actions)"
fi

echo ""
echo "=== Sample Output ==="
echo "Actions sample:"
echo "$ACTIONS_HTML" | head -3 | tail -1
echo ""
echo "People sample:"
echo "$PEOPLE_HTML" | head -3 | tail -1

echo ""
echo "=== Test Complete ==="
