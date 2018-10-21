package validation_test

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/jelmersnoeck/barbossa/apis/barbossa/v1alpha1"
	"github.com/jelmersnoeck/barbossa/apis/barbossa/v1alpha1/validation"

	"k8s.io/api/core/v1"
	"k8s.io/api/extensions/v1beta1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/apimachinery/pkg/util/validation/field"
)

func TestDeploymentValidation(t *testing.T) {
	t.Run("Replicas", func(t *testing.T) {
		hap := v1alpha1.HighAvailabilityPolicy{
			Spec: v1alpha1.HighAvailabilityPolicySpec{
				Replicas: &v1alpha1.HighAvailabilityPolicyReplicas{
					Minimum: 2,
				},
			},
		}

		tcs := map[string]testCase{
			"with a valid spec": {
				dpl: v1beta1.DeploymentSpec{
					Replicas: ptrInt32(3),
				},
			},
			"with an invalid replica count": {
				dpl: v1beta1.DeploymentSpec{
					Replicas: ptrInt32(1),
				},
				errs: []*field.Error{
					field.Invalid(field.NewPath("spec").Child("replicas"), ptrInt32(1), fmt.Sprintf("should be at least %d", hap.Spec.Replicas.Minimum)),
				},
			},
			"with no replica count set": {
				dpl: v1beta1.DeploymentSpec{},
				errs: []*field.Error{
					field.Invalid(field.NewPath("spec").Child("replicas"), nil, "is required"),
				},
			},
		}

		runTests(t, hap, tcs)
	})

	t.Run("UpdateStrategy", func(t *testing.T) {
		hap := v1alpha1.HighAvailabilityPolicy{
			Spec: v1alpha1.HighAvailabilityPolicySpec{
				Strategy: &v1alpha1.HighAvailabilityPolicyStrategy{
					Type: v1beta1.RollingUpdateDeploymentStrategyType,
					RollingUpdate: &v1alpha1.HighAvailabilityPolicyRollingUpdate{
						MinSurge:       fromIntStr("25%"),
						MaxSurge:       fromIntStr("75%"),
						MaxUnavailable: fromIntStr("0"),
					},
				},
			},
		}

		tcs := map[string]testCase{
			"with a valid spec": {
				dpl: v1beta1.DeploymentSpec{
					Replicas: ptrInt32(3),
					Strategy: v1beta1.DeploymentStrategy{
						Type: v1beta1.RollingUpdateDeploymentStrategyType,
						RollingUpdate: &v1beta1.RollingUpdateDeployment{
							MaxSurge:       fromIntStr("50%"),
							MaxUnavailable: fromIntStr("0"),
						},
					},
				},
			},
			"without a strategy defined": {
				dpl: v1beta1.DeploymentSpec{
					Replicas: ptrInt32(3),
				},
				errs: []*field.Error{
					field.Invalid(field.NewPath("spec").Child("strategy").Child("type"), "", fmt.Sprintf("should be '%s'", v1beta1.RollingUpdateDeploymentStrategyType)),
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
							MaxSurge:       fromIntStr("5%"),
							MaxUnavailable: fromIntStr("0"),
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
							MaxSurge:       fromIntStr("95%"),
							MaxUnavailable: fromIntStr("0"),
						},
					},
				},
				errs: []*field.Error{
					field.Invalid(field.NewPath("spec").Child("strategy").Child("rollingUpdate").Child("maxSurge"), "95%", "should be at most 75%"),
				},
			},
			"with different update strategy type": {
				dpl: v1beta1.DeploymentSpec{
					Replicas: ptrInt32(3),
					Strategy: v1beta1.DeploymentStrategy{
						Type: v1beta1.RecreateDeploymentStrategyType,
					},
				},
				errs: []*field.Error{
					field.Invalid(field.NewPath("spec").Child("strategy").Child("type"), string(v1beta1.RecreateDeploymentStrategyType), "should be 'RollingUpdate'"),
				},
			},
			"with a maxUnavailable set too high": {
				dpl: v1beta1.DeploymentSpec{
					Replicas: ptrInt32(3),
					Strategy: v1beta1.DeploymentStrategy{
						Type: v1beta1.RollingUpdateDeploymentStrategyType,
						RollingUpdate: &v1beta1.RollingUpdateDeployment{
							MaxSurge:       fromIntStr("50%"),
							MaxUnavailable: fromIntStr("1"),
						},
					},
				},
				errs: []*field.Error{
					field.Invalid(field.NewPath("spec").Child("strategy").Child("rollingUpdate").Child("maxUnavailable"), "1", "should be at most 0"),
				},
			},
		}

		runTests(t, hap, tcs)
	})

	t.Run("ResourceRequirements", func(t *testing.T) {
		hap := v1alpha1.HighAvailabilityPolicy{
			Spec: v1alpha1.HighAvailabilityPolicySpec{
				Resources: &v1alpha1.HighAvailabilityPolicyResourceRequirements{
					Requests: v1alpha1.ResourceList{
						v1.ResourceCPU: true,
					},
					Limits: v1alpha1.ResourceList{
						v1.ResourceMemory: true,
					},
				},
			},
		}

		cPath := field.NewPath("spec").Child("template").Child("spec").Child("containers")
		tcs := map[string]testCase{
			"with a valid spec": {
				dpl: v1beta1.DeploymentSpec{
					Template: v1.PodTemplateSpec{
						Spec: v1.PodSpec{
							Containers: []v1.Container{
								{
									Resources: v1.ResourceRequirements{
										Requests: v1.ResourceList{
											v1.ResourceCPU: resource.MustParse("5m"),
										},
										Limits: v1.ResourceList{
											v1.ResourceMemory: resource.MustParse("50m"),
										},
									},
								},
							},
						},
					},
				},
			},
			"without requests set": {
				dpl: v1beta1.DeploymentSpec{
					Template: v1.PodTemplateSpec{
						Spec: v1.PodSpec{
							Containers: []v1.Container{
								{
									Name: "test-container",
									Resources: v1.ResourceRequirements{
										Requests: v1.ResourceList{
											v1.ResourceMemory: resource.MustParse("5m"),
										},
										Limits: v1.ResourceList{
											v1.ResourceMemory: resource.MustParse("50m"),
										},
									},
								},
							},
						},
					},
				},
				errs: []*field.Error{
					field.Invalid(cPath.Child("test-container").Child("resources").Child("requests").Child("cpu"), nil, "is required"),
				},
			},
			"without limits set": {
				dpl: v1beta1.DeploymentSpec{
					Template: v1.PodTemplateSpec{
						Spec: v1.PodSpec{
							Containers: []v1.Container{
								{
									Name: "test-container",
									Resources: v1.ResourceRequirements{
										Requests: v1.ResourceList{
											v1.ResourceCPU: resource.MustParse("5m"),
										},
										Limits: v1.ResourceList{
											v1.ResourceCPU: resource.MustParse("50m"),
										},
									},
								},
							},
						},
					},
				},
				errs: []*field.Error{
					field.Invalid(cPath.Child("test-container").Child("resources").Child("limits").Child("memory"), nil, "is required"),
				},
			},
		}

		runTests(t, hap, tcs)
	})
}

func runTests(t *testing.T, hap v1alpha1.HighAvailabilityPolicy, tcs testCases) {
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

type testCases map[string]testCase

type testCase struct {
	dpl  v1beta1.DeploymentSpec
	errs []*field.Error
}

func ptrInt32(i int32) *int32 {
	return &i
}

func fromIntStr(v string) *intstr.IntOrString {
	sv := intstr.FromString(v)
	return &sv
}
