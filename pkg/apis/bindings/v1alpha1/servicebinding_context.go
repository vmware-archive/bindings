package v1alpha1

import (
	"context"

	iduckv1alpha1 "github.com/projectriff/bindings/pkg/apis/duck/v1alpha1"
)

// bindServiceableBindingKey is used as the key for associating information
// with a context.Context.
type bindServiceableBindingKey struct{}

// WithServic notes the resolved image on the context for binding
func WithServiceableBinding(ctx context.Context, binding *iduckv1alpha1.ServiceableBinding) context.Context {
	return context.WithValue(ctx, bindServiceableBindingKey{}, binding)
}

// GetServiceableBinding accesses the resolved image that have been associated
// with this context.
func GetServiceableBinding(ctx context.Context) *iduckv1alpha1.ServiceableBinding {
	value := ctx.Value(bindServiceableBindingKey{})
	if value == nil {
		return nil
	}
	return value.(*iduckv1alpha1.ServiceableBinding)
}
