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
	"fmt"
	"os"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/kruize/kruize-operator/test/utils"
)

// Test parameters that can be set via environment variables
var (
	clusterType   string
	namespace     string
	operatorImage string
	kruizeImage   string
	kruizeUIImage string
)

var _ = BeforeSuite(func() {
	// Get cluster type from environment variable, default to "kind"
	clusterType = os.Getenv("CLUSTER_TYPE")
	if clusterType == "" {
		clusterType = "kind"
	}

	// Get namespace from environment variable, default based on cluster type
	namespace = os.Getenv("KRUIZE_NAMESPACE")
	if namespace == "" {
		if clusterType == "openshift" {
			namespace = "openshift-tuning"
		} else {
			namespace = "monitoring"
		}
	}

	// Get operator image from environment or use Makefile default
	operatorImage = os.Getenv("OPERATOR_IMAGE")
	if operatorImage == "" {
		// Get default from Makefile
		var err error
		operatorImage, err = utils.GetDefaultOperatorImage()
		if err != nil {
			fmt.Fprintf(GinkgoWriter, "Warning: failed to get operator image from Makefile: %v, using fallback\n", err)
			operatorImage = "quay.io/kruize/kruize-operator:0.0.2"
		}
	}

	// Get Kruize (autotune) image from environment (optional)
	kruizeImage = os.Getenv("KRUIZE_IMAGE")

	// Get Kruize UI image from environment (optional)
	kruizeUIImage = os.Getenv("KRUIZE_UI_IMAGE")

	fmt.Fprintf(GinkgoWriter, "Running e2e tests with cluster_type=%s, namespace=%s, operator_image=%s, kruize_image=%s, kruize_ui_image=%s\n",
		clusterType, namespace, operatorImage, kruizeImage, kruizeUIImage)
})

// Run e2e tests using the Ginkgo runner.
func TestE2E(t *testing.T) {
	RegisterFailHandler(Fail)
	fmt.Fprintf(GinkgoWriter, "Starting kruize-operator suite\n")
	RunSpecs(t, "e2e suite")
}
