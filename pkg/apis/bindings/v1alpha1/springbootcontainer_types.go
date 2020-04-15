package v1alpha1

import (
	"context"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/intstr"
	"knative.dev/pkg/apis"
	"knative.dev/pkg/apis/duck"
	duckv1beta1 "knative.dev/pkg/apis/duck/v1beta1"
	"knative.dev/pkg/kmeta"
	"knative.dev/pkg/tracker"
	"knative.dev/pkg/webhook/psbinding"
)

const (
	SpringBootContainerAnnotationKey      = GroupName + "/spring-boot-container"
	SpringBootContainerRulesAnnotationKey = GroupName + "/spring-boot-container-rules"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type SpringBootContainer struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   SpringBootContainerSpec   `json:"spec,omitempty"`
	Status SpringBootContainerStatus `json:"status,omitempty"`
}

var (
	// Check that SpringBootContainer can be validated and defaulted.
	_ apis.Validatable   = (*SpringBootContainer)(nil)
	_ apis.Defaultable   = (*SpringBootContainer)(nil)
	_ kmeta.OwnerRefable = (*SpringBootContainer)(nil)

	// Check is Bindable
	_ psbinding.Bindable  = (*SpringBootContainer)(nil)
	_ duck.BindableStatus = (*SpringBootContainerStatus)(nil)
)

type SpringBootContainerSpec struct {
	// Subject resource to bind into
	Subject *tracker.Reference `json:"subject,omitempty"`

	// TargetContainer is name or index of the container to advise in the template
	// Defaults to the first container
	// +optional
	TargetContainer *intstr.IntOrString `json:"targetContainer,omitempty"`

	// ApplicationProperties to be included in the target application container
	// +optional
	ApplicationProperties map[string]string `json:"applicationProperties,omitempty"`
}

type SpringBootContainerStatus struct {
	duckv1beta1.Status `json:",inline"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type SpringBootContainerList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []SpringBootContainer `json:"items"`
}

func (b *SpringBootContainer) Validate(ctx context.Context) (errs *apis.FieldError) {
	errs = errs.Also(
		b.Spec.Subject.Validate(ctx).ViaField("spec.subject"),
	)
	if b.Spec.Subject.Namespace != b.Namespace {
		errs = errs.Also(
			apis.ErrInvalidValue(b.Spec.Subject.Namespace, "spec.subject.namespace"),
		)
	}

	return errs
}

func (b *SpringBootContainer) SetDefaults(context.Context) {
	if b.Spec.Subject == nil {
		b.Spec.Subject = &tracker.Reference{}
	}
	if b.Spec.Subject.Namespace == "" {
		b.Spec.Subject.Namespace = b.Namespace
	}
	if b.Spec.TargetContainer == nil {
		target := intstr.FromInt(0)
		b.Spec.TargetContainer = &target
	}
	if b.Spec.ApplicationProperties == nil {
		b.Spec.ApplicationProperties = map[string]string{}
	}
}

func (b *SpringBootContainer) GetGroupVersionKind() schema.GroupVersionKind {
	return SchemeGroupVersion.WithKind("SpringBootContainer")
}
