package validation_test

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/jelmersnoeck/barbossa/apis/barbossa/v1alpha1"
	"github.com/jelmersnoeck/barbossa/apis/barbossa/v1alpha1/validation"

	"k8s.io/api/extensions/v1beta1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/apimachinery/pkg/util/validation/field"
)

func TestDeploymentValidation(t *testing.T) {
	hap := v1alpha1.HighAvailabilityPolicy{
		Spec: v1alpha1.HighAvailabilityPolicySpec{
			Replicas: &v1alpha1.HighAvailabilityPolicyReplicas{
				Minimum: 2,
			},
			Strategy: &v1alpha1.HighAvailabilityPolicyStrategy{
				Type: v1beta1.RollingUpdateDeploymentStrategyType,
				RollingUpdate: &v1alpha1.HighAvailabilityPolicyRollingUpdate{
					MinSurge: intstr.FromString("25%"),
					MaxSurge: intstr.FromString("75%"),
				},
			},
		},
	}

	tcs := map[string]struct {
		dpl  v1beta1.DeploymentSpec
		errs []*field.Error
	}{
		"with a valid spec": {
			dpl: v1beta1.DeploymentSpec{
				Replicas: ptrInt32(3),
				Strategy: v1beta1.DeploymentStrategy{
					Type: v1beta1.RollingUpdateDeploymentStrategyType,
					RollingUpdate: &v1beta1.RollingUpdateDeployment{
						MaxSurge: fromIntStr("50%"),
					},
				},
			},
		},
		"with an invalid replica count": {
			dpl: v1beta1.DeploymentSpec{
				Replicas: ptrInt32(1),
				Strategy: v1beta1.DeploymentStrategy{
					Type: v1beta1.RollingUpdateDeploymentStrategyType,
					RollingUpdate: &v1beta1.RollingUpdateDeployment{
						MaxSurge: fromIntStr("50%"),
					},
				},
			},
			errs: []*field.Error{
				field.Invalid(field.NewPath("spec").Child("replicas"), ptrInt32(1), fmt.Sprintf("should be at least %d", hap.Spec.Replicas.Minimum)),
			},
		},
		"with no replica count set": {
			dpl: v1beta1.DeploymentSpec{
				Strategy: v1beta1.DeploymentStrategy{
					Type: v1beta1.RollingUpdateDeploymentStrategyType,
					RollingUpdate: &v1beta1.RollingUpdateDeployment{
						MaxSurge: fromIntStr("50%"),
					},
				},
			},
			errs: []*field.Error{
				field.Invalid(field.NewPath("spec").Child("replicas"), nil, "is required"),
			},
		},
		"without a rolling update configuration": {
			dpl: v1beta1.DeploymentSpec{
				Replicas: ptrInt32(3),
				Strategy: v1beta1.DeploymentStrategy{
					Type: v1beta1.RollingUpdateDeploymentStrategyType,
				},
			},
			errs: []*field.Error{
				field.Invalid(field.NewPath("spec").Child("strategy").Child("rollingUpdate"), nil, "is required"),
			},
		},
		"with a MaxSurge too low": {
			dpl: v1beta1.DeploymentSpec{
				Replicas: ptrInt32(10),
				Strategy: v1beta1.DeploymentStrategy{
					Type: v1beta1.RollingUpdateDeploymentStrategyType,
					RollingUpdate: &v1beta1.RollingUpdateDeployment{
						MaxSurge: fromIntStr("5%"),
					},
				},
			},
			errs: []*field.Error{
				field.Invalid(field.NewPath("spec").Child("strategy").Child("rollingUpdate").Child("maxSurge"), "5%", "should be at least 25%"),
			},
		},
		"with a MaxSurge too high": {
			dpl: v1beta1.DeploymentSpec{
				Replicas: ptrInt32(10),
				Strategy: v1beta1.DeploymentStrategy{
					Type: v1beta1.RollingUpdateDeploymentStrategyType,
					RollingUpdate: &v1beta1.RollingUpdateDeployment{
						MaxSurge: fromIntStr("95%"),
					},
				},
			},
			errs: []*field.Error{
				field.Invalid(field.NewPath("spec").Child("strategy").Child("rollingUpdate").Child("maxSurge"), "95%", "should be at most 75%"),
			},
		},
	}

	for n, tc := range tcs {
		t.Run(n, func(t *testing.T) {
			errs := validation.ValidateDeployment(v1beta1.Deployment{Spec: tc.dpl}, hap)

			if len(errs) != len(tc.errs) {
				t.Errorf("Expected '%d' errors, got '%d'", len(tc.errs), len(errs))
				return
			}

			for i, e := range errs {
				expectedErr := tc.errs[i]
				if !reflect.DeepEqual(e, expectedErr) {
					t.Errorf("Expected\n%v\nbut got \n%v", expectedErr, e)
				}
			}
		})
	}
}

func ptrInt32(i int32) *int32 {
	return &i
}

func fromIntStr(v string) *intstr.IntOrString {
	sv := intstr.FromString(v)
	return &sv
}
