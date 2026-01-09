# kruize-operator

A Kubernetes Operator to automate deployment of [Kruize Autotune](https://github.com/kruize/autotune), a resource optimization tool for Kubernetes workloads.

## Description

The Kruize operator simplifies deployment and management of Kruize on Kubernetes and OpenShift clusters. It provides a declarative way to configure and deploy Kruize components including the core Kruize service and UI through Custom Resource Definitions (CRDs).

For examples of running Kruize and the operator, see [kruize-demos](https://github.com/kruize/kruize-demos/tree/main/monitoring/local_monitoring). You can customize the YAML file (found in config/samples) to deploy with your preferred options.

## SEE ALSO

* [Kruize Autotune](https://github.com/kruize/autotune) - Main Kruize service providing resource optimization recommendations
* [kruize-ui](https://github.com/kruize/kruize-ui) - Web interface for Kruize
* [kruize-demos](https://github.com/kruize/kruize-demos) - Example deployments and demonstrations

## Getting Started

### Prerequisites
- Go version v1.21.0+ (tested with v1.22.5)
- Docker version 17.03+
- kubectl version v1.23.0+
- Access to a Kubernetes v1.23.0+ or OpenShift 4.x+ cluster
- [Prometheus](https://github.com/prometheus/prometheus) (for Minikube, Kind clusters)

### Deployment

The operator uses Kustomize overlays to manage platform-specific configurations:
- **OpenShift** (default): Deploys to `openshift-tuning` namespace
- **Local (Minikube/KIND)**: Deploys to `monitoring` namespace

**Quick Start:**

1. Build and push your image:
   ```sh
   make docker-build docker-push IMG=<some-registry>/kruize-operator:tag
   ```
**NOTE:** Ensure the image is published to a registry accessible from your cluster.

2. Install the CRDs:
   ```sh
   make install
   ```

3. Deploy the operator:
   ```sh
   # OpenShift (default)
   make deploy-openshift IMG=<some-registry>/kruize-operator:tag
   
   # Minikube
   make deploy-minikube IMG=<some-registry>/kruize-operator:tag
   
   # KIND
   make deploy-kind IMG=<some-registry>/kruize-operator:tag
   ```

4. Create instances of your solution:
   ```sh
   # For OpenShift
   kubectl apply -f config/samples/v1alpha1_kruize.yaml -n openshift-tuning
   
   # For Minikube/KIND
   kubectl apply -f config/samples/v1alpha1_kruize.yaml -n monitoring
   ```
   >**NOTE**: Before applying for Minikube/KIND, update [`config/samples/v1alpha1_kruize.yaml`](config/samples/v1alpha1_kruize.yaml):
   >- Set `cluster_type: "minikube"` or `cluster_type: "kind"`
   >- Set `namespace: "monitoring"` (instead of `"openshift-tuning"`)

**For detailed deployment options, overlay configurations, and advanced usage**, see [config/overlays/README.md](config/overlays/README.md).

**NOTE**: If you encounter RBAC errors, you may need to grant yourself cluster-admin privileges or be logged in as admin.

### To Uninstall

1. Delete the instances (CRs):
   ```sh
   # For OpenShift
   kubectl delete -f config/samples/v1alpha1_kruize.yaml -n openshift-tuning
   
   # For Minikube/KIND
   kubectl delete -f config/samples/v1alpha1_kruize.yaml -n monitoring
   ```

2. Undeploy the controller:
   ```sh
   make undeploy-openshift  # For OpenShift
   make undeploy-minikube   # For Minikube
   make undeploy-kind       # For KIND
   ```

3. Delete the CRDs:
   ```sh
   make uninstall
   ```

For more undeployment options, see [config/overlays/README.md](config/overlays/README.md).

## Project Distribution

Following are the steps to build the installer and distribute this project to users.

1. Build the installer for the image built and published in the registry:

```sh
make build-installer IMG=<some-registry>/kruize-operator:tag
```

NOTE: The makefile target mentioned above generates an 'install.yaml'
file in the dist directory. This file contains all the resources built
with Kustomize, which are necessary to install this project without
its dependencies.

2. Using the installer

Users can just run kubectl apply -f <URL for YAML BUNDLE> to install the project, i.e.:

```sh
kubectl apply -f https://raw.githubusercontent.com/<org>/kruize-operator/<tag or branch>/dist/install.yaml
```

## BUILDING

### Requirements
- Go v1.21.0+ (tested with v1.22.5)
- [operator-sdk](https://github.com/operator-framework/operator-sdk) v1.37.0+
- Docker version 17.03+

### Instructions

`make generate manifests` will trigger code/YAML generation and compile the operator controller manager.

`make docker-build` will build an OCI image tagged as `quay.io/kruize/kruize-operator:v0.0.2`. Override with `IMG` variable.

`make bundle` will create an OLM bundle in the `bundle/` directory. `make bundle-build` will create an OCI image of this bundle.

`make catalog-build` will build an OCI image of the operator catalog.

## DEVELOPMENT

Run the operator locally:
```sh
make run
```
This runs the controller manager as a process on your local machine. Note that it will not have access to certain in-cluster resources.

## TESTING

**Run unit tests:**
```sh
make test
```

**Run end-to-end tests:**

The `test-e2e` target supports optional flags for customizing the test environment:

```sh
# Default (OpenShift cluster)
make test-e2e
```
This requires a Kubernetes or OpenShift cluster. Recommended: Minikube, KIND, or CodeReady Containers.

## License

Apache License 2.0, see [LICENSE](/LICENSE).

