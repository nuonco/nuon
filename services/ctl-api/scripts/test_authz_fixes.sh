#!/bin/bash
#
# Authorization Bypass Fix Verification Script
# Tests PRO-2000 through PRO-2013 fixes
#
# This script attempts to access Org A's resources using Org B's credentials.
# EXPECTED RESULT: All requests should return 404 or 403 (access denied)
# VULNERABILITY: If any request returns 200 with data, the fix is incomplete
#

set -euo pipefail

#######################################
# CONFIGURATION - Fill these in before running
#######################################

# API endpoint (local dev or staging)
API_BASE_URL="${API_BASE_URL:-http://localhost:8081}"

# Org A - Owner of test entities (victim)
ORG_A_ID="${ORG_A_ID:-}"
ORG_A_API_KEY="${ORG_A_API_KEY:-}"

# Org B - Attacker trying to access Org A's data
ORG_B_ID="${ORG_B_ID:-}"
ORG_B_API_KEY="${ORG_B_API_KEY:-}"

# Test entity IDs from Org A (to be accessed by Org B)
TEST_APP_ID="${TEST_APP_ID:-}"
TEST_INSTALL_ID="${TEST_INSTALL_ID:-}"
TEST_WORKFLOW_ID="${TEST_WORKFLOW_ID:-}"
TEST_WORKFLOW_STEP_ID="${TEST_WORKFLOW_STEP_ID:-}"
TEST_COMPONENT_ID="${TEST_COMPONENT_ID:-}"
TEST_APP_CONFIG_ID="${TEST_APP_CONFIG_ID:-}"
TEST_ACTION_CONFIG_ID="${TEST_ACTION_CONFIG_ID:-}"

#######################################
# Colors and formatting
#######################################
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

#######################################
# Validation
#######################################
validate_config() {
    local missing=()
    
    [[ -z "$ORG_A_ID" ]] && missing+=("ORG_A_ID")
    [[ -z "$ORG_A_API_KEY" ]] && missing+=("ORG_A_API_KEY")
    [[ -z "$ORG_B_ID" ]] && missing+=("ORG_B_ID")
    [[ -z "$ORG_B_API_KEY" ]] && missing+=("ORG_B_API_KEY")
    [[ -z "$TEST_APP_ID" ]] && missing+=("TEST_APP_ID")
    [[ -z "$TEST_INSTALL_ID" ]] && missing+=("TEST_INSTALL_ID")
    
    if [[ ${#missing[@]} -gt 0 ]]; then
        echo -e "${RED}ERROR: Missing required configuration:${NC}"
        printf '  - %s\n' "${missing[@]}"
        echo ""
        echo "Set these as environment variables or edit this script."
        echo ""
        echo "Example:"
        echo "  export ORG_A_ID='org-xxx'"
        echo "  export ORG_A_API_KEY='api-key-xxx'"
        echo "  export ORG_B_ID='org-yyy'"
        echo "  export ORG_B_API_KEY='api-key-yyy'"
        echo "  export TEST_APP_ID='app-xxx'"
        echo "  export TEST_INSTALL_ID='install-xxx'"
        echo "  ./test_authz_fixes.sh"
        exit 1
    fi
}

#######################################
# Test execution
#######################################
TOTAL_TESTS=0
PASSED_TESTS=0
FAILED_TESTS=0
SKIPPED_TESTS=0

# Make a request using Org B's credentials to access Org A's resource
# Args: $1=method, $2=endpoint, $3=expected_behavior ("block" or "allow")
test_endpoint() {
    local method="$1"
    local endpoint="$2"
    local test_name="$3"
    local issue_id="$4"
    local expected="${5:-block}"
    
    TOTAL_TESTS=$((TOTAL_TESTS + 1))
    
    echo -e "${BLUE}Testing:${NC} $test_name"
    echo -e "  Issue: $issue_id"
    echo -e "  $method $endpoint"
    
    # Make request with Org B credentials to Org A's resource
    local response
    local http_code
    
    response=$(curl -s -w "\n%{http_code}" \
        -X "$method" \
        -H "Authorization: Bearer $ORG_B_API_KEY" \
        -H "X-Nuon-Org-ID: $ORG_B_ID" \
        -H "Content-Type: application/json" \
        "${API_BASE_URL}${endpoint}" 2>&1) || true
    
    http_code=$(echo "$response" | tail -n1)
    local body=$(echo "$response" | sed '$d')
    
    echo -e "  Response: HTTP $http_code"
    
    if [[ "$expected" == "block" ]]; then
        # We expect 404, 403, or 401 (access denied)
        if [[ "$http_code" =~ ^(404|403|401)$ ]]; then
            echo -e "  ${GREEN}✓ PASS${NC} - Access correctly blocked (HTTP $http_code)"
            PASSED_TESTS=$((PASSED_TESTS + 1))
        elif [[ "$http_code" == "200" ]]; then
            echo -e "  ${RED}✗ FAIL${NC} - VULNERABILITY! Cross-org access allowed"
            echo -e "  ${RED}  Response body (truncated):${NC}"
            echo "$body" | head -c 200
            echo ""
            FAILED_TESTS=$((FAILED_TESTS + 1))
        else
            echo -e "  ${YELLOW}? UNKNOWN${NC} - Unexpected response code"
            echo -e "  Response: $body" | head -c 200
            FAILED_TESTS=$((FAILED_TESTS + 1))
        fi
    fi
    
    echo ""
}

# Skip test if required entity ID is missing
skip_if_missing() {
    local var_name="$1"
    local var_value="$2"
    local test_name="$3"
    
    if [[ -z "$var_value" ]]; then
        TOTAL_TESTS=$((TOTAL_TESTS + 1))
        SKIPPED_TESTS=$((SKIPPED_TESTS + 1))
        echo -e "${YELLOW}SKIP:${NC} $test_name"
        echo -e "  Missing: $var_name"
        echo ""
        return 1
    fi
    return 0
}

#######################################
# First, verify Org A can access its own resources
#######################################
verify_org_a_access() {
    echo -e "${BLUE}═══════════════════════════════════════════════════════════${NC}"
    echo -e "${BLUE}STEP 1: Verify Org A can access its own resources${NC}"
    echo -e "${BLUE}═══════════════════════════════════════════════════════════${NC}"
    echo ""
    
    # Test app access
    echo -e "Testing Org A access to its own app..."
    local response
    response=$(curl -s -w "\n%{http_code}" \
        -X GET \
        -H "Authorization: Bearer $ORG_A_API_KEY" \
        -H "X-Nuon-Org-ID: $ORG_A_ID" \
        "${API_BASE_URL}/v1/apps/${TEST_APP_ID}" 2>&1) || true
    
    local http_code=$(echo "$response" | tail -n1)
    
    if [[ "$http_code" == "200" ]]; then
        echo -e "  ${GREEN}✓ Org A can access its app${NC}"
    else
        echo -e "  ${RED}✗ Org A cannot access its own app (HTTP $http_code)${NC}"
        echo -e "  ${RED}  Check that TEST_APP_ID belongs to ORG_A_ID${NC}"
        exit 1
    fi
    
    # Test install access
    echo -e "Testing Org A access to its own install..."
    response=$(curl -s -w "\n%{http_code}" \
        -X GET \
        -H "Authorization: Bearer $ORG_A_API_KEY" \
        -H "X-Nuon-Org-ID: $ORG_A_ID" \
        "${API_BASE_URL}/v1/installs/${TEST_INSTALL_ID}" 2>&1) || true
    
    http_code=$(echo "$response" | tail -n1)
    
    if [[ "$http_code" == "200" ]]; then
        echo -e "  ${GREEN}✓ Org A can access its install${NC}"
    else
        echo -e "  ${RED}✗ Org A cannot access its own install (HTTP $http_code)${NC}"
        echo -e "  ${RED}  Check that TEST_INSTALL_ID belongs to ORG_A_ID${NC}"
        exit 1
    fi
    
    echo ""
    echo -e "${GREEN}Org A access verified. Proceeding with cross-org tests...${NC}"
    echo ""
}

#######################################
# Main test execution
#######################################
run_tests() {
    echo -e "${BLUE}═══════════════════════════════════════════════════════════${NC}"
    echo -e "${BLUE}STEP 2: Test Cross-Org Access (Org B → Org A resources)${NC}"
    echo -e "${BLUE}═══════════════════════════════════════════════════════════${NC}"
    echo ""
    echo -e "Using Org B credentials to access Org A resources."
    echo -e "All tests should ${GREEN}BLOCK${NC} access (return 404/403)."
    echo ""
    
    # PRO-2000: GetApp / findApp ungrouped OR
    test_endpoint "GET" "/v1/apps/${TEST_APP_ID}" \
        "GetApp - ungrouped OR bypass" "PRO-2000"
    
    # PRO-2000 also affects GetAppRunnerLatestConfig
    test_endpoint "GET" "/v1/apps/${TEST_APP_ID}/runner-latest-config" \
        "GetAppRunnerLatestConfig - ungrouped OR bypass" "PRO-2000"
    
    # PRO-2001: GetInstall helper ungrouped OR
    test_endpoint "GET" "/v1/installs/${TEST_INSTALL_ID}" \
        "GetInstall (helper) - ungrouped OR bypass" "PRO-2001"
    
    # PRO-2002: GetAppSecrets missing org check
    test_endpoint "GET" "/v1/apps/${TEST_APP_ID}/secrets" \
        "GetAppSecrets - missing org_id check" "PRO-2002"
    
    # PRO-2003: GetInstall service ungrouped OR
    test_endpoint "GET" "/v1/installs/${TEST_INSTALL_ID}" \
        "GetInstall (service) - ungrouped OR bypass" "PRO-2003"
    
    # PRO-2004: GetAppSandboxConfigs ungrouped OR
    test_endpoint "GET" "/v1/apps/${TEST_APP_ID}/sandbox-configs" \
        "GetAppSandboxConfigs - ungrouped OR bypass" "PRO-2004"
    
    # PRO-2005: GetWorkflow missing org_id
    if skip_if_missing "TEST_WORKFLOW_ID" "$TEST_WORKFLOW_ID" "GetWorkflow - missing org_id"; then
        test_endpoint "GET" "/v1/workflows/${TEST_WORKFLOW_ID}" \
            "GetWorkflow - missing org_id check" "PRO-2005"
    fi
    
    # PRO-2006: GetAppRunnerConfigs ungrouped OR
    test_endpoint "GET" "/v1/apps/${TEST_APP_ID}/runner-configs" \
        "GetAppRunnerConfigs - ungrouped OR bypass" "PRO-2006"
    
    # PRO-2007: GetAppInputConfigs ungrouped OR
    test_endpoint "GET" "/v1/apps/${TEST_APP_ID}/input-configs" \
        "GetAppInputConfigs - ungrouped OR bypass" "PRO-2007"
    
    # PRO-2008: GetWorkflowStep missing org_id
    if skip_if_missing "TEST_WORKFLOW_ID" "$TEST_WORKFLOW_ID" "GetWorkflowStep"; then
        if skip_if_missing "TEST_WORKFLOW_STEP_ID" "$TEST_WORKFLOW_STEP_ID" "GetWorkflowStep"; then
            test_endpoint "GET" "/v1/workflows/${TEST_WORKFLOW_ID}/steps/${TEST_WORKFLOW_STEP_ID}" \
                "GetWorkflowStep - missing org_id check" "PRO-2008"
        fi
    fi
    
    # PRO-2009: GetAppComponentLatestConfig missing app validation
    if skip_if_missing "TEST_COMPONENT_ID" "$TEST_COMPONENT_ID" "GetAppComponentLatestConfig"; then
        test_endpoint "GET" "/v1/apps/${TEST_APP_ID}/components/${TEST_COMPONENT_ID}/configs/latest" \
            "GetAppComponentLatestConfig - missing app_id validation" "PRO-2009"
    fi
    
    # PRO-2010: GetAppConfig recurse=true bypass
    if skip_if_missing "TEST_APP_CONFIG_ID" "$TEST_APP_CONFIG_ID" "GetAppConfig recurse"; then
        test_endpoint "GET" "/v1/apps/${TEST_APP_ID}/configs/${TEST_APP_CONFIG_ID}?recurse=true" \
            "GetAppConfig recurse=true - org bypass" "PRO-2010"
    fi
    
    # PRO-2011: GetInstallEvents missing org_id
    test_endpoint "GET" "/v1/installs/${TEST_INSTALL_ID}/events" \
        "GetInstallEvents - missing org_id check" "PRO-2011"
    
    # PRO-2012: GetAppActionConfig missing org validation
    if skip_if_missing "TEST_ACTION_CONFIG_ID" "$TEST_ACTION_CONFIG_ID" "GetAppActionConfig"; then
        test_endpoint "GET" "/v1/apps/${TEST_APP_ID}/actions/configs/${TEST_ACTION_CONFIG_ID}" \
            "GetAppActionConfig - missing org_id check" "PRO-2012"
    fi
    
    # PRO-2013: GetAppInstalls missing org check on app
    test_endpoint "GET" "/v1/apps/${TEST_APP_ID}/installs" \
        "GetAppInstalls - missing org ownership check" "PRO-2013"
}

#######################################
# Summary
#######################################
print_summary() {
    echo -e "${BLUE}═══════════════════════════════════════════════════════════${NC}"
    echo -e "${BLUE}TEST SUMMARY${NC}"
    echo -e "${BLUE}═══════════════════════════════════════════════════════════${NC}"
    echo ""
    echo -e "Total tests:   $TOTAL_TESTS"
    echo -e "${GREEN}Passed:        $PASSED_TESTS${NC}"
    echo -e "${RED}Failed:        $FAILED_TESTS${NC}"
    echo -e "${YELLOW}Skipped:       $SKIPPED_TESTS${NC}"
    echo ""
    
    if [[ $FAILED_TESTS -eq 0 && $PASSED_TESTS -gt 0 ]]; then
        echo -e "${GREEN}═══════════════════════════════════════════════════════════${NC}"
        echo -e "${GREEN}ALL TESTS PASSED! Authorization fixes verified.${NC}"
        echo -e "${GREEN}═══════════════════════════════════════════════════════════${NC}"
        exit 0
    elif [[ $FAILED_TESTS -gt 0 ]]; then
        echo -e "${RED}═══════════════════════════════════════════════════════════${NC}"
        echo -e "${RED}SOME TESTS FAILED! Vulnerabilities may still exist.${NC}"
        echo -e "${RED}═══════════════════════════════════════════════════════════${NC}"
        exit 1
    else
        echo -e "${YELLOW}═══════════════════════════════════════════════════════════${NC}"
        echo -e "${YELLOW}NO TESTS RUN. Check configuration.${NC}"
        echo -e "${YELLOW}═══════════════════════════════════════════════════════════${NC}"
        exit 1
    fi
}

#######################################
# Main
#######################################
main() {
    echo ""
    echo -e "${BLUE}═══════════════════════════════════════════════════════════${NC}"
    echo -e "${BLUE}Authorization Bypass Fix Verification Script${NC}"
    echo -e "${BLUE}Testing PRO-2000 through PRO-2013${NC}"
    echo -e "${BLUE}═══════════════════════════════════════════════════════════${NC}"
    echo ""
    echo "API Base URL: $API_BASE_URL"
    echo "Org A (victim): $ORG_A_ID"
    echo "Org B (attacker): $ORG_B_ID"
    echo ""
    
    validate_config
    verify_org_a_access
    run_tests
    print_summary
}

main "$@"
