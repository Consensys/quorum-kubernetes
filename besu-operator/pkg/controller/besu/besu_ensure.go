package besu

import (
	"context"

	hyperledgerv1alpha1 "github.com/Sumaid/besu-kubernetes/besu-operator/pkg/apis/hyperledger/v1alpha1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

func (r *ReconcileBesu) ensureBesuNode(request reconcile.Request,
	instance *hyperledgerv1alpha1.Besu,
	sfs *hyperledgerv1alpha1.BesuNode,
) (*reconcile.Result, error) {

	// See if BesuNode already exists and create if it doesn't
	err := r.client.Get(context.TODO(), types.NamespacedName{
		Name:      sfs.Name,
		Namespace: instance.Namespace,
	}, sfs)
	if err != nil && errors.IsNotFound(err) {

		// Create the BesuNode
		log.Info("Creating a new BesuNode", "BesuNode.Namespace", sfs.Namespace, "BesuNode.Name", sfs.Name)
		err = r.client.Create(context.TODO(), sfs)

		if err != nil {
			// BesuNode failed
			log.Error(err, "Failed to create new BesuNode", "BesuNode.Namespace", sfs.Namespace, "BesuNode.Name", sfs.Name)
			return &reconcile.Result{}, err
		} else {
			// BesuNode was successful
			return nil, nil
		}
	} else if err != nil {
		// Error that isn't due to the BesuNode not existing
		log.Error(err, "Failed to get BesuNode")
		return &reconcile.Result{}, err
	}

	return nil, nil
}

func (r *ReconcileBesu) ensureConfigMap(request reconcile.Request,
	instance *hyperledgerv1alpha1.Besu,
	s *corev1.ConfigMap,
) (*reconcile.Result, error) {
	found := &corev1.ConfigMap{}
	err := r.client.Get(context.TODO(), types.NamespacedName{
		Name:      s.Name,
		Namespace: instance.Namespace,
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

func (r *ReconcileBesu) ensureServiceAccount(request reconcile.Request,
	instance *hyperledgerv1alpha1.Besu,
	sfs *corev1.ServiceAccount,
) (*reconcile.Result, error) {

	// See if SecretAccount already exists and create if it doesn't
	found := &corev1.ServiceAccount{}
	err := r.client.Get(context.TODO(), types.NamespacedName{
		Name:      sfs.Name,
		Namespace: instance.Namespace,
	}, found)
	if err != nil && errors.IsNotFound(err) {

		// Create the SecretAccount
		log.Info("Creating a new SecretAccount", "SecretAccount.Namespace", sfs.Namespace, "SecretAccount.Name", sfs.Name)
		err = r.client.Create(context.TODO(), sfs)

		if err != nil {
			// SecretAccount failed
			log.Error(err, "Failed to create new SecretAccount", "SecretAccount.Namespace", sfs.Namespace, "SecretAccount.Name", sfs.Name)
			return &reconcile.Result{}, err
		} else {
			// SecretAccount was successful
			return nil, nil
		}
	} else if err != nil {
		// Error that isn't due to the SecretAccount not existing
		log.Error(err, "Failed to get SecretAccount")
		return &reconcile.Result{}, err
	}

	return nil, nil
}

func (r *ReconcileBesu) ensureRole(request reconcile.Request,
	instance *hyperledgerv1alpha1.Besu,
	sfs *rbacv1.Role,
) (*reconcile.Result, error) {

	// See if Role already exists and create if it doesn't
	found := &rbacv1.Role{}
	err := r.client.Get(context.TODO(), types.NamespacedName{
		Name:      sfs.Name,
		Namespace: instance.Namespace,
	}, found)
	if err != nil && errors.IsNotFound(err) {

		// Create the Role
		log.Info("Creating a new Role", "Role.Namespace", sfs.Namespace, "Role.Name", sfs.Name)
		err = r.client.Create(context.TODO(), sfs)

		if err != nil {
			// Role failed
			log.Error(err, "Failed to create new Role", "Role.Namespace", sfs.Namespace, "Role.Name", sfs.Name)
			return &reconcile.Result{}, err
		} else {
			// Role was successful
			return nil, nil
		}
	} else if err != nil {
		// Error that isn't due to the Role not existing
		log.Error(err, "Failed to get Role")
		return &reconcile.Result{}, err
	}

	return nil, nil
}

func (r *ReconcileBesu) ensureRoleBinding(request reconcile.Request,
	instance *hyperledgerv1alpha1.Besu,
	sfs *rbacv1.RoleBinding,
) (*reconcile.Result, error) {

	// See if RoleBinding already exists and create if it doesn't
	found := &rbacv1.RoleBinding{}
	err := r.client.Get(context.TODO(), types.NamespacedName{
		Name:      sfs.Name,
		Namespace: instance.Namespace,
	}, found)
	if err != nil && errors.IsNotFound(err) {

		// Create the RoleBinding
		log.Info("Creating a new RoleBinding", "RoleBinding.Namespace", sfs.Namespace, "RoleBinding.Name", sfs.Name)
		err = r.client.Create(context.TODO(), sfs)

		if err != nil {
			// RoleBinding failed
			log.Error(err, "Failed to create new RoleBinding", "RoleBinding.Namespace", sfs.Namespace, "RoleBinding.Name", sfs.Name)
			return &reconcile.Result{}, err
		} else {
			// RoleBinding was successful
			return nil, nil
		}
	} else if err != nil {
		// Error that isn't due to the RoleBinding not existing
		log.Error(err, "Failed to get RoleBinding")
		return &reconcile.Result{}, err
	}

	return nil, nil
}

func (r *ReconcileBesu) ensureJob(request reconcile.Request,
	instance *hyperledgerv1alpha1.Besu,
	sfs *batchv1.Job,
) (*reconcile.Result, error) {

	// See if Job already exists and create if it doesn't
	found := &batchv1.Job{}
	err := r.client.Get(context.TODO(), types.NamespacedName{
		Name:      sfs.Name,
		Namespace: instance.Namespace,
	}, found)
	if err != nil && errors.IsNotFound(err) {

		// Create the Job
		log.Info("Creating a new Job", "Job.Namespace", sfs.Namespace, "Job.Name", sfs.Name)
		err = r.client.Create(context.TODO(), sfs)

		if err != nil {
			// Job failed
			log.Error(err, "Failed to create new Job", "Job.Namespace", sfs.Namespace, "Job.Name", sfs.Name)
			return &reconcile.Result{}, err
		} else {
			// Job was successful
			return nil, nil
		}
	} else if err != nil {
		// Error that isn't due to the Job not existing
		log.Error(err, "Failed to get Job")
		return &reconcile.Result{}, err
	}

	return nil, nil
}

func (r *ReconcileBesu) ensurePrometheus(request reconcile.Request,
	instance *hyperledgerv1alpha1.Besu,
	sfs *hyperledgerv1alpha1.Prometheus,
) (*reconcile.Result, error) {

	// See if Prometheus already exists and create if it doesn't
	err := r.client.Get(context.TODO(), types.NamespacedName{
		Name:      sfs.Name,
		Namespace: instance.Namespace,
	}, sfs)
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
	err := r.client.Get(context.TODO(), types.NamespacedName{
		Name:      sfs.Name,
		Namespace: instance.Namespace,
	}, sfs)
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
