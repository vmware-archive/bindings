package v1alpha1

import (
	"context"

	iduckv1alpha1 "github.com/projectriff/bindings/pkg/apis/duck/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"knative.dev/pkg/apis"
	duckv1beta1 "knative.dev/pkg/apis/duck/v1beta1"
	"knative.dev/pkg/kmeta"
)

const (
	BindableServiceAnnotationKey = GroupName + "/bindable-service"
)

// +genclient
// +genreconciler
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type BindableService struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   BindableServiceSpec   `json:"spec,omitempty"`
	Status BindableServiceStatus `json:"status,omitempty"`
}

var (
	// Check that BindableService can be validated and defaulted.
	_ apis.Validatable   = (*BindableService)(nil)
	_ apis.Defaultable   = (*BindableService)(nil)
	_ kmeta.OwnerRefable = (*BindableService)(nil)
)

type BindableServiceSpec struct {
	Binding iduckv1alpha1.ServiceableBinding `json:"binding,omitempty"`
}

type BindableServiceStatus struct {
	duckv1beta1.Status `json:",inline"`
	Binding            iduckv1alpha1.ServiceableBinding `json:"binding,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type BindableServiceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []BindableService `json:"items"`
}

func (b *BindableService) Validate(ctx context.Context) (errs *apis.FieldError) {
	if b.Spec.Binding.Metadata.Name == "" {
		errs = errs.Also(
			apis.ErrMissingField("spec.binding.metadata.name"),
		)
	}

	return errs
}

func (b *BindableService) SetDefaults(context.Context) {
	// nothing to do
}

func (b *BindableService) GetGroupVersionKind() schema.GroupVersionKind {
	return SchemeGroupVersion.WithKind("BindableService")
}
