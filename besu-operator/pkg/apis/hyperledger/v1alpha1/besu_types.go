package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// BesuSpec defines the desired state of Besu
type BesuSpec struct {

	// Number of bootnodes in the network
	// +kubebuider:default:2
	BootnodesCount int `json:"bootnodescount"`

	// Number of validators in the network
	// +kubebuider:default:4
	ValidatorsCount int `json:"validatorscount"`

	// Number of members in the network
	// +optional
	// +kubebuider:default:0
	Members int32 `json:"members"`

	// Bootnodes configurations
	// +optional
	Bootnodes []BesuNodeSpec `json:"bootnodes,omitempty"`

	// Validators configurations
	// +optional
	Validators []BesuNodeSpec `json:"validators,omitempty"`

	// Besu Image Configuration
	// +optional
	// +kubebuider:default: {repository: hyperledger/besu; tag: 1.4.6; pullPolicy: IfNotPresent}
	Image Image `json:"image,omitempty"`

	// Besu Network Genesis Configuration
	// +optional
	Genesis Genesis `json:"genesis,omitempty"`
}

// Image defines the desired Besu Image configurations
type Image struct {

	// Besu container image repository
	// +kubebuider:default:hyperledger/besu
	Repository string `json:"repository"`

	// Besu container image tag
	// +kubebuider:default: 1.4.6
	Tag string `json:"tag"`

	// Besu container image pull policy
	// +kubebuider:default:IfNotPresent
	// +optional
	PullPolicy string `json:"pullPolicy"`
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
	// 	requesttimeoutseconds: 10
	GenesisConfig GenesisConfig `json:"config"`

	// Nonce
	// +kubebuider:default:0x0
	Nonce string `json:"nonce"`

	// Timestamp
	// +kubebuider:default:0x58ee40ba
	Timestamp string `json:"timestamp"`

	// Set the block size limit (measured in gas)
	// +kubebuider:default:0x47b760
	GasLimit string `json:"gasLimit"`

	// Specify a fixed difficulty in private networks
	// +kubebuider:default:0x1
	Difficulty string `json:"difficulty"`

	// Hash for Istanbul block identification (IBFT 2.0).
	// +kubebuider:default:0x63746963616c2062797a616e74696e65206661756c7420746f6c6572616e6365
	MixHash string `json:"mixHash"`

	// The coinbase address is the account to which mining rewards are paid.
	// +kubebuider:default:0x0000000000000000000000000000000000000000
	CoinBase string `json:"coinbase"`

	// Predeploy contracts when starting Besu with Ether
	Alloc []Transaction `json:"alloc,omitempty"`
}

// GenesisConfig defines config options in genesis
type GenesisConfig struct {

	// The identifier of the private Ethereum network
	// +kubebuider:default:2018
	ChainID int `json:"chainId"`

	// In private networks; the milestone block defines the protocol version for the network
	// +kubebuider:default:0
	ConstantinopleFixBlock int `json:"constantinoplefixblock"`

	// Ibft2 configurations
	// +kubebuider:default: {blockperiodseconds:2; epochlength:30000; requesttimeoutseconds:10}
	Ibft2 Ibft2 `json:"ibft2"`
}

// Ibft2 options
type Ibft2 struct {

	// Minimum block time in seconds.
	// +kubebuider:default:2
	BlockPeriodSeconds int `json:"blockperiodseconds"`

	// Number of blocks after which to reset all votes.
	// +kubebuider:default:30000
	EpochLength int `json:"epochlength"`

	// 	Timeout for each consensus round before a round change.
	// +kubebuider:default:10
	RequestTimeoutSeconds int `json:"requesttimeoutseconds"`
}

// Transaction defines alloc
type Transaction struct {

	// Address
	Address string `json:"address"`

	// Balance
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
