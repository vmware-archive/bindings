package resolver

import (
	"context"
	"errors"
	"fmt"

	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	pkgapisduck "knative.dev/pkg/apis/duck"
	duckv1 "knative.dev/pkg/apis/duck/v1"
	"knative.dev/pkg/client/injection/ducks/duck/v1/podspecable"
	"knative.dev/pkg/controller"
	"knative.dev/pkg/tracker"
)

// PodSpecableResolver resolves Destinations and ObjectReferences into a PodSpec.
type PodSpecableResolver struct {
	tracker         tracker.Interface
	informerFactory pkgapisduck.InformerFactory
}

// NewPodSpecableResolver constructs a LatestImageResolver with context and a callback
// for a given podspecableType passed to the PodSpecResolver's tracker.
func NewPodSpecableResolver(ctx context.Context, callback func(types.NamespacedName)) *PodSpecableResolver {
	ret := &PodSpecableResolver{}

	ret.tracker = tracker.New(callback, controller.GetTrackerLease(ctx))
	ret.informerFactory = &pkgapisduck.CachedInformerFactory{
		Delegate: &pkgapisduck.EnqueueInformerFactory{
			Delegate:     podspecable.Get(ctx),
			EventHandler: controller.HandleAll(ret.tracker.OnChanged),
		},
	}
	return ret
}

func (r *PodSpecableResolver) PodSpecableFromObjectReference(ref *tracker.Reference, parent interface{}) ([]*duckv1.WithPod, error) {
	if ref == nil {
		return nil, errors.New("ref is nil")
	}
	if err := r.tracker.TrackReference(*ref, parent); err != nil {
		return nil, fmt.Errorf("failed to track %+v: %v", ref, err)
	}
	gvr, _ := meta.UnsafeGuessKindToResource(ref.GroupVersionKind())
	_, lister, err := r.informerFactory.Get(gvr)
	if err != nil {
		return nil, err
	}

	var referents []*duckv1.WithPod
	if ref.Name != "" {
		psObj, err := lister.ByNamespace(ref.Namespace).Get(ref.Name)
		if err != nil {
			return nil, fmt.Errorf("failed to get resource for %+v: %v", gvr, err)
		}
		referents = append(referents, psObj.(*duckv1.WithPod))
	} else {
		selector, err := metav1.LabelSelectorAsSelector(ref.Selector)
		if err != nil {
			return nil, err
		}
		psObjs, err := lister.ByNamespace(ref.Namespace).List(selector)
		if err != nil {
			return nil, fmt.Errorf("failed to list resources for %+v: %v", gvr, err)
		}
		for _, psObj := range psObjs {
			referents = append(referents, psObj.(*duckv1.WithPod))
		}
	}
	return referents, nil
}
