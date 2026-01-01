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
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

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
	flag.StringVar(&clusterType, "cluster-type", constants.ClusterTypeKind, "Cluster type: kind, minikube, or openshift")
	flag.StringVar(&namespace, "namespace", "", "Target namespace for Kruize deployment (default: auto-detected based on cluster type)")
	flag.StringVar(&operatorImage, "operator-image", "", "Operator image (default: read from Makefile)")
	flag.StringVar(&kruizeImage, "kruize-image", "", "Kruize/Autotune image (optional)")
	flag.StringVar(&kruizeUIImage, "kruize-ui-image", "", "Kruize UI image (optional)")
}

var _ = BeforeSuite(func() {
	// Auto-detect namespace if not specified via flag
	if namespace == "" {
		if clusterType == constants.ClusterTypeOpenShift {
			namespace = "openshift-tuning"
		} else {
			namespace = "monitoring"
		}
	}

	// Get operator image from Makefile if not specified via flag
	if operatorImage == "" {
		var err error
		operatorImage, err = utils.ExtractImageFromMakefile()
		if err != nil {
			// Use default if Makefile not found
			operatorImage = "quay.io/kruize/kruize-operator:0.0.2"
		}
	}

	fmt.Fprintf(GinkgoWriter, "Running e2e tests with cluster_type=%s, namespace=%s, operator_image=%s, kruize_image=%s, kruize_ui_image=%s\n",
		clusterType, namespace, operatorImage, kruizeImage, kruizeUIImage)
})

// Run e2e tests using the Ginkgo runner.
func TestE2E(t *testing.T) {
	RegisterFailHandler(Fail)
	fmt.Fprintf(GinkgoWriter, "Starting kruize-operator suite\n")
	RunSpecs(t, "e2e suite")
}
