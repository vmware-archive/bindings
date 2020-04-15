package v1alpha1

import (
	"context"
	"fmt"
	"strings"

	"github.com/projectriff/bindings/pkg/mononoke/opinions"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"knative.dev/pkg/apis"
	"knative.dev/pkg/apis/duck"
	v1 "knative.dev/pkg/apis/duck/v1"
	"knative.dev/pkg/tracker"
)

const (
	SpringBootContainerConditionReady            = apis.ConditionReady
	SpringBootContainerConditionBindingAvailable = "BindingAvailable"
)

var sbcCondSet = apis.NewLivingConditionSet(
	SpringBootContainerConditionBindingAvailable,
)

func (b *SpringBootContainer) GetSubject() tracker.Reference {
	return *b.Spec.Subject
}

func (b *SpringBootContainer) GetBindingStatus() duck.BindableStatus {
	return &b.Status
}

func (b *SpringBootContainer) FindTargetContainer(podSpec corev1.PodSpec) (string, int, error) {
	switch b.Spec.TargetContainer.Type {
	case intstr.Int:
		idx := int(b.Spec.TargetContainer.IntVal)
		if l := len(podSpec.Containers); idx < l {
			c := podSpec.Containers[idx]
			return c.Name, idx, nil
		}
	case intstr.String:
		name := b.Spec.TargetContainer.StrVal
		for i, c := range podSpec.Containers {
			if c.Name == name {
				return name, i, nil
			}
		}
	}

	return "", 0, fmt.Errorf("Unable to find container %q", b.Spec.TargetContainer)
}

func (b *SpringBootContainer) Do(ctx context.Context, ps *v1.WithPod) {
	// undo existing bindings so we can start clean
	b.Undo(ctx, ps)

	imageMetadata := GetSpringBootContainerMetadata(ctx)[ps.Name]
	if b.Spec.ApplicationProperties == nil {
		b.Spec.ApplicationProperties = map[string]string{}
	}
	c := opinions.StashSpringApplicationProperties(ctx, b.Spec.ApplicationProperties)

	_, containerIdx, err := b.FindTargetContainer(ps.Spec.Template.Spec)
	if err != nil {
		// TODO log error
		return
	}
	applied, err := opinions.SpringBoot.Apply(c, ps, containerIdx, imageMetadata)
	if err != nil {
		// TODO log error
		return
	}

	ps.Annotations[SpringBootContainerRulesAnnotationKey] = strings.Join(applied, ",")
	ps.Annotations[SpringBootContainerAnnotationKey] = "bound"
}

func (b *SpringBootContainer) Undo(ctx context.Context, ps *v1.WithPod) {
	if ps.Annotations == nil {
		ps.Annotations = map[string]string{}
	}

	ps.Annotations[SpringBootContainerAnnotationKey] = "unbound"
}

func (bs *SpringBootContainerStatus) InitializeConditions() {
	sbcCondSet.Manage(bs).InitializeConditions()
}

func (bs *SpringBootContainerStatus) MarkBindingAvailable() {
	sbcCondSet.Manage(bs).MarkTrue(SpringBootContainerConditionBindingAvailable)
}

func (bs *SpringBootContainerStatus) MarkBindingUnavailable(reason string, message string) {
	sbcCondSet.Manage(bs).MarkFalse(SpringBootContainerConditionBindingAvailable, reason, message)
}

func (bs *SpringBootContainerStatus) SetObservedGeneration(gen int64) {
	bs.ObservedGeneration = gen
}
