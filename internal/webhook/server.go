package webhook

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/jelmersnoeck/barbossa/apis/barbossa/v1alpha1"
	"github.com/jelmersnoeck/barbossa/apis/barbossa/v1alpha1/validation"
	"github.com/jelmersnoeck/barbossa/pkg/client/generated/clientset/versioned"

	"k8s.io/api/admission/v1beta1"
	ev1beta1 "k8s.io/api/extensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/client-go/kubernetes"
)

var (
	runtimeScheme = runtime.NewScheme()
	codecs        = serializer.NewCodecFactory(runtimeScheme)
	deserializer  = codecs.UniversalDeserializer()

	ignoredNamespaces = []string{
		metav1.NamespaceSystem,
		metav1.NamespacePublic,
	}
)

type Server struct {
	srv       *http.Server
	crdClient versioned.Interface
}

func NewServer(k8sClient kubernetes.Interface, crdClient versioned.Interface, addr string, cert tls.Certificate) (*Server, error) {
	srv := &Server{
		srv: &http.Server{
			Addr: addr,
			TLSConfig: &tls.Config{
				Certificates: []tls.Certificate{cert},
			},
		},
		crdClient: crdClient,
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/validate", srv.admissionHandler)
	srv.srv.Handler = mux

	return srv, nil
}

func (s *Server) Run(stopCh <-chan struct{}) error {
	log.Printf("Starting the Webhook Server...")

	errCh := make(chan error)
	go func() {
		// we configure the server with a proper certificate when we create it,
		// no need to specify file paths here. We don't want to be bound to
		// loading the cert from file.
		if err := s.srv.ListenAndServeTLS("", ""); err != nil {
			errCh <- err
		}
	}()

	select {
	case err := <-errCh:
		return err
	case <-stopCh:
		log.Printf("Stopping the Webhook Server...")
		s.srv.Shutdown(context.Background())
	}

	return nil
}

func (s *Server) admissionHandler(w http.ResponseWriter, r *http.Request) {
	var body []byte
	if r.Body != nil {
		data, err := ioutil.ReadAll(r.Body)

		if err != nil {
			http.Error(w, "could not read out body", http.StatusInternalServerError)
			return
		}

		body = data
	}

	if len(body) == 0 {
		http.Error(w, "empty body", http.StatusBadRequest)
		return
	}

	if r.Header.Get("Content-Type") != "application/json" {
		http.Error(w, "invalid Content-Type, expect `application/json`", http.StatusUnsupportedMediaType)
		return
	}

	var admissionResponse *v1beta1.AdmissionResponse
	ar := v1beta1.AdmissionReview{}
	if _, _, err := deserializer.Decode(body, nil, &ar); err != nil {
		admissionResponse = &v1beta1.AdmissionResponse{
			Result: &metav1.Status{
				Message: err.Error(),
			},
		}
	} else if !ignoredNamespace(ar.Request.Namespace) {
		switch r.URL.Path {
		case "/validate":
			admissionResponse = s.validate(&ar)
		}
	}

	admissionReview := v1beta1.AdmissionReview{}
	admissionReview.Response = admissionResponse
	if ar.Request != nil {
		admissionReview.Response.UID = ar.Request.UID
	}

	resp, err := json.Marshal(admissionReview)
	if err != nil {
		log.Printf("Can't encode response: %v", err)
		http.Error(w, fmt.Sprintf("could not encode response: %v", err), http.StatusInternalServerError)
		return
	}

	if _, err := w.Write(resp); err != nil {
		log.Printf("Can't write response: %v", err)
		http.Error(w, fmt.Sprintf("could not write response: %v", err), http.StatusInternalServerError)
	}
}

func (s *Server) validate(ar *v1beta1.AdmissionReview) *v1beta1.AdmissionResponse {
	var dpl ev1beta1.Deployment
	if err := json.Unmarshal(ar.Request.Object.Raw, &dpl); err != nil {
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

	hapList, err := s.crdClient.Barbossa().
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

func ignoredNamespace(ns string) bool {
	for _, namespace := range ignoredNamespaces {
		if namespace == ns {
			return true
		}
	}

	return false
}
