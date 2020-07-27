package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// BesuNodeSpec defines the desired state of BesuNode
type BesuNodeSpec struct {

	// Type of node, takes one of the values : Bootnode, Validator, Member
	// +kubebuilder:validation:Enum:["Member", "Bootnode", "Validator"]
	// +kubebuilder:default:=Member
	// +optional
	Type string `json:"type"`

	// Number of replica pods corresponding to this node
	// +optional
	// +kubebuilder:default:=2
	Replicas int32 `json:"replicas"`

	// Besu Image Configuration
	// +optional
	// +kubebuilder:default:={repository: hyperledger/besu, tag: "1.4.6", pullPolicy: IfNotPresent}
	Image Image `json:"image"`

	// Requests and limits
	// +optional
	// +kubebuilder:default:={memRequest: "1024Mi", cpuRequest: "100m", memLimit: "2048Mi", cpuLimit: "500m"}
	Resources Resources `json:"resources"`

	// P2P Port configuration
	// +optional
	// +kubebuilder:default:={enabled: true, host: "0.0.0.0", port: 30303, discovery: true, authenticationEnabled: false}
	P2P PortConfig `json:"p2p"`

	// RPC Port Configuration
	// +optional
	// +kubebuilder:default:={enabled: true, host: "0.0.0.0", port: 8545, authenticationEnabled: false}
	RPC PortConfig `json:"rpc"`

	// WS
	// +optional
	// +kubebuilder:default:={enabled: false, host: "0.0.0.0", port: 8546, authenticationEnabled: false}
	WS PortConfig `json:"ws"`

	// GraphQl
	// +optional
	// +kubebuilder:default:={enabled: false, host: "0.0.0.0", port: 8547, authenticationEnabled: false}
	GraphQl PortConfig `json:"graphql"`

	// +optional
	// +kubebuilder:default:={enabled: true, host: "0.0.0.0", port: 9545}
	Metrics PortConfig `json:"metrics"`

	// Defaults to ["*"]
	// +optional
	HTTPWhitelist string `json:"httpwhitelist"`

	// +optional
	Bootnodes int `json:"bootnodes"`

	// Size of the Volume
	// +kubebuider:default:="1Gi"
	// +optional
	PVCSizeLimit string `json:"pvcSizeLimit"`

	// Storage class of the Volume
	// +kubebuider:default:="standard"
	// +optional
	PVCStorageClass string `json:"pvcStorageClass"`
}

// PortConfig defines port configurations of different types of ports
type PortConfig struct {
	// Port is enabled or not
	// +optional
	Enabled bool `json:"enabled"`

	// Host
	// +kubebuider:default=0.0.0.0
	// +optional
	Host string `json:"host"`

	// Port
	// +optional
	Port int `json:"port"`

	// +optional
	API string `json:"api"`

	// +optional
	CorsOrigins string `json:"corsOrigins"`

	// +optional
	AuthenticationEnabled bool `json:"authenticationEnabled"`

	// +optional
	Discovery bool `json:"discovery"`
}

// Resources defines requests and limits of CPU and memory
type Resources struct {

	// Memory Request
	// +optional
	MemRequest string `json:"memRequest"`

	// CPU Request
	// +optional
	CPURequest string `json:"cpuRequest"`

	// Memory Limit
	// +optional
	MemLimit string `json:"memLimit"`

	// CPU Limit
	// +optional
	CPULimit string `json:"cpuLimit"`
}

// BesuNodeStatus defines the observed state of BesuNode
type BesuNodeStatus struct {
	Replicas int32 `json:"replicas"`

	ReadyReplicas int32 `json:"readyreplicas"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// BesuNode is the Schema for the besunodes API
// +kubebuilder:subresource:status
// +kubebuilder:resource:path=besunodes,scope=Namespaced
// +kubebuilder:printcolumn:name="Replicas",type=string,JSONPath=`.status.replicas`
// +kubebuilder:printcolumn:name="ReadyReplicas",type=string,JSONPath=`.status.readyreplicas`
type BesuNode struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   BesuNodeSpec   `json:"spec,omitempty"`
	Status BesuNodeStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// BesuNodeList contains a list of BesuNode
type BesuNodeList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []BesuNode `json:"items"`
}

func init() {
	SchemeBuilder.Register(&BesuNode{}, &BesuNodeList{})
}
