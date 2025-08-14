#!/bin/bash

# Test Runner Script for Chat Application
echo "🧪 Running Chat Application Tests"
echo "================================="

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Function to print colored output
print_status() {
    if [ $2 -eq 0 ]; then
        echo -e "${GREEN}✓ $1${NC}"
    else
        echo -e "${RED}✗ $1${NC}"
    fi
}

# Test counters
TOTAL_TESTS=0
PASSED_TESTS=0

echo ""
echo "📋 Running Go Unit Tests..."
echo "----------------------------"

# Test authentication handler
echo "Testing Authentication Handler..."
if command -v go &> /dev/null; then
    go test -v auth_handler_test.go 2>/dev/null
    AUTH_RESULT=$?
    print_status "Authentication Handler Tests" $AUTH_RESULT
    if [ $AUTH_RESULT -eq 0 ]; then
        PASSED_TESTS=$((PASSED_TESTS + 1))
    fi
else
    echo -e "${YELLOW}⚠ Go not found, skipping Go tests${NC}"
fi
TOTAL_TESTS=$((TOTAL_TESTS + 1))

# Test message handler
echo ""
echo "Testing Message Handler..."
if command -v go &> /dev/null; then
    go test -v message_handler_test.go 2>/dev/null
    MESSAGE_RESULT=$?
    print_status "Message Handler Tests" $MESSAGE_RESULT
    if [ $MESSAGE_RESULT -eq 0 ]; then
        PASSED_TESTS=$((PASSED_TESTS + 1))
    fi
fi
TOTAL_TESTS=$((TOTAL_TESTS + 1))

echo ""
echo "📊 Test Summary"
echo "==============="
echo "Total Test Suites: $TOTAL_TESTS"
echo "Passed: $PASSED_TESTS"
echo "Failed: $((TOTAL_TESTS - PASSED_TESTS))"

if [ $PASSED_TESTS -eq $TOTAL_TESTS ]; then
    echo -e "${GREEN}🎉 All tests passed!${NC}"
    exit 0
else
    echo -e "${YELLOW}⚠ Some tests failed or were skipped${NC}"
    exit 1
fi
