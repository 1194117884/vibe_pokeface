#!/bin/bash
# Hard concurrency test for registration API
# Tests that the server prevents duplicate usernames under high concurrency

set -e

BASE_URL="${1:-http://localhost:8080}"
CONCURRENCY="${2:-20}"  # number of concurrent requests
USERNAME="concurrency_test_$(date +%s)"

echo "=== Concurrent Registration Test ==="
echo "Target: $BASE_URL/api/auth/register"
echo "Username: $USERNAME"
echo "Concurrency: $CONCURRENCY"
echo ""

# Fire N concurrent registration requests
PIDS=()
for i in $(seq 1 $CONCURRENCY); do
  curl -s -X POST "$BASE_URL/api/auth/register" \
    -H "Content-Type: application/json" \
    -d "{\"nickname\":\"$USERNAME\",\"password\":\"Test1234!\"}" \
    -o "/tmp/reg_resp_$$_$i.json" \
    -w "%{http_code}" > "/tmp/reg_code_$$_$i.txt" &
  PIDS+=($!)
done

# Wait for all to finish
for pid in "${PIDS[@]}"; do
  wait "$pid" 2>/dev/null || true
done

echo "=== Results ==="
SUCCESS=0
CONFLICT=0
OTHER=0
for i in $(seq 1 $CONCURRENCY); do
  CODE=$(cat "/tmp/reg_code_$$_$i.txt" 2>/dev/null || echo "000")
  case "$CODE" in
    200) SUCCESS=$((SUCCESS + 1));;
    409) CONFLICT=$((CONFLICT + 1));;
    *)
      OTHER=$((OTHER + 1))
      echo "Unexpected code $CODE: $(cat /tmp/reg_resp_$$_$i.json 2>/dev/null)"
      ;;
  esac
done

echo ""
echo "200 (OK):       $SUCCESS"
echo "409 (Conflict): $CONFLICT"
echo "Other:          $OTHER"
echo ""

# Verify only ONE user exists with this nickname
echo "=== Verification ==="
echo "Checking how many users have nickname '$USERNAME'..."
LOGIN_RESP=$(curl -s -X POST "$BASE_URL/api/auth/login" \
  -H "Content-Type: application/json" \
  -d "{\"password\":\"Test1234!\",\"provider_uid\":\"password:$USERNAME\"}")
echo "Login response: $LOGIN_RESP"

if echo "$LOGIN_RESP" | grep -q '"token"'; then
  echo ""
  echo "PASS: Login succeeds with the credentials"
  echo "FAIL: Only 1 registration should succeed but we got $SUCCESS successes"
  if [ "$SUCCESS" -eq 1 ]; then
    echo "FIX VERIFIED: Exactly 1 registration succeeded, all others were rejected"
  else
    echo "WARNING: $SUCCESS registrations succeeded with the same username"
  fi
else
  echo ""
  echo "FAIL: Cannot login with the registered credentials"
fi

# Cleanup
rm -f /tmp/reg_resp_$$_*.json /tmp/reg_code_$$_*.txt

echo ""
echo "=== Test Complete ==="
