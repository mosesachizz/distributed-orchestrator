#!/bin/bash

set -e

# Deployment script for the orchestrator

ENVIRONMENT=${1:-development}
VERSION=${2:-latest}

echo "Deploying orchestrator to $ENVIRONMENT environment, version: $VERSION"

if [ "$ENVIRONMENT" = "production" ]; then
    # Production deployment
    kubectl apply -f deployments/kubernetes/
    
    # Update images
    kubectl set image deployment/orchestrator-master master=orchestrator-master:$VERSION
    kubectl set image deployment/orchestrator-worker worker=orchestrator-worker:$VERSION
    
    # Wait for rollout
    kubectl rollout status deployment/orchestrator-master
    kubectl rollout status deployment/orchestrator-worker
    
elif [ "$ENVIRONMENT" = "development" ]; then
    # Development deployment with Docker Compose
    docker-compose up -d --build
    
else
    echo "Unknown environment: $ENVIRONMENT"
    exit 1
fi

echo "Deployment complete!"