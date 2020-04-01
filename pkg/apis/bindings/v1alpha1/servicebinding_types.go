package v1alpha1

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"knative.dev/pkg/apis"
	"knative.dev/pkg/apis/duck"
	duckv1beta1 "knative.dev/pkg/apis/duck/v1beta1"
	"knative.dev/pkg/kmeta"
	"knative.dev/pkg/tracker"
	"knative.dev/pkg/webhook/psbinding"
)

const (
	ServiceBindingAnnotationKey = GroupName + "/service-binding"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type ServiceBinding struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ServiceBindingSpec   `json:"spec,omitempty"`
	Status ServiceBindingStatus `json:"status,omitempty"`
}

var (
	// Check that ServiceBinding can be validated and defaulted.
	_ apis.Validatable   = (*ServiceBinding)(nil)
	_ apis.Defaultable   = (*ServiceBinding)(nil)
	_ kmeta.OwnerRefable = (*ServiceBinding)(nil)

	// Check is Bindable
	_ psbinding.Bindable  = (*ServiceBinding)(nil)
	_ duck.BindableStatus = (*ServiceBindingStatus)(nil)
)

type ServiceBindingSpec struct {
	// Subject resource to bind into
	Subject *tracker.Reference `json:"subject,omitempty"`
	// Bining reference to the binding metadata and secret
	Binding ServiceCredentialReference `json:"binding,omitempty"`
	// ContainerName targets a specific container within the subject to inject
	// into. If not set, all container will be injected
	ContainerName string `json:"containerName,omitempty"`
	// BindingMode restricts which aspects of the provisioned binding are
	// exposed to the subject:
	// - Metadata: mounts only the metadata
	// - Secret: mounts the metadata and secret
	BindingMode ServiceBindingMode `json:"bindingMode,omitempty"`
}

type ServiceCredentialReference struct {
	Metadata corev1.LocalObjectReference `json:"metadata"`
	Secret   corev1.LocalObjectReference `json:"secret"`
}

type ServiceBindingMode string

const (
	MetadataServiceBinding ServiceBindingMode = "Metadata"
	SecretServiceBinding   ServiceBindingMode = "Secret"
)

type ServiceBindingStatus struct {
	duckv1beta1.Status `json:",inline"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type ServiceBindingList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ServiceBinding `json:"items"`
}

func (b *ServiceBinding) Validate(ctx context.Context) (errs *apis.FieldError) {
	b.Spec.Subject.Validate(ctx)
	if b.Spec.Subject.Namespace != b.Namespace {
		errs = errs.Also(
			apis.ErrInvalidValue(b.Spec.Subject.Namespace, "spec.subject.namespace"),
		)
	}
	if b.Spec.Binding.Metadata.Name == "" {
		errs = errs.Also(
			apis.ErrMissingField("spec.binding.metadata.name"),
		)
	}
	if b.Spec.Binding.Secret.Name == "" && b.Spec.BindingMode == SecretServiceBinding {
		errs = errs.Also(
			apis.ErrMissingField("spec.binding.secret.name"),
		)
	}
	if b.Spec.BindingMode != MetadataServiceBinding && b.Spec.BindingMode != SecretServiceBinding {
		errs = errs.Also(
			apis.ErrInvalidValue(b.Spec.BindingMode, "spec.bindingMode"),
		)
	}

	return errs
}

func (b *ServiceBinding) SetDefaults(context.Context) {
	if b.Spec.Subject.Namespace == "" {
		// Default the subject's namespace to our namespace.
		b.Spec.Subject.Namespace = b.Namespace
	}
	if b.Spec.BindingMode == "" {
		b.Spec.BindingMode = SecretServiceBinding
	}
}

func (b *ServiceBinding) GetGroupVersionKind() schema.GroupVersionKind {
	return SchemeGroupVersion.WithKind("ServiceBinding")
}
