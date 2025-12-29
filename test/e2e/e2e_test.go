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
	"os/exec"
	"strings"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/kruize/kruize-operator/test/utils"
)

const operatorNamespace = "kruize-operator-system"

var _ = Describe("controller", Ordered, func() {
	BeforeAll(func() {
		// Skip Prometheus Operator installation on OpenShift as it's pre-installed
		if clusterType != "openshift" {
			By(fmt.Sprintf("installing prometheus operator for %s cluster", clusterType))
			err := utils.InstallPrometheusOperator(clusterType)
			if err != nil {
				// Log the error but don't fail - Prometheus might already be installed
				fmt.Fprintf(GinkgoWriter, "Warning: Prometheus installation encountered an issue: %v\n", err)
				fmt.Fprintf(GinkgoWriter, "Continuing with tests - Prometheus may already be installed\n")
			}
		} else {
			By("skipping prometheus operator installation on OpenShift (pre-installed)")
		}

		By("creating manager namespace")
		cmd := exec.Command("kubectl", "create", "ns", operatorNamespace)
		_, _ = utils.Run(cmd)
	})

	AfterAll(func() {

		By("removing manager namespace")
		cmd := exec.Command("kubectl", "delete", "ns", operatorNamespace)
		_, _ = utils.Run(cmd)
	})

	Context("Operator", func() {
		It("should run successfully", func() {
			var controllerPodName string
			var err error

			// Update CR BEFORE deploying controller so it uses correct images
			By(fmt.Sprintf("updating sample CR for cluster_type=%s and namespace=%s", clusterType, namespace))
			err = utils.UpdateKruizeSampleYAML(clusterType, namespace, kruizeImage, kruizeUIImage)
			ExpectWithOffset(1, err).NotTo(HaveOccurred())

			By(fmt.Sprintf("using operator image: %s", operatorImage))

			By("installing CRDs")
			cmd := exec.Command("make", "install")
			_, err = utils.Run(cmd)
			ExpectWithOffset(1, err).NotTo(HaveOccurred())

			By("deploying the controller-manager")
			cmd = exec.Command("make", "deploy", fmt.Sprintf("IMG=%s", operatorImage))
			_, err = utils.Run(cmd)
			ExpectWithOffset(1, err).NotTo(HaveOccurred())

			By("validating that the controller-manager pod is running as expected")
			verifyControllerUp := func() error {
				// Get pod name

				cmd = exec.Command("kubectl", "get",
					"pods", "-l", "control-plane=controller-manager",
					"-o", "go-template={{ range .items }}"+
						"{{ if not .metadata.deletionTimestamp }}"+
						"{{ .metadata.name }}"+
						"{{ \"\\n\" }}{{ end }}{{ end }}",
					"-n", operatorNamespace,
				)

				podOutput, err := utils.Run(cmd)
				ExpectWithOffset(2, err).NotTo(HaveOccurred())
				podNames := utils.GetNonEmptyLines(string(podOutput))
				if len(podNames) != 1 {
					return fmt.Errorf("expect 1 controller pods running, but got %d", len(podNames))
				}
				controllerPodName = podNames[0]
				ExpectWithOffset(2, controllerPodName).Should(ContainSubstring("controller-manager"))

				// Validate pod status
				cmd = exec.Command("kubectl", "get",
					"pods", controllerPodName, "-o", "jsonpath={.status.phase}",
					"-n", operatorNamespace,
				)
				status, err := utils.Run(cmd)
				ExpectWithOffset(2, err).NotTo(HaveOccurred())
				if string(status) != "Running" {
					return fmt.Errorf("controller pod in %s status", status)
				}
				return nil
			}
			EventuallyWithOffset(1, verifyControllerUp, time.Minute, time.Second).Should(Succeed())

		})

		It("should deploy Kruize components successfully", func() {
			By(fmt.Sprintf("creating namespace %s if it doesn't exist", namespace))
			cmd := exec.Command("kubectl", "create", "namespace", namespace)
			_, _ = utils.Run(cmd) // Ignore error if namespace already exists
			
			By("creating a Kruize custom resource")
			cmd = exec.Command("kubectl", "apply", "-f", "config/samples/v1alpha1_kruize.yaml", "-n", namespace)
			_, err := utils.Run(cmd)
			ExpectWithOffset(1, err).NotTo(HaveOccurred())

			By("checking that kruize namespace is created")
			Eventually(func() error {
				cmd := exec.Command("kubectl", "get", "namespace", namespace)
				_, err := utils.Run(cmd)
				return err
			}, 2*time.Minute, 5*time.Second).Should(Succeed())

			// Determine expected service account based on cluster type
			expectedSA := "default"
			if clusterType == "openshift" {
				expectedSA = "kruize-sa"
			}
			
			By(fmt.Sprintf("checking that Kruize ServiceAccount '%s' is created", expectedSA))
			Eventually(func() error {
				cmd := exec.Command("kubectl", "get", "serviceaccount", expectedSA, "-n", namespace)
				_, err := utils.Run(cmd)
				return err
			}, 2*time.Minute, 5*time.Second).Should(Succeed())

			By("checking that Kruize database is ready")
			Eventually(func() error {
				// Try kruize-db-deployment first, then kruize-db
				var cmd *exec.Cmd
				var output []byte
				var err error
				
				cmd = exec.Command("kubectl", "get", "deployment", "kruize-db-deployment", "-n", namespace, "-o", "jsonpath={.status.readyReplicas}")
				output, err = utils.Run(cmd)
				if err != nil {
					// Try kruize-db if kruize-db-deployment doesn't exist
					cmd = exec.Command("kubectl", "get", "deployment", "kruize-db", "-n", namespace, "-o", "jsonpath={.status.readyReplicas}")
					output, err = utils.Run(cmd)
					if err != nil {
						return fmt.Errorf("neither kruize-db-deployment nor kruize-db found: %w", err)
					}
				}
				
				if string(output) != "1" {
					return fmt.Errorf("database deployment not ready")
				}
				return nil
			}, 3*time.Minute, 10*time.Second).Should(Succeed())

			By("checking that Kruize deployment is ready")
			Eventually(func() error {
				cmd := exec.Command("kubectl", "get", "deployment", "kruize", "-n", namespace, "-o", "jsonpath={.status.readyReplicas}")
				output, err := utils.Run(cmd)
				if err != nil {
					return err
				}
				if string(output) != "1" {
					return fmt.Errorf("deployment not ready")
				}
				return nil
			}, 3*time.Minute, 10*time.Second).Should(Succeed())

			By("verifying deployed Kruize image")
			cmd = exec.Command("kubectl", "get", "deployment", "kruize", "-n", namespace, "-o", "jsonpath={.spec.template.spec.containers[0].image}")
			output, err := utils.Run(cmd)
			ExpectWithOffset(1, err).NotTo(HaveOccurred())
			deployedImage := strings.TrimSpace(string(output))
			fmt.Fprintf(GinkgoWriter, "Deployed Kruize image: %s\n", deployedImage)
			
			// If custom Kruize image was specified, verify it matches
			if kruizeImage != "" {
				if deployedImage == kruizeImage {
					fmt.Fprintf(GinkgoWriter, "✓ Deployed image matches specified KRUIZE_IMAGE: %s\n", kruizeImage)
				} else {
					fmt.Fprintf(GinkgoWriter, "⚠ Warning: Deployed image %s does not match specified KRUIZE_IMAGE %s\n", deployedImage, kruizeImage)
				}
			} else {
				fmt.Fprintf(GinkgoWriter, "Using default Kruize image from CR: %s\n", deployedImage)
			}

			By("verifying Kruize API is responding")
			Eventually(func() error {
				cmd := exec.Command("kubectl", "exec", "deployment/kruize", "-n", namespace, "--", "curl", "-s", "localhost:8080/health")
				_, err := utils.Run(cmd)
				return err
			}, 3*time.Minute, 10*time.Second).Should(Succeed())

			By("checking that datasource is configured via Kruize API")
			var datasourceOutput string
			Eventually(func() error {
				cmd := exec.Command("kubectl", "exec", "deployment/kruize", "-n", namespace, "--", "curl", "-s", "localhost:8080/datasources")
				output, err := utils.Run(cmd)
				if err != nil {
					return err
				}
				
				datasourceOutput = string(output)
				
				// Check for prometheus-1 datasource
				if !strings.Contains(datasourceOutput, "prometheus-1") {
					return fmt.Errorf("prometheus-1 datasource not found in API response")
				}
				return nil
			}, 3*time.Minute, 20*time.Second).Should(Succeed())

			By("printing datasources API output")
			fmt.Fprintf(GinkgoWriter, "\n=== Datasources API Output ===\n%s\n==============================\n", datasourceOutput)

			By("cleaning up the Kruize custom resource")
			cmd = exec.Command("kubectl", "delete", "-f", "config/samples/v1alpha1_kruize.yaml", "-n", namespace)
			_, _ = utils.Run(cmd)

			By("cleaning up the Kruize operator")
            			cmd = exec.Command("make", "undeploy")
            			_, _ = utils.Run(cmd)
		})
	})
})
