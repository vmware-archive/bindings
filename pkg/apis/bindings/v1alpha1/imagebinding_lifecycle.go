package v1alpha1

import (
	"context"

	"knative.dev/pkg/apis"
	"knative.dev/pkg/apis/duck"
	v1 "knative.dev/pkg/apis/duck/v1"
	"knative.dev/pkg/tracker"
)

const (
	// ImageBindingConditionReady is set when the binding has been applied to the subjects.
	ImageBindingConditionReady = apis.ConditionReady
)

var imgCondSet = apis.NewLivingConditionSet()

func (ib *ImageBinding) GetSubject() tracker.Reference {
	return *ib.Spec.Subject
}

func (ib *ImageBinding) GetBindingStatus() duck.BindableStatus {
	return &ib.Status
}

func (ib *ImageBinding) Do(ctx context.Context, ps *v1.WithPod) {
	images := GetImages(ctx)
	for _, p := range ib.Spec.Providers {
		for i, c := range ps.Spec.Template.Spec.Containers {
			if c.Name == p.ContainerName {
				img, ok := images[p.ContainerName]
				if !ok {
					continue
				}
				ps.Spec.Template.Spec.Containers[i].Image = img
				ps.Annotations["io.projectriff.bindings"] = "bound"
			}
		}
	}
}

func (ib *ImageBinding) Undo(ctx context.Context, ps *v1.WithPod) {
	//cannot undo
	ps.Annotations["io.projectriff.bindings"] = "unbound"
}

func (ibs *ImageBindingStatus) InitializeConditions() {
	imgCondSet.Manage(ibs).InitializeConditions()
}

func (ibs *ImageBindingStatus) MarkBindingAvailable() {
	imgCondSet.Manage(ibs).MarkTrue(ImageBindingConditionReady)
}

func (ibs *ImageBindingStatus) MarkBindingUnavailable(reason string, message string) {
	imgCondSet.Manage(ibs).MarkFalse(
		ImageBindingConditionReady, reason, message)
}

func (ibs *ImageBindingStatus) SetObservedGeneration(gen int64) {
	ibs.ObservedGeneration = gen
}
