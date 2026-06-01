#!/bin/bash
set -euo pipefail

zookeeper="${KAFKA_ZOOKEEPER_CONNECT:-zookeeper:2181}"
broker_id="${KAFKA_BROKER_ID:-1}"
broker_path="/brokers/ids/${broker_id}"

# Wait for Zookeeper before registering the broker.
cub zk-ready "$zookeeper" 120

# After an unclean Kafka exit, /brokers/ids/<id> can linger and block startup (NodeExists).
# cp-kafka ships zookeeper-shell (not kafka-zookeeper-shell).
clean_stale_broker_registration() {
  if ! command -v zookeeper-shell >/dev/null 2>&1; then
    echo "zookeeper-shell not found; skipping stale broker cleanup"
    return 0
  fi
  echo "Clearing stale Zookeeper registration at ${broker_path} ..."
  echo "delete ${broker_path}" | zookeeper-shell "$zookeeper" >/dev/null 2>&1 || true
  echo "deleteall ${broker_path}" | zookeeper-shell "$zookeeper" >/dev/null 2>&1 || true
}

clean_stale_broker_registration

exec /etc/confluent/docker/run
