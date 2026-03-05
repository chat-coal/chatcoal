#!/bin/sh
set -e

echo "Running database migrations..."
./migrate -cmd up

echo "Starting server..."
exec ./server
