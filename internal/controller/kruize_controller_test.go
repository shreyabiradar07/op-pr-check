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

package controller

import (
	"context"
	"os"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	mydomainv1alpha1 "github.com/kruize/kruize-operator/api/v1alpha1"
	"github.com/kruize/kruize-operator/internal/utils"
)

var _ = Describe("Kruize Controller", func() {
	ctx := context.Background()

	//setting test mode for the controller
	BeforeEach(func() {
		os.Setenv("KRUIZE_TEST_MODE", "true")
	})

	AfterEach(func() {
		os.Unsetenv("KRUIZE_TEST_MODE")
	})
	Context("Test mode behavior", func() {
		var reconciler *KruizeReconciler

		BeforeEach(func() {
			reconciler = &KruizeReconciler{
				Client: k8sClient,
				Scheme: k8sClient.Scheme(),
			}
		})

		It("should detect test mode correctly", func() {
			os.Setenv("KRUIZE_TEST_MODE", "true")
			Expect(reconciler.isTestMode()).To(BeTrue())

			os.Setenv("KRUIZE_TEST_MODE", "false")
			Expect(reconciler.isTestMode()).To(BeFalse())

			os.Unsetenv("KRUIZE_TEST_MODE")
			Expect(reconciler.isTestMode()).To(BeFalse())
		})

		It("should skip pod waiting in test mode", func() {
			os.Setenv("KRUIZE_TEST_MODE", "true")

			// This should return immediately without error
			err := reconciler.waitForKruizePods(ctx, "test-namespace", time.Second*1)
			Expect(err).NotTo(HaveOccurred())
		})
	})
	Context("When reconciling different cluster types", func() {
		It("should handle OpenShift cluster type", func() {
			kruize := &mydomainv1alpha1.Kruize{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-kruize-openshift",
					Namespace: "default",
				},
				Spec: mydomainv1alpha1.KruizeSpec{
					Cluster_type: "openshift",
					Namespace:    "openshift-tuning",
					Size:         1,
				},
			}
			Expect(k8sClient.Create(ctx, kruize)).To(Succeed())

			defer func() {
				Expect(k8sClient.Delete(ctx, kruize)).To(Succeed())
			}()

			controllerReconciler := &KruizeReconciler{
				Client: k8sClient,
				Scheme: k8sClient.Scheme(),
			}

			_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: types.NamespacedName{
					Name:      "test-kruize-openshift",
					Namespace: "default",
				},
			})
			Expect(err).NotTo(HaveOccurred())
		})

		It("should reject unsupported cluster types", func() {
			kruize := &mydomainv1alpha1.Kruize{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-kruize-invalid",
					Namespace: "default",
				},
				Spec: mydomainv1alpha1.KruizeSpec{
					Cluster_type: "invalid-cluster",
					Namespace:    "test",
					Size:         1,
				},
			}
			Expect(k8sClient.Create(ctx, kruize)).To(Succeed())

			defer func() {
				Expect(k8sClient.Delete(ctx, kruize)).To(Succeed())
			}()

			controllerReconciler := &KruizeReconciler{
				Client: k8sClient,
				Scheme: k8sClient.Scheme(),
			}

			_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: types.NamespacedName{
					Name:      "test-kruize-invalid",
					Namespace: "default",
				},
			})
			Expect(err).To(HaveOccurred())
		})
	})

	Context("Resource generation", func() {
		It("should generate cluster-scoped resources for OpenShift", func() {
			generator := utils.NewKruizeResourceGenerator("test-namespace", "", "", "openshift")

			clusterResources := generator.ClusterScopedResources()
			Expect(clusterResources).NotTo(BeEmpty())
			Expect(len(clusterResources)).To(BeNumerically(">", 0))
		})

		It("should generate namespaced resources for OpenShift", func() {
			generator := utils.NewKruizeResourceGenerator("test-namespace", "", "", "openshift")

			namespacedResources := generator.NamespacedResources()
			Expect(namespacedResources).NotTo(BeEmpty())
			Expect(len(namespacedResources)).To(BeNumerically(">", 0))
		})

		It("should generate Kubernetes cluster-scoped resources", func() {
			generator := utils.NewKruizeResourceGenerator("test-namespace", "", "", "minikube")

			clusterResources := generator.KubernetesClusterScopedResources()
			Expect(clusterResources).NotTo(BeEmpty())
			Expect(len(clusterResources)).To(BeNumerically(">", 0))
		})

		It("should generate Kubernetes namespaced resources", func() {
			generator := utils.NewKruizeResourceGenerator("test-namespace", "", "", "minikube")

			namespacedResources := generator.KubernetesNamespacedResources()
			Expect(namespacedResources).NotTo(BeEmpty())
			Expect(len(namespacedResources)).To(BeNumerically(">", 0))
		})

		It("should use default images when not specified", func() {
			generator := utils.NewKruizeResourceGenerator("test-namespace", "", "", "openshift")

			Expect(generator.Autotune_image).To(Equal("quay.io/kruize/autotune_operator:latest"))
			Expect(generator.Autotune_ui_image).To(Equal("quay.io/kruize/kruize-ui:0.0.9"))
		})

		It("should use custom images when specified", func() {
			customImage := "custom/image:v1.0"
			customUIImage := "custom/ui:v1.0"
			generator := utils.NewKruizeResourceGenerator("test-namespace", customImage, customUIImage, "openshift")

			Expect(generator.Autotune_image).To(Equal(customImage))
			Expect(generator.Autotune_ui_image).To(Equal(customUIImage))
		})
	})

})
