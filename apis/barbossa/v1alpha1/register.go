package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

var (
	schemeBuilder = runtime.NewSchemeBuilder(addKnownTypes)

	// AddToScheme applies the SchemeBuilder functions to a specified scheme
	AddToScheme = schemeBuilder.AddToScheme
)

const (
	groupName  = "barbossa.sphc.io"
	apiVersion = "v1alpha1"
)

// SchemeGroupVersion is the GroupVersion for the Monitor CRD.
var SchemeGroupVersion = schema.GroupVersion{Group: groupName, Version: apiVersion}

// Resource gets an Monitor GroupResource for a specified resource.
func Resource(resource string) schema.GroupResource {
	return SchemeGroupVersion.WithResource(resource).GroupResource()
}

func addKnownTypes(scheme *runtime.Scheme) error {
	scheme.AddKnownTypes(SchemeGroupVersion,
		&HighAvailabilityPolicy{},
		&HighAvailabilityPolicyList{},
	)
	metav1.AddToGroupVersion(scheme, SchemeGroupVersion)
	return nil
}
