package v1alpha1

import (
	"knative.dev/pkg/apis"
)

const (
	ProvisionedServiceConditionReady = apis.ConditionReady
)

var psCondSet = apis.NewLivingConditionSet()

func (ps *ProvisionedServiceStatus) MarkReady() {
	psCondSet.Manage(ps).MarkTrue(ProvisionedServiceConditionReady)
}

func (ps *ProvisionedServiceStatus) InitializeConditions() {
	psCondSet.Manage(ps).InitializeConditions()
}
