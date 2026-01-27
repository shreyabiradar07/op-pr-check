/*
Copyright 2025.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package constants

import "strings"

// SupportedClusterTypes defines all cluster types supported by Kruize
var SupportedClusterTypes = []string{
	ClusterTypeOpenShift,
	ClusterTypeMinikube,
	ClusterTypeKind,
}

// Cluster type constants
const (
	ClusterTypeOpenShift = "openshift"
	ClusterTypeMinikube  = "minikube"
	ClusterTypeKind      = "kind"
)

// Default container image versions
const (
	// DefaultAutotuneImage is the default container image for Kruize Autotune
	DefaultAutotuneImage = "quay.io/kruize/autotune_operator:0.8.1"

	// DefaultAutotuneUIImage is the default container image for Kruize UI
	DefaultAutotuneUIImage = "quay.io/kruize/kruize-ui:0.0.9"
)

// NormalizeClusterType converts user input to lowercase for internal use
// This allows case-insensitive cluster type matching (e.g., "OpenShift", "OPENSHIFT", "openshift" all work)
func NormalizeClusterType(clusterType string) string {
	return strings.ToLower(strings.TrimSpace(clusterType))
}

// IsValidClusterType checks if the given cluster type is supported (case-insensitive)
func IsValidClusterType(clusterType string) bool {
	normalized := NormalizeClusterType(clusterType)
	for _, validType := range SupportedClusterTypes {
		if normalized == validType {
			return true
		}
	}
	return false
}

// GetKubePrometheusVersion returns the appropriate kube-prometheus version based on cluster type
func GetKubePrometheusVersion(clusterType string) string {
	normalized := NormalizeClusterType(clusterType)
	switch normalized {
	case ClusterTypeKind:
		return "v0.13.0"
	case ClusterTypeMinikube:
		return "v0.16.0"
	default:
		return "v0.16.0" // Default to latest
	}
}
