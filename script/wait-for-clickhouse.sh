#!/bin/sh

CLICKHOUSE_HOST="${DB_ADDRESS}"
CLICKHOUSE_PORT="${DB_PORT}"
RETRY_INTERVAL=5
MAX_RETRIES=60

retry_count=0
while [ $retry_count -lt $MAX_RETRIES ]; do
  if nc -z -v -w5 $CLICKHOUSE_HOST $CLICKHOUSE_PORT; then
    echo "ClickHouse is ready!"
    exit 0
  else
    echo "Failed to connect to ClickHouse. Retrying in $RETRY_INTERVAL seconds..."
    retry_count=$((retry_count + 1))
    sleep $RETRY_INTERVAL
  fi
done

echo "Failed to connect to ClickHouse after $MAX_RETRIES retries. Exiting."
exit 1
