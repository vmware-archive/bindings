package v1alpha1

import (
	"context"

	"github.com/projectriff/bindings/pkg/mononoke/cnb"
)

// springBootContainerKey is used as the key for associating information with a
// context.Context.
type springBootContainerMetatdataKey struct{}

// WithSpringBootContainerMetadata notes the resolved image metadata on the
// context for binding
func WithSpringBootContainerMetadata(ctx context.Context, value map[string]cnb.BuildMetadata) context.Context {
	return context.WithValue(ctx, springBootContainerMetatdataKey{}, value)
}

// GetSpringBootContainerMetadata accesses the resolved image metadata that has
// been associated with this context.
func GetSpringBootContainerMetadata(ctx context.Context) map[string]cnb.BuildMetadata {
	value := ctx.Value(springBootContainerMetatdataKey{})
	if value == nil {
		return nil
	}
	return value.(map[string]cnb.BuildMetadata)
}
