/*
Copyright 2020 the original author or authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package opinions

import (
	"context"

	"github.com/projectriff/bindings/pkg/mononoke/cnb"
	corev1 "k8s.io/api/core/v1"
	v1 "knative.dev/pkg/apis/duck/v1"
)

type Opinion interface {
	GetId() string
	Applicable(applied AppliedOpinions, imageMetadata cnb.BuildMetadata) bool
	Apply(ctx context.Context, target *v1.WithPod, containerIdx int, metadata cnb.BuildMetadata) error
}

type Opinions []Opinion

func (os Opinions) Apply(ctx context.Context, target *v1.WithPod, containerIdx int, imageMetadata cnb.BuildMetadata) ([]string, error) {
	applied := AppliedOpinions{}
	for _, o := range os {
		if o.Applicable(applied, imageMetadata) {
			applied = append(applied, o.GetId())
			if err := o.Apply(ctx, target, containerIdx, imageMetadata); err != nil {
				return nil, err
			}
		}
	}
	return applied, nil
}

type AppliedOpinions []string

func (os AppliedOpinions) Has(id string) bool {
	for _, o := range os {
		if o == id {
			return true
		}
	}
	return false
}

type BasicOpinion struct {
	Id             string
	ApplicableFunc func(applied AppliedOpinions, metadata cnb.BuildMetadata) bool
	ApplyFunc      func(ctx context.Context, target *v1.WithPod, containerIdx int, metadata cnb.BuildMetadata) error
}

func (o *BasicOpinion) GetId() string {
	return o.Id
}

func (o *BasicOpinion) Applicable(applied AppliedOpinions, metadata cnb.BuildMetadata) bool {
	if o.ApplicableFunc == nil {
		return true
	}
	return o.ApplicableFunc(applied, metadata)
}

func (o *BasicOpinion) Apply(ctx context.Context, target *v1.WithPod, containerIdx int, metadata cnb.BuildMetadata) error {
	return o.ApplyFunc(ctx, target, containerIdx, metadata)
}

func findContainerPort(ps corev1.PodSpec, port int32) (string, *corev1.ContainerPort) {
	for _, c := range ps.Containers {
		for _, p := range c.Ports {
			if p.ContainerPort == port {
				return c.Name, &p
			}
		}
	}
	return "", nil
}
