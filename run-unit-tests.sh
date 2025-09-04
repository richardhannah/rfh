#!/bin/bash

# RFH Unit Test Runner
# This script runs Go unit tests for the RFH (RuleStack) project

set -e

# Default values
COVERAGE=false
VERBOSE=false
RACE=false
PACKAGE="./..."
TEST_NAME=""
SHORT=false

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Function to print colored output
print_color() {
    local color=$1
    local message=$2
    echo -e "${color}${message}${NC}"
}

# Function to show usage
show_usage() {
    echo "Usage: $0 [OPTIONS]"
    echo ""
    echo "Run unit tests for the RFH project with various options."
    echo ""
    echo "Options:"
    echo "  -c, --coverage      Generate test coverage report"
    echo "  -v, --verbose       Enable verbose test output"
    echo "  -r, --race          Run tests with race detection"
    echo "  -p, --package PKG   Run tests for specific package (default: ./...)"
    echo "  -t, --test NAME     Run specific test by name pattern"
    echo "  -s, --short         Run tests in short mode (skip long-running tests)"
    echo "  -h, --help          Show this help message"
    echo ""
    echo "Examples:"
    echo "  $0                                    # Run all unit tests"
    echo "  $0 --coverage                        # Run tests with coverage report"
    echo "  $0 --package ./internal/client -v    # Run client package tests with verbose output"
    echo "  $0 --test 'TestHTTPClient*' --race   # Run specific test pattern with race detection"
}

# Parse command line arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        -c|--coverage)
            COVERAGE=true
            shift
            ;;
        -v|--verbose)
            VERBOSE=true
            shift
            ;;
        -r|--race)
            RACE=true
            shift
            ;;
        -p|--package)
            PACKAGE="$2"
            shift 2
            ;;
        -t|--test)
            TEST_NAME="$2"
            shift 2
            ;;
        -s|--short)
            SHORT=true
            shift
            ;;
        -h|--help)
            show_usage
            exit 0
            ;;
        *)
            echo "Unknown option: $1"
            show_usage
            exit 1
            ;;
    esac
done

# Check if Go is installed
if ! command -v go &> /dev/null; then
    print_color $RED "Error: Go is not installed or not in PATH"
    exit 1
fi

# Display Go version
GO_VERSION=$(go version)
print_color $BLUE "Using: $GO_VERSION"

# Check if we're in the correct directory
if [ ! -f "go.mod" ]; then
    print_color $RED "Error: Please run this script from the RFH root directory (where go.mod exists)"
    exit 1
fi

# Build test command
TEST_ARGS=("test" "$PACKAGE")

# Add test flags
if [ "$COVERAGE" = true ]; then
    TEST_ARGS+=("-cover" "-coverprofile=coverage.out")
fi

if [ "$VERBOSE" = true ]; then
    TEST_ARGS+=("-v")
fi

if [ "$RACE" = true ]; then
    TEST_ARGS+=("-race")
fi

if [ "$SHORT" = true ]; then
    TEST_ARGS+=("-short")
fi

if [ -n "$TEST_NAME" ]; then
    TEST_ARGS+=("-run" "$TEST_NAME")
fi

# Always add timeout to prevent hanging tests
TEST_ARGS+=("-timeout" "30s")

# Display what we're running
print_color $BLUE "\\n🧪 Running Go unit tests..."
print_color $YELLOW "Command: go ${TEST_ARGS[*]}"

# Run the tests
echo ""
start_time=$(date +%s.%N)
if go "${TEST_ARGS[@]}"; then
    exit_code=0
else
    exit_code=$?
fi
end_time=$(date +%s.%N)

# Calculate duration (using awk since bc may not be available)
duration=$(awk "BEGIN {print $end_time - $start_time}")

# Display results
echo ""
if [ $exit_code -eq 0 ]; then
    print_color $GREEN "✅ Tests passed in $(printf "%.2f" $duration) seconds"
    
    # If coverage was requested, show coverage report
    if [ "$COVERAGE" = true ] && [ -f "coverage.out" ]; then
        print_color $BLUE "\\n📊 Coverage Report:"
        go tool cover -func=coverage.out
        
        # Generate HTML coverage report
        print_color $BLUE "\\nGenerating HTML coverage report..."
        if go tool cover -html=coverage.out -o=coverage.html; then
            print_color $GREEN "HTML coverage report generated: coverage.html"
        fi
    fi
else
    print_color $RED "❌ Tests failed after $(printf "%.2f" $duration) seconds"
    print_color $RED "Exit code: $exit_code"
fi

# Clean up
if [ "$COVERAGE" = true ] && [ -f "coverage.out" ]; then
    print_color $YELLOW "\\nCoverage file: coverage.out"
fi

exit $exit_code