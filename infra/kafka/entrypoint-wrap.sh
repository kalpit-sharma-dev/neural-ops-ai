#!/bin/bash
set -euo pipefail

zookeeper="${KAFKA_ZOOKEEPER_CONNECT:-zookeeper:2181}"
broker_id="${KAFKA_BROKER_ID:-1}"

# Wait for Zookeeper before registering the broker.
cub zk-ready "$zookeeper" 120

# After an unclean Kafka exit, /brokers/ids/<id> can linger until the old ZK
# session expires and blocks startup with NodeExists. Safe for single-broker dev.
if command -v kafka-zookeeper-shell >/dev/null 2>&1; then
  echo "delete /brokers/ids/${broker_id}" | kafka-zookeeper-shell "$zookeeper" 2>/dev/null || true
fi

exec /etc/confluent/docker/run
