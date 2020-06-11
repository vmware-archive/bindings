/*
Copyright 2020 The Knative Authors

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

package resources

import (
	"fmt"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"knative.dev/pkg/kmeta"

	servicev1alpha1 "github.com/projectriff/bindings/pkg/apis/service/v1alpha1"
)

func MakeProjectedSecret(binding *servicev1alpha1.ServiceBinding, reference *corev1.Secret) *corev1.Secret {
	projection := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			// TODO generate the secret name
			Name:      fmt.Sprintf("%s-projection", binding.Name),
			Namespace: binding.Namespace,
			Labels: kmeta.UnionMaps(binding.GetLabels(), map[string]string{
				servicev1alpha1.ServiceBindingLabelKey: binding.Name,
			}),
			OwnerReferences: []metav1.OwnerReference{*kmeta.NewControllerRef(binding)},
		},
	}

	projection.Data = reference.DeepCopy().Data
	if binding.Spec.Type != "" {
		projection.Data["type"] = []byte(binding.Spec.Type)
	}
	if binding.Spec.Provider != "" {
		projection.Data["provider"] = []byte(binding.Spec.Provider)
	}

	return projection
}
