package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// GrafanaSpec defines the desired state of Grafana
type GrafanaSpec struct {
	// Owner metadata name
	// +optional
	Owner string `json:"owner,omitempty"`

	// Grafana Requests and limits
	// +optional
	// +kubebuilder:default:={memRequest: "256Mi", cpuRequest: "100m", memLimit: "512Mi", cpuLimit: "500m"}
	Resources Resources `json:"resources"`

	// Grafana Image Configuration
	// +optional
	// +kubebuilder:default:={repository: grafana/grafana, tag: "6.2.5", pullPolicy: IfNotPresent}
	Image Image `json:"image,omitempty"`

	// Number of replica pods corresponding to grafana node
	// +optional
	// +kubebuilder:default:=1
	Replicas int32 `json:"replicas,omitempty"`

	// NodePort
	// +optional
	// +kubebuilder:default:=30030
	NodePort int32 `json:"nodeport,omitempty"`
}

// GrafanaStatus defines the observed state of Grafana
type GrafanaStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book-v1.book.kubebuilder.io/beyond_basics/generating_crd.html
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Grafana is the Schema for the grafanas API
// +kubebuilder:subresource:status
// +kubebuilder:resource:path=grafanas,scope=Namespaced
type Grafana struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// +kubebuilder:default:={resources: {memRequest: "256Mi", cpuRequest: "100m", memLimit: "512Mi", cpuLimit: "500m"}, image:{repository: grafana/grafana, tag: "6.2.5", pullPolicy: IfNotPresent}, replicas:1, nodeport:30030}
	Spec   GrafanaSpec   `json:"spec,omitempty"`
	Status GrafanaStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// GrafanaList contains a list of Grafana
type GrafanaList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Grafana `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Grafana{}, &GrafanaList{})
}
