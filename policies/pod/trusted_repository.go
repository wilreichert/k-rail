// Copyright 2019 Cruise LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//    https://www.apache.org/licenses/LICENSE-2.0
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package pod

import (
	"context"
	"regexp"

	admissionv1 "k8s.io/api/admission/v1"
	corev1 "k8s.io/api/core/v1"

	"github.com/cruise-automation/k-rail/v3/policies"
	"github.com/cruise-automation/k-rail/v3/resource"
)

type PolicyTrustedRepository struct{}

func (p PolicyTrustedRepository) Name() string {
	return "pod_trusted_repository"
}

func (p PolicyTrustedRepository) Validate(ctx context.Context, config policies.Config, ar *admissionv1.AdmissionRequest) ([]policies.ResourceViolation, []policies.PatchOperation) {

	resourceViolations := []policies.ResourceViolation{}

	podResource := resource.GetPodResource(ctx, ar)
	if podResource == nil {
		return resourceViolations, nil
	}

	validateContainer := func(container corev1.Container) {
		matches := 0
		for _, pattern := range config.PolicyTrustedRepositoryRegexes {
			if matched, _ := regexp.MatchString(pattern, container.Image); matched {
				matches++
			}
		}

		if matches == 0 {
			violationText := "Trusted Image Repository: image must be sourced from a trusted repository. Untrusted Images: " + container.Image
			resourceViolations = append(resourceViolations, policies.ResourceViolation{
				Namespace:    ar.Namespace,
				ResourceName: podResource.ResourceName,
				ResourceKind: podResource.ResourceKind,
				Violation:    violationText,
				Policy:       p.Name(),
			})
		}
	}

	for _, container := range podResource.PodSpec.Containers {
		validateContainer(container)
	}

	for _, container := range podResource.PodSpec.InitContainers {
		validateContainer(container)
	}

	return resourceViolations, nil
}
