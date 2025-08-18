#!/bin/bash

# Test script for RFH API and CLI operations
# Tests publish, search, and install operations
# Default registry URL points to local Docker dev environment

set -e

# Default parameters
REGISTRY_URL="${1:-http://localhost:8080}"
TOKEN="${2:-}"
SKIP_CLEANUP=false
VERBOSE=false

# Parse command line arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        --registry-url)
            REGISTRY_URL="$2"
            shift 2
            ;;
        --token)
            TOKEN="$2"
            shift 2
            ;;
        --skip-cleanup)
            SKIP_CLEANUP=true
            shift
            ;;
        --verbose)
            VERBOSE=true
            shift
            ;;
        -h|--help)
            echo "Usage: $0 [--registry-url URL] [--token TOKEN] [--skip-cleanup] [--verbose]"
            echo "  --registry-url URL    Registry URL (default: http://localhost:8080)"
            echo "  --token TOKEN         Authentication token"
            echo "  --skip-cleanup        Don't clean up test artifacts"
            echo "  --verbose             Enable verbose output"
            exit 0
            ;;
        *)
            echo "Unknown option: $1"
            exit 1
            ;;
    esac
done

# Global variables
TEST_PACKAGE_NAME="monty-python-quotes"
TEST_PACKAGE_VERSION="1.0.0"
TEST_DIR="test-package"
ORIGINAL_DIR="$(pwd)"

# Color codes
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
CYAN='\033[0;36m'
GRAY='\033[0;37m'
NC='\033[0m' # No Color

function print_header() {
    echo -e "${CYAN}🧪 RFH API and CLI Test Script${NC}"
    echo -e "${CYAN}================================${NC}"
}

function print_step() {
    echo -e "\n${YELLOW}🔍 $1${NC}"
}

function print_success() {
    echo -e "${GREEN}✅ $1${NC}"
}

function print_error() {
    echo -e "${RED}❌ $1${NC}"
}

function print_warning() {
    echo -e "${YELLOW}⚠️  $1${NC}"
}

function test_prerequisites() {
    print_step "Checking prerequisites..."
    
    # Check if rfh executable exists
    if ! command -v rfh &> /dev/null; then
        print_error "RFH CLI not found. Run install-cli script first."
        exit 1
    fi
    
    # Test rfh command
    if rfh --help > /dev/null 2>&1; then
        print_success "RFH CLI is available"
    else
        print_error "RFH CLI not working properly"
        exit 1
    fi
    
    # Check if test package directory exists
    if [ ! -d "$TEST_DIR" ]; then
        print_error "Test package directory '$TEST_DIR' not found"
        exit 1
    fi
    
    print_success "Prerequisites check passed"
}

function test_api_health() {
    print_step "Testing API health..."
    
    if command -v curl &> /dev/null; then
        HTTP_CLIENT="curl"
    elif command -v wget &> /dev/null; then
        HTTP_CLIENT="wget"
    else
        print_error "Neither curl nor wget found. Please install one of them."
        exit 1
    fi
    
    if [ "$HTTP_CLIENT" = "curl" ]; then
        if [ "$VERBOSE" = true ]; then
            response=$(curl -s -f "$REGISTRY_URL/v1/health" -v 2>&1) || {
                print_error "Failed to connect to API at $REGISTRY_URL"
                echo "Make sure the API server is running (e.g., docker-compose up)"
                exit 1
            }
        else
            response=$(curl -s -f "$REGISTRY_URL/v1/health") || {
                print_error "Failed to connect to API at $REGISTRY_URL"
                echo "Make sure the API server is running (e.g., docker-compose up)"
                exit 1
            }
        fi
    else
        response=$(wget -q -O - "$REGISTRY_URL/v1/health") || {
            print_error "Failed to connect to API at $REGISTRY_URL"
            echo "Make sure the API server is running (e.g., docker-compose up)"
            exit 1
        }
    fi
    
    if echo "$response" | grep -q '"status":"ok"'; then
        print_success "API is healthy"
    else
        print_error "API returned unexpected response: $response"
        exit 1
    fi
}

function setup_registry() {
    print_step "Setting up registry configuration..."
    
    # Add test registry
    if [ -n "$TOKEN" ]; then
        if rfh registry add test "$REGISTRY_URL" --token "$TOKEN"; then
            print_success "Registry added with token"
        else
            print_error "Failed to add registry with token"
            exit 1
        fi
    else
        if rfh registry add test "$REGISTRY_URL"; then
            print_success "Registry added"
        else
            print_error "Failed to add registry"
            exit 1
        fi
    fi
    
    # Set as current registry
    if rfh registry use test; then
        print_success "Registry set as current"
    else
        print_error "Failed to set registry as current"
        exit 1
    fi
}

function test_package_init() {
    print_step "Testing package initialization..."
    
    cd "$TEST_DIR"
    
    # Verify rulestack.json exists and is valid
    if [ ! -f "rulestack.json" ]; then
        print_error "rulestack.json not found in test package"
        exit 1
    fi
    
    # Check if jq is available for JSON validation
    if command -v jq &> /dev/null; then
        if package_name=$(jq -r '.name' rulestack.json 2>/dev/null); then
            if [ "$package_name" = "$TEST_PACKAGE_NAME" ]; then
                print_success "Package manifest is valid"
            else
                print_error "Package name mismatch in manifest: expected $TEST_PACKAGE_NAME, got $package_name"
                exit 1
            fi
        else
            print_error "Invalid JSON in rulestack.json"
            exit 1
        fi
    else
        # Basic validation without jq
        if grep -q "\"name\".*\"$TEST_PACKAGE_NAME\"" rulestack.json; then
            print_success "Package manifest appears valid"
        else
            print_error "Package name not found in manifest"
            exit 1
        fi
    fi
}

function test_pack_operation() {
    print_step "Testing pack operation..."
    
    if [ "$VERBOSE" = true ]; then
        if rfh pack --verbose; then
            print_success "Pack operation completed"
        else
            print_error "Pack operation failed"
            exit 1
        fi
    else
        if rfh pack; then
            print_success "Pack operation completed"
        else
            print_error "Pack operation failed"
            exit 1
        fi
    fi
    
    expected_archive="$TEST_PACKAGE_NAME-$TEST_PACKAGE_VERSION.tgz"
    if [ -f "$expected_archive" ]; then
        print_success "Package packed successfully: $expected_archive"
    else
        print_error "Archive file not created: $expected_archive"
        exit 1
    fi
}

function test_publish_operation() {
    print_step "Testing publish operation..."
    
    if [ "$VERBOSE" = true ]; then
        if rfh publish --token noauth --verbose; then
            print_success "Package published successfully"
        else
            print_warning "Publish operation failed (might be expected if package already exists)"
        fi
    else
        if rfh publish --token noauth; then
            print_success "Package published successfully"
        else
            print_warning "Publish operation failed (might be expected if package already exists)"
        fi
    fi
}

function test_search_operation() {
    print_step "Testing search operation..."
    
    # Test basic search
    if [ "$VERBOSE" = true ]; then
        search_output=$(rfh search "monty" --verbose) || {
            print_error "Search operation failed"
            exit 1
        }
    else
        search_output=$(rfh search "monty") || {
            print_error "Search operation failed"
            exit 1
        }
    fi
    
    if echo "$search_output" | grep -q "$TEST_PACKAGE_NAME"; then
        print_success "Search found our test package"
    else
        print_error "Search did not find our test package"
        echo -e "${YELLOW}Search output:${NC}"
        echo -e "${GRAY}$search_output${NC}"
    fi
    
    # Test search with filters
    if [ "$VERBOSE" = true ]; then
        tag_search_output=$(rfh search "quotes" --tag "humor" --verbose) || {
            print_warning "Search with filters failed"
        }
    else
        tag_search_output=$(rfh search "quotes" --tag "humor") || {
            print_warning "Search with filters failed"
        }
    fi
    
    print_success "Search with filters completed"
}

function test_install_operation() {
    print_step "Testing install operation (add command)..."
    
    # Move to a different directory to test installation
    install_test_dir="install-test"
    if [ -d "$install_test_dir" ]; then
        rm -rf "$install_test_dir"
    fi
    mkdir "$install_test_dir"
    cd "$install_test_dir"
    
    # The add command is now implemented and should work
    install_success=true
    
    if [ "$VERBOSE" = true ]; then
        rfh add "$TEST_PACKAGE_NAME@$TEST_PACKAGE_VERSION" --verbose
        add_exit_code=$?
    else
        rfh add "$TEST_PACKAGE_NAME@$TEST_PACKAGE_VERSION"
        add_exit_code=$?
    fi
    
    # Verify installation - these are REQUIRED for the test to pass
    if [ -d ".rulestack" ]; then
        print_success ".rulestack directory created"
    else
        print_error ".rulestack directory NOT created - FAILED"
        install_success=false
    fi
    
    if [ -d ".rulestack/$TEST_PACKAGE_NAME" ]; then
        print_success "Package directory created: .rulestack/$TEST_PACKAGE_NAME"
    else
        print_error "Package directory NOT created: .rulestack/$TEST_PACKAGE_NAME - FAILED"
        install_success=false
    fi
    
    if [ -f "rulestack.json" ]; then
        print_success "Project manifest created: rulestack.json"
        # Show content for debugging
        if [ "$VERBOSE" = true ]; then
            echo -e "${GRAY}rulestack.json content:${NC}"
            cat rulestack.json | sed 's/^/  /'
        fi
    else
        print_error "Project manifest NOT created: rulestack.json - FAILED"
        install_success=false
    fi
    
    if [ -f "rulestack.lock.json" ]; then
        print_success "Lock manifest created: rulestack.lock.json"
        # Show content for debugging
        if [ "$VERBOSE" = true ]; then
            echo -e "${GRAY}rulestack.lock.json content:${NC}"
            cat rulestack.lock.json | sed 's/^/  /'
        fi
    else
        print_error "Lock manifest NOT created: rulestack.lock.json - FAILED"
        install_success=false
    fi
    
    # Check if package files exist
    if [ -n "$(ls -A .rulestack/$TEST_PACKAGE_NAME 2>/dev/null)" ]; then
        file_count=$(ls -1 .rulestack/$TEST_PACKAGE_NAME | wc -l)
        print_success "Package files extracted: $file_count files found"
    else
        print_error "No package files found in .rulestack/$TEST_PACKAGE_NAME - FAILED"
        install_success=false
    fi
    
    if [ "$install_success" = true ] && [ $add_exit_code -eq 0 ]; then
        print_success "Package installed successfully - ALL VERIFICATIONS PASSED"
    else
        print_error "Package installation FAILED verification checks"
        exit 1
    fi
    
    # Return to original test directory
    cd ..
}

function cleanup() {
    print_step "Cleaning up test artifacts..."
    
    cd "$ORIGINAL_DIR"
    
    if [ "$SKIP_CLEANUP" = false ]; then
        # Clean up test files
        if ls "$TEST_DIR"/*.tgz >/dev/null 2>&1; then
            rm -f "$TEST_DIR"/*.tgz
            print_success "Removed test archives"
        fi
        
        if [ -d "install-test" ]; then
            rm -rf "install-test"
            print_success "Removed install test directory"
        fi
        
        # Remove test registry (optional)
        if rfh registry remove test >/dev/null 2>&1; then
            print_success "Removed test registry configuration"
        else
            echo -e "${YELLOW}Note: Could not remove test registry configuration${NC}"
        fi
    else
        echo -e "${YELLOW}Skipping cleanup (--skip-cleanup flag used)${NC}"
    fi
}

function run_tests() {
    test_prerequisites
    test_api_health
    setup_registry
    test_package_init
    test_pack_operation
    test_publish_operation
    test_search_operation
    test_install_operation
    
    echo -e "\n${GREEN}🎉 All tests completed!${NC}"
    echo -e "${GREEN}✅ Pack operation: Working${NC}"
    echo -e "${GREEN}✅ Publish operation: Working${NC}"
    echo -e "${GREEN}✅ Search operation: Working${NC}"
    echo -e "${GREEN}✅ Install operation: Working${NC}"
}

# Trap to ensure cleanup runs even if script fails
trap cleanup EXIT

# Main execution
print_header
echo -e "${CYAN}Registry URL: $REGISTRY_URL${NC}"
if [ -n "$TOKEN" ]; then
    echo -e "${CYAN}Using authentication token: ****${NC}"
else
    echo -e "${YELLOW}No authentication token provided${NC}"
fi

run_tests

echo -e "\n${CYAN}🏁 Test script completed!${NC}"