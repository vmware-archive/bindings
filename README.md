# Bindings

## ImageBinding

The ImageBinding resource binds an Imageable resource to a resources that embeds a PodSpec.

```
apiVersion: bindings.projectriff.io/v1alpha1
kind: ImageBinding
metadata:
  name: test-binding
spec:
  subject:
    apiVersion: ""
    kind: ""
    name: ""
    namespace: ""

  providers:
    - containerName: ""
      imageableRef:
        apiVersion: ""
        kind: ""
        name: ""
        namespace: ""
```

The ImageBinding resource must be configured in the following way:
* `subject`: a reference to a `PodSpecable` resource
    * `apiVersion`: resource API Version
    * `kind`: resource kind
    * `name`: resource name
    * `namespace`: resource namespace (defaults to namespace of the `ImageBinding`)
* `providers`
    * `containerName`: name of container within the subject's pod spec to update with image from `imageableRef`
    * `imageableRef`: a reference to an `Imageable` resource
        * `apiVersion`: resource API Version
        * `kind`: resource kind
        * `name`: resource name
        * `namespace`: resource namespace (defaults to namespace of the `ImageBinding`)


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