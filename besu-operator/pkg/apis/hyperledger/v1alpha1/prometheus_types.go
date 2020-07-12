package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// PrometheusSpec defines the desired state of Prometheus
type PrometheusSpec struct {

	// Requests and limits
	// +optional
	// +kubebuilder:default:={memRequest: "256Mi", cpuRequest: "100m", memLimit: "512Mi", cpuLimit: "500m"}
	Resources Resources `json:"resources"`

	// Prometheus Image Configuration
	// +optional
	// +kubebuilder:default:={repository: prom/prometheus, tag: v2.11.1, pullPolicy: IfNotPresent}
	Image Image `json:"image,omitempty"`

	// Number of replica pods corresponding to prometheus node
	// +optional
	// +kubebuilder:default:=1
	Replicas int32 `json:"replicas,omitempty"`

	// NodePort
	// +optional
	// +kubebuilder:default:=30090
	NodePort int32 `json:"nodeport,omitempty"`
}

// PrometheusStatus defines the observed state of Prometheus
type PrometheusStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book-v1.book.kubebuilder.io/beyond_basics/generating_crd.html
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Prometheus is the Schema for the prometheus API
// +kubebuilder:subresource:status
// +kubebuilder:resource:path=prometheus,scope=Namespaced
type Prometheus struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// +kubebuilder:default:={resources: {memRequest: "256Mi", cpuRequest: "100m", memLimit: "512Mi", cpuLimit: "500m"}, image:{repository: prom/prometheus, tag: v2.11.1, pullPolicy: IfNotPresent}, replicas:1, nodeport:30090}
	Spec   PrometheusSpec   `json:"spec,omitempty"`
	Status PrometheusStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// PrometheusList contains a list of Prometheus
type PrometheusList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Prometheus `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Prometheus{}, &PrometheusList{})
}
