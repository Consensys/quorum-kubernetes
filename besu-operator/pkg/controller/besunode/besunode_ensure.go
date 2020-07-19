package besunode

import (
	"context"

	hyperledgerv1alpha1 "github.com/Sumaid/besu-kubernetes/besu-operator/pkg/apis/hyperledger/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

func (r *ReconcileBesuNode) ensureStatefulSet(request reconcile.Request,
	instance *hyperledgerv1alpha1.BesuNode,
	sfs *appsv1.StatefulSet,
) (*reconcile.Result, error) {

	// See if StatefulSet already exists and create if it doesn't
	found := &appsv1.StatefulSet{}
	err := r.client.Get(context.TODO(), types.NamespacedName{
		Name:      sfs.ObjectMeta.Name,
		Namespace: sfs.ObjectMeta.Namespace,
	}, found)
	if err != nil && errors.IsNotFound(err) {

		// Create the StatefulSet
		log.Info("Creating a new StatefulSet", "StatefulSet.Namespace", sfs.Namespace, "StatefulSet.Name", sfs.Name)
		err = r.client.Create(context.TODO(), sfs)

		if err != nil {
			// StatefulSet failed
			log.Error(err, "Failed to create new StatefulSet", "StatefulSet.Namespace", sfs.Namespace, "StatefulSet.Name", sfs.Name)
			return &reconcile.Result{}, err
		} else {
			// StatefulSet was successful
			return nil, nil
		}
	} else if err != nil {
		// Error that isn't due to the StatefulSet not existing
		log.Error(err, "Failed to get StatefulSet")
		return &reconcile.Result{}, err
	}

	var result *reconcile.Result
	result, err = r.handleStatefulSetChanges(instance, found)
	if result != nil || err != nil {
		return result, err
	}
	err = r.updateBesuNodeStatus(instance, found.Status.Replicas, found.Status.ReadyReplicas)
	if err != nil {
		log.Error(err, "Failed to update besu node status")
	}
	return nil, nil
}

func (r *ReconcileBesuNode) updateBesuNodeStatus(instance *hyperledgerv1alpha1.BesuNode,
	replicas int32,
	readyreplicas int32) error {
	instance.Status.Replicas = replicas
	instance.Status.ReadyReplicas = readyreplicas
	err := r.client.Status().Update(context.TODO(), instance)
	return err
}

func (r *ReconcileBesuNode) handleStatefulSetChanges(instance *hyperledgerv1alpha1.BesuNode, found *appsv1.StatefulSet) (*reconcile.Result, error) {
	updated := false
	if instance.Spec.Image.Repository+":"+instance.Spec.Image.Tag != found.Spec.Template.Spec.Containers[0].Image {
		found.Spec.Template.Spec.Containers[0].Image = instance.Spec.Image.Repository + ":" + instance.Spec.Image.Tag
		updated = true
	}
	if instance.Spec.Replicas != *found.Spec.Replicas {
		found.Spec.Replicas = &instance.Spec.Replicas
		updated = true
	}
	if updated {
		err := r.client.Update(context.TODO(), found)
		if err != nil {
			log.Error(err, "Failed to update Statefulset.", "Statefulset.Namespace", found.Namespace, "Statefulset.Name", found.Name)
			return &reconcile.Result{}, err
		}
		// Spec updated - return and requeue
		return &reconcile.Result{Requeue: true}, nil
	}
	return nil, nil
}

func (r *ReconcileBesuNode) ensureServiceAccount(request reconcile.Request,
	instance *hyperledgerv1alpha1.BesuNode,
	sfs *corev1.ServiceAccount,
) (*reconcile.Result, error) {

	// See if SecretAccount already exists and create if it doesn't
	found := &corev1.ServiceAccount{}
	err := r.client.Get(context.TODO(), types.NamespacedName{
		Name:      sfs.ObjectMeta.Name,
		Namespace: sfs.ObjectMeta.Namespace,
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

func (r *ReconcileBesuNode) ensureRole(request reconcile.Request,
	instance *hyperledgerv1alpha1.BesuNode,
	sfs *rbacv1.Role,
) (*reconcile.Result, error) {

	// See if Role already exists and create if it doesn't
	found := &rbacv1.Role{}
	err := r.client.Get(context.TODO(), types.NamespacedName{
		Name:      sfs.ObjectMeta.Name,
		Namespace: sfs.ObjectMeta.Namespace,
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

func (r *ReconcileBesuNode) ensureRoleBinding(request reconcile.Request,
	instance *hyperledgerv1alpha1.BesuNode,
	sfs *rbacv1.RoleBinding,
) (*reconcile.Result, error) {

	// See if RoleBinding already exists and create if it doesn't
	found := &rbacv1.RoleBinding{}
	err := r.client.Get(context.TODO(), types.NamespacedName{
		Name:      sfs.ObjectMeta.Name,
		Namespace: sfs.ObjectMeta.Namespace,
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
func (r *ReconcileBesuNode) ensureService(request reconcile.Request,
	instance *hyperledgerv1alpha1.BesuNode,
	s *corev1.Service,
) (*reconcile.Result, error) {
	found := &corev1.Service{}
	err := r.client.Get(context.TODO(), types.NamespacedName{
		Name:      s.ObjectMeta.Name,
		Namespace: s.ObjectMeta.Namespace,
	}, found)
	if err != nil && errors.IsNotFound(err) {

		// Create the service
		log.Info("Creating a new Service", "Service.Namespace", s.Namespace, "Service.Name", s.Name)
		err = r.client.Create(context.TODO(), s)

		if err != nil {
			// Creation failed
			log.Error(err, "Failed to create new Service", "Service.Namespace", s.Namespace, "Service.Name", s.Name)
			return &reconcile.Result{}, err
		} else {
			// Creation was successful
			return nil, nil
		}
	} else if err != nil {
		// Error that isn't due to the service not existing
		log.Error(err, "Failed to get Service")
		return &reconcile.Result{}, err
	}

	return nil, nil
}

func (r *ReconcileBesuNode) ensureConfigMap(request reconcile.Request,
	instance *hyperledgerv1alpha1.BesuNode,
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
