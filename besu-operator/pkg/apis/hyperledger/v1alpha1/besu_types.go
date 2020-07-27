package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// BesuSpec defines the desired state of Besu
type BesuSpec struct {

	// Number of bootnodes in the network
	// +kubebuilder:default:=2
	BootnodesCount int `json:"bootnodescount"`

	// Number of validators in the network
	// +kubebuilder:default:=4
	ValidatorsCount int `json:"validatorscount"`

	// Number of members in the network
	// +optional
	// +kubebuilder:default:=0
	Members int32 `json:"members"`

	// Bootnodes keys
	// +optional
	BootnodeKeys []Key `json:"bootnodeKeys,omitempty"`

	// Validators keys
	// +optional
	ValidatorKeys []Key `json:"validatorKeys,omitempty"`

	// Common Besu nodes configuration
	// +optional
	// +kubebuilder:default:={replicas:2, image:{repository: hyperledger/besu, tag: "1.4.6", pullPolicy: IfNotPresent}, resources:{memRequest: "1024Mi", cpuRequest: "100m", memLimit: "2048Mi", cpuLimit: "500m"}, p2p: {enabled: true, host: "0.0.0.0", port: 30303, discovery: true, authenticationEnabled: false}, rpc:{enabled: true, host: "0.0.0.0", port: 8545, authenticationEnabled: false}, ws: {enabled: false, host: "0.0.0.0", port: 8546, authenticationEnabled: false}, graphql: {enabled: false, host: "0.0.0.0", port: 8547, authenticationEnabled: false}, metrics: {enabled: true, host: "0.0.0.0", port: 9545}}
	BesuNodeSpec BesuNodeSpec `json:"besunodespec,omitempty"`

	// Besu Network Genesis Configuration
	// +optional
	// +kubebuilder:default:={genesis: {config: {chainId: 2018, constantinoplefixblock: 0, ibft2: {blockperiodseconds: 2, epochlength: 30000, requesttimeoutseconds: 10}}, nonce: "0x0", timestamp: "0x58ee40ba", gasLimit: "0x47b760", difficulty: "0x1", mixHash: "0x63746963616c2062797a616e74696e65206661756c7420746f6c6572616e6365", coinbase: "0x0000000000000000000000000000000000000000"}}
	GenesisJSON GenesisJSON `json:"genesis.json,omitempty"`

	// Bootnodes are validators or not
	// +optional
	// +kubebuilder:default:=false
	BootnodesAreValidators bool `json:"bootnodesarevalidators,omitempty"`

	// Deploy Grafana/Prometheus or not
	// +optional
	// +kubebuilder:default:=true
	Monitoring bool `json:"monitoring,omitempty"`

	// Defines prometheus spec
	// +optional
	// +kubebuilder:default:={resources: {memRequest: "256Mi", cpuRequest: "100m", memLimit: "512Mi", cpuLimit: "500m"}, image:{repository: prom/prometheus, tag: v2.11.1, pullPolicy: IfNotPresent}, replicas:1, nodeport:30090}
	PrometheusSpec PrometheusSpec `json:"prometheusspec,omitempty"`

	// Defines grafana spec
	// +optional
	// +kubebuilder:default:={resources: {memRequest: "256Mi", cpuRequest: "100m", memLimit: "512Mi", cpuLimit: "500m"}, image:{repository: grafana/grafana, tag: "6.2.5", pullPolicy: IfNotPresent}, replicas:1, nodeport:30030}
	GrafanaSpec GrafanaSpec `json:"grafanaspec,omitempty"`
}

// Key defines the private & public keys of bootnodes & validators
type Key struct {
	// Public key
	PubKey string `json:"pubkey"`

	// Private key
	PrivKey string `json:"privkey"`
}

// Image defines the desired Besu Image configurations
type Image struct {

	// Image repository
	// +optional
	Repository string `json:"repository"`

	// Image tag
	// +optional
	Tag string `json:"tag"`

	// Image pull policy
	// +kubebuilder:default:=IfNotPresent
	// +optional
	PullPolicy string `json:"pullPolicy"`
}

// GenesisJSON defines the genesis.json file
type GenesisJSON struct {
	// +optional
	Genesis Genesis `json:"genesis,omitempty"`
	// +optional
	Blockchain Blockchain `json:"blockchain,omitempty"`
}

// Genesis defines the desired configurations of genesis
type Genesis struct {

	// GenesisConfig
	// +optional
	GenesisConfig GenesisConfig `json:"config"`

	// Nonce
	// +kubebuilder:default:="0x0"
	// +optional
	Nonce string `json:"nonce"`

	// Timestamp
	// +kubebuilder:default:="0x58ee40ba"
	// +optional
	Timestamp string `json:"timestamp"`

	// Set the block size limit (measured in gas)
	// +kubebuilder:default:="0x47b760"
	// +optional
	GasLimit string `json:"gasLimit"`

	// Specify a fixed difficulty in private networks
	// +kubebuilder:default:="0x1"
	// +optional
	Difficulty string `json:"difficulty"`

	// Hash for Istanbul block identification (IBFT 2.0).
	// +kubebuilder:default:="\"0x63746963616c2062797a616e74696e65206661756c7420746f6c6572616e6365\""
	// +optional
	MixHash string `json:"mixHash"`

	// The coinbase address is the account to which mining rewards are paid.
	// +kubebuilder:default:="0x0000000000000000000000000000000000000000"
	// +optional
	CoinBase string `json:"coinbase"`

	// Predeploy contracts when starting Besu with Ether
	// +optional
	Alloc map[string]Transaction `json:"alloc,omitempty"`

	// +optional
	ExtraData string `json:"extraData,omitempty"`
}

// Blockchain defines number of network nodes
type Blockchain struct {
	Nodes Nodes `json:"nodes,omitempty"`
}

// Nodes defines number of nodes in the network
type Nodes struct {
	Generate bool `json:"generate,omitempty"`
	Count    int  `json:"count,omitempty"`
}

// GenesisConfig defines config options in genesis
type GenesisConfig struct {

	// The identifier of the private Ethereum network
	// +kubebuilder:default:=2018
	// +optional
	ChainID int `json:"chainId"`

	// In private networks, the milestone block defines the protocol version for the network
	// +kubebuilder:default:=0
	// +optional
	ConstantinopleFixBlock int `json:"constantinoplefixblock"`

	// Ibft2 configurations
	// +kubebuilder:default:={blockperiodseconds:2, epochlength:30000, requesttimeoutseconds:10}
	// +optional
	Ibft2 Ibft2 `json:"ibft2"`
}

// Ibft2 options
type Ibft2 struct {

	// Minimum block time in seconds.
	// +kubebuilder:default:=2
	// +optional
	BlockPeriodSeconds int `json:"blockperiodseconds"`

	// Number of blocks after which to reset all votes.
	// +kubebuilder:default:=30000
	// +optional
	EpochLength int `json:"epochlength"`

	// 	Timeout for each consensus round before a round change.
	// +kubebuilder:default:=10
	// +optional
	RequestTimeoutSeconds int `json:"requesttimeoutseconds"`
}

// Transaction defines alloc
type Transaction struct {

	// privateKey
	// +optional
	PrivateKey string `json:"privateKey"`

	//Comment
	// +optional
	Comment string `json:"comment"`

	// Balance
	// +optional
	Balance string `json:"balance"`
}

// BesuStatus defines the observed state of Besu
type BesuStatus struct {

	// Shows how many bootnodes are ready
	BootnodesReady string `json:"bootnodesready,omitempty"`

	// Shows how many validators are ready
	ValidatorsReady string `json:"validatorsready,omitempty"`

	// Shows how many members are ready
	MembersReady string `json:"membersready,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Besu is the Schema for the besus API
// +kubebuilder:subresource:status
// +kubebuilder:resource:path=besus,scope=Namespaced
// +kubebuilder:printcolumn:name="BootnodesReady",type=string,JSONPath=`.status.bootnodesready`
// +kubebuilder:printcolumn:name="ValidatorsReady",type=string,JSONPath=`.status.validatorsready`
// +kubebuilder:printcolumn:name="MembersReady",type=string,JSONPath=`.status.membersready`
// +kubebuilder:printcolumn:name="Repository",type=string,JSONPath=`.spec.besunodespec.image.repository`
// +kubebuilder:printcolumn:name="Tag",type=string,JSONPath=`.spec.besunodespec.image.tag`
type Besu struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   BesuSpec   `json:"spec,omitempty"`
	Status BesuStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// BesuList contains a list of Besu
type BesuList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Besu `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Besu{}, &BesuList{})
}
