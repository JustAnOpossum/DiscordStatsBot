#!/usr/bin/env bash
set -euo pipefail

DATA_DIR="$HOME/.mongodb-data"
mkdir -p "$DATA_DIR"

if pgrep -x mongod >/dev/null; then
	echo "mongod already running"
else
	mongod --fork --logpath "$DATA_DIR/mongod.log" --dbpath "$DATA_DIR" --bind_ip 127.0.0.1
fi
