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
	el = validateResourceRequirements(el, dpl, hap)

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
		return append(el, field.Invalid(path.Child("type"), string(dplStrategy.Type), fmt.Sprintf("should be '%s'", hapStrategy.Type)))
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

		el = validateMaxSurge(el, upPath, reps, dplStrategy, hapStrategy)
		el = validateMaxUnavailable(el, upPath, reps, dplStrategy, hapStrategy)
	}

	return el
}

// validate the maxSurge value. It needs to be hap.minSurge <= dpl.maxSurge <= hap.maxSurge
func validateMaxSurge(el field.ErrorList, upPath *field.Path, reps int, dplStrategy v1beta1.DeploymentStrategy, hapStrategy *v1alpha1.HighAvailabilityPolicyStrategy) field.ErrorList {
	dplVal, err := intstr.GetValueFromIntOrPercent(dplStrategy.RollingUpdate.MaxSurge, reps, true)
	if err != nil {
		return append(el, field.Invalid(upPath.Child("maxSurge"), dplStrategy.RollingUpdate.MaxSurge.String(), err.Error()))
	}

	if hapStrategy.RollingUpdate.MinSurge != nil {
		hapMinVal, err := intstr.GetValueFromIntOrPercent(hapStrategy.RollingUpdate.MinSurge, reps, true)
		if err != nil {
			return append(el, field.Invalid(upPath.Child("minSurge"), hapStrategy.RollingUpdate.MinSurge.String(), err.Error()))
		}

		if dplVal < hapMinVal {
			val := hapStrategy.RollingUpdate.MinSurge
			return append(el, field.Invalid(upPath.Child("maxSurge"), dplStrategy.RollingUpdate.MaxSurge.String(), fmt.Sprintf("should be at least %s", val.String())))
		}
	}

	if hapStrategy.RollingUpdate.MaxSurge != nil {
		hapMaxVal, err := intstr.GetValueFromIntOrPercent(hapStrategy.RollingUpdate.MaxSurge, reps, true)
		if err != nil {
			return append(el, field.Invalid(upPath.Child("maxSurge"), hapStrategy.RollingUpdate.MaxSurge.String(), err.Error()))
		}

		if dplVal > hapMaxVal {
			return append(el, field.Invalid(upPath.Child("maxSurge"), dplStrategy.RollingUpdate.MaxSurge.String(), fmt.Sprintf("should be at most %s", hapStrategy.RollingUpdate.MaxSurge.String())))
		}
	}

	return el
}

// time to validate the maxUnavailable. It needs to be dpl.maxUnavailable <= hap.maxUnavailable
func validateMaxUnavailable(el field.ErrorList, upPath *field.Path, reps int, dplStrategy v1beta1.DeploymentStrategy, hapStrategy *v1alpha1.HighAvailabilityPolicyStrategy) field.ErrorList {
	if hapStrategy.RollingUpdate.MaxUnavailable == nil {
		return el
	}

	dplVal, err := intstr.GetValueFromIntOrPercent(dplStrategy.RollingUpdate.MaxUnavailable, reps, true)
	if err != nil {
		return append(el, field.Invalid(upPath.Child("maxUnavailable"), dplStrategy.RollingUpdate.MaxUnavailable.String(), err.Error()))
	}

	hapVal, err := intstr.GetValueFromIntOrPercent(hapStrategy.RollingUpdate.MaxUnavailable, reps, true)
	if err != nil {
		return append(el, field.Invalid(upPath.Child("maxUnavailable"), hapStrategy.RollingUpdate.MaxUnavailable.String(), err.Error()))
	}

	if dplVal > hapVal {
		return append(el, field.Invalid(upPath.Child("maxUnavailable"), dplStrategy.RollingUpdate.MaxUnavailable.String(), fmt.Sprintf("should be at most %s", hapStrategy.RollingUpdate.MaxUnavailable.String())))
	}
	return el
}

func validateResourceRequirements(el field.ErrorList, dpl v1beta1.Deployment, hap v1alpha1.HighAvailabilityPolicy) field.ErrorList {
	resources := hap.Spec.Resources
	if resources == nil {
		return el
	}

	cPath := specPath.Child("template").Child("spec").Child("containers")
	for _, container := range dpl.Spec.Template.Spec.Containers {
		path := cPath.Child(container.Name).Child("resources")

		rPath := path.Child("requests")
		for name, rc := range resources.Requests {
			if !rc {
				continue
			}

			if _, ok := container.Resources.Requests[name]; !ok {
				el = append(el, field.Invalid(rPath.Child(string(name)), nil, "is required"))
			}
		}

		lPath := path.Child("limits")
		for name, rc := range resources.Limits {
			if !rc {
				continue
			}

			if _, ok := container.Resources.Limits[name]; !ok {
				el = append(el, field.Invalid(lPath.Child(string(name)), nil, "is required"))
			}
		}
	}

	return el
}
