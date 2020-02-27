package main

import (
	"github.com/projectriff/bindings/pkg/reconciler/imagebinding"
	"knative.dev/pkg/injection/sharedmain"
)

func main() {
	sharedmain.Main("controller", imagebinding.NewController)
}
