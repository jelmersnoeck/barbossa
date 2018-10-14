package validation

import (
	"fmt"

	"github.com/jelmersnoeck/barbossa/apis/barbossa/v1alpha1"

	"k8s.io/api/extensions/v1beta1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/apimachinery/pkg/util/validation/field"
)

var specPath = field.NewPath("spec")

// ValidateDeployment validates the deployment based on a HighAvailabilityPolicy
// and ensures that all fields that are required are set correctly.
func ValidateDeployment(dpl v1beta1.Deployment, hap v1alpha1.HighAvailabilityPolicy) field.ErrorList {
	el := field.ErrorList{}

	el = validateReplicaCount(el, dpl, hap)
	el = validateUpdateStrategy(el, dpl, hap)

	return el
}

func validateReplicaCount(el field.ErrorList, dpl v1beta1.Deployment, hap v1alpha1.HighAvailabilityPolicy) field.ErrorList {
	if hap.Spec.Replicas == nil {
		return el
	}

	if dpl.Spec.Replicas == nil {
		return append(el, field.Invalid(specPath.Child("replicas"), nil, "is required"))
	}

	if *dpl.Spec.Replicas < hap.Spec.Replicas.Minimum {
		return append(el, field.Invalid(specPath.Child("replicas"), dpl.Spec.Replicas, fmt.Sprintf("should be at least %d", hap.Spec.Replicas.Minimum)))
	}

	return el
}

func validateUpdateStrategy(el field.ErrorList, dpl v1beta1.Deployment, hap v1alpha1.HighAvailabilityPolicy) field.ErrorList {
	if hap.Spec.Strategy == nil {
		return el
	}

	path := specPath.Child("strategy")
	dplStrategy := dpl.Spec.Strategy
	hapStrategy := hap.Spec.Strategy

	if dplStrategy.Type != hapStrategy.Type {
		return append(el, field.Invalid(path.Child("type"), dplStrategy.Type, fmt.Sprintf("should be %s", hapStrategy.Type)))
	}

	if dplStrategy.Type == v1beta1.RollingUpdateDeploymentStrategyType {
		upPath := path.Child("rollingUpdate")
		if dplStrategy.RollingUpdate == nil {
			return append(el, field.Invalid(upPath, nil, "is required"))
		}

		// we don't need to add the validation here, the replicas validation
		// takes care of that.
		if dpl.Spec.Replicas == nil {
			return el
		}

		reps := int(*dpl.Spec.Replicas)
		dplVal, err := intstr.GetValueFromIntOrPercent(dplStrategy.RollingUpdate.MaxSurge, reps, true)
		if err != nil {
			return append(el, field.Invalid(upPath.Child("maxSurge"), dplStrategy.RollingUpdate.MaxSurge.String(), err.Error()))
		}

		hapMinVal, err := intstr.GetValueFromIntOrPercent(&hapStrategy.RollingUpdate.MinSurge, reps, true)
		if err != nil {
			return append(el, field.Invalid(upPath.Child("minSurge"), hapStrategy.RollingUpdate.MinSurge.String(), err.Error()))
		}

		hapMaxVal, err := intstr.GetValueFromIntOrPercent(&hapStrategy.RollingUpdate.MaxSurge, reps, true)
		if err != nil {
			return append(el, field.Invalid(upPath.Child("maxSurge"), hapStrategy.RollingUpdate.MaxSurge.String(), err.Error()))
		}

		if dplVal < hapMinVal {
			val := &hapStrategy.RollingUpdate.MinSurge
			fmt.Println(val.String())
			return append(el, field.Invalid(upPath.Child("maxSurge"), dplStrategy.RollingUpdate.MaxSurge.String(), fmt.Sprintf("should be at least %s", val.String())))
		}

		if dplVal > hapMaxVal {
			return append(el, field.Invalid(upPath.Child("maxSurge"), dplStrategy.RollingUpdate.MaxSurge.String(), fmt.Sprintf("should be at most %s", hapStrategy.RollingUpdate.MaxSurge.String())))
		}
	}

	return el
}
