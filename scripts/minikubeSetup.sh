#!/bin/bash

# Install the Knative Serving and Eventing components
helm install knative zeiss-staging/knative --wait
# Install the Eventing components
helm install eventing zeiss-staging/eventing --wait
# Install the NATZ operator
helm install natz-operator natz-operator/natz-operator --wait --namespace knative-eventing
# Create operator resources
kubectl apply -f ../examples/operator.yaml
# Create account resources
kubectl apply -f ../examples/account.yaml
# Create user resources
kubectl apply -f ../examples/user.yaml
# Install NATS.io
helm install nats nats/nats --values examples/values.yaml
# Install the NATZ accounts-server
helm install account-server natz-operator/account-server --wait --namespace knative-eventing --values examples/account-server.yaml
