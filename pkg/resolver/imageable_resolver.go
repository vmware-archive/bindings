package resolver

import (
	"context"
	"errors"
	"fmt"

	iduckv1alpha1 "github.com/projectriff/bindings/pkg/apis/duck/v1alpha1"
	"github.com/projectriff/bindings/pkg/client/injection/ducks/duck/v1alpha1/imageable"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/types"
	pkgapisduck "knative.dev/pkg/apis/duck"
	"knative.dev/pkg/controller"
	"knative.dev/pkg/tracker"
)

// URIResolver resolves Destinations and ObjectReferences into a URI.
type LatestImageResolver struct {
	tracker         tracker.Interface
	informerFactory pkgapisduck.InformerFactory
}

// NewLatestImageResolver constructs a LatestImageResolver with context and a callback
// for a given imageableType (Imageable) passed to the LatestImageResolver's tracker.
func NewLatestImageResolver(ctx context.Context, callback func(types.NamespacedName)) *LatestImageResolver {
	ret := &LatestImageResolver{}

	ret.tracker = tracker.New(callback, controller.GetTrackerLease(ctx))
	ret.informerFactory = &pkgapisduck.CachedInformerFactory{
		Delegate: &pkgapisduck.EnqueueInformerFactory{
			Delegate:     imageable.Get(ctx),
			EventHandler: controller.HandleAll(ret.tracker.OnChanged),
		},
	}
	return ret
}

func (r *LatestImageResolver) LatestImageFromObjectReference(ref *tracker.Reference, parent interface{}) (string, error) {
	if ref == nil {
		return "", errors.New("ref is nil")
	}
	if err := r.tracker.TrackReference(*ref, parent); err != nil {
		return "", fmt.Errorf("failed to track %+v: %v", ref, err)
	}
	gvr, _ := meta.UnsafeGuessKindToResource(ref.GroupVersionKind())
	_, lister, err := r.informerFactory.Get(gvr)
	if err != nil {
		return "", err
	}
	obj, err := lister.ByNamespace(ref.Namespace).Get(ref.Name)
	if err != nil {
		return "", fmt.Errorf("failed to get resource for %+v: %v", gvr, err)
	}
	imageable, ok := obj.(*iduckv1alpha1.ImageableType)
	if !ok {
		return "", fmt.Errorf("%+v (%T) is not an ImageableType", ref, ref)
	}
	return imageable.Status.LatestImage, nil
}
