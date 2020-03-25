# Bindings

## ImageBinding

The ImageBinding resource binds an Imageable resource to a resources that embeds a PodSpec.

```
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
