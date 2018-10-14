package v1alpha1

import (
	"k8s.io/api/extensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

// HighAvailabilityPolicySpec defines the detailed configuration we want our
// Deployment objects to adhere to with regards to High Availability.
type HighAvailabilityPolicySpec struct {
	// The weight of the HighAvailabilityPolicy determines the priority of this
	// Policy. If a resource has multiple policies that target it, the policy
	// with the highest weight will be used to validate that resource.
	Weight int `json:"weight"`

	// Selector is a LabelSelector to select a set of Deployments which fall
	// under this Policy for validation.
	Selector *metav1.LabelSelector `json:"selector"`

	// The Replica Configuration allows us to configure specific values for the
	// Deployment Replica count.
	Replicas *HighAvailabilityPolicyReplicas `json:"replicas,omitempty"`

	// The Strategy allows us to configure specific boundaries in which the
	// Strategy for the selected Deployments should fall.
	Strategy *HighAvailabilityPolicyStrategy `json:"strategy,omitempty"`
}

// HighAvailabilityPolicyReplicas is the configuration to validate the Replica
// count of a Deployment configuration.
type HighAvailabilityPolicyReplicas struct {
	// Minimum defines the minimum of Replicas we want our Deployments to have
	// configured.
	Minimum int32 `json:"minimum"`
}

// HighAvailabilityPolicyStrategy is the configuration to validate the
// Strategy for a Deployment Resource.
type HighAvailabilityPolicyStrategy struct {
	// Type of Deployment. The selected deployments must be of this type to pass
	// the validation.
	Type v1beta1.DeploymentStrategyType `json:"type"`

	// Rolling Update configuration parameters. If the Type is RollingUpdate,
	// this will be used to validate the linked RollingUpdate configuration.
	RollingUpdate *HighAvailabilityPolicyRollingUpdate `json:"rollingUpdate,omitempty"`
}

// HighAvailabilityPolicyRollingUpdate is the configuration to validate the
// RollingUpdate Strategy of a Deployment Resource.
type HighAvailabilityPolicyRollingUpdate struct {
	// MinSurge is used to enforce that the `maxForce` on the Deployment
	// Resource is at least this value.
	MinSurge intstr.IntOrString `json:"minSurge"`

	// MaxSurge is used to enforce that the `maxForce` on the Deployment
	// Resource is at most this value.
	MaxSurge intstr.IntOrString `json:"maxSurge"`
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// HighAvailabilityPolicy is used to enforce some availability configuration for
// Deployments. The Policy object will be used as a Validation mechanism through
// webhooks and prevent non-adhering Deployments from being added to the cluster.
type HighAvailabilityPolicy struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec HighAvailabilityPolicySpec `json:"spec"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// HighAvailabilityPolicyList is a list of HighAvailabilityPolicies which are
// available in the cluster.
type HighAvailabilityPolicyList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`
	Items           []HighAvailabilityPolicy `json:"items"`
}
