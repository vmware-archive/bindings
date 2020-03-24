package v1alpha1

import (
	"context"
	"strings"

	jsoniter "github.com/json-iterator/go"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	"knative.dev/pkg/apis"
	"knative.dev/pkg/apis/duck"
	duckv1 "knative.dev/pkg/apis/duck/v1"
	"knative.dev/pkg/tracker"
)

const (
	ImageBindingConditionReady            = apis.ConditionReady
	ImageBindingConditionBindingAvailable = "BindingAvailable"
	ImageBindingAnnotationKeyPrefix       = "imagebindings." + GroupName + "/"
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

type imageBindingAnnotationData struct {
	Container string `json:"container"`
	Provider  string `json:"provider"`
}

func (b *ImageBinding) annotationKey() string {
	return ImageBindingAnnotationKeyPrefix + b.Name
}

func (b *ImageBinding) annotationVal() string {
	gvr, _ := meta.UnsafeGuessKindToResource(b.Spec.Provider.GroupVersionKind())
	ba := &imageBindingAnnotationData{
		Container: b.Spec.ContainerName,
		Provider:  gvr.GroupResource().String() + "/" + b.Spec.Provider.Name,
	}
	data, _ := jsoniter.MarshalToString(ba)
	return data
}

func (b *ImageBinding) Do(ctx context.Context, ps *duckv1.WithPod) {
	latestImage := GetLatestImage(ctx)
	if latestImage == "" {
		return
	}
	b.Undo(ctx, ps)
	b.doContainers(ps.Spec.Template.Spec.Containers, ps, latestImage)
	b.doContainers(ps.Spec.Template.Spec.InitContainers, ps, latestImage)
}

func (b *ImageBinding) doContainers(ctrs []corev1.Container, ps *duckv1.WithPod, latestImage string) {
	for i, c := range ctrs {
		if c.Name != b.Spec.ContainerName {
			continue
		}
		if belongsToAnotherBinding(c.Name, ps) {
			return
		}
		ctrs[i].Image = latestImage
		if ps.Annotations == nil {
			ps.Annotations = map[string]string{}
		}
		ps.Annotations[b.annotationKey()] = b.annotationVal()
	}
}

func belongsToAnotherBinding(container string, ps *duckv1.WithPod) bool {
	for k, v := range ps.Annotations {
		if !strings.HasPrefix(k, ImageBindingAnnotationKeyPrefix) {
			continue
		}
		var a imageBindingAnnotationData
		if err := jsoniter.UnmarshalFromString(v, &a); err != nil {
			continue
		}
		if a.Container == container {
			return true
		}
	}
	return false
}

func (b *ImageBinding) Undo(ctx context.Context, ps *duckv1.WithPod) {
	// cannot undo image update
	// removes binding annotation
	delete(ps.Annotations, b.annotationKey())
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
