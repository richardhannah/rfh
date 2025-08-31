#!/bin/bash
# RFH Cucumber Test Runner
# This script runs the Cucumber BDD tests for RFH

set -e

echo "🧪 RFH Cucumber Test Runner"
echo "=========================="

# Default to fail-fast enabled
FAIL_FAST_FLAG="--fail-fast"
FAIL_FAST_ENABLED=true

# Check if we're in the correct directory
if [ ! -f "go.mod" ] || [ ! -d "cucumber-testing" ]; then
    echo "❌ Error: Please run this script from the RFH root directory"
    exit 1
fi

# Check if Node.js is available
if ! command -v node &> /dev/null; then
    echo "❌ Error: Node.js is not installed or not in PATH"
    echo "   Please install Node.js to run Cucumber tests"
    exit 1
fi

# Check if npm is available
if ! command -v npm &> /dev/null; then
    echo "❌ Error: npm is not installed or not in PATH"
    exit 1
fi

# Check if Docker is available and running
if command -v docker &> /dev/null; then
    if docker info &> /dev/null; then
        echo "🐳 Docker is available - managing test infrastructure..."
        
        # Clean up any existing test containers and start fresh
        echo "   Cleaning up existing test containers..."
        docker-compose -f docker-compose.test.yml down -v &> /dev/null || true
        
        # Build and start test infrastructure
        echo "   Building and starting test API server..."
        docker-compose -f docker-compose.test.yml up --build -d &> /dev/null
        
        # Wait for API to be healthy
        echo -n "   Waiting for test API to be ready"
        for i in {1..60}; do
            if curl -s http://localhost:8081/v1/health &> /dev/null; then
                echo " ✅"
                break
            fi
            echo -n "."
            sleep 1
        done
        
        # Check if API is responding
        if ! curl -s http://localhost:8081/v1/health &> /dev/null; then
            echo ""
            echo "   ⚠️  Warning: API server not responding - some tests may fail"
        fi
    else
        echo "⚠️  Docker is installed but not running - skipping API server setup"
        echo "   Some registry/auth tests may fail without the API server"
    fi
else
    echo "⚠️  Docker not found - skipping API server setup"
    echo "   Some registry/auth tests may fail without the API server"
fi

# Navigate to cucumber testing directory
cd cucumber-testing

# Check if dependencies are installed
if [ ! -d "node_modules" ]; then
    echo "📦 Installing test dependencies..."
    npm install
fi

echo ""
echo "🏗️  Building RFH binary..."
cd ..
go build -o dist/rfh.exe ./cmd/cli
echo "✅ RFH binary built successfully"

cd cucumber-testing

echo ""
echo "🧪 Running Cucumber tests..."
echo ""

# Parse command line arguments for fail-fast option
for arg in "$@"; do
    if [[ "$arg" == "-fail-fast=false" ]]; then
        FAIL_FAST_FLAG=""
        FAIL_FAST_ENABLED=false
        shift
    fi
done

# Display fail-fast status
if [ "$FAIL_FAST_ENABLED" = true ]; then
    echo "⚡ Fail-fast mode: ENABLED (stop on first failure)"
else
    echo "🔄 Fail-fast mode: DISABLED (run full test suite)"
fi
echo ""

# Parse command line arguments for test target
TEST_PATTERN=""
case "${1:-all}" in
    "init")
        echo "🎯 Running rfh init tests only..."
        TEST_PATTERN="features/01-init-*.feature"
        ;;
    "actual")
        echo "🎯 Running actual behavior tests only..."
        TEST_PATTERN="features/01-init-empty-directory.feature"
        ;;
    "working")
        echo "🎯 Running only working/passing scenarios..."
        npm run test:working
        exit 0
        ;;
    "all"|"")
        echo "🎯 Running all available tests..."
        TEST_PATTERN="features/*.feature"
        ;;
    *)
        echo "❌ Unknown test target: $1"
        echo "Usage: $0 [init|actual|working|all] [-fail-fast=false]"
        echo "  init    - Run rfh init tests only"
        echo "  actual  - Run actual behavior tests only" 
        echo "  working - Run only passing scenarios"
        echo "  all     - Run all tests (default)"
        echo ""
        echo "Options:"
        echo "  -fail-fast=false - Disable fail-fast (run full suite even on failures)"
        echo "                    Default: fail-fast enabled (stop on first failure)"
        exit 1
        ;;
esac

# Run the tests
if [ -n "$TEST_PATTERN" ]; then
    npx cucumber-js $TEST_PATTERN --format progress $FAIL_FAST_FLAG
else
    npm test
fi

# Store exit code immediately after tests
TEST_EXIT_CODE=$?

echo ""
echo "✅ Test execution completed!"
echo ""
echo "📊 Test Results Summary:"
echo "   - Check output above for pass/fail status"
echo "   - Detailed JSON report available in cucumber-report.json"
echo ""
echo "💡 Tips:"
echo "   - Run './run-tests.sh actual' for only working scenarios"
echo "   - Run './run-tests.sh init' for all init-related tests"
echo "   - Failed tests may indicate missing RFH features"

# Clean up test infrastructure
if command -v docker &> /dev/null && docker info &> /dev/null; then
    echo ""
    echo "🧹 Cleaning up test infrastructure..."
    docker-compose -f docker-compose.test.yml down -v &> /dev/null || true
    echo "   Test containers cleaned up ✅"
fi

exit $TEST_EXIT_CODE