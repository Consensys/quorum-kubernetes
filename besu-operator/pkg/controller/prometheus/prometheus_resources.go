package prometheus

import (
	hyperledgerv1alpha1 "github.com/Sumaid/besu-kubernetes/besu-operator/pkg/apis/hyperledger/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"

	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

func (r *ReconcilePrometheus) prometheusConfigMap(instance *hyperledgerv1alpha1.Prometheus) *corev1.ConfigMap {
	confmap := &corev1.ConfigMap{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ConfigMap",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      instance.ObjectMeta.Name + "-configmap",
			Namespace: instance.Namespace,
			Labels:    r.getLabels(instance, instance.ObjectMeta.Name+"-configmap"),
		},
		Data: map[string]string{
			"prometheus.yml": `
          global:
            scrape_interval:     15s
            evaluation_interval: 15s
          alerting:
          rule_files:
          scrape_configs:
            - job_name: 'kubernetes-apiservers'
              kubernetes_sd_configs:
              - role: endpoints
              scheme: https
              tls_config:
                ca_file: /var/run/secrets/kubernetes.io/serviceaccount/ca.crt
              bearer_token_file: /var/run/secrets/kubernetes.io/serviceaccount/token
              relabel_configs:
              - source_labels: [__meta_kubernetes_namespace, __meta_kubernetes_service_name, __meta_kubernetes_endpoint_port_name]
                action: keep
                regex: default;kubernetes;https
      
            - job_name: 'kubernetes-nodes'
              scheme: https
              tls_config:
                ca_file: /var/run/secrets/kubernetes.io/serviceaccount/ca.crt
              bearer_token_file: /var/run/secrets/kubernetes.io/serviceaccount/token
              kubernetes_sd_configs:
              - role: node
              relabel_configs:
              - action: labelmap
                regex: __meta_kubernetes_node_label_(.+)
              - target_label: __address__
                replacement: kubernetes.default.svc:443
              - source_labels: [__meta_kubernetes_node_name]
                regex: (.+)
                target_label: __metrics_path__
                replacement: /api/v1/nodes/${1}/proxy/metrics
      
            - job_name: 'kubernetes-pods'
              kubernetes_sd_configs:
              - role: pod
              relabel_configs:
              - source_labels: [__meta_kubernetes_pod_annotation_prometheus_io_scrape]
                action: keep
                regex: true
              - source_labels: [__meta_kubernetes_pod_annotation_prometheus_io_path]
                action: replace
                target_label: __metrics_path__
                regex: (.+)
              - source_labels: [__address__, __meta_kubernetes_pod_annotation_prometheus_io_port]
                action: replace
                regex: ([^:]+)(?::\d+)?;(\d+)
                replacement: $1:$2
                target_label: __address__
              - action: labelmap
                regex: __meta_kubernetes_pod_label_(.+)
              - source_labels: [__meta_kubernetes_namespace]
                action: replace
                target_label: kubernetes_namespace
              - source_labels: [__meta_kubernetes_pod_name]
                action: replace
                target_label: kubernetes_pod_name`,
		},
	}
	controllerutil.SetControllerReference(instance, confmap, r.scheme)
	return confmap
}

func (r *ReconcilePrometheus) prometheusService(instance *hyperledgerv1alpha1.Prometheus) *corev1.Service {

	serv := &corev1.Service{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Service",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      instance.ObjectMeta.Name,
			Namespace: instance.Namespace,
			Labels:    r.getLabels(instance, instance.ObjectMeta.Name),
		},
		Spec: corev1.ServiceSpec{
			Type:     "NodePort",
			Selector: r.getLabels(instance, instance.ObjectMeta.Name),
			Ports: []corev1.ServicePort{
				{
					Name:       instance.ObjectMeta.Name,
					Protocol:   "TCP",
					Port:       int32(9090),
					TargetPort: intstr.FromInt(int(9090)),
					NodePort:   int32(instance.Spec.NodePort),
				},
			},
		},
	}
	controllerutil.SetControllerReference(instance, serv, r.scheme)
	return serv
}

func (r *ReconcilePrometheus) prometheusDeployment(instance *hyperledgerv1alpha1.Prometheus) *appsv1.Deployment {
	depl := &appsv1.Deployment{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Deployment",
			APIVersion: "apps/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      instance.ObjectMeta.Name,
			Namespace: instance.Namespace,
			Labels:    r.getLabels(instance, instance.ObjectMeta.Name),
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &instance.Spec.Replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: r.getLabels(instance, instance.ObjectMeta.Name),
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels:      r.getLabels(instance, instance.ObjectMeta.Name),
					Annotations: r.getPrometheusAnnotations(instance),
				},
				Spec: corev1.PodSpec{
					ServiceAccountName: instance.ObjectMeta.Name + "-sa",
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
							Env: []corev1.EnvVar{
								{
									Name: "POD_IP",
									ValueFrom: &corev1.EnvVarSource{
										FieldRef: &corev1.ObjectFieldSelector{
											FieldPath: "status.podIP",
										},
									},
								},
								{
									Name: "POD_NAME",
									ValueFrom: &corev1.EnvVarSource{
										FieldRef: &corev1.ObjectFieldSelector{
											FieldPath: "metadata.name",
										},
									},
								},
							},
							VolumeMounts: []corev1.VolumeMount{
								corev1.VolumeMount{
									Name:      instance.ObjectMeta.Name + "-config",
									MountPath: "/etc/prometheus/",
									ReadOnly:  true,
								},
							},
							Ports: []corev1.ContainerPort{
								{
									Name:          "http",
									ContainerPort: int32(9090),
									Protocol:      "TCP",
								},
							},
							Command: []string{
								"/bin/prometheus",
							},
							Args: []string{
								"--config.file=/etc/prometheus/prometheus.yml",
							},
						},
					},
					Volumes: []corev1.Volume{
						corev1.Volume{
							Name: instance.ObjectMeta.Name + "-config",
							VolumeSource: corev1.VolumeSource{
								ConfigMap: &corev1.ConfigMapVolumeSource{
									LocalObjectReference: corev1.LocalObjectReference{
										Name: instance.ObjectMeta.Name + "-configmap",
									},
									Items: []corev1.KeyToPath{
										{
											Key:  "prometheus.yml",
											Path: "prometheus.yml",
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}
	controllerutil.SetControllerReference(instance, depl, r.scheme)
	return depl
}

func (r *ReconcilePrometheus) prometheusRole(instance *hyperledgerv1alpha1.Prometheus) *rbacv1.Role {

	role := &rbacv1.Role{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Role",
			APIVersion: "rbac.authorization.k8s.io/v1beta1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      instance.ObjectMeta.Name + "-role",
			Namespace: instance.ObjectMeta.Namespace,
		},
		Rules: []rbacv1.PolicyRule{
			{
				APIGroups: []string{
					"",
				},
				Resources: []string{
					"nodes",
					"nodes/proxy",
					"services",
					"endpoints",
					"pods",
				},
				Verbs: []string{
					"get",
					"list",
					"watch",
				},
			},
			{
				APIGroups: []string{
					"extensions",
				},
				Resources: []string{
					"ingresses",
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

func (r *ReconcilePrometheus) prometheusRoleBinding(instance *hyperledgerv1alpha1.Prometheus) *rbacv1.RoleBinding {

	rb := &rbacv1.RoleBinding{
		TypeMeta: metav1.TypeMeta{
			Kind:       "RoleBinding",
			APIVersion: "rbac.authorization.k8s.io/v1beta1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      instance.ObjectMeta.Name + "-rb",
			Namespace: instance.ObjectMeta.Namespace,
		},
		RoleRef: rbacv1.RoleRef{
			APIGroup: "rbac.authorization.k8s.io",
			Kind:     "Role",
			Name:     instance.ObjectMeta.Name + "-role",
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

func (r *ReconcilePrometheus) getLabels(instance *hyperledgerv1alpha1.Prometheus, name string) map[string]string {
	labels := make(map[string]string)
	labels["app"] = name
	return labels
}

func (r *ReconcilePrometheus) getPrometheusAnnotations(instance *hyperledgerv1alpha1.Prometheus) map[string]string {
	annotations := make(map[string]string)
	annotations["prometheus.io/scrape"] = "true"
	annotations["prometheus.io/port"] = "9090"
	return annotations
}
