#!/bin/bash

# Integration test suite for a running CDN.
# Usage: ./test.sh [port] [host]
# Defaults: port=8080, host=localhost, asset from the example project

PORT=${1:-8080}
HOST=${2:-localhost}
BASE_URL="http://$HOST:$PORT"
ASSET=${ASSET:-/karako/logos/logo-icon-black.png}

echo "🧪 CDN integration tests"
echo "📍 Target: $BASE_URL"
echo "────────────────────────────────────────"

RED='\033[0;31m'
GREEN='\033[0;32m'
NC='\033[0m'

PASSED=0
FAILED=0

test_result() {
    local test_name="$1"
    local result="$2"
    local expected="$3"
    local actual="$4"

    if [ "$result" = "PASS" ]; then
        echo -e "${GREEN}✅ $test_name${NC}"
        ((PASSED++))
    else
        echo -e "${RED}❌ $test_name${NC}"
        echo -e "   Expected: $expected"
        echo -e "   Got: $actual"
        ((FAILED++))
    fi
}

# Test 1: /health endpoint
echo -e "\n🏥 Endpoint /health"
response=$(curl -s -w "HTTPSTATUS:%{http_code}" "$BASE_URL/health")
http_code=$(echo "$response" | tr -d '\n' | sed -e 's/.*HTTPSTATUS://')
body=$(echo "$response" | sed -e 's/HTTPSTATUS:.*//g')

if [ "$http_code" = "200" ] && [ "$body" = "healthy" ]; then
    test_result "/health returns 200 and 'healthy'" "PASS"
else
    test_result "/health returns 200 and 'healthy'" "FAIL" "200, healthy" "$http_code, $body"
fi

# Test 2: /health headers
echo -e "\n📋 /health headers"
response=$(curl -s -I "$BASE_URL/health")
content_type=$(echo "$response" | grep -i "content-type" | tr -d '\r')
cors_origin=$(echo "$response" | grep -i "access-control-allow-origin" | tr -d '\r')

if echo "$content_type" | grep -q "text/plain"; then
    test_result "/health has the right Content-Type" "PASS"
else
    test_result "/health has the right Content-Type" "FAIL" "text/plain" "$content_type"
fi

if [ "$cors_origin" = "Access-Control-Allow-Origin: *" ]; then
    test_result "/health has the CORS headers" "PASS"
else
    test_result "/health has the CORS headers" "FAIL" "Access-Control-Allow-Origin: *" "$cors_origin"
fi

# Test 3: known file
echo -e "\n🖼️  Known file ($ASSET)"
http_code=$(curl -s -w "%{http_code}" -o /dev/null "$BASE_URL$ASSET")
content_length=$(curl -s -I "$BASE_URL$ASSET" | grep -i "content-length" | tr -d '\r' | sed 's/.*: //')

if [ "$http_code" = "200" ] && [ -n "$content_length" ] && [ "$content_length" -gt 0 ]; then
    test_result "The file answers 200 with content" "PASS"
else
    test_result "The file answers 200 with content" "FAIL" "200, content" "$http_code, length=$content_length"
fi

# Test 4: file headers
echo -e "\n🏷️  File headers"
response=$(curl -s -I "$BASE_URL$ASSET")
content_type=$(echo "$response" | grep -i "content-type" | tr -d '\r')
etag=$(echo "$response" | grep -i "etag" | tr -d '\r')
cache_control=$(echo "$response" | grep -i "cache-control" | tr -d '\r')
cors_origin=$(echo "$response" | grep -i "access-control-allow-origin" | tr -d '\r')

if echo "$content_type" | grep -q "image/png"; then
    test_result "Content-Type image/png" "PASS"
else
    test_result "Content-Type image/png" "FAIL" "image/png" "$content_type"
fi

if echo "$etag" | grep -i -q "etag:"; then
    test_result "ETag header present" "PASS"
else
    test_result "ETag header present" "FAIL" "an ETag header" "$etag"
fi

if [ "$cache_control" = "Cache-Control: public, max-age=31536000, immutable" ]; then
    test_result "Cache-Control is immutable" "PASS"
else
    test_result "Cache-Control is immutable" "FAIL" "public, max-age=31536000, immutable" "$cache_control"
fi

if [ "$cors_origin" = "Access-Control-Allow-Origin: *" ]; then
    test_result "CORS headers present" "PASS"
else
    test_result "CORS headers present" "FAIL" "Access-Control-Allow-Origin: *" "$cors_origin"
fi

# Test 5: ETag caching (304)
echo -e "\n🏷️  Cache ETag"
response=$(curl -s -I "$BASE_URL$ASSET")
etag=$(echo "$response" | grep -i "etag" | tr -d '\r' | sed 's/ETag: //I')
response=$(curl -s -w "HTTPSTATUS:%{http_code}" -H "If-None-Match: $etag" "$BASE_URL$ASSET")
http_code=$(echo "$response" | tr -d '\n' | sed -e 's/.*HTTPSTATUS://')

if [ "$http_code" = "304" ]; then
    test_result "If-None-Match returns 304 Not Modified" "PASS"
else
    test_result "If-None-Match returns 304 Not Modified" "FAIL" "304" "$http_code"
fi

# Test 6: root redirect
echo -e "\n🏠 Redirect on /"
response=$(curl -s -w "HTTPSTATUS:%{http_code}" -I "$BASE_URL/")
http_code=$(echo "$response" | tr -d '\n' | sed -e 's/.*HTTPSTATUS://')
location=$(echo "$response" | grep -i "location" | tr -d '\r')

if [ "$http_code" = "302" ] && echo "$location" | grep -q "https://"; then
    test_result "/ redirects to the configured domain" "PASS"
else
    test_result "/ redirects to the configured domain" "FAIL" "302, https://<domain>" "$http_code, $location"
fi

# Test 7: unknown path with an extension -> cacheable 404
echo -e "\n❓ 404 on an unknown file"
http_code=$(curl -s -w "%{http_code}" -o /dev/null "$BASE_URL/inexistant.json")

if [ "$http_code" = "404" ]; then
    test_result "Unknown path with an extension returns 404" "PASS"
else
    test_result "Unknown path with an extension returns 404" "FAIL" "404" "$http_code"
fi

# Test 7b: unknown path without an extension -> redirect
response=$(curl -s -w "HTTPSTATUS:%{http_code}" -I "$BASE_URL/inexistant")
http_code=$(echo "$response" | tr -d '\n' | sed -e 's/.*HTTPSTATUS://')
location=$(echo "$response" | grep -i "location" | tr -d '\r')

if [ "$http_code" = "302" ] && echo "$location" | grep -q "https://"; then
    test_result "Unknown path without an extension redirects" "PASS"
else
    test_result "Unknown path without an extension redirects" "FAIL" "302, https://<domain>" "$http_code, $location"
fi

# Test 8: OPTIONS preflight (CORS)
echo -e "\n🌐 CORS preflight"
response=$(curl -s -w "HTTPSTATUS:%{http_code}" -X OPTIONS "$BASE_URL$ASSET")
http_code=$(echo "$response" | tr -d '\n' | sed -e 's/.*HTTPSTATUS://')

if [ "$http_code" = "204" ]; then
    test_result "OPTIONS returns 204 No Content" "PASS"
else
    test_result "OPTIONS returns 204 No Content" "FAIL" "204" "$http_code"
fi

# Test 9: HEAD request
echo -e "\n📋 HEAD request"
response=$(curl -s -w "HTTPSTATUS:%{http_code}" -I "$BASE_URL$ASSET")
http_code=$(echo "$response" | tr -d '\n' | sed -e 's/.*HTTPSTATUS://')

if [ "$http_code" = "200" ]; then
    test_result "HEAD returns 200 OK" "PASS"
else
    test_result "HEAD returns 200 OK" "FAIL" "200" "$http_code"
fi

# Test 10: discover.json endpoint
echo -e "\n🗺️  Endpoint discover.json"
response=$(curl -s -w "HTTPSTATUS:%{http_code}" "$BASE_URL/discover.json")
http_code=$(echo "$response" | tr -d '\n' | sed -e 's/.*HTTPSTATUS://')
body=$(echo "$response" | sed -e 's/HTTPSTATUS:.*//g')
content_type=$(curl -s -I "$BASE_URL/discover.json" | grep -i "content-type" | tr -d '\r')

if [ "$http_code" = "200" ] && echo "$body" | grep -q '"resources"' && echo "$body" | grep -q "$ASSET"; then
    test_result "discover.json lists the resources" "PASS"
else
    test_result "discover.json lists the resources" "FAIL" "200 + resources + $ASSET" "$http_code"
fi

if echo "$content_type" | grep -q "application/json"; then
    test_result "discover.json has the right Content-Type" "PASS"
else
    test_result "discover.json has the right Content-Type" "FAIL" "application/json" "$content_type"
fi

# Test 11: /metrics endpoint
echo -e "\n📈 Endpoint /metrics"
response=$(curl -s -w "HTTPSTATUS:%{http_code}" "$BASE_URL/metrics")
http_code=$(echo "$response" | tr -d '\n' | sed -e 's/.*HTTPSTATUS://')
body=$(echo "$response" | sed -e 's/HTTPSTATUS:.*//g')

if [ "$http_code" = "200" ] && echo "$body" | grep -q "karako_cdn_responses_total"; then
    test_result "/metrics exposes the counters" "PASS"
else
    test_result "/metrics exposes the counters" "FAIL" "200 + karako_cdn_responses_total" "$http_code"
fi

# Test 12: gzip on discover.json
echo -e "\n🗜️  gzip compression"
encoding=$(curl -s -I -H "Accept-Encoding: gzip" "$BASE_URL/discover.json" | grep -i "content-encoding" | tr -d '\r')

if echo "$encoding" | grep -qi "gzip"; then
    test_result "discover.json is served gzipped when accepted" "PASS"
else
    test_result "discover.json is served gzipped when accepted" "FAIL" "Content-Encoding: gzip" "$encoding"
fi

# Summary
echo -e "\n────────────────────────────────────────"
echo -e "📊 Summary:"
echo -e "   ${GREEN}Passed: $PASSED${NC}"
echo -e "   ${RED}Failed: $FAILED${NC}"
echo -e "   Total: $((PASSED + FAILED))"

if [ $FAILED -eq 0 ]; then
    echo -e "\n${GREEN}🎉 All tests pass. The CDN works as expected.${NC}"
    exit 0
else
    echo -e "\n${RED}⚠️  Some tests failed. Check the implementation.${NC}"
    exit 1
fi
