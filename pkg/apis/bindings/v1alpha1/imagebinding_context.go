package v1alpha1

import (
	"context"
)

// bindLatestImageKey is used as the key for associating information
// with a context.Context.
type bindLatestImageKey struct{}

// WithLatestImage notes the resolved image on the context for binding
func WithLatestImage(ctx context.Context, img string) context.Context {
	return context.WithValue(ctx, bindLatestImageKey{}, img)
}

// GetLatestImage accesses the resolved image that have been associated
// with this context.
func GetLatestImage(ctx context.Context) string {
	value := ctx.Value(bindLatestImageKey{})
	if value == nil {
		return ""
	}
	return value.(string)
}
