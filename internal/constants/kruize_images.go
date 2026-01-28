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

import "os"

// Default container image versions
const (
	// defaultAutotuneImageTag is the default tag for Kruize Autotune image
	defaultAutotuneImageTag = "0.8.1"
	
	// defaultAutotuneUIImageTag is the default tag for Kruize UI image
	defaultAutotuneUIImageTag = "0.0.9"
	
	// defaultAutotuneImageRepo is the default repository for Kruize Autotune image
	defaultAutotuneImageRepo = "quay.io/kruize/autotune_operator"
	
	// defaultAutotuneUIImageRepo is the default repository for Kruize UI image
	defaultAutotuneUIImageRepo = "quay.io/kruize/kruize-ui"
)

// GetDefaultAutotuneImage returns the default Autotune image, checking environment variables first
func GetDefaultAutotuneImage() string {
	// Check for environment variable override
	if envImage := os.Getenv("DEFAULT_AUTOTUNE_IMAGE"); envImage != "" {
		return envImage
	}
	return defaultAutotuneImageRepo + ":" + defaultAutotuneImageTag
}

// GetDefaultAutotuneUIImage returns the default Autotune UI image, checking environment variables first
func GetDefaultAutotuneUIImage() string {
	// Check for environment variable override
	if envImage := os.Getenv("DEFAULT_AUTOTUNE_UI_IMAGE"); envImage != "" {
		return envImage
	}
	return defaultAutotuneUIImageRepo + ":" + defaultAutotuneUIImageTag
}
