/*
Copyright 2024.

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

package e2e

import (
	"flag"
	"fmt"
	"os"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"gopkg.in/yaml.v3"

	"github.com/kruize/kruize-operator/internal/constants"
	"github.com/kruize/kruize-operator/test/utils"
)

// Test parameters that can be set via command-line flags
var (
	clusterType   string
	namespace     string
	operatorImage string
	kruizeImage   string
	kruizeUIImage string
)

func init() {
	// Read defaults from sample YAML file
	defaultClusterType, defaultNamespace, defaultKruizeImage, defaultKruizeUIImage := readSampleYAMLDefaults()

	flag.StringVar(&clusterType, "cluster-type", defaultClusterType, "Cluster type: kind, minikube, or openshift")
	flag.StringVar(&namespace, "namespace", defaultNamespace, "Target namespace for Kruize deployment (default: from sample YAML or auto-detected based on cluster type)")
	flag.StringVar(&operatorImage, "operator-image", "", "Operator image (default: read from Makefile)")
	flag.StringVar(&kruizeImage, "kruize-image", defaultKruizeImage, "Kruize/Autotune image (default: from sample YAML)")
	flag.StringVar(&kruizeUIImage, "kruize-ui-image", defaultKruizeUIImage, "Kruize UI image (default: from sample YAML)")
}

// readSampleYAMLDefaults reads default values from the sample YAML file
func readSampleYAMLDefaults() (clusterType, namespace, kruizeImage, kruizeUIImage string) {
	// Set fallback defaults
	clusterType = constants.ClusterTypeKind
	namespace = ""
	kruizeImage = ""
	kruizeUIImage = ""

	// Get project directory to construct absolute path
	projectDir, err := utils.GetProjectDir()
	if err != nil {
		return
	}

	sourcePath := projectDir + "/config/samples/v1alpha1_kruize.yaml"
	content, err := os.ReadFile(sourcePath)
	if err != nil {
		return
	}

	// Parse YAML
	var yamlData map[string]interface{}
	if err := yaml.Unmarshal(content, &yamlData); err != nil {
		return
	}

	// Extract spec section
	spec, ok := yamlData["spec"].(map[string]interface{})
	if !ok {
		return
	}

	// Extract values from sample YAML with type assertions
	if ct, ok := spec["cluster_type"].(string); ok {
		clusterType = ct
	}
	if ns, ok := spec["namespace"].(string); ok {
		namespace = ns
	}
	if ai, ok := spec["autotune_image"].(string); ok {
		kruizeImage = ai
	}
	if aui, ok := spec["autotune_ui_image"].(string); ok {
		kruizeUIImage = aui
	}

	return
}

var _ = BeforeSuite(func() {
	// Auto-detect namespace based on cluster type if not explicitly set
	if namespace == "" {
		namespace = getDefaultNamespace(clusterType)
	} else {
		// If namespace was set from sample YAML but cluster type was changed via flag,
		// override namespace to match the new cluster type
		defaultClusterType, defaultNamespace, _, _ := readSampleYAMLDefaults()
		if namespace == defaultNamespace && clusterType != defaultClusterType {
			namespace = getDefaultNamespace(clusterType)
		}
	}

	// Get operator image from Makefile if not specified via flag
	if operatorImage == "" {
		var err error
		operatorImage, err = utils.ExtractImageFromMakefile()
		if err != nil {
			operatorImage = "quay.io/kruize/kruize-operator:0.0.2"
		}
	}

	fmt.Fprintf(GinkgoWriter, "Running e2e tests with cluster_type=%s, namespace=%s, operator_image=%s, kruize_image=%s, kruize_ui_image=%s\n",
		clusterType, namespace, operatorImage, kruizeImage, kruizeUIImage)
})

// getDefaultNamespace returns the default namespace for the given cluster type
func getDefaultNamespace(clusterType string) string {
	if clusterType == constants.ClusterTypeOpenShift {
		return "openshift-tuning"
	}
	return "monitoring"
}

// Run e2e tests using the Ginkgo runner.
func TestE2E(t *testing.T) {
	RegisterFailHandler(Fail)
	fmt.Fprintf(GinkgoWriter, "Starting kruize-operator suite\n")
	RunSpecs(t, "e2e suite")
}
