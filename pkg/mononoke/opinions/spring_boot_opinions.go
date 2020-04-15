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
	"encoding/json"
	"fmt"
	"math"
	"regexp"
	"strconv"
	"strings"

	"github.com/Masterminds/semver"
	"github.com/projectriff/bindings/pkg/mononoke/cnb"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/apimachinery/pkg/util/sets"
	v1 "knative.dev/pkg/apis/duck/v1"
)

var SpringBoot = Opinions{
	&BasicOpinion{
		Id: "spring-boot",
		ApplicableFunc: func(applied AppliedOpinions, imageMetadata cnb.BuildMetadata) bool {
			bootMetadata := NewSpringBootBOMMetadata(imageMetadata)
			return bootMetadata.HasDependency("spring-boot")
		},
		ApplyFunc: func(ctx context.Context, target *v1.WithPod, containerIdx int, imageMetadata cnb.BuildMetadata) error {
			bootMetadata := NewSpringBootBOMMetadata(imageMetadata)
			setLabel(target, "apps.mononoke.local/framework", "spring-boot")
			for _, d := range bootMetadata.Dependencies {
				if d.Name == "spring-boot" {
					setAnnotation(target, "boot.spring.io/version", d.Version)
					break
				}
			}
			return nil
		},
	},
	&BasicOpinion{
		Id: "spring-boot-graceful-shutdown",
		ApplicableFunc: func(applied AppliedOpinions, imageMetadata cnb.BuildMetadata) bool {
			bootMetadata := NewSpringBootBOMMetadata(imageMetadata)
			return bootMetadata.HasDependencyConstraint("spring-boot", ">= 2.3.0-0") && bootMetadata.HasDependency(
				"spring-boot-starter-tomcat",
				"spring-boot-starter-jetty",
				"spring-boot-starter-reactor-netty",
				"spring-boot-starter-undertow",
			)
		},
		ApplyFunc: func(ctx context.Context, target *v1.WithPod, containerIdx int, imageMetadata cnb.BuildMetadata) error {
			applicationProperties := GetSpringApplicationProperties(ctx)
			if _, ok := applicationProperties["server.shutdown.grace-period"]; ok {
				// boot grace period is already defined, skipping
				return nil
			}
			var k8sGracePeriodSeconds int64 = 30 // default k8s grace period is 30 seconds
			if target.Spec.Template.Spec.TerminationGracePeriodSeconds != nil {
				k8sGracePeriodSeconds = *target.Spec.Template.Spec.TerminationGracePeriodSeconds
			}
			target.Spec.Template.Spec.TerminationGracePeriodSeconds = &k8sGracePeriodSeconds
			// allocate 80% of the k8s grace period to boot
			bootGracePeriodSeconds := int(math.Floor(0.8 * float64(k8sGracePeriodSeconds)))
			applicationProperties["server.shutdown.grace-period"] = fmt.Sprintf("%ds", bootGracePeriodSeconds)
			return nil
		},
	},
	&BasicOpinion{
		Id: "spring-web-port",
		ApplicableFunc: func(applied AppliedOpinions, imageMetadata cnb.BuildMetadata) bool {
			bootMetadata := NewSpringBootBOMMetadata(imageMetadata)
			return bootMetadata.HasDependency("spring-web")
		},
		ApplyFunc: func(ctx context.Context, target *v1.WithPod, containerIdx int, imageMetadata cnb.BuildMetadata) error {
			applicationProperties := GetSpringApplicationProperties(ctx)

			serverPort := applicationProperties.Default("server.port", "8080")
			port, err := strconv.Atoi(serverPort)
			if err != nil {
				return err
			}

			c := &target.Spec.Template.Spec.Containers[containerIdx]

			if name, cp := findContainerPort(target.Spec.Template.Spec, int32(port)); cp == nil {
				c.Ports = append(c.Ports, corev1.ContainerPort{
					ContainerPort: int32(port),
					Protocol:      corev1.ProtocolTCP,
				})
			} else if name != c.Name {
				// port is in use by a different container
				return fmt.Errorf("desired port %s is in use by container %q, set 'server.port' boot property to an open port", serverPort, name)
			}

			return nil
		},
	},
	&BasicOpinion{
		Id: "spring-boot-actuator",
		ApplicableFunc: func(applied AppliedOpinions, imageMetadata cnb.BuildMetadata) bool {
			bootMetadata := NewSpringBootBOMMetadata(imageMetadata)
			return bootMetadata.HasDependency("spring-boot-actuator")
		},
		ApplyFunc: func(ctx context.Context, target *v1.WithPod, containerIdx int, imageMetadata cnb.BuildMetadata) error {
			applicationProperties := GetSpringApplicationProperties(ctx)

			managementPort := applicationProperties.Default("management.server.port", applicationProperties["server.port"])
			managementBasePath := applicationProperties.Default("management.endpoints.web.base-path", "/actuator")
			managementScheme := corev1.URISchemeHTTP
			if applicationProperties["management.server.ssl.enabled"] == "true" {
				managementScheme = corev1.URISchemeHTTPS
			}

			setAnnotation(target, "boot.spring.io/actuator", fmt.Sprintf("%s://:%s%s", strings.ToLower(string(managementScheme)), managementPort, managementBasePath))

			return nil
		},
	},
	&BasicOpinion{
		Id: "spring-boot-actuator-probes",
		ApplicableFunc: func(applied AppliedOpinions, imageMetadata cnb.BuildMetadata) bool {
			bootMetadata := NewSpringBootBOMMetadata(imageMetadata)
			return bootMetadata.HasDependency("spring-boot-actuator")
		},
		ApplyFunc: func(ctx context.Context, target *v1.WithPod, containerIdx int, imageMetadata cnb.BuildMetadata) error {
			bootMetadata := NewSpringBootBOMMetadata(imageMetadata)
			applicationProperties := GetSpringApplicationProperties(ctx)

			if v := applicationProperties.Default("management.health.probes.enabled", "true"); v != "true" {
				// management health probes were disabled by the user, skip
				return nil
			}

			managementBasePath := applicationProperties["management.endpoints.web.base-path"]
			managementPort, err := strconv.Atoi(applicationProperties["management.server.port"])
			if err != nil {
				return err
			}
			managementScheme := corev1.URISchemeHTTP
			if applicationProperties["management.server.ssl.enabled"] == "true" {
				managementScheme = corev1.URISchemeHTTPS
			}

			var livenessEndpoint, readinessEndpoint string
			if bootMetadata.HasDependencyConstraint("spring-boot-actuator", ">= 2.3.0-0") {
				livenessEndpoint = "/health/liveness"
				readinessEndpoint = "/health/readiness"
			} else {
				livenessEndpoint = "/info"
				readinessEndpoint = "/info"
			}

			c := &target.Spec.Template.Spec.Containers[containerIdx]

			// define probes
			if c.StartupProbe == nil {
				// currently alpha in k8s 1.16+
				// TODO(scothis) add if k8s can handle it
			}
			if c.LivenessProbe == nil {
				c.LivenessProbe = &corev1.Probe{
					// increase default to give more time to start
					// TODO(scothis) remove if a StartupProbe is defined
					InitialDelaySeconds: 30,
				}
			}
			if c.LivenessProbe.Handler == (corev1.Handler{}) {
				c.LivenessProbe.Handler = corev1.Handler{
					HTTPGet: &corev1.HTTPGetAction{
						Path:   managementBasePath + livenessEndpoint,
						Port:   intstr.FromInt(managementPort),
						Scheme: managementScheme,
					},
				}
			}
			if c.ReadinessProbe == nil {
				c.ReadinessProbe = &corev1.Probe{}
			}
			if c.ReadinessProbe.Handler == (corev1.Handler{}) {
				c.ReadinessProbe.Handler = corev1.Handler{
					HTTPGet: &corev1.HTTPGetAction{
						Path:   managementBasePath + readinessEndpoint,
						Port:   intstr.FromInt(managementPort),
						Scheme: managementScheme,
					},
				}
			}

			return nil
		},
	},

	// service intents
	&SpringBootServiceIntent{
		Id:        "service-intent-mysql",
		LabelName: "services.mononoke.local/mysql",
		Dependencies: sets.NewString(
			"mysql-connector-java",
			"r2dbc-mysql",
		),
	},
	&SpringBootServiceIntent{
		Id:        "service-intent-postgres",
		LabelName: "services.mononoke.local/postgres",
		Dependencies: sets.NewString(
			"postgresql",
			"r2dbc-postgresql",
		),
	},
	&SpringBootServiceIntent{
		Id:        "service-intent-mongodb",
		LabelName: "services.mononoke.local/mongodb",
		Dependencies: sets.NewString(
			"mongodb-driver-core",
		),
	},
	&SpringBootServiceIntent{
		Id:        "service-intent-rabbitmq",
		LabelName: "services.mononoke.local/rabbitmq",
		Dependencies: sets.NewString(
			"amqp-client",
		),
	},
	&SpringBootServiceIntent{
		Id:        "service-intent-redis",
		LabelName: "services.mononoke.local/redis",
		Dependencies: sets.NewString(
			"jedis",
		),
	},
	&SpringBootServiceIntent{
		Id:        "service-intent-kafka",
		LabelName: "services.mononoke.local/kafka",
		Dependencies: sets.NewString(
			"kafka-clients",
		),
	},
	&SpringBootServiceIntent{
		Id:        "service-intent-kafka-streams",
		LabelName: "services.mononoke.local/kafka-streams",
		Dependencies: sets.NewString(
			"kafka-streams",
		),
	},

	// TODO add a whole lot more opinions
}

func NewSpringBootBOMMetadata(imageMetadata cnb.BuildMetadata) SpringBootBOMMetadata {
	// TODO(scothis) find a better way to convert map[string]interface{} to SpringBootBOMMetadata{}
	bom := imageMetadata.FindBOM("spring-boot")
	bootMetadata := SpringBootBOMMetadata{}
	bytes, err := json.Marshal(bom.Metadata)
	if err != nil {
		panic(err)
	}
	json.Unmarshal(bytes, &bootMetadata)
	return bootMetadata
}

type SpringBootBOMMetadata struct {
	Classes      string                            `json:"classes"`
	ClassPath    []string                          `json:"classpath"`
	Dependencies []SpringBootBOMMetadataDependency `json:"dependencies"`
}

func (m *SpringBootBOMMetadata) Dependency(name string) *SpringBootBOMMetadataDependency {
	for _, d := range m.Dependencies {
		if d.Name == name {
			return &d
		}
	}
	return nil
}

func (m *SpringBootBOMMetadata) HasDependency(names ...string) bool {
	n := sets.NewString(names...)
	for _, d := range m.Dependencies {
		if n.Has(d.Name) {
			return true
		}
	}
	return false
}

func (m *SpringBootBOMMetadata) HasDependencyConstraint(name, constraint string) bool {
	d := m.Dependency(name)
	if d == nil {
		return false
	}
	v, err := semver.NewVersion(m.normalizeVersion(d.Version))
	if err != nil {
		return false
	}
	c, err := semver.NewConstraint(constraint)
	if err != nil {
		return false
	}
	return c.Check(v)
}

func (m *SpringBootBOMMetadata) normalizeVersion(version string) string {
	r := regexp.MustCompile(`^([0-9]+\.[0-9]+\.[0-9]+)\.`)
	return r.ReplaceAllString(version, "$1-")
}

type SpringBootBOMMetadataDependency struct {
	Name    string `json:"name"`
	Sha256  string `json:"sha256"`
	Version string `json:"version"`
}

type springApplicationPropertiesKey struct{}

type SpringApplicationProperties map[string]string

func StashSpringApplicationProperties(ctx context.Context, props SpringApplicationProperties) context.Context {
	return context.WithValue(ctx, springApplicationPropertiesKey{}, props)
}

func GetSpringApplicationProperties(ctx context.Context) SpringApplicationProperties {
	value := ctx.Value(springApplicationPropertiesKey{})
	if props, ok := value.(SpringApplicationProperties); ok {
		return props
	}
	return nil
}

func (props SpringApplicationProperties) Default(key, value string) string {
	if _, ok := props[key]; !ok {
		props[key] = value
	}
	return props[key]
}

type SpringBootServiceIntent struct {
	Id           string
	LabelName    string
	Dependencies sets.String
}

func (o *SpringBootServiceIntent) GetId() string {
	return o.Id
}

func (o *SpringBootServiceIntent) Applicable(applied AppliedOpinions, metadata cnb.BuildMetadata) bool {
	bootMetadata := NewSpringBootBOMMetadata(metadata)
	for _, d := range bootMetadata.Dependencies {
		if o.Dependencies.Has(d.Name) {
			return true
		}
	}
	return false
}

func (o *SpringBootServiceIntent) Apply(ctx context.Context, target *v1.WithPod, containerIdx int, metadata cnb.BuildMetadata) error {
	bootMetadata := NewSpringBootBOMMetadata(metadata)
	for _, d := range bootMetadata.Dependencies {
		if o.Dependencies.Has(d.Name) {
			setLabel(target, o.LabelName, target.Spec.Template.Spec.Containers[containerIdx].Name)
			setAnnotation(target, o.LabelName, fmt.Sprintf("%s/%s", d.Name, d.Version))
			break
		}
	}
	return nil
}
