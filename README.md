# Bindings

## ServiceBinding

The ServiceBinding resource projects a provisioned service into an application. This is an implementation of the draft [Service Binding Specification for Kubernetes](https://github.com/application-stacks/service-binding-specification/tree/3d5eff32392f757de82d7719ea2869860ed803fa). This implementation will continue to evolve alongside the spec.

```yaml
apiVersion: service.binding/v1alpha1
kind: ServiceBinding
metadata:
  name:                 # string
spec:
  name:                 # string, optional, default: .metadata.name
  type:                 # string, optional
  provider:             # string, optional

  application:          # PodSpec-able resource ObjectReference-able
    apiVersion:         # string
    kind:               # string
    name:               # string
    containers:         # []intstr.IntOrString, optional
    ...

  service:              # Provisioned Service-able resource ObjectReference-able
    apiVersion:         # string
    kind:               # string
    name:               # string
    ...

status:
  conditions:           # []Condition containing at least one entry for `Ready`
  - type:               # string
    status:             # string
    lastTransitionTime: # Time
    reason:             # string
    message:            # string
```

A `ProvisionedService` resource is included which implements the Provisioned Service duck-type. This resource can expose an existing secret as a provisioned service.

```yaml
apiVersion: bindings.projectriff.io/v1alpha1
kind: ProvisionedService
metadata:
  name: account-db
spec:
  binding:
    name: account-db-service
#status: # status is created by the ProvisionedService controller, and is what implements the duck-type
#  binding:
#    name: account-db-service
```

## ImageBinding

The ImageBinding resource binds an Imageable resource to a resources that embeds a PodSpec.

```yaml
apiVersion: bindings.projectriff.io/v1alpha1
kind: ImageBinding
metadata:
  name: test-binding
spec:
  containerName: ""
  provider:
    apiVersion: ""
    kind: ""
    name: ""
  subject:
    apiVersion: ""
    kind: ""
    name: ""
```

The ImageBinding resource must be configured in the following way:
* `containerName`: name of container within the subject's pod spec to update with image from the provider
* `provider`: a reference to an `Imageable` resource
    * `apiVersion`: resource API Version
    * `kind`: resource kind
    * `name`: resource name
* `subject`: a reference to a `PodSpecable` resource
    * `apiVersion`: resource API Version
    * `kind`: resource kind
    * `name`: resource name


## Imageable
An Imageable resource is any resource that provides an image by implementing the following partial schema
```
status:
  latestImage: <image-reference>
```

Examples of Imageable resources
* kpack Image resource
* riff Application resource
* riff Function resource
* riff Container resource

## Install

The CRDs and controller can be installed into a cluster using [`ko`](https://github.com/google/ko)

```sh
ko apply --strict -f config
```
