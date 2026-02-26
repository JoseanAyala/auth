#!/usr/bin/env bash
# Test the Guardian rate limiter.
# Usage:
#   ./scripts/test-rate-limit.sh              # burst test only
#   ./scripts/test-rate-limit.sh --sustained  # burst + sustained drip

BASE_URL="${BASE_URL:-http://localhost:8080}"
ENDPOINT="$BASE_URL/health"

BURST=25        # send more than the default burst of 20 to trigger 429s
DRIP_COUNT=15   # number of drip requests after waiting for tokens to refill
DRIP_DELAY=0.15 # seconds between drip requests (just above 1/RPS=0.1s)

pass=0
fail=0

run_request() {
  local label="$1"
  response=$(curl -s -o /dev/null -w "%{http_code} %{time_total}" "$ENDPOINT")
  code=$(echo "$response" | awk '{print $1}')
  time=$(echo "$response" | awk '{print $2}')

  if [ "$code" = "200" ]; then
    echo "  $label → 200 OK           (${time}s)"
    ((pass++))
  elif [ "$code" = "429" ]; then
    retry=$(curl -s -I "$ENDPOINT" | grep -i retry-after | awk '{print $2}' | tr -d '\r')
    echo "  $label → 429 RATE LIMITED  (${time}s) Retry-After: ${retry}s"
    ((fail++))
  else
    echo "  $label → $code"
  fi
}

echo "========================================"
echo " Guardian Rate Limiter Test"
echo " Target: $ENDPOINT"
echo "========================================"

# ── Phase 1: Burst ────────────────────────────────────────────────────────────
echo ""
echo "Phase 1: Burst ($BURST rapid requests, default burst=20)"
echo "  Expect: first 20 → 200, remainder → 429"
echo ""

for i in $(seq 1 $BURST); do
  run_request "Request $(printf '%02d' $i)"
done

echo ""
echo "  Passed: $pass  |  Rate-limited: $fail"

if [ "$1" != "--sustained" ]; then
  echo ""
  echo "Run with --sustained to also test token drip recovery."
  exit 0
fi

# ── Phase 2: Sustained (drip after cooldown) ──────────────────────────────────
echo ""
echo "========================================"
echo "Phase 2: Sustained drip (${DRIP_DELAY}s between requests)"
echo "  Waiting 3s for bucket to partially refill..."
echo "========================================"
sleep 3

pass=0
fail=0

for i in $(seq 1 $DRIP_COUNT); do
  run_request "Drip   $(printf '%02d' $i)"
  sleep "$DRIP_DELAY"
done

echo ""
echo "  Passed: $pass  |  Rate-limited: $fail"
echo ""
echo "Done."
