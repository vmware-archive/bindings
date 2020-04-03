package v1alpha1

import (
	"knative.dev/pkg/apis"
)

const (
	BindableServiceConditionReady = apis.ConditionReady
)

var bsCondSet = apis.NewLivingConditionSet()

func (bs *BindableServiceStatus) MarkReady() {
	bsCondSet.Manage(bs).MarkTrue(BindableServiceConditionReady)
}

func (bs *BindableServiceStatus) InitializeConditions() {
	bsCondSet.Manage(bs).InitializeConditions()
}
