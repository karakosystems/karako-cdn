#!/bin/bash

# Script de test d'intégration du CDN Karako
# Usage: ./test.sh [port] [host]
# Défaut: port=8080, host=localhost

PORT=${1:-8080}
HOST=${2:-localhost}
BASE_URL="http://$HOST:$PORT"
ASSET="/images/karako/logos/logo-icon-black.png"

echo "🧪 Test du CDN Karako"
echo "📍 Cible: $BASE_URL"
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
        echo -e "   Attendu: $expected"
        echo -e "   Obtenu: $actual"
        ((FAILED++))
    fi
}

# Test 1 : endpoint /health
echo -e "\n🏥 Endpoint /health"
response=$(curl -s -w "HTTPSTATUS:%{http_code}" "$BASE_URL/health")
http_code=$(echo "$response" | tr -d '\n' | sed -e 's/.*HTTPSTATUS://')
body=$(echo "$response" | sed -e 's/HTTPSTATUS:.*//g')

if [ "$http_code" = "200" ] && [ "$body" = "healthy" ]; then
    test_result "/health retourne 200 et 'healthy'" "PASS"
else
    test_result "/health retourne 200 et 'healthy'" "FAIL" "200, healthy" "$http_code, $body"
fi

# Test 2 : en-têtes de /health
echo -e "\n📋 En-têtes de /health"
response=$(curl -s -I "$BASE_URL/health")
content_type=$(echo "$response" | grep -i "content-type" | tr -d '\r')
cors_origin=$(echo "$response" | grep -i "access-control-allow-origin" | tr -d '\r')

if echo "$content_type" | grep -q "text/plain"; then
    test_result "/health a le bon Content-Type" "PASS"
else
    test_result "/health a le bon Content-Type" "FAIL" "text/plain" "$content_type"
fi

if [ "$cors_origin" = "Access-Control-Allow-Origin: *" ]; then
    test_result "/health a les en-têtes CORS" "PASS"
else
    test_result "/health a les en-têtes CORS" "FAIL" "Access-Control-Allow-Origin: *" "$cors_origin"
fi

# Test 3 : asset Karako
echo -e "\n🖼️  Asset Karako ($ASSET)"
http_code=$(curl -s -w "%{http_code}" -o /dev/null "$BASE_URL$ASSET")
content_length=$(curl -s -I "$BASE_URL$ASSET" | grep -i "content-length" | tr -d '\r' | sed 's/.*: //')

if [ "$http_code" = "200" ] && [ -n "$content_length" ] && [ "$content_length" -gt 0 ]; then
    test_result "L'asset répond 200 avec du contenu" "PASS"
else
    test_result "L'asset répond 200 avec du contenu" "FAIL" "200, contenu" "$http_code, length=$content_length"
fi

# Test 4 : en-têtes de l'asset
echo -e "\n🏷️  En-têtes de l'asset"
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
    test_result "En-tête ETag présent" "PASS"
else
    test_result "En-tête ETag présent" "FAIL" "ETag présent" "$etag"
fi

if [ "$cache_control" = "Cache-Control: public, max-age=31536000, immutable" ]; then
    test_result "Cache-Control correct" "PASS"
else
    test_result "Cache-Control correct" "FAIL" "public, max-age=31536000, immutable" "$cache_control"
fi

if [ "$cors_origin" = "Access-Control-Allow-Origin: *" ]; then
    test_result "En-têtes CORS présents" "PASS"
else
    test_result "En-têtes CORS présents" "FAIL" "Access-Control-Allow-Origin: *" "$cors_origin"
fi

# Test 5 : cache ETag (304)
echo -e "\n🏷️  Cache ETag"
response=$(curl -s -I "$BASE_URL$ASSET")
etag=$(echo "$response" | grep -i "etag" | tr -d '\r' | sed 's/ETag: //I')
response=$(curl -s -w "HTTPSTATUS:%{http_code}" -H "If-None-Match: $etag" "$BASE_URL$ASSET")
http_code=$(echo "$response" | tr -d '\n' | sed -e 's/.*HTTPSTATUS://')

if [ "$http_code" = "304" ]; then
    test_result "If-None-Match retourne 304 Not Modified" "PASS"
else
    test_result "If-None-Match retourne 304 Not Modified" "FAIL" "304" "$http_code"
fi

# Test 6 : redirection de la racine
echo -e "\n🏠 Redirection de /"
response=$(curl -s -w "HTTPSTATUS:%{http_code}" -I "$BASE_URL/")
http_code=$(echo "$response" | tr -d '\n' | sed -e 's/.*HTTPSTATUS://')
location=$(echo "$response" | grep -i "location" | tr -d '\r')

if [ "$http_code" = "302" ] && echo "$location" | grep -q "https://"; then
    test_result "/ redirige vers le domaine configuré" "PASS"
else
    test_result "/ redirige vers le domaine configuré" "FAIL" "302, https://<domaine>" "$http_code, $location"
fi

# Test 7 : chemin inconnu avec extension -> 404 cacheable
echo -e "\n❓ 404 sur asset inconnu"
http_code=$(curl -s -w "%{http_code}" -o /dev/null "$BASE_URL/inexistant.json")

if [ "$http_code" = "404" ]; then
    test_result "Chemin inconnu avec extension retourne 404" "PASS"
else
    test_result "Chemin inconnu avec extension retourne 404" "FAIL" "404" "$http_code"
fi

# Test 7b : chemin inconnu sans extension -> redirection
response=$(curl -s -w "HTTPSTATUS:%{http_code}" -I "$BASE_URL/inexistant")
http_code=$(echo "$response" | tr -d '\n' | sed -e 's/.*HTTPSTATUS://')
location=$(echo "$response" | grep -i "location" | tr -d '\r')

if [ "$http_code" = "302" ] && echo "$location" | grep -q "https://"; then
    test_result "Chemin inconnu sans extension redirige vers le domaine" "PASS"
else
    test_result "Chemin inconnu sans extension redirige vers le domaine" "FAIL" "302, https://<domaine>" "$http_code, $location"
fi

# Test 8 : préflight OPTIONS (CORS)
echo -e "\n🌐 Préflight CORS"
response=$(curl -s -w "HTTPSTATUS:%{http_code}" -X OPTIONS "$BASE_URL$ASSET")
http_code=$(echo "$response" | tr -d '\n' | sed -e 's/.*HTTPSTATUS://')

if [ "$http_code" = "204" ]; then
    test_result "OPTIONS retourne 204 No Content" "PASS"
else
    test_result "OPTIONS retourne 204 No Content" "FAIL" "204" "$http_code"
fi

# Test 9 : requête HEAD
echo -e "\n📋 Requête HEAD"
response=$(curl -s -w "HTTPSTATUS:%{http_code}" -I "$BASE_URL$ASSET")
http_code=$(echo "$response" | tr -d '\n' | sed -e 's/.*HTTPSTATUS://')

if [ "$http_code" = "200" ]; then
    test_result "HEAD retourne 200 OK" "PASS"
else
    test_result "HEAD retourne 200 OK" "FAIL" "200" "$http_code"
fi

# Test 10 : endpoint discover.json
echo -e "\n🗺️  Endpoint discover.json"
response=$(curl -s -w "HTTPSTATUS:%{http_code}" "$BASE_URL/discover.json")
http_code=$(echo "$response" | tr -d '\n' | sed -e 's/.*HTTPSTATUS://')
body=$(echo "$response" | sed -e 's/HTTPSTATUS:.*//g')
content_type=$(curl -s -I "$BASE_URL/discover.json" | grep -i "content-type" | tr -d '\r')

if [ "$http_code" = "200" ] && echo "$body" | grep -q '"resources"' && echo "$body" | grep -q "$ASSET"; then
    test_result "discover.json liste les ressources" "PASS"
else
    test_result "discover.json liste les ressources" "FAIL" "200 + resources + $ASSET" "$http_code"
fi

if echo "$content_type" | grep -q "application/json"; then
    test_result "discover.json a le bon Content-Type" "PASS"
else
    test_result "discover.json a le bon Content-Type" "FAIL" "application/json" "$content_type"
fi

# Test 11 : endpoint /metrics
echo -e "\n📈 Endpoint /metrics"
response=$(curl -s -w "HTTPSTATUS:%{http_code}" "$BASE_URL/metrics")
http_code=$(echo "$response" | tr -d '\n' | sed -e 's/.*HTTPSTATUS://')
body=$(echo "$response" | sed -e 's/HTTPSTATUS:.*//g')

if [ "$http_code" = "200" ] && echo "$body" | grep -q "karako_cdn_responses_total"; then
    test_result "/metrics expose les compteurs" "PASS"
else
    test_result "/metrics expose les compteurs" "FAIL" "200 + karako_cdn_responses_total" "$http_code"
fi

# Test 12 : gzip sur discover.json
echo -e "\n🗜️  Compression gzip"
encoding=$(curl -s -I -H "Accept-Encoding: gzip" "$BASE_URL/discover.json" | grep -i "content-encoding" | tr -d '\r')

if echo "$encoding" | grep -qi "gzip"; then
    test_result "discover.json servi en gzip quand accepté" "PASS"
else
    test_result "discover.json servi en gzip quand accepté" "FAIL" "Content-Encoding: gzip" "$encoding"
fi

# Bilan
echo -e "\n────────────────────────────────────────"
echo -e "📊 Bilan:"
echo -e "   ${GREEN}Réussis: $PASSED${NC}"
echo -e "   ${RED}Échoués: $FAILED${NC}"
echo -e "   Total: $((PASSED + FAILED))"

if [ $FAILED -eq 0 ]; then
    echo -e "\n${GREEN}🎉 Tous les tests passent. Le CDN fonctionne correctement.${NC}"
    exit 0
else
    echo -e "\n${RED}⚠️  Des tests ont échoué. Vérifier l'implémentation.${NC}"
    exit 1
fi
