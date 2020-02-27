package v1alpha1

import (
	"context"
)

// bindImagesKey is used as the key for associating information
// with a context.Context.
type bindImagesKey struct{}

// WithSinkURI notes on the context for binding that the resolved SinkURI
// is the provided apis.URL.
func WithImages(ctx context.Context, images map[string]string) context.Context {
	return context.WithValue(ctx, bindImagesKey{}, images)
}

// GetSinkURI accesses the apis.URL for the Sink URI that has been associated
// with this context.
func GetImages(ctx context.Context) map[string]string {
	value := ctx.Value(bindImagesKey{})
	if value == nil {
		return nil
	}
	return value.(map[string]string)
}
