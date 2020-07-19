package besu

import (
	"context"

	hyperledgerv1alpha1 "github.com/Sumaid/besu-kubernetes/besu-operator/pkg/apis/hyperledger/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

func (r *ReconcileBesu) ensureBesuNode(request reconcile.Request,
	instance *hyperledgerv1alpha1.Besu,
	sfs *hyperledgerv1alpha1.BesuNode,
) (*reconcile.Result, error, int, int) {

	// See if BesuNode already exists and create if it doesn't
	found := &hyperledgerv1alpha1.BesuNode{}
	err := r.client.Get(context.TODO(), types.NamespacedName{
		Name:      sfs.ObjectMeta.Name,
		Namespace: sfs.ObjectMeta.Namespace,
	}, found)
	if err != nil && errors.IsNotFound(err) {

		// Create the BesuNode
		log.Info("Creating a new BesuNode", "BesuNode.Namespace", sfs.Namespace, "BesuNode.Name", sfs.Name)
		err = r.client.Create(context.TODO(), sfs)

		if err != nil {
			// BesuNode failed
			log.Error(err, "Failed to create new BesuNode", "BesuNode.Namespace", sfs.Namespace, "BesuNode.Name", sfs.Name)
			return &reconcile.Result{}, err, int(found.Status.Replicas), int(found.Status.ReadyReplicas)
		} else {
			// BesuNode was successful
			return nil, nil, int(found.Status.Replicas), int(found.Status.ReadyReplicas)
		}
	} else if err != nil {
		// Error that isn't due to the BesuNode not existing
		log.Error(err, "Failed to get BesuNode", "BesuNode.Namespace", sfs.Namespace, "BesuNode.Name", sfs.Name)
		return &reconcile.Result{}, err, int(found.Status.Replicas), int(found.Status.ReadyReplicas)
	}

	var result *reconcile.Result
	result, err = r.handleBesuNodeChanges(instance, found)
	if result != nil || err != nil {
		return result, err, int(found.Status.Replicas), int(found.Status.ReadyReplicas)
	}
	log.Info("ensureBesuNode", "All went  :", "well", "BesuNode.Namespace", sfs.Namespace, "BesuNode.Name", sfs.Name)
	return nil, nil, int(found.Status.Replicas), int(found.Status.ReadyReplicas)
}

func (r *ReconcileBesu) handleBesuNodeChanges(instance *hyperledgerv1alpha1.Besu, found *hyperledgerv1alpha1.BesuNode) (*reconcile.Result, error) {
	updated := false
	if instance.Spec.BesuNodeSpec.Image.Tag != found.Spec.Image.Tag || instance.Spec.BesuNodeSpec.Image.Repository != found.Spec.Image.Repository {
		found.Spec.Image = instance.Spec.BesuNodeSpec.Image
		updated = true
	}

	// replicas handling
	if found.Spec.Type == "Member" {
		if instance.Spec.Members != found.Spec.Replicas {
			found.Spec.Replicas = instance.Spec.Members
			updated = true
		}

	} else {
		if instance.Spec.BesuNodeSpec.Replicas != found.Spec.Replicas {
			found.Spec.Replicas = instance.Spec.BesuNodeSpec.Replicas
			updated = true
		}
	}

	if updated {
		err := r.client.Update(context.TODO(), found)
		if err != nil {
			log.Error(err, "Failed to update BesuNode.", "BesuNode.Namespace", found.Namespace, "BesuNode.Name", found.Name)
			return &reconcile.Result{}, err
		}
		// Spec updated - return and requeue
		return &reconcile.Result{Requeue: true}, nil
	}
	return nil, nil
}

func (r *ReconcileBesu) ensureConfigMap(request reconcile.Request,
	instance *hyperledgerv1alpha1.Besu,
	s *corev1.ConfigMap,
) (*reconcile.Result, error) {
	found := &corev1.ConfigMap{}
	err := r.client.Get(context.TODO(), types.NamespacedName{
		Name:      s.ObjectMeta.Name,
		Namespace: s.ObjectMeta.Namespace,
	}, found)
	if err != nil && errors.IsNotFound(err) {
		// Create the ConfigMap
		log.Info("Creating a new ConfigMap", "ConfigMap.Namespace", s.Namespace, "ConfigMap.Name", s.Name)
		err = r.client.Create(context.TODO(), s)

		if err != nil {
			// Creation failed
			log.Error(err, "Failed to create new ConfigMap", "ConfigMap.Namespace", s.Namespace, "ConfigMap.Name", s.Name)
			return &reconcile.Result{}, err
		} else {
			// Creation was successful
			return nil, nil
		}
	} else if err != nil {
		// Error that isn't due to the ConfigMap not existing
		log.Error(err, "Failed to get ConfigMap")
		return &reconcile.Result{}, err
	}

	return nil, nil
}

func (r *ReconcileBesu) ensurePrometheus(request reconcile.Request,
	instance *hyperledgerv1alpha1.Besu,
	sfs *hyperledgerv1alpha1.Prometheus,
) (*reconcile.Result, error) {

	// See if Prometheus already exists and create if it doesn't
	found := &hyperledgerv1alpha1.Prometheus{}
	err := r.client.Get(context.TODO(), types.NamespacedName{
		Name:      sfs.ObjectMeta.Name,
		Namespace: sfs.ObjectMeta.Namespace,
	}, found)
	if err != nil && errors.IsNotFound(err) {

		// Create the Prometheus
		log.Info("Creating a new Prometheus", "Prometheus.Namespace", sfs.Namespace, "Prometheus.Name", sfs.Name)
		err = r.client.Create(context.TODO(), sfs)

		if err != nil {
			// Prometheus failed
			log.Error(err, "Failed to create new Prometheus", "Prometheus.Namespace", sfs.Namespace, "Prometheus.Name", sfs.Name)
			return &reconcile.Result{}, err
		} else {
			// Prometheus was successful
			return nil, nil
		}
	} else if err != nil {
		// Error that isn't due to the Prometheus not existing
		log.Error(err, "Failed to get Prometheus")
		return &reconcile.Result{}, err
	}

	return nil, nil
}

func (r *ReconcileBesu) ensureGrafana(request reconcile.Request,
	instance *hyperledgerv1alpha1.Besu,
	sfs *hyperledgerv1alpha1.Grafana,
) (*reconcile.Result, error) {

	// See if Grafana already exists and create if it doesn't
	found := &hyperledgerv1alpha1.Grafana{}
	err := r.client.Get(context.TODO(), types.NamespacedName{
		Name:      sfs.ObjectMeta.Name,
		Namespace: sfs.ObjectMeta.Namespace,
	}, found)
	if err != nil && errors.IsNotFound(err) {

		// Create the Grafana
		log.Info("Creating a new Grafana", "Grafana.Namespace", sfs.Namespace, "Grafana.Name", sfs.Name)
		err = r.client.Create(context.TODO(), sfs)

		if err != nil {
			// Grafana failed
			log.Error(err, "Failed to create new Grafana", "Grafana.Namespace", sfs.Namespace, "Grafana.Name", sfs.Name)
			return &reconcile.Result{}, err
		} else {
			// Grafana was successful
			return nil, nil
		}
	} else if err != nil {
		// Error that isn't due to the Grafana not existing
		log.Error(err, "Failed to get Grafana")
		return &reconcile.Result{}, err
	}

	return nil, nil
}

func (r *ReconcileBesu) ensureSecret(request reconcile.Request,
	instance *hyperledgerv1alpha1.Besu,
	s *corev1.Secret,
) (*reconcile.Result, error) {
	found := &corev1.Secret{}
	err := r.client.Get(context.TODO(), types.NamespacedName{
		Name:      s.ObjectMeta.Name,
		Namespace: s.ObjectMeta.Namespace,
	}, found)
	if err != nil && errors.IsNotFound(err) {
		// Create the secret
		log.Info("Creating a new secret", "Secret.Namespace", s.Namespace, "Secret.Name", s.Name)
		err = r.client.Create(context.TODO(), s)

		if err != nil {
			// Creation failed
			log.Error(err, "Failed to create new Secret", "Secret.Namespace", s.Namespace, "Secret.Name", s.Name)
			return &reconcile.Result{}, err
		} else {
			// Creation was successful
			return nil, nil
		}
	} else if err != nil {
		// Error that isn't due to the secret not existing
		log.Error(err, "Failed to get Secret")
		return &reconcile.Result{}, err
	}

	return nil, nil
}
