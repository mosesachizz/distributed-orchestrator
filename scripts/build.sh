#!/bin/bash

set -e

# Build script for the orchestrator

VERSION=${1:-latest}
PLATFORM=${2:-linux/amd64}

echo "Building orchestrator version: $VERSION"

# Build master
echo "Building master..."
docker build -f Dockerfile.master -t orchestrator-master:$VERSION --platform $PLATFORM .

# Build worker
echo "Building worker..."
docker build -f Dockerfile.worker -t orchestrator-worker:$VERSION --platform $PLATFORM .

echo "Build complete!"