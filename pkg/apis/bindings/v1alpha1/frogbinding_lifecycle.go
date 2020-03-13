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
	FrogBindingConditionReady            = apis.ConditionReady
	FrogBindingConditionBindingAvailable = "BindingAvailable"
)

var frogCondSet = apis.NewLivingConditionSet(
	FrogBindingConditionBindingAvailable,
)

func (b *FrogBinding) GetSubject() tracker.Reference {
	return *b.Spec.Subject
}

func (b *FrogBinding) GetBindingStatus() duck.BindableStatus {
	return &b.Status
}

func (b *FrogBinding) Do(ctx context.Context, ps *v1.WithPod) {
	// undo existing bindings so we can start clean
	b.Undo(ctx, ps)

	existingVolumes := sets.NewString()
	for _, v := range ps.Spec.Template.Spec.Volumes {
		existingVolumes.Insert(v.Name)
	}

	newVolumes := sets.NewString()
	for _, p := range b.Spec.Providers {
		// TODO ensure unique volume names
		// TODO limit volume name length
		metadataVolume := fmt.Sprintf("%s-metadata", p.Ref.Metadata.Name)
		secretVolume := fmt.Sprintf("%s-secret", p.Ref.Secret.Name)
		if !existingVolumes.Has(metadataVolume) {
			ps.Spec.Template.Spec.Volumes = append(ps.Spec.Template.Spec.Volumes, corev1.Volume{
				Name: metadataVolume,
				VolumeSource: corev1.VolumeSource{
					ConfigMap: &corev1.ConfigMapVolumeSource{
						LocalObjectReference: p.Ref.Metadata,
					},
				},
			})
			existingVolumes.Insert(metadataVolume)
			newVolumes.Insert(metadataVolume)
		}
		if p.BindingMode == SecretFrogBinding && !existingVolumes.Has(secretVolume) {
			ps.Spec.Template.Spec.Volumes = append(ps.Spec.Template.Spec.Volumes, corev1.Volume{
				Name: secretVolume,
				VolumeSource: corev1.VolumeSource{
					Secret: &corev1.SecretVolumeSource{
						SecretName: p.Ref.Secret.Name,
					},
				},
			})
			existingVolumes.Insert(secretVolume)
			newVolumes.Insert(secretVolume)
		}
		for i := range ps.Spec.Template.Spec.InitContainers {
			b.DoContainer(ctx, ps, &ps.Spec.Template.Spec.InitContainers[i], metadataVolume, secretVolume, p)
		}
		for i := range ps.Spec.Template.Spec.Containers {
			b.DoContainer(ctx, ps, &ps.Spec.Template.Spec.Containers[i], metadataVolume, secretVolume, p)
		}
	}

	// track which volumes are injected, so they can be removed when no longer used
	ps.Annotations[FrogBindingAnnotationKey] = strings.Join(newVolumes.List(), ",")
}

func (b *FrogBinding) DoContainer(ctx context.Context, ps *v1.WithPod, c *corev1.Container, metadataVolume, secretVolume string, p FrogProvider) {
	if c.Name != p.ContainerName && p.ContainerName != "" {
		// ignore the container
		return
	}

	mountPath := ""
	// lookup predefined mount path
	for _, e := range c.Env {
		if e.Name == "CNB_BINDINGS" {
			mountPath = e.Value
			break
		}
	}
	if mountPath == "" {
		// default mount path
		// TODO is there a better default path?
		mountPath = "/var/bindings"
		c.Env = append(c.Env, corev1.EnvVar{
			Name:  "CNB_BINDINGS",
			Value: mountPath,
		})
	}

	containerVolumes := sets.NewString()
	for _, vm := range c.VolumeMounts {
		containerVolumes.Insert(vm.Name)
	}

	if !containerVolumes.Has(metadataVolume) {
		// inject metadata
		c.VolumeMounts = append(c.VolumeMounts, corev1.VolumeMount{
			Name:      metadataVolume,
			MountPath: fmt.Sprintf("%s/%s/metadata", mountPath, p.Name),
			ReadOnly:  true,
		})
	}
	if !containerVolumes.Has(secretVolume) && p.BindingMode == SecretFrogBinding {
		// inject secret
		c.VolumeMounts = append(c.VolumeMounts, corev1.VolumeMount{
			Name:      secretVolume,
			MountPath: fmt.Sprintf("%s/%s/secret", mountPath, p.Name),
			ReadOnly:  true,
		})
	}
}


func (b *FrogBinding) Undo(ctx context.Context, ps *v1.WithPod) {
	if ps.Annotations == nil {
		ps.Annotations = map[string]string{}
	}

	boundVolumes := sets.NewString(strings.Split(ps.Annotations[FrogBindingAnnotationKey], ",")...)

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

	delete(ps.Annotations, FrogBindingAnnotationKey)
}

func (b *FrogBinding) UndoContainer(ctx context.Context, ps *v1.WithPod, c *corev1.Container, boundVolumes *sets.String) {
	preservedMounts := []corev1.VolumeMount{}
	for _, vm := range c.VolumeMounts {
		if !boundVolumes.Has(vm.Name) {
			preservedMounts = append(preservedMounts, vm)
		}
	}
	c.VolumeMounts = preservedMounts
}

func (bs *FrogBindingStatus) InitializeConditions() {
	frogCondSet.Manage(bs).InitializeConditions()
}

func (bs *FrogBindingStatus) MarkBindingAvailable() {
	frogCondSet.Manage(bs).MarkTrue(FrogBindingConditionBindingAvailable)
}

func (bs *FrogBindingStatus) MarkBindingUnavailable(reason string, message string) {
	frogCondSet.Manage(bs).MarkFalse(
		FrogBindingConditionBindingAvailable, reason, message)
}

func (bs *FrogBindingStatus) SetObservedGeneration(gen int64) {
	bs.ObservedGeneration = gen
}
