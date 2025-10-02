#!/bin/sh
set -e

hostport=$1
shift

echo "Waiting for database at $hostport..."

until pg_isready -h ${hostport%:*} -p ${hostport#*:}; do
  echo "Database not ready yet..."
  sleep 1
done

echo "Database is ready!"
exec "$@"
