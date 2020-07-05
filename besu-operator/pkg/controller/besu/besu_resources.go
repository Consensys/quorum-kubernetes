package besu

import (
	"fmt"
	"strconv"

	hyperledgerv1alpha1 "github.com/Sumaid/besu-kubernetes/besu-operator/pkg/apis/hyperledger/v1alpha1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

func (r *ReconcileBesu) besuRole(instance *hyperledgerv1alpha1.Besu) *rbacv1.Role {

	role := &rbacv1.Role{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Role",
			APIVersion: "rbac.authorization.k8s.io/v1beta1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      instance.ObjectMeta.Name + "-genesis-role",
			Namespace: instance.Namespace,
		},
		Rules: []rbacv1.PolicyRule{
			{
				APIGroups: []string{
					"",
				},
				Resources: []string{
					"secrets",
					"configmaps",
				},
				Verbs: []string{
					"get",
					"create",
					"list",
					"update",
					"delete",
				},
			},
			{
				APIGroups: []string{
					"",
				},
				Resources: []string{
					"pods",
				},
				Verbs: []string{
					"get",
					"list",
					"watch",
				},
			},
		},
	}
	controllerutil.SetControllerReference(instance, role, r.scheme)
	return role
}

func (r *ReconcileBesu) besuRoleBinding(instance *hyperledgerv1alpha1.Besu) *rbacv1.RoleBinding {

	rb := &rbacv1.RoleBinding{
		TypeMeta: metav1.TypeMeta{
			Kind:       "RoleBinding",
			APIVersion: "rbac.authorization.k8s.io/v1beta1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      instance.ObjectMeta.Name + "-rb",
			Namespace: instance.Namespace,
		},
		RoleRef: rbacv1.RoleRef{
			APIGroup: "rbac.authorization.k8s.io",
			Kind:     "Role",
			Name:     instance.ObjectMeta.Name + "-genesis-role",
		},
		Subjects: []rbacv1.Subject{
			{
				Kind:      "ServiceAccount",
				Name:      instance.ObjectMeta.Name + "-sa",
				Namespace: instance.Namespace,
			},
		},
	}
	controllerutil.SetControllerReference(instance, rb, r.scheme)
	return rb
}

func (r *ReconcileBesu) besuConfigMap(instance *hyperledgerv1alpha1.Besu) *corev1.ConfigMap {
	data := make(map[string]string)
	data["genesis.json"] =
		fmt.Sprintf(`{
			"genesis": {
			  "config": {
				 "chainId": 2018,
				 "constantinoplefixblock": 0,
				 "ibft2": {
				   "blockperiodseconds": 2,
				   "epochlength": 30000,
				   "requesttimeoutseconds": 10
				 }
			   },
			   "nonce": "0x0",
			   "timestamp": "0x58ee40ba",
			   "gasLimit": "0x47b760",
			   "difficulty": "0x1",
			   "mixHash": "0x63746963616c2062797a616e74696e65206661756c7420746f6c6572616e6365",
			   "coinbase": "0x0000000000000000000000000000000000000000",
			   "alloc": {
				  "fe3b557e8fb62b89f4916b721be55ceb828dbd73": {
					 "privateKey": "8f2a55949038a9610f50fb23b5883af3b4ecb3c3bb792cbcefbd1542c692be63",
					 "comment": "private key and this comment are ignored.  In a real chain, the private key should NOT be stored",
					 "balance": "0xad78ebc5ac6200000"
				  },
				  "627306090abaB3A6e1400e9345bC60c78a8BEf57": {
					"privateKey": "c87509a1c067bbde78beb793e6fa76530b6382a4c0241e5e4a9ec0a0f44dc0d3",
					"comment": "private key and this comment are ignored.  In a real chain, the private key should NOT be stored",
					"balance": "90000000000000000000000"
				  },
				  "f17f52151EbEF6C7334FAD080c5704D77216b732": {
					"privateKey": "ae6ae8e5ccbfb04590405997ee2d52d2b330726137b875053c36d94e974d162f",
					"comment": "private key and this comment are ignored.  In a real chain, the private key should NOT be stored",
					"balance": "90000000000000000000000"
				  }
				 }
			},
			"blockchain": {
			  "nodes": {
				"generate": true,
				  "count": %d
			  }
			}
		}`, instance.Spec.BootnodesCount+instance.Spec.ValidatorsCount)

	data["genesisnode"] = `
	{
		"config": {
		  "chainId": 2018,
		  "constantinoplefixblock": 0,
		  "ibft2": {
			"blockperiodseconds": 2,
			"epochlength": 30000,
			"requesttimeoutseconds": 10
		  }
		},
		"nonce": "0x0",
		"timestamp": "0x58ee40ba",
		"gasLimit": "0x47b760",
		"difficulty": "0x1",
		"mixHash": "0x63746963616c2062797a616e74696e65206661756c7420746f6c6572616e6365",
		"coinbase": "0x0000000000000000000000000000000000000000",
		"alloc": {
		  "fe3b557e8fb62b89f4916b721be55ceb828dbd73": {
			"privateKey": "8f2a55949038a9610f50fb23b5883af3b4ecb3c3bb792cbcefbd1542c692be63",
			"comment": "private key and this comment are ignored.  In a real chain, the private key should NOT be stored",
			"balance": "0xad78ebc5ac6200000"
		  },
		  "627306090abaB3A6e1400e9345bC60c78a8BEf57": {
			"privateKey": "c87509a1c067bbde78beb793e6fa76530b6382a4c0241e5e4a9ec0a0f44dc0d3",
			"comment": "private key and this comment are ignored.  In a real chain, the private key should NOT be stored",
			"balance": "90000000000000000000000"
		  },
		  "f17f52151EbEF6C7334FAD080c5704D77216b732": {
			"privateKey": "ae6ae8e5ccbfb04590405997ee2d52d2b330726137b875053c36d94e974d162f",
			"comment": "private key and this comment are ignored.  In a real chain, the private key should NOT be stored",
			"balance": "90000000000000000000000"
		  }
		},
		"extraData": "0xf87ea00000000000000000000000000000000000000000000000000000000000000000f85494ca6e9704586eb1fb38194308e2192e43b1e1979c94ce2276efc33fee3c321e634eac28a9476e53b71c94f466a7174230056004d11178d2647c12740fa58b94b83820d6cf4b7e5aa67a2b57969caa5cdf6dff49808400000000c0"
	  }`

	conf := &corev1.ConfigMap{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ConfigMap",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "besu-" + "configmap",
			Namespace: instance.Namespace,
			Labels: map[string]string{
				"app": "besu-" + "configmap",
			},
		},
		Data: data,
	}
	controllerutil.SetControllerReference(instance, conf, r.scheme)
	return conf
}

func (r *ReconcileBesu) besuInitJob(instance *hyperledgerv1alpha1.Besu) *batchv1.Job {

	var backofflimit int32 = 3
	var completions int32 = 1

	labels := make(map[string]string)
	labels["app"] = instance.ObjectMeta.Name + "-init-job"

	job := &batchv1.Job{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Job",
			APIVersion: "batch/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      instance.ObjectMeta.Name + "-init-job",
			Namespace: instance.Namespace,
		},
		Spec: batchv1.JobSpec{
			BackoffLimit: &backofflimit,
			Completions:  &completions,
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labels,
				},
				Spec: corev1.PodSpec{
					ServiceAccountName: instance.ObjectMeta.Name + "-sa",
					RestartPolicy:      "Never",
					Containers: []corev1.Container{
						corev1.Container{
							Name:            instance.ObjectMeta.Name + "-generate-genesis",
							Image:           "hyperledger/besu:1.4.6",
							ImagePullPolicy: "IfNotPresent",
							VolumeMounts: []corev1.VolumeMount{
								corev1.VolumeMount{
									Name:      "raw-config",
									MountPath: "/raw-config",
								},
								corev1.VolumeMount{
									Name:      "generated-config",
									MountPath: "/generated-config",
								},
							},
							Command: []string{
								"/bin/bash",
								"-c",
							},
							Args: []string{
								`
								apt-get update && apt-get install -y curl
								curl -LO https://storage.googleapis.com/kubernetes-release/release/$(curl -s https://storage.googleapis.com/kubernetes-release/release/stable.txt)/bin/linux/amd64/kubectl && chmod +x ./kubectl
								echo "Creating config ..."
								/opt/besu/bin/besu operator generate-blockchain-config --config-file=/raw-config/genesis.json --to=/generated-config
								echo "Creating genesis configmap in k8s ..."
								./kubectl create configmap --namespace ${NAMESPACE} besu-genesis --from-file=genesis.json=/generated-config/genesis.json
								echo "Creating validator secrets in k8s ..."
								i=1
								for f in /generated-config/keys/*; do
								  if [ -d ${f} ]; then
									echo $f
									sed 's/^0x//' ${f}/key.pub > ${f}/enode.key
									if [[ i -le ${BOOTNODES} ]]
									then
										./kubectl create secret --namespace ${NAMESPACE} generic besu-bootnode${i}-key --from-file=private.key=${f}/key.priv --from-file=public.key=${f}/key.pub --from-file=enode.key=${f}/enode.key
									else
									    j=$((i - BOOTNODES))
										./kubectl create secret --namespace ${NAMESPACE} generic besu-validator${j}-key --from-file=private.key=${f}/key.priv --from-file=public.key=${f}/key.pub --from-file=enode.key=${f}/enode.key
									fi
									i=$((i+1))
								  fi
								done
								echo "Completed ..."`,
							},
							Env: []corev1.EnvVar{
								{
									Name:  "NAMESPACE",
									Value: instance.ObjectMeta.Namespace,
								},
								{
									Name:  "BOOTNODES",
									Value: strconv.Itoa(instance.Spec.BootnodesCount),
								},
							},
						},
					},
					Volumes: []corev1.Volume{
						corev1.Volume{
							Name: "raw-config",
							VolumeSource: corev1.VolumeSource{
								ConfigMap: &corev1.ConfigMapVolumeSource{
									LocalObjectReference: corev1.LocalObjectReference{
										Name: "besu-configmap",
									},
								},
							},
						},
						corev1.Volume{
							Name: "generated-config",
							VolumeSource: corev1.VolumeSource{
								EmptyDir: &corev1.EmptyDirVolumeSource{},
							},
						},
					},
				},
			},
		},
	}
	controllerutil.SetControllerReference(instance, job, r.scheme)
	return job
}

func (r *ReconcileBesu) besuCleanupJob(instance *hyperledgerv1alpha1.Besu) *batchv1.Job {

	var backofflimit int32 = 3
	var completions int32 = 1

	labels := make(map[string]string)
	labels["app"] = instance.ObjectMeta.Name + "-cleanup-job"

	job := &batchv1.Job{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Job",
			APIVersion: "batch/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      instance.ObjectMeta.Name + "-cleanup-job",
			Namespace: instance.Namespace,
		},
		Spec: batchv1.JobSpec{
			BackoffLimit: &backofflimit,
			Completions:  &completions,
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labels,
				},
				Spec: corev1.PodSpec{
					ServiceAccountName: instance.ObjectMeta.Name + "-sa",
					RestartPolicy:      "Never",
					Containers: []corev1.Container{
						corev1.Container{
							Name:            instance.ObjectMeta.Name + "-delete-genesis",
							Image:           "pegasyseng/k8s-helper:v1.18.4",
							ImagePullPolicy: "IfNotPresent",
							Command: []string{
								"/bin/bash",
								"-c",
							},
							Args: []string{
								`
								apt-get update && apt-get install -y curl
								curl -LO https://storage.googleapis.com/kubernetes-release/release/$(curl -s https://storage.googleapis.com/kubernetes-release/release/stable.txt)/bin/linux/amd64/kubectl && chmod +x ./kubectl
								echo "Deleting genesis configmap in k8s ..."
								./kubectl delete configmap --namespace ${NAMESPACE} besu-genesis
								for (( f=1; f<=${TOTAL}; f++ )); do
								  echo $f
								  if [[ f -le ${BOOTNODES} ]]
								  then
								      echo "Deleting bootnode secret"
									  ./kubectl delete secret --namespace ${NAMESPACE} besu-bootnode${f}-key
								  else
									  echo "Deleting validator secret"
									  j=$((f - BOOTNODES))
									  ./kubectl delete secret --namespace ${NAMESPACE} besu-validator${j}-key
								  fi
								done
								echo "Completed ..."`,
							},
							Env: []corev1.EnvVar{
								{
									Name:  "NAMESPACE",
									Value: instance.ObjectMeta.Namespace,
								},
								{
									Name:  "BOOTNODES",
									Value: strconv.Itoa(instance.Spec.BootnodesCount),
								},
								{
									Name:  "TOTAL",
									Value: strconv.Itoa(instance.Spec.BootnodesCount + instance.Spec.ValidatorsCount),
								},
							},
						},
					},
				},
			},
		},
	}
	controllerutil.SetControllerReference(instance, job, r.scheme)
	return job
}

func (r *ReconcileBesu) newBesuNode(instance *hyperledgerv1alpha1.Besu,
	name string,
	nodeType string,
	bootsCount int,
) *hyperledgerv1alpha1.BesuNode {
	node := &hyperledgerv1alpha1.BesuNode{
		TypeMeta: metav1.TypeMeta{
			Kind:       "BesuNode",
			APIVersion: "hyperledger.org/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: instance.Namespace,
		},
		Spec: hyperledgerv1alpha1.BesuNodeSpec{
			Type:      nodeType,
			Bootnodes: bootsCount,
			Replicas:  1,
			Image: hyperledgerv1alpha1.Image{
				Repository: "hyperledger/besu",
				Tag:        "1.4.6",
				PullPolicy: "IfNotPresent",
			},
			Resources: hyperledgerv1alpha1.Resources{
				MemRequest: "1024Mi",
				CPURequest: "100m",
				MemLimit:   "2048Mi",
				CPULimit:   "500m",
			},
			P2P: hyperledgerv1alpha1.PortConfig{
				Enabled:               true,
				Host:                  "0.0.0.0",
				Port:                  30303,
				Discovery:             true,
				AuthenticationEnabled: false,
			},
			RPC: hyperledgerv1alpha1.PortConfig{
				Enabled:               true,
				Host:                  "0.0.0.0",
				Port:                  8545,
				AuthenticationEnabled: false,
			},
			WS: hyperledgerv1alpha1.PortConfig{
				Enabled:               false,
				Host:                  "0.0.0.0",
				Port:                  8546,
				AuthenticationEnabled: false,
			},
			GraphQl: hyperledgerv1alpha1.PortConfig{
				Enabled:               false,
				Host:                  "0.0.0.0",
				Port:                  8547,
				AuthenticationEnabled: false,
			},
		},
	}
	controllerutil.SetControllerReference(instance, node, r.scheme)
	return node
}
