package besunode

import (
	// "io/ioutil"
	// "encoding/json"

	"fmt"
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
	keyFile := "--node-private-key-file=/secrets/private.key "
	rpcOptions := "--rpc-http-enabled --rpc-http-host=0.0.0.0 --rpc-http-port=8545 --rpc-http-cors-origins=${NODES_HTTP_CORS_ORIGINS} --rpc-http-api=ETH,NET,IBFT "
	graphqlOptions := "--graphql-http-enabled --graphql-http-host=0.0.0.0 --graphql-http-port=8547 --graphql-http-cors-origins=${NODES_HTTP_CORS_ORIGINS} "
	rpcWsOptions := "--rpc-ws-enabled --rpc-ws-host=0.0.0.0 --rpc-ws-port=8546 "
	metricsOptions := "--metrics-enabled=true --metrics-host=0.0.0.0 --metrics-port=9545 "
	hostWhitelist := "--host-whitelist=${NODES_HOST_WHITELIST} "

	bootEnodes := ""
	curlCommandTemplate := "curl -X GET --connect-timeout 30 --max-time 10 --retry 6 --retry-delay 0 --retry-max-time 300 ${%s}:8545/liveness "
	curlCommand := ""
	for i := 1; i < instance.Spec.Bootnodes+1; i++ {
		bootEnodes += "enode://${BOOTNODE" + strconv.Itoa(i) + "_PUBKEY}@"
		bootEnodes += "${BESU_BOOTNODE" + strconv.Itoa(i) + "_SERVICE_HOST}:"
		bootEnodes += "${BESU_BOOTNODE" + strconv.Itoa(i) + "_SERVICE_PORT}"
		curlCommand += fmt.Sprintf(curlCommandTemplate, "BESU_BOOTNODE"+strconv.Itoa(i)+"_SERVICE_HOST")
		if i < instance.Spec.Bootnodes {
			bootEnodes += ","
			curlCommand += "&& "
		}
	}
	reqLogger.Info("curlCommand : ")
	reqLogger.Info(curlCommand)

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
					curlCommand,
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
							Key:  "genesisnode",
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
		reqLogger.Info("Bootnodes key environment binder : ")
		envVars = append(envVars, corev1.EnvVar{
			Name: "BOOTNODE" + strconv.Itoa(i) + "_PUBKEY",
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: &corev1.SecretKeySelector{
					Key: "enode.key",
					LocalObjectReference: corev1.LocalObjectReference{
						Name: "besu-bootnode" + strconv.Itoa(i) + "-key",
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

	replicas := instance.Spec.Replicas
	if replicas == 0 {
		replicas = 1
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
			Replicas:            &replicas,
			PodManagementPolicy: "OrderedReady",
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
			"private.key": instance.Spec.PrivKey,
			"public.key":  instance.Spec.PubKey,
			"enode.key":   instance.Spec.PubKey,
		},
	}
	controllerutil.SetControllerReference(instance, secr, r.scheme)
	return secr
}
