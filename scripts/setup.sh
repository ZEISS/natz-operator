#!/bin/bash

set -x

# Get the current directory
dir=$(pwd)

# Setup charts
helm repo add nats https://nats-io.github.io/k8s/helm/charts
helm repo update

# Pull images
docker pull nats:2.10.25-alpine
docker pull natsio/nats-server-config-reloader:0.16.1
docker pull natsio/nats-box:0.16.0

# Load images into kind
kind load docker-image nats:2.10.25-alpine
kind load docker-image natsio/nats-server-config-reloader:0.16.1
kind load docker-image natsio/nats-box:0.16.0