package webhook

import (
	"encoding/json"
	"net/http"

	admissionv1beta1 "k8s.io/api/admission/v1beta1"
	"k8s.io/api/extensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/rest"
)

type DeploymentAdmissionHook struct{}

func (c *DeploymentAdmissionHook) Initialize(kubeClientConfig *rest.Config, stopCh <-chan struct{}) error {
	return nil
}

func (c *DeploymentAdmissionHook) ValidatingResource() (plural schema.GroupVersionResource, singular string) {
	gv := v1beta1.SchemeGroupVersion
	gv.Group = "admission." + gv.Group
	// override version to be the version of the admissionresponse resource
	gv.Version = "v1beta1"
	return gv.WithResource("deployments"), "deployment"
}

func (c *DeploymentAdmissionHook) Validate(admissionSpec *admissionv1beta1.AdmissionRequest) *admissionv1beta1.AdmissionResponse {
	status := &admissionv1beta1.AdmissionResponse{}

	dpl := &v1beta1.Deployment{}
	err := json.Unmarshal(admissionSpec.Object.Raw, dpl)
	if err != nil {
		status.Allowed = false
		status.Result = &metav1.Status{
			Status: metav1.StatusFailure, Code: http.StatusBadRequest, Reason: metav1.StatusReasonBadRequest,
			Message: err.Error(),
		}
		return status
	}

	status.Allowed = true

	return status
}
