package v1alpha1

import (
	"context"

	corev1 "k8s.io/api/core/v1"
)

// bindServiceableBindingKey is used as the key for associating information
// with a context.Context.
type bindServiceableBindingKey struct{}

// WithServic notes the resolved image on the context for binding
func WithServiceableBinding(ctx context.Context, binding *corev1.LocalObjectReference) context.Context {
	return context.WithValue(ctx, bindServiceableBindingKey{}, binding)
}

// GetServiceableBinding accesses the resolved image that have been associated
// with this context.
func GetServiceableBinding(ctx context.Context) *corev1.LocalObjectReference {
	value := ctx.Value(bindServiceableBindingKey{})
	if value == nil {
		return nil
	}
	return value.(*corev1.LocalObjectReference)
}
