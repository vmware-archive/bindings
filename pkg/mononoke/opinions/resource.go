/*
Copyright 2020 the original author or authors.

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

package opinions

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1 "knative.dev/pkg/apis/duck/v1"
)

type Resource interface {
	metav1.ObjectMetaAccessor
	PodTemplate() *corev1.PodTemplateSpec
}

// setAnnotation sets the annotation on both the resource and the resource's
// PodTemplateSpec
func setAnnotation(r *v1.WithPod, key, value string) {
	if r.Annotations == nil {
		r.Annotations = map[string]string{}
	}
	r.Annotations[key] = value

	if r.Spec.Template.Annotations == nil {
		r.Spec.Template.Annotations = map[string]string{}
	}
	r.Spec.Template.Annotations[key] = value
}

// setLabel sets the label on both the resource and the resource's
// PodTemplateSpec
func setLabel(r *v1.WithPod, key, value string) {
	if r.Labels == nil {
		r.Labels = map[string]string{}
	}
	r.Labels[key] = value

	if r.Spec.Template.Labels == nil {
		r.Spec.Template.Labels = map[string]string{}
	}
	r.Spec.Template.Labels[key] = value
}
