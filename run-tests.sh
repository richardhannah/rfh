#!/bin/bash
# RFH Cucumber Test Runner
# This script runs the Cucumber BDD tests for RFH

set -e

echo "🧪 RFH Cucumber Test Runner"
echo "=========================="

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

# Parse command line arguments
TEST_PATTERN=""
case "${1:-all}" in
    "init")
        echo "🎯 Running rfh init tests only..."
        TEST_PATTERN="features/init-*.feature"
        ;;
    "actual")
        echo "🎯 Running actual behavior tests only..."
        TEST_PATTERN="features/init-actual-behavior.feature"
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
        echo "Usage: $0 [init|actual|working|all]"
        echo "  init    - Run rfh init tests only"
        echo "  actual  - Run actual behavior tests only" 
        echo "  working - Run only passing scenarios"
        echo "  all     - Run all tests (default)"
        exit 1
        ;;
esac

# Run the tests
if [ -n "$TEST_PATTERN" ]; then
    npx cucumber-js $TEST_PATTERN --format progress
else
    npm test
fi

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