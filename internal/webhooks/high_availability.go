package webhooks

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/jelmersnoeck/barbossa/apis/barbossa/v1alpha1"
	"github.com/jelmersnoeck/barbossa/apis/barbossa/v1alpha1/validation"
	"github.com/jelmersnoeck/barbossa/pkg/client/generated/clientset/versioned"

	"k8s.io/api/admission/v1beta1"
	ev1beta1 "k8s.io/api/extensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/rest"
)

type HighAvailabilityAdmissionHook struct {
	crdClient versioned.Interface
}

func (h *HighAvailabilityAdmissionHook) Initialize(cfg *rest.Config, stopCh <-chan struct{}) error {
	crdClient, err := versioned.NewForConfig(cfg)
	if err != nil {
		return err
	}

	h.crdClient = crdClient
	return nil
}

func (h *HighAvailabilityAdmissionHook) ValidatingResource() (plural schema.GroupVersionResource, singular string) {
	gv := v1alpha1.SchemeGroupVersion
	gv.Group = "admission." + gv.Group
	return gv.WithResource("highavailabilitypolicies"), "highavailabilitypolicy"
}

func (h *HighAvailabilityAdmissionHook) Validate(ar *v1beta1.AdmissionRequest) *v1beta1.AdmissionResponse {
	var dpl ev1beta1.Deployment
	if err := json.Unmarshal(ar.Object.Raw, &dpl); err != nil {
		return &v1beta1.AdmissionResponse{
			Allowed: false,
			Result: &metav1.Status{
				Status:  metav1.StatusFailure,
				Code:    http.StatusBadRequest,
				Reason:  metav1.StatusReasonBadRequest,
				Message: err.Error(),
			},
		}
	}

	hapList, err := h.crdClient.Barbossa().
		HighAvailabilityPolicies(dpl.Namespace).List(metav1.ListOptions{})
	if err != nil {
		return &v1beta1.AdmissionResponse{
			Allowed: false,
			Result: &metav1.Status{
				Status:  metav1.StatusFailure,
				Code:    http.StatusInternalServerError,
				Reason:  metav1.StatusReasonInternalError,
				Message: err.Error(),
			},
		}
	}

	var hap *v1alpha1.HighAvailabilityPolicy

	// go over all items and see if the selector matches this deployment
	for _, dhap := range hapList.Items {
		// the currently selected hap has a higher weight than this one, don't
		// bother checking anything.
		if hap != nil {
			if hap.Spec.Weight > dhap.Spec.Weight {
				continue
			}
		}

		shap := &dhap
		lblSelector, err := metav1.LabelSelectorAsSelector(shap.Spec.Selector)
		if err != nil {
			log.Printf("Could not get label selector for %s:%s: %s", shap.Namespace, shap.Name, err)

			return &v1beta1.AdmissionResponse{
				Allowed: false,
				Result: &metav1.Status{
					Status:  metav1.StatusFailure,
					Code:    http.StatusInternalServerError,
					Reason:  metav1.StatusReasonInternalError,
					Message: err.Error(),
				},
			}
		}

		// the labels match and we know this hap has a higher weight than the
		// currently selected one, mark this one to be used!
		if lblSelector.Matches(labels.Set(dpl.Labels)) {
			hap = shap
		}
	}

	// no hap which selects this resource, ignore it!
	if hap == nil {
		return &v1beta1.AdmissionResponse{
			Allowed: true,
		}
	}

	log.Printf("Validating %s:%s", dpl.Namespace, dpl.Name)
	if err := validation.ValidateDeployment(dpl, *hap).ToAggregate(); err != nil {
		return &v1beta1.AdmissionResponse{
			Allowed: false,
			Result: &metav1.Status{
				Status:  metav1.StatusFailure,
				Code:    http.StatusNotAcceptable,
				Reason:  metav1.StatusReasonNotAcceptable,
				Message: err.Error(),
			},
		}
	}

	return &v1beta1.AdmissionResponse{
		Allowed: true,
	}
}
