# E2E Tests

## Overview

The e2e tests validate the complete Kruize operator deployment workflow on existing Kubernetes clusters. Tests are designed to work with **kind**, **minikube** and **openshift** clusters without requiring cluster creation/deletion. The tests automatically install cluster-appropriate Prometheus versions.

## Prerequisites

### Required
1. **Kubernetes Cluster**: kind, minikube or openshift cluster running and accessible
2. **kubectl**: Configured to access the cluster

## Command-Line Flags

### Core Flags

| Flag | Default | Description                                                                                                                                          |
|------|---------|------------------------------------------------------------------------------------------------------------------------------------------------------|
| `--cluster-type` | `kind` | Cluster type: `kind`, `minikube` or openshift                                                                                                        |
| `--namespace` | Auto-detected | Target namespace for Kruize deployment<br>• kind/minikube: `monitoring`<br>• openshift: `openshift-tuning` <br>• Can be overridden to any custom namespace |
| `--operator-image` | From Makefile | Operator image (reads IMAGE_TAG_BASE:VERSION from Makefile)                                                                                          |
| `--kruize-image` | From YAML | Kruize/Autotune image (overrides `autotune_image` in CR)                                                                                             |
| `--kruize-ui-image` | From YAML | Kruize UI image (overrides `autotune_ui_image` in CR)                                                                                                |

## Usage Examples

### Basic Usage

```bash
# Run on kind cluster (default)
go test ./test/e2e/... -v

# Run on openshift cluster
go test ./test/e2e/... -v -- -cluster-type=openshift
```

### Custom Configuration

```bash
# Custom namespace
go test ./test/e2e/... -v -- -cluster-type=kind -namespace=my-namespace

# Custom operator image
go test ./test/e2e/... -v -- -operator-image=quay.io/kruize/kruize-operator:0.0.2

# Custom Kruize images
go test ./test/e2e/... -v -- \
  -kruize-image=quay.io/kruize/autotune_operator:0.8 \
  -kruize-ui-image=quay.io/kruize/kruize-ui:latest

# Complete custom setup
go test ./test/e2e/... -v -- \
  -cluster-type=minikube \
  -namespace=my-namespace \
  -operator-image=quay.io/kruize/kruize-operator:latest \
  -kruize-image=quay.io/kruize/autotune_operator:latest \
  -kruize-ui-image=quay.io/kruize/kruize-ui:latest
```

### With Ginkgo Flags

```bash
# Verbose Ginkgo output
go test ./test/e2e/... -v -ginkgo.v -- -cluster-type=kind

# Run specific test
go test ./test/e2e/... -v -ginkgo.focus="should deploy Kruize" -- -cluster-type=kind

# View available flags
go test -c ./test/e2e/...
./e2e.test -help
```