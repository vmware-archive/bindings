package main

import (
	"github.com/projectriff/bindings/pkg/reconciler/imagebinding"
	"github.com/projectriff/bindings/pkg/reconciler/servicebinding"
	"knative.dev/pkg/injection/sharedmain"
)

func main() {
	sharedmain.Main("controller",
		imagebinding.NewController,
		servicebinding.NewController,
	)
}
