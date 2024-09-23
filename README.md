# Kubernetes Operator for NATS Accounting

[![Release](https://github.com/ZEISS/natz-operator/actions/workflows/release.yml/badge.svg)](https://github.com/ZEISS/natz-operator/actions/workflows/release.yml)
[![Taylor Swift](https://img.shields.io/badge/secured%20by-taylor%20swift-brightgreen.svg)](https://twitter.com/SwiftOnSecurity)
[![Volkswagen](https://auchenberg.github.io/volkswagen/volkswargen_ci.svg?v=1)](https://github.com/auchenberg/volkswagen)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)

A Kubernetes operator for [NATS](https://nats.io/) accounting.

## Installation

[Helm](https://helm.sh/) can be used to install the `natz-operator` to your Kubernetes cluster.

```shell
helm repo add natz-operator https://zeiss.github.io/natz-operator/helm/charts
helm repo update
helm search repo natz-operator
```

## Usage

Creating the operator for the [NATS](https://nats.io/) accounting.

```yaml
apiVersion: natz.zeiss.com/v1alpha1
kind: NatsOperator
metadata:
  namespace: knative-eventing
  name: natsoperator-sample
spec:
```

Creating an account that supports the use of jetstream.

```yaml
apiVersion: natz.zeiss.com/v1alpha1
kind: NatsAccount
metadata:
  namespace: knative-eventing 
  name: knative-eventing-account
spec:
  operatorRef:
    name: natsoperator-sample 
  allowedUserNamespaces:
  - knative-eventing
  imports: []
  exports: []
  limits: 
    disk_storage: -1
    streams: -1
    conn: -1
    imports: -1
    exports: -1
    subs: -1
    payload: -1
    data: -1
```

Creating a user account.

```yaml
apiVersion: natz.zeiss.com/v1alpha1
kind: NatsUser
metadata:
  namespace: knative-eventing
  name: knative-eventing-user
spec:
  accountRef:
    namespace: knative-eventing
    name: knative-eventing-account
  limits:
    payload: -1
    subs: -1
    data: -1
```

## NATS Operator

The operator can be integrated with the NATS operator.

```yaml
config:
  jetstream:
    enabled: true
    fileStore:
      pvc:
        size: 2Gi
  resolver:
    enabled: true
    merge:
      type: full
      interval: "2m"
      timeout: "1.9s"
  merge:
    00$include: "../custom-auth/auth.conf"
    debug: true
container:
  patch:
  - op: add
    path: "/volumeMounts/-"
    value:
      name: auth-config
      mountPath: "/etc/custom-auth"
statefulSet:
  patch:
  - op: add
    path: /spec/template/spec/volumes/-
    value:
      name: "auth-config"
      secret:
        defaultMode: 420
        secretName: "natsoperator-sample-server-config"
```

## License

[Apache 2.0](/LICENSE)