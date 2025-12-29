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

package utils

import (
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"

	. "github.com/onsi/ginkgo/v2" //nolint:golint,revive
)

const (
	prometheusOperatorVersion = "v0.72.0"
	prometheusOperatorURL     = "https://github.com/prometheus-operator/prometheus-operator/" +
		"releases/download/%s/bundle.yaml"

	certmanagerVersion = "v1.14.4"
	certmanagerURLTmpl = "https://github.com/jetstack/cert-manager/releases/download/%s/cert-manager.yaml"
)

func warnError(err error) {
	fmt.Fprintf(GinkgoWriter, "warning: %v\n", err)
}

// InstallPrometheusOperator installs the prometheus Operator to be used to export the enabled metrics.
func InstallPrometheusOperator() error {
	// Check if prometheus-operator deployment already exists
	checkCmd := exec.Command("kubectl", "get", "deployment", "prometheus-operator", "-n", "default")
	_, err := Run(checkCmd)
	if err == nil {
		fmt.Fprintf(GinkgoWriter, "Prometheus Operator already installed, skipping installation\n")
		return nil
	}

	url := fmt.Sprintf(prometheusOperatorURL, prometheusOperatorVersion)
	// Use server-side apply to handle large CRDs
	cmd := exec.Command("kubectl", "apply", "--server-side", "-f", url)
	_, err = Run(cmd)
	return err
}

// Run executes the provided command within this context
func Run(cmd *exec.Cmd) ([]byte, error) {
	dir, _ := GetProjectDir()
	cmd.Dir = dir

	if err := os.Chdir(cmd.Dir); err != nil {
		fmt.Fprintf(GinkgoWriter, "chdir dir: %s\n", err)
	}

	cmd.Env = append(os.Environ(), "GO111MODULE=on")
	command := strings.Join(cmd.Args, " ")
	fmt.Fprintf(GinkgoWriter, "running: %s\n", command)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return output, fmt.Errorf("%s failed with error: (%v) %s", command, err, string(output))
	}

	return output, nil
}

// UninstallPrometheusOperator uninstalls the prometheus
func UninstallPrometheusOperator() {
	url := fmt.Sprintf(prometheusOperatorURL, prometheusOperatorVersion)
	cmd := exec.Command("kubectl", "delete", "-f", url)
	if _, err := Run(cmd); err != nil {
		warnError(err)
	}
}

// UninstallCertManager uninstalls the cert manager
func UninstallCertManager() {
	url := fmt.Sprintf(certmanagerURLTmpl, certmanagerVersion)
	cmd := exec.Command("kubectl", "delete", "-f", url)
	if _, err := Run(cmd); err != nil {
		warnError(err)
	}
}

// InstallCertManager installs the cert manager bundle.
func InstallCertManager() error {
	url := fmt.Sprintf(certmanagerURLTmpl, certmanagerVersion)
	cmd := exec.Command("kubectl", "apply", "-f", url)
	if _, err := Run(cmd); err != nil {
		return err
	}
	// Wait for cert-manager-webhook to be ready, which can take time if cert-manager
	// was re-installed after uninstalling on a cluster.
	cmd = exec.Command("kubectl", "wait", "deployment.apps/cert-manager-webhook",
		"--for", "condition=Available",
		"--namespace", "cert-manager",
		"--timeout", "5m",
	)

	_, err := Run(cmd)
	return err
}

// LoadImageToKindCluster loads a local docker image to the kind cluster
func LoadImageToKindClusterWithName(name string) error {
	cluster := "kind"
	if v, ok := os.LookupEnv("KIND_CLUSTER"); ok {
		cluster = v
	}
	kindOptions := []string{"load", "docker-image", name, "--name", cluster}
	cmd := exec.Command("kind", kindOptions...)
	_, err := Run(cmd)
	return err
}

// GetNonEmptyLines converts given command output string into individual objects
// according to line breakers, and ignores the empty elements in it.
func GetNonEmptyLines(output string) []string {
	var res []string
	elements := strings.Split(output, "\n")
	for _, element := range elements {
		if element != "" {
			res = append(res, element)
		}
	}

	return res
}

// GetProjectDir will return the directory where the project is
func GetProjectDir() (string, error) {
	wd, err := os.Getwd()
	if err != nil {
		return wd, err
	}
	wd = strings.Replace(wd, "/test/e2e", "", -1)
	return wd, nil
}

// UpdateKruizeSampleYAML updates the v1alpha1_kruize.yaml file with the specified cluster type, namespace, and optional images
func UpdateKruizeSampleYAML(clusterType, namespace, kruizeImage, kruizeUIImage string) error {
	// Update cluster_type
	cmd := exec.Command("sed", "-i",
		fmt.Sprintf("s/cluster_type: \"[^\"]*\"/cluster_type: \"%s\"/g", clusterType),
		"config/samples/v1alpha1_kruize.yaml")
	if _, err := Run(cmd); err != nil {
		return fmt.Errorf("failed to update cluster_type: %w", err)
	}

	// Update namespace
	cmd = exec.Command("sed", "-i",
		fmt.Sprintf("s/namespace: \"[^\"]*\"/namespace: \"%s\"/g", namespace),
		"config/samples/v1alpha1_kruize.yaml")
	if _, err := Run(cmd); err != nil {
		return fmt.Errorf("failed to update namespace: %w", err)
	}

	// Update autotune_image if specified
	if kruizeImage != "" {
		cmd = exec.Command("sed", "-i",
			fmt.Sprintf("s|autotune_image: \"[^\"]*\"|autotune_image: \"%s\"|g", kruizeImage),
			"config/samples/v1alpha1_kruize.yaml")
		if _, err := Run(cmd); err != nil {
			return fmt.Errorf("failed to update autotune_image: %w", err)
		}
		fmt.Fprintf(GinkgoWriter, "Updated autotune_image to %s\n", kruizeImage)
	}

	// Update autotune_ui_image if specified
	if kruizeUIImage != "" {
		cmd = exec.Command("sed", "-i",
			fmt.Sprintf("s|autotune_ui_image: \"[^\"]*\"|autotune_ui_image: \"%s\"|g", kruizeUIImage),
			"config/samples/v1alpha1_kruize.yaml")
		if _, err := Run(cmd); err != nil {
			return fmt.Errorf("failed to update autotune_ui_image: %w", err)
		}
		fmt.Fprintf(GinkgoWriter, "Updated autotune_ui_image to %s\n", kruizeUIImage)
	}

	fmt.Fprintf(GinkgoWriter, "Updated v1alpha1_kruize.yaml with cluster_type=%s, namespace=%s\n", clusterType, namespace)
	return nil
}


// GetDefaultOperatorImage reads the default operator image from Makefile
func GetDefaultOperatorImage() (string, error) {
	// Use make to print the IMG variable value
	cmd := exec.Command("make", "-s", "--no-print-directory", "-f", "Makefile", "print-img")
	output, err := cmd.CombinedOutput()
	if err != nil {
		// Fallback: try to extract from Makefile directly
		return extractImageFromMakefile()
	}
	
	image := strings.TrimSpace(string(output))
	if image != "" {
		return image, nil
	}
	
	// Fallback if make command didn't work
	return extractImageFromMakefile()
}

// extractImageFromMakefile extracts the operator image from Makefile
func extractImageFromMakefile() (string, error) {
	content, err := os.ReadFile("Makefile")
	if err != nil {
		return "", fmt.Errorf("failed to read Makefile: %w", err)
	}
	
	// Extract IMAGE_TAG_BASE and VERSION
	imageTagBaseRe := regexp.MustCompile(`IMAGE_TAG_BASE\s*\?=\s*(.+)`)
	versionRe := regexp.MustCompile(`VERSION\s*\?=\s*(.+)`)
	
	imageTagBaseMatch := imageTagBaseRe.FindStringSubmatch(string(content))
	versionMatch := versionRe.FindStringSubmatch(string(content))
	
	if len(imageTagBaseMatch) < 2 || len(versionMatch) < 2 {
		return "", fmt.Errorf("failed to extract IMAGE_TAG_BASE or VERSION from Makefile")
	}
	
	imageTagBase := strings.TrimSpace(imageTagBaseMatch[1])
	version := strings.TrimSpace(versionMatch[1])
	
	return fmt.Sprintf("%s:%s", imageTagBase, version), nil
}
