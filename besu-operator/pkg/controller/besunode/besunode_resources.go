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
	besunodeImage                = "hyperledger/besu:1.4.6"
	ImagePullPolicy              = "IfNotPresent"
	VolumesReadOnly              = true
	MountPath                    = "/configs"
	SecretsVolumeName            = "key"
	SecretsVolumeMountPath       = "/secrets"
	GenesisConfigVolumeName      = "genesis-config"
	GenesisConfigVolumeMountPath = "/etc/genesis"
	PodManagementPolicy          = "OrderedReady"
)

func (r *ReconcileBesuNode) besunodeStatefulSet(instance *hyperledgerv1alpha1.BesuNode) *appsv1.StatefulSet {
	reqLogger := log.WithValues("Resource : ", "Statefulset")
	reqLogger.Info("Ensuring BesuNode Statefulset")

	initContainers := []corev1.Container{}
	if instance.Spec.Type != "Bootnode" {
		initContainers = r.getInitContainer(instance)
	}

	volumes := r.getVolumes(instance)
	volumeMounts := r.getVolumeMounts(instance)
	envVars := r.getEnvVars(instance)
	readinessProbe := r.getReadinessProbe(instance)

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
			Labels:    r.getLabels(instance),
		},
		Spec: appsv1.StatefulSetSpec{
			Replicas:            &instance.Spec.Replicas,
			PodManagementPolicy: PodManagementPolicy,
			Selector: &metav1.LabelSelector{
				MatchLabels: r.getLabels(instance),
			},
			ServiceName: "besu-" + instance.ObjectMeta.Name,
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels:      r.getLabels(instance),
					Annotations: r.getPrometheusAnnotations(instance),
				},
				Spec: corev1.PodSpec{
					ServiceAccountName: instance.ObjectMeta.Name + "-sa",
					InitContainers:     initContainers,
					Containers: []corev1.Container{
						corev1.Container{
							Name:            instance.ObjectMeta.Name,
							Image:           instance.Spec.Image.Repository + ":" + instance.Spec.Image.Tag,
							ImagePullPolicy: corev1.PullPolicy(instance.Spec.Image.PullPolicy),
							Resources: corev1.ResourceRequirements{
								Requests: corev1.ResourceList{
									corev1.ResourceCPU:    resource.MustParse(instance.Spec.Resources.CPURequest),
									corev1.ResourceMemory: resource.MustParse(instance.Spec.Resources.MemRequest),
								},
								Limits: corev1.ResourceList{
									corev1.ResourceCPU:    resource.MustParse(instance.Spec.Resources.CPULimit),
									corev1.ResourceMemory: resource.MustParse(instance.Spec.Resources.MemLimit),
								},
							},
							Env:          envVars,
							VolumeMounts: volumeMounts,
							Ports: []corev1.ContainerPort{
								{
									Name:          "json-rpc",
									ContainerPort: int32(instance.Spec.RPC.Port),
									Protocol:      "TCP",
								},
								{
									Name:          "ws",
									ContainerPort: int32(instance.Spec.WS.Port),
									Protocol:      "TCP",
								},
								{
									Name:          "graphql",
									ContainerPort: int32(instance.Spec.GraphQl.Port),
									Protocol:      "TCP",
								},
								{
									Name:          "rlpx",
									ContainerPort: int32(instance.Spec.P2P.Port),
									Protocol:      "TCP",
								},
								{
									Name:          "discovery",
									ContainerPort: int32(instance.Spec.P2P.Port),
									Protocol:      "UDP",
								},
								{
									Name:          "metrics",
									ContainerPort: int32(instance.Spec.Metrics.Port),
									Protocol:      "TCP",
								},
							},
							Command: []string{
								"/bin/sh",
								"-c",
							},
							Args: []string{
								r.getBesuCommand(instance),
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

	serv := &corev1.Service{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Service",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "besu-" + instance.ObjectMeta.Name,
			Namespace: instance.Namespace,
			Labels:    r.getLabels(instance),
		},
		Spec: corev1.ServiceSpec{
			Type:     "ClusterIP",
			Selector: r.getLabels(instance),
			Ports: []corev1.ServicePort{
				{
					Name:       "discovery",
					Protocol:   "UDP",
					Port:       int32(instance.Spec.P2P.Port),
					TargetPort: intstr.FromInt(int(instance.Spec.P2P.Port)),
				},
				{
					Name:       "rlpx",
					Protocol:   "TCP",
					Port:       int32(instance.Spec.P2P.Port),
					TargetPort: intstr.FromInt(int(instance.Spec.P2P.Port)),
				},
				{
					Name:       "json-rpc",
					Protocol:   "TCP",
					Port:       int32(instance.Spec.RPC.Port),
					TargetPort: intstr.FromInt(int(instance.Spec.RPC.Port)),
				},
				{
					Name:       "ws",
					Protocol:   "TCP",
					Port:       int32(instance.Spec.WS.Port),
					TargetPort: intstr.FromInt(int(instance.Spec.WS.Port)),
				},
				{
					Name:       "graphql",
					Protocol:   "TCP",
					Port:       int32(instance.Spec.GraphQl.Port),
					TargetPort: intstr.FromInt(int(instance.Spec.GraphQl.Port)),
				},
			},
		},
	}
	controllerutil.SetControllerReference(instance, serv, r.scheme)
	return serv
}

func (r *ReconcileBesuNode) getBesuCommand(instance *hyperledgerv1alpha1.BesuNode) string {
	commandInitial := "exec /opt/besu/bin/besu "
	genesisFile := fmt.Sprintf("--genesis-file=%s ", GenesisConfigVolumeMountPath+"/genesis.json")
	keyFile := fmt.Sprintf("--node-private-key-file=%s ", "/secrets/private.key")
	rpcOptions := fmt.Sprintf("--rpc-http-enabled=%t --rpc-http-host=%s --rpc-http-port=%d --rpc-http-cors-origins=${NODES_HTTP_CORS_ORIGINS} --rpc-http-api=ETH,NET,IBFT ",
		instance.Spec.RPC.Enabled, instance.Spec.RPC.Host, instance.Spec.RPC.Port)
	graphqlOptions := fmt.Sprintf("--graphql-http-enabled=%t --graphql-http-host=%s --graphql-http-port=%d --graphql-http-cors-origins=${NODES_HTTP_CORS_ORIGINS} ",
		instance.Spec.GraphQl.Enabled, instance.Spec.GraphQl.Host, instance.Spec.GraphQl.Port)
	rpcWsOptions := fmt.Sprintf("--rpc-ws-enabled=%t --rpc-ws-host=%s --rpc-ws-port=%d ",
		instance.Spec.WS.Enabled, instance.Spec.WS.Host, instance.Spec.WS.Port)
	metricsOptions := fmt.Sprintf("--metrics-enabled=%t --metrics-host=%s --metrics-port=%d ",
		instance.Spec.Metrics.Enabled, instance.Spec.Metrics.Host, instance.Spec.Metrics.Port)
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
	enodeOption := "--bootnodes=" + bootEnodes

	keyString := ""
	if instance.Spec.Type != "Member" {
		keyString = keyFile
	}
	command := commandInitial + genesisFile + keyString + rpcOptions + graphqlOptions + rpcWsOptions + metricsOptions + hostWhitelist + enodeOption
	return command
}

func (r *ReconcileBesuNode) getCurlCommand(instance *hyperledgerv1alpha1.BesuNode) string {
	curlCommandTemplate := "curl -X GET --connect-timeout 30 --max-time 10 --retry 6 --retry-delay 0 --retry-max-time 300 ${%s}:8545/liveness "
	curlCommand := ""
	for i := 1; i < instance.Spec.Bootnodes+1; i++ {
		curlCommand += fmt.Sprintf(curlCommandTemplate, "BESU_BOOTNODE"+strconv.Itoa(i)+"_SERVICE_HOST")
		if i < instance.Spec.Bootnodes {
			curlCommand += "&& "
		}
	}
	return curlCommand
}

func (r *ReconcileBesuNode) getPrometheusAnnotations(instance *hyperledgerv1alpha1.BesuNode) map[string]string {
	annotations := make(map[string]string)
	annotations["prometheus.io/scrape"] = "true"
	annotations["prometheus.io/port"] = "9545"
	annotations["prometheus.io/path"] = "/metrics"
	return annotations
}

func (r *ReconcileBesuNode) getEnvVars(instance *hyperledgerv1alpha1.BesuNode) []corev1.EnvVar {
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
				SecretKeyRef: &corev1.SecretKeySelector{
					Key: "enode.key",
					LocalObjectReference: corev1.LocalObjectReference{
						Name: "besu-bootnode" + strconv.Itoa(i) + "-key",
					},
				},
			},
		})
	}

	return envVars
}

func (r *ReconcileBesuNode) getInitContainer(instance *hyperledgerv1alpha1.BesuNode) []corev1.Container {
	initContainers := []corev1.Container{
		corev1.Container{
			Name:  "init-bootnode",
			Image: "pegasyseng/k8s-helper:v1.18.4",
			Command: []string{
				"sh",
				"-c",
				r.getCurlCommand(instance),
			},
		},
	}
	return initContainers
}

func (r *ReconcileBesuNode) getVolumes(instance *hyperledgerv1alpha1.BesuNode) []corev1.Volume {
	volumes := []corev1.Volume{
		corev1.Volume{
			Name: GenesisConfigVolumeName,
			VolumeSource: corev1.VolumeSource{
				ConfigMap: &corev1.ConfigMapVolumeSource{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: "besu-genesis",
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
	return volumes
}

func (r *ReconcileBesuNode) getVolumeMounts(instance *hyperledgerv1alpha1.BesuNode) []corev1.VolumeMount {
	volumeMounts := []corev1.VolumeMount{
		corev1.VolumeMount{
			Name:      GenesisConfigVolumeName,
			MountPath: GenesisConfigVolumeMountPath,
			ReadOnly:  VolumesReadOnly,
		},
	}
	return volumeMounts
}

func (r *ReconcileBesuNode) getReadinessProbe(instance *hyperledgerv1alpha1.BesuNode) *corev1.Probe {
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
	return readinessProbe
}

func (r *ReconcileBesuNode) getLabels(instance *hyperledgerv1alpha1.BesuNode) map[string]string {
	labels := make(map[string]string)
	labels["app"] = instance.ObjectMeta.Name
	return labels
}
