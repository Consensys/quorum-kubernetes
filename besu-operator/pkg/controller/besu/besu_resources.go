package besu

import (
	"encoding/json"
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

	GenesisObject := instance.Spec.GenesisJSON
	count := instance.Spec.ValidatorsCount
	if instance.Spec.BootnodesAreValidators {
		count = count + instance.Spec.BootnodesCount
	}
	GenesisObject.Blockchain = hyperledgerv1alpha1.Blockchain{
		Nodes: hyperledgerv1alpha1.Nodes{
			Generate: true,
			Count:    count,
		},
	}
	b, err := json.Marshal(GenesisObject)
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
							Image:           instance.Spec.BesuNodeSpec.Image.Repository + ":" + instance.Spec.BesuNodeSpec.Image.Tag,
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
									if [ ${BOOTNODESAREVALIDATORS} = "true" ];
									then
										if [[ i -le ${BOOTNODES} ]]
										then
											./kubectl create secret --namespace ${NAMESPACE} generic besu-bootnode${i}-key --from-file=private.key=${f}/key.priv --from-file=public.key=${f}/key.pub --from-file=enode.key=${f}/enode.key
										else
											j=$((i - BOOTNODES))
											./kubectl create secret --namespace ${NAMESPACE} generic besu-validator${j}-key --from-file=private.key=${f}/key.priv --from-file=public.key=${f}/key.pub --from-file=enode.key=${f}/enode.key
										fi
									else
										./kubectl create secret --namespace ${NAMESPACE} generic besu-validator${i}-key --from-file=private.key=${f}/key.priv --from-file=public.key=${f}/key.pub --from-file=enode.key=${f}/enode.key
									fi
									i=$((i+1))
								  fi
								done

								if [ ${BOOTNODESAREVALIDATORS} = "false" ];
								then
									echo "Creating bootnode keys ..."
									for (( j=1; j<=${BOOTNODES}; j++ ));
									do
										/opt/besu/bin/besu public-key export --to=public${j}.key
										sed 's/^0x//' ./public${j}.key > enode${j}.key
										echo "Creating bootnode ${j} secrets in k8s ..."
										./kubectl create secret generic besu-bootnode${j}-key --namespace ${NAMESPACE} --from-file=private.key=/opt/besu/key --from-file=public.key=./public${j}.key --from-file=enode.key=./enode${j}.key
										rm ./public${j}.key ./enode${j}.key /opt/besu/key
									done
								fi
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
									Name:  "BOOTNODESAREVALIDATORS",
									Value: strconv.FormatBool(instance.Spec.BootnodesAreValidators),
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
