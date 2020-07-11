package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// BesuSpec defines the desired state of Besu
type BesuSpec struct {

	// Number of bootnodes in the network
	// +kubebuilder:default:2
	BootnodesCount int `json:"bootnodescount"`

	// Number of validators in the network
	// +kubebuilder:default:4
	ValidatorsCount int `json:"validatorscount"`

	// Number of members in the network
	// +optional
	// +kubebuilder:default:0
	Members int32 `json:"members"`

	// Bootnodes configurations
	// +optional
	Bootnodes []BesuNodeSpec `json:"bootnodes,omitempty"`

	// Validators configurations
	// +optional
	Validators []BesuNodeSpec `json:"validators,omitempty"`

	// Besu Image Configuration
	// +optional
	// +kubebuilder:default: {repository: hyperledger/besu; tag: 1.4.6; pullPolicy: IfNotPresent}
	Image Image `json:"image,omitempty"`

	// Besu Network Genesis Configuration
	// +optional
	// +kubebuilder:default: {genesis: {config: {chainId: 2018; constantinoplefixblock: 0; ibft2: {blockperiodseconds: 5; epochlength: 30000; requesttimeoutseconds: 10}}; nonce: 0x0; timestamp: 0x58ee40ba; extraData: 0xf83ea00000000000000000000000000000000000000000000000000000000000000000d5949811ebc35d7b06b3fa8dc5809a1f9c52751e1deb808400000000c0; gasLimit: 0x1fffffffffffff; difficulty: 0x1; mixHash: 0x63746963616c2062797a616e74696e65206661756c7420746f6c6572616e6365; coinbase: 0x0000000000000000000000000000000000000000; alloc: {9811ebc35d7b06b3fa8dc5809a1f9c52751e1deb: {balance: 0xad78ebc5ac6200000}}}}
	GenesisJSON GenesisJSON `json:"genesis.json,omitempty"`

	// Bootnodes are validators or not
	// +optional
	// +kubebuilder:default:false
	BootnodesAreValidators bool `json:"bootnodesarevalidators,omitempty"`

	// Deploy Grafana/Prometheus or not
	// +optional
	// +kubebuilder:default:=true
	Monitoring bool `json:"monitoring,omitempty"`

	// Defines prometheus spec
	// +optional
	Prometheus PrometheusSpec `json:"prometheus,omitempty"`

	// Defines grafana spec
	// +optional
	Grafna GrafanaSpec `json:"grafana,omitempty"`
}

// Image defines the desired Besu Image configurations
type Image struct {

	// Besu container image repository
	// +kubebuilder:default:hyperledger/besu
	Repository string `json:"repository"`

	// Besu container image tag
	// +kubebuilder:default: 1.4.6
	Tag string `json:"tag"`

	// Besu container image pull policy
	// +kubebuilder:default:IfNotPresent
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
	// Defaults to :
	// chainId: 2018
	//   constantinoplefixblock: 0
	//   ibft2:
	//     blockperiodseconds: 2
	//     epochlength: 30000
	// 	   requesttimeoutseconds: 10
	// +optional
	GenesisConfig GenesisConfig `json:"config"`

	// Nonce
	// +kubebuilder:default:0x0
	// +optional
	Nonce string `json:"nonce"`

	// Timestamp
	// +kubebuilder:default:0x58ee40ba
	// +optional
	Timestamp string `json:"timestamp"`

	// Set the block size limit (measured in gas)
	// +kubebuilder:default:0x47b760
	// +optional
	GasLimit string `json:"gasLimit"`

	// Specify a fixed difficulty in private networks
	// +kubebuilder:default:0x1
	// +optional
	Difficulty string `json:"difficulty"`

	// Hash for Istanbul block identification (IBFT 2.0).
	// +kubebuilder:default:0x63746963616c2062797a616e74696e65206661756c7420746f6c6572616e6365
	// +optional
	MixHash string `json:"mixHash"`

	// The coinbase address is the account to which mining rewards are paid.
	// +kubebuilder:default:0x0000000000000000000000000000000000000000
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
	// +kubebuilder:default:2018
	// +optional
	ChainID int `json:"chainId"`

	// In private networks; the milestone block defines the protocol version for the network
	// +kubebuilder:default:0
	// +optional
	ConstantinopleFixBlock int `json:"constantinoplefixblock"`

	// Ibft2 configurations
	// +kubebuilder:default: {blockperiodseconds:2; epochlength:30000; requesttimeoutseconds:10}
	// +optional
	Ibft2 Ibft2 `json:"ibft2"`
}

// Ibft2 options
type Ibft2 struct {

	// Minimum block time in seconds.
	// +kubebuilder:default:2
	// +optional
	BlockPeriodSeconds int `json:"blockperiodseconds"`

	// Number of blocks after which to reset all votes.
	// +kubebuilder:default:30000
	// +optional
	EpochLength int `json:"epochlength"`

	// 	Timeout for each consensus round before a round change.
	// +kubebuilder:default:10
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

	// Field to show whether child besu node resources have keys or not
	HaveKeys bool `json:"havekeys"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Besu is the Schema for the besus API
// +kubebuilder:subresource:status
// +kubebuilder:resource:path=besus,scope=Namespaced
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
