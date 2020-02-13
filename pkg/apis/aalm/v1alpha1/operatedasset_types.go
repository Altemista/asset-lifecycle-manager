package v1alpha1

import (
	olmv1alpha1 "github.com/operator-framework/operator-lifecycle-manager/pkg/api/apis/operators/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// OperatedAssetSpec defines the desired state of OperatedAsset
type OperatedAssetSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book-v1.book.kubebuilder.io/beyond_basics/generating_crd.html
	Operator olmv1alpha1.SubscriptionSpec       `json:"operator"`
	Asset    runtime.RawExtension               `json:"asset"`
}

type OperatedAssetState string

const (
	OperatedAssetStateNone             = ""
	OperatedAssetStateInstallFailed    = "FAILED"
	OperatedAssetStateInstallPending   = "PENDING"
	OperatedAssetStateInstallRunning   = "RUNNING"
	OperatedAssetStateInstallCompleted = "COMPLETED"
)

// OperatedAssetStatus defines the observed state of OperatedAsset
type OperatedAssetStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book-v1.book.kubebuilder.io/beyond_basics/generating_crd.html
	State OperatedAssetState `json:"state,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// OperatedAsset is the Schema for the operatedassets API
// +kubebuilder:subresource:status
// +kubebuilder:resource:path=operatedassets,scope=Namespaced
type OperatedAsset struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   OperatedAssetSpec   `json:"spec,omitempty"`
	Status OperatedAssetStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// OperatedAssetList contains a list of OperatedAsset
type OperatedAssetList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []OperatedAsset `json:"items"`
}

func init() {
	SchemeBuilder.Register(&OperatedAsset{}, &OperatedAssetList{})
}
