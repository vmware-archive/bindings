package v1alpha1

import (
	"context"

	"knative.dev/pkg/apis"
	"knative.dev/pkg/apis/duck"
	v1 "knative.dev/pkg/apis/duck/v1"
	"knative.dev/pkg/tracker"
)

const (
	ImageBindingConditionReady            = apis.ConditionReady
	ImageBindingConditionBindingAvailable = "BindingAvailable"
)

var imgCondSet = apis.NewLivingConditionSet(
	ImageBindingConditionBindingAvailable,
)

func (b *ImageBinding) GetSubject() tracker.Reference {
	return *b.Spec.Subject
}

func (b *ImageBinding) GetBindingStatus() duck.BindableStatus {
	return &b.Status
}

func (b *ImageBinding) Do(ctx context.Context, ps *v1.WithPod) {
	images := GetImages(ctx)
	for _, p := range b.Spec.Providers {
		for i, c := range ps.Spec.Template.Spec.Containers {
			if c.Name == p.ContainerName {
				img, ok := images[p.ContainerName]
				if !ok {
					continue
				}
				ps.Spec.Template.Spec.Containers[i].Image = img
				ps.Annotations[ImageBindingAnnotationKey] = "bound"
			}
		}
	}
}

func (b *ImageBinding) Undo(ctx context.Context, ps *v1.WithPod) {
	// cannot undo
	delete(ps.Annotations, ImageBindingAnnotationKey)
}

func (bs *ImageBindingStatus) InitializeConditions() {
	imgCondSet.Manage(bs).InitializeConditions()
}

func (bs *ImageBindingStatus) MarkBindingAvailable() {
	imgCondSet.Manage(bs).MarkTrue(ImageBindingConditionBindingAvailable)
}

func (bs *ImageBindingStatus) MarkBindingUnavailable(reason string, message string) {
	imgCondSet.Manage(bs).MarkFalse(
		ImageBindingConditionBindingAvailable, reason, message)
}

func (bs *ImageBindingStatus) SetObservedGeneration(gen int64) {
	bs.ObservedGeneration = gen
}
