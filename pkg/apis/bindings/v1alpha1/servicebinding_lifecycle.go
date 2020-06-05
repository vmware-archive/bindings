package v1alpha1

import (
	"context"
	"fmt"
	"strings"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/sets"
	"knative.dev/pkg/apis"
	"knative.dev/pkg/apis/duck"
	v1 "knative.dev/pkg/apis/duck/v1"
	"knative.dev/pkg/tracker"
)

const (
	ServiceBindingConditionReady            = apis.ConditionReady
	ServiceBindingConditionBindingAvailable = "BindingAvailable"
)

var serviceCondSet = apis.NewLivingConditionSet(
	ServiceBindingConditionBindingAvailable,
)

func (b *ServiceBinding) GetSubject() tracker.Reference {
	return *b.Spec.Subject
}

func (b *ServiceBinding) GetBindingStatus() duck.BindableStatus {
	return &b.Status
}

func (b *ServiceBinding) Do(ctx context.Context, ps *v1.WithPod) {
	// undo existing bindings so we can start clean
	b.Undo(ctx, ps)

	existingVolumes := sets.NewString()
	for _, v := range ps.Spec.Template.Spec.Volumes {
		existingVolumes.Insert(v.Name)
	}

	newVolumes := sets.NewString()
	sb := GetServiceableBinding(ctx)
	if sb == nil {
		return
	}
	// TODO ensure unique volume names
	// TODO limit volume name length
	bindingVolume := fmt.Sprintf("%s-binding", sb.Name)
	if !existingVolumes.Has(bindingVolume) {
		ps.Spec.Template.Spec.Volumes = append(ps.Spec.Template.Spec.Volumes, corev1.Volume{
			Name: bindingVolume,
			VolumeSource: corev1.VolumeSource{
				ConfigMap: &corev1.ConfigMapVolumeSource{
					LocalObjectReference: *sb,
				},
			},
		})
		existingVolumes.Insert(bindingVolume)
		newVolumes.Insert(bindingVolume)
	}
	for i := range ps.Spec.Template.Spec.InitContainers {
		b.DoContainer(ctx, ps, &ps.Spec.Template.Spec.InitContainers[i], bindingVolume)
	}
	for i := range ps.Spec.Template.Spec.Containers {
		b.DoContainer(ctx, ps, &ps.Spec.Template.Spec.Containers[i], bindingVolume)
	}

	// track which volumes are injected, so they can be removed when no longer used
	ps.Annotations[b.annotationKey()] = strings.Join(newVolumes.List(), ",")
}

func (b *ServiceBinding) DoContainer(ctx context.Context, ps *v1.WithPod, c *corev1.Container, bindingVolume string) {
	if c.Name != b.Spec.ContainerName && b.Spec.ContainerName != "" {
		// ignore the container
		return
	}

	mountPath := ""
	// lookup predefined mount path
	for _, e := range c.Env {
		if e.Name == "SERVICE_BINDINGS_ROOT" {
			mountPath = e.Value
			break
		}
	}
	if mountPath == "" {
		// default mount path
		mountPath = "/bindings"
		c.Env = append(c.Env, corev1.EnvVar{
			Name:  "SERVICE_BINDINGS_ROOT",
			Value: mountPath,
		})
	}

	containerVolumes := sets.NewString()
	for _, vm := range c.VolumeMounts {
		containerVolumes.Insert(vm.Name)
	}

	if !containerVolumes.Has(bindingVolume) {
		// inject metadata
		c.VolumeMounts = append(c.VolumeMounts, corev1.VolumeMount{
			Name:      bindingVolume,
			MountPath: fmt.Sprintf("%s/%s", mountPath, b.Name),
			ReadOnly:  true,
		})
	}
}

func (b *ServiceBinding) Undo(ctx context.Context, ps *v1.WithPod) {
	if ps.Annotations == nil {
		ps.Annotations = map[string]string{}
	}

	boundVolumes := sets.NewString(strings.Split(ps.Annotations[b.annotationKey()], ",")...)

	preservedVolumes := []corev1.Volume{}
	for _, v := range ps.Spec.Template.Spec.Volumes {
		if !boundVolumes.Has(v.Name) {
			preservedVolumes = append(preservedVolumes, v)
		}
	}
	ps.Spec.Template.Spec.Volumes = preservedVolumes

	for i := range ps.Spec.Template.Spec.InitContainers {
		b.UndoContainer(ctx, ps, &ps.Spec.Template.Spec.InitContainers[i], &boundVolumes)
	}
	for i := range ps.Spec.Template.Spec.Containers {
		b.UndoContainer(ctx, ps, &ps.Spec.Template.Spec.Containers[i], &boundVolumes)
	}

	delete(ps.Annotations, b.annotationKey())
}

func (b *ServiceBinding) UndoContainer(ctx context.Context, ps *v1.WithPod, c *corev1.Container, boundVolumes *sets.String) {
	preservedMounts := []corev1.VolumeMount{}
	for _, vm := range c.VolumeMounts {
		if !boundVolumes.Has(vm.Name) {
			preservedMounts = append(preservedMounts, vm)
		}
	}
	c.VolumeMounts = preservedMounts
}

func (b *ServiceBinding) annotationKey() string {
	return fmt.Sprintf("%s-%s", ServiceBindingAnnotationKey, b.Name)
}

func (bs *ServiceBindingStatus) InitializeConditions() {
	serviceCondSet.Manage(bs).InitializeConditions()
}

func (bs *ServiceBindingStatus) MarkBindingAvailable() {
	serviceCondSet.Manage(bs).MarkTrue(ServiceBindingConditionBindingAvailable)
}

func (bs *ServiceBindingStatus) MarkBindingUnavailable(reason string, message string) {
	serviceCondSet.Manage(bs).MarkFalse(
		ServiceBindingConditionBindingAvailable, reason, message)
}

func (bs *ServiceBindingStatus) SetObservedGeneration(gen int64) {
	bs.ObservedGeneration = gen
}
