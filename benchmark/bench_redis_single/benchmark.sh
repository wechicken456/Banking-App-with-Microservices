#!/bin/bash

# Remember to start the docker containers and migrate database in other terminals before running this script

# ============= CONFIG =================
API_GW="http://localhost:18000"
N_USERS=100
READS_PER_USER=10
DEPOSITS_PER_USER=10
DURATION="10s"
WARM_UP_DURATION="2s"
JWT_FILE="jwts.txt"
READ_TARGETS="read_targets.txt"
DEPOSIT_TARGETS="deposit_targets.txt"

echo "Cleaning up old files: " $JWT_FILE $READ_TARGETS $DEPOSIT_TARGETS
rm -r $JWT_FILE $READ_TARGETS $DEPOSIT_TARGETS

set -e

# =========== CREATE N_USERS with accounts + JWTs =================
for i in $(seq 1 $N_USERS); do
    EMAIL="testuser$i@example.com"
    PASSWORD="pass"

    # Register
    ikey=$(uuidgen) # idempotency key
    curl -s -X POST "$API_GW/api/register" \
        -H "Idempotency-Key: $ikey" \
        -H "Content-Type: application/json" \
        -d "{\"email\": \"$EMAIL\", \"password\": \"$PASSWORD\"}" >/dev/null

    # Login to get JWT and fingerprint
    ikey=$(uuidgen) # idempotency key
    LOGIN_RES=$(curl -i -s -X POST "$API_GW/api/login" \
        -H "Idempotency-Key: $ikey" \
        -H "Content-Type: application/json" \
        -d "{\"email\": \"$EMAIL\", \"password\": \"$PASSWORD\"}")

    #echo "LOGIN_RES: $LOGIN_RES"

    JWT=$(echo $LOGIN_RES | grep '"accessToken"' | grep -o '"accessToken":"[^"]*' | cut -d '"' -f 4)
    #JWT=$(echo $LOGIN_RES | grep '"accessToken"' | jq -r '.accessToken')
    #echo "JWT token for user $i: $JWT"

    FINGERPRINT=$(echo $LOGIN_RES | grep -i "set-cookie: fingerprint=" | head -1 | cut -d '=' -f 2 | cut -d ';' -f 1)
    #echo "FINGERPRINT for user $i: $FINGERPRINT"

    if [ -z "$JWT" ] || [ -z "$FINGERPRINT" ]; then
        echo "Failed to get JWT or Fingerprint... Exiting"
        exit 1
    fi

    # Create account
    ikey=$(uuidgen) # idempotency key
    ACC_RES=$(curl -s -X POST "$API_GW/api/create-account" \
        -H "Idempotency-Key: $ikey" \
        -H "Authorization: Bearer $JWT" \
        -b "fingerprint=$FINGERPRINT" \
        -H "Content-Type: application/json" \
        -d "{\"balance\":10$i}")

    # echo "ACC_RES: $ACC_RES"

    ACC_ID=$(echo $ACC_RES | jq -r '.accountId')
    ACC_NUM=$(echo $ACC_RES | jq -r '.accountNumber')

    #echo "Writing the following line to file $JWT_FILE: $JWT|$FINGERPRINT|$ACC_ID|$ACC_NUM"
    echo "$JWT|$FINGERPRINT|$ACC_ID|$ACC_NUM" >>$JWT_FILE
done

# =============== Create READ account balance targets ====================
# not running any queries yet

while IFS='|' read -r JWT FP ACC_ID ACC_NUM; do
    for _ in $(seq 1 $READS_PER_USER); do
        cat <<EOF >>$READ_TARGETS
GET $API_GW/api/account?accountID=$ACC_ID
Authorization: Bearer $JWT
Cookie: fingerprint=$FP

EOF
    done
done <$JWT_FILE

# =============== Create WRITE deposit targets ===================
while IFS='|' read -r JWT FP ACC_ID ACC_NUM; do

    body_file="./json_body/deposit_$ACC_ID.json"

    cat >$body_file <<EOF
{"accountId":"$ACC_ID","amount":100,"transactionType":"CREDIT"}
EOF

    for _ in $(seq 1 $DEPOSITS_PER_USER); do
        ikey=$(uuidgen)
        cat <<EOF >>$DEPOSIT_TARGETS
POST $API_GW/api/create-transaction
Authorization: Bearer $JWT
Cookie: fingerprint=$FP
Idempotency-Key: $ikey
Content-Type: application/json
@$body_file
EOF
    done
    printf '\n' >>$DEPOSIT_TARGETS
done <$JWT_FILE

# ================= Run benchmarks =========================
echo ""
echo "=== CACHE WARM UP ==="
while IFS='|' read -r JWT FP ACC_ID ACC_NUM; do
    curl -s -o /dev/null -H "Authorization: Bearer $JWT" -b"fingerprint=$FP" "$API_GW/api/account?accountID=$ACC_ID" &
done <$JWT_FILE

echo ""
echo "=== BALANCE READ (CACHE) ==="
vegeta attack -targets=$READ_TARGETS -duration=$DURATION >bench_read.bin
vegeta report -type=json bench_read.bin >bench_read.json
echo "Results saved to bench_read.json"

echo ""
echo "=== DEPOSIT (WRITE + INVALIDATION) ==="
vegeta attack -targets=$DEPOSIT_TARGETS -duration=$DURATION >bench_write.bin
vegeta report -type=json bench_write.bin >bench_write.json
echo "Results saved to bench_write.json"
