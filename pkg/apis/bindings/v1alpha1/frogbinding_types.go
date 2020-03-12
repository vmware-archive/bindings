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
	FrogBindingAnnotationKey = GroupName + "/frog-binding"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type FrogBinding struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   FrogBindingSpec   `json:"spec,omitempty"`
	Status FrogBindingStatus `json:"status,omitempty"`
}

var (
	// Check that FrogBinding can be validated and defaulted.
	_ apis.Validatable   = (*FrogBinding)(nil)
	_ apis.Defaultable   = (*FrogBinding)(nil)
	_ kmeta.OwnerRefable = (*FrogBinding)(nil)

	// Check is Bindable
	_ psbinding.Bindable  = (*FrogBinding)(nil)
	_ duck.BindableStatus = (*FrogBindingStatus)(nil)
)

type FrogBindingSpec struct {
	Subject   *tracker.Reference `json:"subject,omitempty"`
	Providers []FrogProvider     `json:"providers,omitempty"`
}

// FrogProvider represents a single logical binding to inject into the subject
type FrogProvider struct {
	// TODO switch to using remote references
	// Ref           *tracker.Reference `json:"ref, omitempty"`
	// Ref holds names for the metadata and secret
	Ref           FrogReference   `json:"ref,omitempty"`
	// Name within the CNB_BINDINGS directory to mount the metadata and secrets
	Name          string          `json:"name"`
	// ContainerName targets a specific container within the subject to inject
	// into. If not set, all container will be injected
	ContainerName string          `json:"containerName,omitempty"`
	// BindingMode restricts which aspects of the provisioned binding are
	// exposed to the subject:
	// - Metadata: mounts only the metadata 
	// - Secret: mounts the metadata and secret
	BindingMode   FrogBindingMode `json:"bindingMode,omitempty"`
}

type FrogReference struct {
	Metadata corev1.LocalObjectReference `json:"metadata"`
	Secret   corev1.LocalObjectReference `json:"secret"`
}

type FrogBindingMode string

const (
	MetadataFrogBinding FrogBindingMode = "Metadata"
	SecretFrogBinding   FrogBindingMode = "Secret"
)

type FrogBindingStatus struct {
	duckv1beta1.Status `json:",inline"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type FrogBindingList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []FrogBinding `json:"items"`
}

func (b *FrogBinding) Validate(ctx context.Context) (errs *apis.FieldError) {
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
		if p.Ref.Secret.Name == "" && p.BindingMode == SecretFrogBinding {
			errs = errs.Also(
				apis.ErrMissingField("ref.secret.name").ViaFieldIndex("spec.providers", i),
			)
		}
		if p.BindingMode != MetadataFrogBinding && p.BindingMode != SecretFrogBinding {
			errs = errs.Also(
				apis.ErrInvalidValue(p.BindingMode, "bindingMode").ViaFieldIndex("spec.providers", i),
			)
		}
	}

	return errs
}

func (b *FrogBinding) SetDefaults(context.Context) {
	if b.Spec.Subject.Namespace == "" {
		// Default the subject's namespace to our namespace.
		b.Spec.Subject.Namespace = b.Namespace
	}
	for i, p := range b.Spec.Providers {
		if p.BindingMode == "" {
			b.Spec.Providers[i].BindingMode = SecretFrogBinding
		}
	}
}

func (b *FrogBinding) GetGroupVersionKind() schema.GroupVersionKind {
	return SchemeGroupVersion.WithKind("FrogBinding")
}
