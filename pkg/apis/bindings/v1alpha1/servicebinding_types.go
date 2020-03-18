package v1alpha1

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/validation"
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
	Subject   *tracker.Reference          `json:"subject,omitempty"`
	Providers []ServiceCredentialProvider `json:"providers,omitempty"`
}

// ServiceCredentialProvider represents a single logical binding to inject into the subject
type ServiceCredentialProvider struct {
	// TODO switch to using remote references
	// Ref           *tracker.Reference `json:"ref, omitempty"`
	// Ref holds names for the metadata and secret
	Ref ServiceCredentialReference `json:"ref,omitempty"`
	// Name within the CNB_BINDINGS directory to mount the metadata and secrets
	Name string `json:"name"`
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
	for i, p := range b.Spec.Providers {
		if p.Name == "" {
			// TODO force uniqueness between providers
			errs = errs.Also(
				apis.ErrMissingField("name").ViaFieldIndex("spec.providers", i),
			)
		} else if out := validation.NameIsDNSLabel(p.Name, false); len(out) != 0 {
			errs = errs.Also(
				apis.ErrInvalidValue(p.Name, "name").ViaFieldIndex("spec.providers", i),
			)
		}
		if p.Ref.Metadata.Name == "" {
			errs = errs.Also(
				apis.ErrMissingField("ref.metadata.name").ViaFieldIndex("spec.providers", i),
			)
		}
		if p.Ref.Secret.Name == "" && p.BindingMode == SecretServiceBinding {
			errs = errs.Also(
				apis.ErrMissingField("ref.secret.name").ViaFieldIndex("spec.providers", i),
			)
		}
		if p.BindingMode != MetadataServiceBinding && p.BindingMode != SecretServiceBinding {
			errs = errs.Also(
				apis.ErrInvalidValue(p.BindingMode, "bindingMode").ViaFieldIndex("spec.providers", i),
			)
		}
	}

	return errs
}

func (b *ServiceBinding) SetDefaults(context.Context) {
	if b.Spec.Subject.Namespace == "" {
		// Default the subject's namespace to our namespace.
		b.Spec.Subject.Namespace = b.Namespace
	}
	for i, p := range b.Spec.Providers {
		if p.BindingMode == "" {
			b.Spec.Providers[i].BindingMode = SecretServiceBinding
		}
	}
}

func (b *ServiceBinding) GetGroupVersionKind() schema.GroupVersionKind {
	return SchemeGroupVersion.WithKind("ServiceBinding")
}
