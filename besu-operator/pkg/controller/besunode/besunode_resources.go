package besunode

import (
	// "io/ioutil"
	// "encoding/json"

	"strconv"

	hyperledgerv1alpha1 "github.com/Sumaid/besu-kubernetes/besu-operator/pkg/apis/hyperledger/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"

	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

// Constants common across any instance
const (
	besunodeImage                = "hyperledger/besu:latest"
	ImagePullPolicy              = "IfNotPresent"
	VolumesReadOnly              = true
	MountPath                    = "/configs"
	SecretsVolumeName            = "key"
	SecretsVolumeMountPath       = "/secrets"
	GenesisConfigVolumeName      = "genesis-config"
	GenesisConfigVolumeMountPath = "/configs"
)

func (r *ReconcileBesuNode) besunodeStatefulSet(instance *hyperledgerv1alpha1.BesuNode) *appsv1.StatefulSet {
	reqLogger := log.WithValues("Resource : ", "Statefulset")
	reqLogger.Info("Ensuring BesuNode Statefulset")

	annotations := make(map[string]string)
	annotations["prometheus.io/scrape"] = "true"
	annotations["prometheus.io/port"] = "9545"
	annotations["prometheus.io/path"] = "/metrics"

	labels := make(map[string]string)
	labels["app"] = instance.ObjectMeta.Name

	commandInitial := "exec /opt/besu/bin/besu --genesis-file=/configs/genesis.json "
	keyFile := "--node-private-key-file=/secrets/key "
	rpcOptions := "--rpc-http-enabled --rpc-http-host=0.0.0.0 --rpc-http-port=8545 --rpc-http-cors-origins=${NODES_HTTP_CORS_ORIGINS} --rpc-http-api=ETH,NET,IBFT "
	graphqlOptions := "--graphql-http-enabled --graphql-http-host=0.0.0.0 --graphql-http-port=8547 --graphql-http-cors-origins=${NODES_HTTP_CORS_ORIGINS} "
	rpcWsOptions := "--rpc-ws-enabled --rpc-ws-host=0.0.0.0 --rpc-ws-port=8546 "
	metricsOptions := "--metrics-enabled=true --metrics-host=0.0.0.0 --metrics-port=9545 "
	hostWhitelist := "--host-whitelist=${NODES_HOST_WHITELIST} "

	bootEnodes := ""
	for i := 1; i < instance.Spec.Bootnodes+1; i++ {
		bootEnodes += "enode://${BOOTNODE" + strconv.Itoa(i) + "_PUBKEY}@"
		bootEnodes += "${BESU_BOOTNODE" + strconv.Itoa(i) + "_SERVICE_HOST}:"
		bootEnodes += "${BESU_BOOTNODE" + strconv.Itoa(i) + "_SERVICE_PORT}"
		if i < instance.Spec.Bootnodes {
			bootEnodes += ","
		}
	}
	reqLogger.Info("Enodes string")
	reqLogger.Info(bootEnodes)

	enodeOption := "--bootnodes=" + bootEnodes

	keyString := ""
	if instance.Spec.Type != "Member" {
		keyString = keyFile
	}
	command := commandInitial + keyString + rpcOptions + graphqlOptions + rpcWsOptions + metricsOptions + hostWhitelist + enodeOption

	initContainers := []corev1.Container{}
	if instance.Spec.Type != "Bootnode" {
		initContainers = []corev1.Container{
			corev1.Container{
				Name:  "init-bootnode",
				Image: "byrnedo/alpine-curl",
				Command: []string{
					"sh",
					"-c",
					"curl -X GET --connect-timeout 30 --max-time 10 --retry 6 --retry-delay 0 --retry-max-time 300 ${BESU_BOOTNODE1_SERVICE_HOST}:8545/liveness",
				},
			},
		}
	}

	volumes := []corev1.Volume{
		corev1.Volume{
			Name: "genesis-config",
			VolumeSource: corev1.VolumeSource{
				ConfigMap: &corev1.ConfigMapVolumeSource{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: "besu-configmap",
					},
					Items: []corev1.KeyToPath{
						{
							Key:  "genesis.json",
							Path: "genesis.json",
						},
					},
				},
			},
		},
	}

	volumeMounts := []corev1.VolumeMount{
		corev1.VolumeMount{
			Name:      GenesisConfigVolumeName,
			MountPath: GenesisConfigVolumeMountPath,
			ReadOnly:  VolumesReadOnly,
		},
	}

	envVars := []corev1.EnvVar{
		{
			Name: "POD_IP",
			ValueFrom: &corev1.EnvVarSource{
				FieldRef: &corev1.ObjectFieldSelector{
					FieldPath: "status.podIP",
				},
			},
		},
		{
			Name:  "NODES_HTTP_CORS_ORIGINS",
			Value: "*",
		},
		{
			Name:  "NODES_HOST_WHITELIST",
			Value: "*",
		},
	}

	for i := 1; i < instance.Spec.Bootnodes+1; i++ {
		envVars = append(envVars, corev1.EnvVar{
			Name: "BOOTNODE" + strconv.Itoa(i) + "_PUBKEY",
			ValueFrom: &corev1.EnvVarSource{
				ConfigMapKeyRef: &corev1.ConfigMapKeySelector{
					Key: "bootnode" + strconv.Itoa(i) + "pubkey",
					LocalObjectReference: corev1.LocalObjectReference{
						Name: "besu-configmap",
					},
				},
			},
		})
	}

	readinessProbe := &corev1.Probe{
		Handler: corev1.Handler{
			HTTPGet: &corev1.HTTPGetAction{
				Path: "/readiness",
				Port: intstr.FromInt(8545),
			},
		},
		InitialDelaySeconds: int32(50),
		PeriodSeconds:       int32(30),
	}

	if instance.Spec.Type != "Member" {
		envVars = append(envVars, corev1.EnvVar{
			Name: "POD_NAME",
			ValueFrom: &corev1.EnvVarSource{
				FieldRef: &corev1.ObjectFieldSelector{
					FieldPath: "metadata.name",
				},
			},
		})

		readinessProbe = nil

		volumes = append(volumes, corev1.Volume{
			Name: "key",
			VolumeSource: corev1.VolumeSource{
				Secret: &corev1.SecretVolumeSource{
					SecretName: "besu-" + instance.ObjectMeta.Name + "-key",
				},
			},
		})

		volumeMounts = append(volumeMounts, corev1.VolumeMount{
			Name:      SecretsVolumeName,
			MountPath: SecretsVolumeMountPath,
			ReadOnly:  VolumesReadOnly,
		})
	}

	sfs := &appsv1.StatefulSet{
		TypeMeta: metav1.TypeMeta{
			Kind:       "StatefulSet",
			APIVersion: "apps/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      instance.ObjectMeta.Name,
			Namespace: instance.Namespace,
			Labels:    labels,
		},
		Spec: appsv1.StatefulSetSpec{
			Replicas: &instance.Spec.Replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: labels,
			},
			ServiceName: "besu-" + instance.ObjectMeta.Name,
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels:      labels,
					Annotations: annotations,
				},
				Spec: corev1.PodSpec{
					ServiceAccountName: instance.ObjectMeta.Name + "-sa",
					InitContainers:     initContainers,
					Containers: []corev1.Container{
						corev1.Container{
							Name:            instance.ObjectMeta.Name,
							Image:           besunodeImage,
							ImagePullPolicy: ImagePullPolicy,
							Resources: corev1.ResourceRequirements{
								Requests: corev1.ResourceList{
									corev1.ResourceCPU:    resource.MustParse("100m"),
									corev1.ResourceMemory: resource.MustParse("1024Mi"),
								},
								Limits: corev1.ResourceList{
									corev1.ResourceCPU:    resource.MustParse("500m"),
									corev1.ResourceMemory: resource.MustParse("2048Mi"),
								},
							},
							Env:          envVars,
							VolumeMounts: volumeMounts,
							Ports: []corev1.ContainerPort{
								{
									Name:          "json-rpc",
									ContainerPort: int32(8545),
									Protocol:      "TCP",
								},
								{
									Name:          "ws",
									ContainerPort: int32(8546),
									Protocol:      "TCP",
								},
								{
									Name:          "graphql",
									ContainerPort: int32(8547),
									Protocol:      "TCP",
								},
								{
									Name:          "rlpx",
									ContainerPort: int32(30303),
									Protocol:      "TCP",
								},
								{
									Name:          "discovery",
									ContainerPort: int32(30303),
									Protocol:      "UDP",
								},
							},
							Command: []string{
								"/bin/sh",
								"-c",
							},
							Args: []string{
								command,
							},
							LivenessProbe: &corev1.Probe{
								Handler: corev1.Handler{
									HTTPGet: &corev1.HTTPGetAction{
										Path: "/liveness",
										Port: intstr.FromInt(8545),
									},
								},
								InitialDelaySeconds: int32(60),
								PeriodSeconds:       int32(30),
							},
							ReadinessProbe: readinessProbe,
						},
					},
					Volumes: volumes,
				},
			},
		},
	}
	controllerutil.SetControllerReference(instance, sfs, r.scheme)
	return sfs
}

func (r *ReconcileBesuNode) besunodeRole(instance *hyperledgerv1alpha1.BesuNode) *rbacv1.Role {

	role := &rbacv1.Role{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Role",
			APIVersion: "rbac.authorization.k8s.io/v1beta1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      instance.ObjectMeta.Name + "-key-read-role",
			Namespace: instance.Namespace,
		},
		Rules: []rbacv1.PolicyRule{
			{
				APIGroups: []string{
					"",
				},
				Resources: []string{
					"secrets",
				},
				ResourceNames: []string{
					"besu-" + instance.ObjectMeta.Name + "-key",
				},
				Verbs: []string{
					"get",
				},
			},
		},
	}
	controllerutil.SetControllerReference(instance, role, r.scheme)
	return role
}

func (r *ReconcileBesuNode) besunodeRoleBinding(instance *hyperledgerv1alpha1.BesuNode) *rbacv1.RoleBinding {

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
			Name:     instance.ObjectMeta.Name + "-key-read-role",
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

func (r *ReconcileBesuNode) besunodeService(instance *hyperledgerv1alpha1.BesuNode) *corev1.Service {

	labels := make(map[string]string)
	labels["app"] = instance.ObjectMeta.Name

	serv := &corev1.Service{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Service",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "besu-" + instance.ObjectMeta.Name,
			Namespace: instance.Namespace,
			Labels:    labels,
		},
		Spec: corev1.ServiceSpec{
			Type:     "ClusterIP",
			Selector: labels,
			Ports: []corev1.ServicePort{
				{
					Name:       "discovery",
					Protocol:   "UDP",
					Port:       int32(30303),
					TargetPort: intstr.FromInt(int(30303)),
				},
				{
					Name:       "rlpx",
					Protocol:   "TCP",
					Port:       int32(30303),
					TargetPort: intstr.FromInt(int(30303)),
				},
				{
					Name:       "json-rpc",
					Protocol:   "TCP",
					Port:       int32(8545),
					TargetPort: intstr.FromInt(int(8545)),
				},
				{
					Name:       "ws",
					Protocol:   "TCP",
					Port:       int32(8546),
					TargetPort: intstr.FromInt(int(8546)),
				},
				{
					Name:       "graphql",
					Protocol:   "TCP",
					Port:       int32(8547),
					TargetPort: intstr.FromInt(int(8547)),
				},
			},
		},
	}
	controllerutil.SetControllerReference(instance, serv, r.scheme)
	return serv
}

func (r *ReconcileBesuNode) besunodeSecret(instance *hyperledgerv1alpha1.BesuNode) *corev1.Secret {
	secr := &corev1.Secret{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Secret",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "besu-" + instance.ObjectMeta.Name + "-key",
			Namespace: instance.Namespace,
			Labels: map[string]string{
				"app": "besu-" + instance.ObjectMeta.Name + "-key",
			},
		},
		Type: "Opaque",
		StringData: map[string]string{
			"key": instance.Spec.PrivKey,
		},
	}
	controllerutil.SetControllerReference(instance, secr, r.scheme)
	return secr
}

// func (r *ReconcileBesuNode) besunodeConfigMap(instance *hyperledgerv1alpha1.BesuNode) *corev1.ConfigMap {
// 	conf := &corev1.ConfigMap{
// 		TypeMeta: metav1.TypeMeta{
// 			Kind:       "ConfigMap",
// 			APIVersion: "v1",
// 		},
// 		ObjectMeta: metav1.ObjectMeta{
// 			Name:      "besu-" + "configmap",
// 			Namespace: instance.Namespace,
// 			Labels: map[string]string{
// 				"app": "besu-" + "configmap",
// 			},
// 		},
// 		Data: map[string]string{
// 			"bootnode1pubkey":      "5d812c3c25ff398ab416968fce9009c2be7ed70a87abc8ea30bd667ce17a9287a6341fbf6ce757bb8148436c39c71296639ea81afcc94cdf908b6e1344f26188",
// 			"bootnode2pubkey":      "b2fba529681ea7f4619556753d40c8689b936fb1c621bc91f94d2938eb58c285d4911457ae4887b9c3bd593b2d608d319c6dc384d6acae2d043a4657029178d3",
// 			"bootnode3pubkey":      "00b20ab6a385a2403d64637b3d93cb6d83215a08f29adb6feb4b8bf03387b734444e8b060f53150dea4b9b897823540d19918c13d6f57a5153d190b5fad7bf51",
// 			"bootnode4pubkey":      "5fc1f8dc9f0c03087128e4bd724530e883d7de1a431269876dff9c95b8952f73c7e85ac7b49d85a2ad4950e967319482af435e07a0eab0a98d98449437787a00",
// 			"nodesHttpCorsOrigins": "*",
// 			"nodesHostWhitelist":   "*",
// 			"genesis.json": `
// 				{
// 				"config": {
// 				  "chainId": 2018,
// 				  "constantinoplefixblock": 0,
// 				  "ibft2": {
// 					"blockperiodseconds": 2,
// 					"epochlength": 30000,
// 					"requesttimeoutseconds": 10
// 				  }
// 				},
// 				"nonce": "0x0",
// 				"timestamp": "0x58ee40ba",
// 				"gasLimit": "0x47b760",
// 				"difficulty": "0x1",
// 				"mixHash": "0x63746963616c2062797a616e74696e65206661756c7420746f6c6572616e6365",
// 				"coinbase": "0x0000000000000000000000000000000000000000",
// 				"alloc": {
// 				  "fe3b557e8fb62b89f4916b721be55ceb828dbd73": {
// 					"privateKey": "8f2a55949038a9610f50fb23b5883af3b4ecb3c3bb792cbcefbd1542c692be63",
// 					"comment": "private key and this comment are ignored.  In a real chain, the private key should NOT be stored",
// 					"balance": "0xad78ebc5ac6200000"
// 				  },
// 				  "627306090abaB3A6e1400e9345bC60c78a8BEf57": {
// 					"privateKey": "c87509a1c067bbde78beb793e6fa76530b6382a4c0241e5e4a9ec0a0f44dc0d3",
// 					"comment": "private key and this comment are ignored.  In a real chain, the private key should NOT be stored",
// 					"balance": "90000000000000000000000"
// 				  },
// 				  "f17f52151EbEF6C7334FAD080c5704D77216b732": {
// 					"privateKey": "ae6ae8e5ccbfb04590405997ee2d52d2b330726137b875053c36d94e974d162f",
// 					"comment": "private key and this comment are ignored.  In a real chain, the private key should NOT be stored",
// 					"balance": "90000000000000000000000"
// 				  }
// 				},
// 				"extraData": "0xf87ea00000000000000000000000000000000000000000000000000000000000000000f85494ca6e9704586eb1fb38194308e2192e43b1e1979c94ce2276efc33fee3c321e634eac28a9476e53b71c94f466a7174230056004d11178d2647c12740fa58b94b83820d6cf4b7e5aa67a2b57969caa5cdf6dff49808400000000c0"
// 			  }`,
// 		},
// 	}
// 	controllerutil.SetControllerReference(instance, conf, r.scheme)
// 	return conf
// }
