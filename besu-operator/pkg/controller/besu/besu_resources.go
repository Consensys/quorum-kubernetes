package besu

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"

	hyperledgerv1alpha1 "github.com/Sumaid/besu-kubernetes/besu-operator/pkg/apis/hyperledger/v1alpha1"
	corev1 "k8s.io/api/core/v1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	"encoding/hex"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/rlp"
	"golang.org/x/crypto/sha3"
)

func (r *ReconcileBesu) besuGenesisConfigMap(instance *hyperledgerv1alpha1.Besu) *corev1.ConfigMap {
	data := make(map[string]string)

	GenesisObject := instance.Spec.GenesisJSON
	GenesisObject.Genesis.ExtraData = r.getExtraData(instance)

	b, err := json.Marshal(GenesisObject.Genesis)
	if err != nil {
		log.Error(err, "Failed to convert genesis to json", "Namespace", instance.Namespace, "Name", instance.Name)
		return nil
	}
	data["genesis.json"] = string(b)

	conf := &corev1.ConfigMap{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ConfigMap",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "besu-" + "genesis",
			Namespace: instance.Namespace,
			Labels: map[string]string{
				"app": "besu-" + "genesis",
			},
		},
		Data: data,
	}
	controllerutil.SetControllerReference(instance, conf, r.scheme)
	return conf
}

func getAddress(pubkey string) string {
	// Public key to address
	publicKeyBytes, _ := hex.DecodeString(pubkey)
	var buf []byte
	hash := sha3.NewLegacyKeccak256()
	hash.Write(publicKeyBytes[1:]) // remove EC prefix 04
	buf = hash.Sum(nil)
	return hexutil.Encode(buf[12:])
}

func (r *ReconcileBesu) getExtraData(instance *hyperledgerv1alpha1.Besu) string {
	addresses := []string{}
	for _, key := range instance.Spec.ValidatorKeys {
		pubkey := key.PubKey
		if pubkey[:2] == "0x" {
			pubkey = pubkey[2:]
		}
		addresses = append(addresses, getAddress(pubkey)[2:])
	}
	vanity := [32]byte{}
	adds := []interface{}{}
	vote := []byte{}
	buf := new(bytes.Buffer)
	binary.Write(buf, binary.LittleEndian, uint32(0))
	round := buf.Bytes()
	seals := []interface{}{}
	for _, v := range addresses {
		b, err := hex.DecodeString(v)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
		}
		adds = append(adds, b)
	}
	out, _ := rlp.EncodeToBytes([]interface{}{vanity, adds, vote, round, seals})
	extraData := hexutil.Encode(out)
	return extraData
}

func (r *ReconcileBesu) newBesuNode(instance *hyperledgerv1alpha1.Besu,
	name string,
	nodeType string,
	bootsCount int) *hyperledgerv1alpha1.BesuNode {
	node := &hyperledgerv1alpha1.BesuNode{
		TypeMeta: metav1.TypeMeta{
			Kind:       "BesuNode",
			APIVersion: "hyperledger.org/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: instance.Namespace,
		},
		Spec: instance.Spec.BesuNodeSpec,
	}
	node.Spec.Bootnodes = bootsCount
	node.Spec.Type = nodeType
	controllerutil.SetControllerReference(instance, node, r.scheme)
	return node
}

func (r *ReconcileBesu) newPrometheus(instance *hyperledgerv1alpha1.Besu) *hyperledgerv1alpha1.Prometheus {
	prometheusNode := &hyperledgerv1alpha1.Prometheus{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Prometheus",
			APIVersion: "hyperledger.org/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      instance.ObjectMeta.Name + "-prometheus",
			Namespace: instance.Namespace,
		},
		Spec: instance.Spec.PrometheusSpec,
	}
	controllerutil.SetControllerReference(instance, prometheusNode, r.scheme)
	return prometheusNode
}

func (r *ReconcileBesu) newGrafana(instance *hyperledgerv1alpha1.Besu) *hyperledgerv1alpha1.Grafana {
	grafanaSpec := instance.Spec.GrafanaSpec
	grafanaSpec.Owner = instance.ObjectMeta.Name
	grafanaNode := &hyperledgerv1alpha1.Grafana{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Grafana",
			APIVersion: "hyperledger.org/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      instance.ObjectMeta.Name + "-grafana",
			Namespace: instance.Namespace,
		},
		Spec: grafanaSpec,
	}
	controllerutil.SetControllerReference(instance, grafanaNode, r.scheme)
	return grafanaNode
}

func (r *ReconcileBesu) besuSecret(instance *hyperledgerv1alpha1.Besu,
	name string,
	privkey string,
	pubkey string) *corev1.Secret {
	var en, pb string
	pb = ""
	en = ""
	if pubkey != "" {
		if pubkey[:2] == "0x" {
			en = pubkey[2:]
			pb = pubkey
		} else {
			en = pubkey
			pb = "0x" + pubkey
		}
	}
	secr := &corev1.Secret{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Secret",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "besu-" + name + "-key",
			Namespace: instance.Namespace,
			Labels: map[string]string{
				"app": "besu-" + name + "-key",
			},
		},
		Type: "Opaque",
		StringData: map[string]string{
			"private.key": privkey,
			"public.key":  pb,
			"enode.key":   en,
		},
	}
	controllerutil.SetControllerReference(instance, secr, r.scheme)
	return secr
}
