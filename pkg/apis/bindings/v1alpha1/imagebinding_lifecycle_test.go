package v1alpha1

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	v1 "knative.dev/pkg/apis/duck/v1"
	"knative.dev/pkg/tracker"
)

func TestImageBindingDo(t *testing.T) {
	for _, tc := range []struct {
		testName      string
		bindingName   string
		latestImage   string
		containerName string
		provider      *tracker.Reference
		providerGR    *schema.GroupResource
		ps            *v1.WithPod
		expectedPS    *v1.WithPod
	}{
		{
			testName:      "updates container image",
			latestImage:   "some/image",
			containerName: "some-container",
			provider: &tracker.Reference{
				APIVersion: "some.group/v1",
				Kind:       "Imageable",
				Name:       "some-imageable",
			},
			bindingName: "some-binding",
			ps:          withContainer("some-container", "old/image"),
			expectedPS: withAnnotation(
				withContainer("some-container", "some/image"),
				"imagebindings.bindings.projectriff.io/some-binding",
				`{"container":"some-container","provider":"imageables.some.group/some-imageable"}`,
			),
		},
		{
			testName:      "updates init container image",
			latestImage:   "some/image",
			containerName: "some-container",
			provider: &tracker.Reference{
				APIVersion: "some.group/v1",
				Kind:       "Imageable",
				Name:       "some-imageable",
			},
			bindingName: "some-binding",
			ps:          withInitContainer("some-container", "old/image"),
			expectedPS: withAnnotation(
				withInitContainer("some-container", "some/image"),
				"imagebindings.bindings.projectriff.io/some-binding",
				`{"container":"some-container","provider":"imageables.some.group/some-imageable"}`,
			),
		},
		{
			testName:      "no such container",
			latestImage:   "some/image",
			containerName: "other-container",
			provider: &tracker.Reference{
				APIVersion: "some.group/v1",
				Kind:       "Imageable",
				Name:       "some-imageable",
			},
			bindingName: "some-binding",
			ps:          withContainer("some-container", "old/image"),
			expectedPS:  withContainer("some-container", "old/image"),
		},
		{
			testName:      "no latest image",
			containerName: "some-container",
			provider: &tracker.Reference{
				APIVersion: "some.group/v1",
				Kind:       "Imageable",
				Name:       "some-imageable",
			},
			bindingName: "some-binding",
			ps:          withContainer("some-container", "old/image"),
			expectedPS: withContainer("some-container", "old/image"),
		},
		{
			testName:      "container already bound",
			latestImage:   "some/image",
			containerName: "some-container",
			provider: &tracker.Reference{
				APIVersion: "some.group/v1",
				Kind:       "Imageable",
				Name:       "some-imageable",
			},
			bindingName: "some-binding",
			ps: withAnnotation(
				withContainer("some-container", "other/image"),
				"imagebindings.bindings.projectriff.io/other-binding",
				`{"container":"some-container"}`,
			),
			expectedPS: withAnnotation(
				withContainer("some-container", "other/image"),
				"imagebindings.bindings.projectriff.io/other-binding",
				`{"container":"some-container"}`,
			),
		},
	} {
		b := &ImageBinding{
			ObjectMeta: metav1.ObjectMeta{
				Name: tc.bindingName,
			},
			Spec: ImageBindingSpec{
				ContainerName: tc.containerName,
				Provider:      tc.provider,
			},
		}
		ctx := context.Background()
		ctx = WithLatestImage(ctx, tc.latestImage)
		b.Do(ctx, tc.ps)
		if diff := cmp.Diff(tc.expectedPS, tc.ps); diff != "" {
			t.Errorf("Do(%s) (-expected, +actual) = %v", tc.testName, diff)
		}
	}
}

func TestImageBindingUndo(t *testing.T) {
	for _, tc := range []struct {
		name          string
		containerName string
		bindingName   string
		ps            *v1.WithPod
		expectedPS    *v1.WithPod
	}{
		{
			name:          "removes annotation",
			bindingName:   "some-binding",
			containerName: "some-container",
			ps: withAnnotation(
				withContainer("some-container", "some/image"),
				"imagebindings.bindings.projectriff.io/some-binding",
				"foo",
			),
			expectedPS: withContainer("some-container", "some/image"),
		},
	} {
		b := &ImageBinding{
			ObjectMeta: metav1.ObjectMeta{
				Name: "some-binding",
			},
			Spec: ImageBindingSpec{
				ContainerName: tc.containerName,
			},
		}
		ctx := context.Background()
		b.Undo(ctx, tc.ps)
		if diff := cmp.Diff(tc.expectedPS, tc.ps); diff != "" {
			t.Errorf("Do(%s) (-expected, +actual) = %v", tc.name, diff)
		}
	}
}

func withContainer(container, image string) *v1.WithPod {
	return &v1.WithPod{
		ObjectMeta: metav1.ObjectMeta{
			Annotations: map[string]string{},
		},
		Spec: v1.WithPodSpec{
			Template: v1.PodSpecable{
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  container,
							Image: image,
						},
					},
				},
			},
		},
	}
}

func withInitContainer(container, image string) *v1.WithPod {
	return &v1.WithPod{
		ObjectMeta: metav1.ObjectMeta{
			Annotations: map[string]string{},
		},
		Spec: v1.WithPodSpec{
			Template: v1.PodSpecable{
				Spec: corev1.PodSpec{
					InitContainers: []corev1.Container{
						{
							Name:  container,
							Image: image,
						},
					},
				},
			},
		},
	}
}

func withAnnotation(ps *v1.WithPod, k, v string) *v1.WithPod {
	ps.Annotations[k] = v
	return ps
}
