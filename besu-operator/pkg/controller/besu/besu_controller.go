package besu

import (
	"context"
	"strconv"

	hyperledgerv1alpha1 "github.com/Sumaid/besu-kubernetes/besu-operator/pkg/apis/hyperledger/v1alpha1"
	"github.com/Sumaid/besu-kubernetes/besu-operator/pkg/resources"
	"github.com/go-logr/logr"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

const besuFinalizer = "finalizer.besu.hyperleger.org"

var log = logf.Log.WithName("controller_besu")

// Add creates a new Besu Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileBesu{client: mgr.GetClient(), scheme: mgr.GetScheme()}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("besu-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource Besu
	err = c.Watch(&source.Kind{Type: &hyperledgerv1alpha1.Besu{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	// Watch for changes to secondary crd resource Besu Node and requeue the owner Besu
	err = c.Watch(&source.Kind{Type: &hyperledgerv1alpha1.BesuNode{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &hyperledgerv1alpha1.Besu{},
	})
	if err != nil {
		return err
	}

	err = c.Watch(&source.Kind{Type: &corev1.ConfigMap{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &hyperledgerv1alpha1.Besu{},
	})
	if err != nil {
		return err
	}

	err = c.Watch(&source.Kind{Type: &hyperledgerv1alpha1.Grafana{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &hyperledgerv1alpha1.Besu{},
	})
	if err != nil {
		return err
	}

	err = c.Watch(&source.Kind{Type: &hyperledgerv1alpha1.Prometheus{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &hyperledgerv1alpha1.Besu{},
	})
	if err != nil {
		return err
	}

	return nil
}

// blank assignment to verify that ReconcileBesu implements reconcile.Reconciler
var _ reconcile.Reconciler = &ReconcileBesu{}

// ReconcileBesu reconciles a Besu object
type ReconcileBesu struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client client.Client
	scheme *runtime.Scheme
}

// Reconcile reads that state of the cluster for a Besu object and makes changes based on the state read
// and what is in the Besu.Spec
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileBesu) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Reconciling Besu")

	// Fetch the Besu instance
	instance := &hyperledgerv1alpha1.Besu{}
	err := r.client.Get(context.TODO(), request.NamespacedName, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			reqLogger.Info("Not found so maybe deleted")
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}
	var result *reconcile.Result

	result, err = r.handleCleanupFinalizer(reqLogger, instance)
	if result != nil {
		return *result, err
	}

	if len(instance.Spec.BootnodeKeys) > 0 {
		instance.Status.HaveKeys = true
	} else {
		instance.Status.HaveKeys = false
	}

	if instance.Status.HaveKeys == false {
		result, err = r.ensureRole(request, instance, r.besuRole(instance))
		if result != nil {
			return *result, err
		}

		result, err = r.ensureRoleBinding(request, instance, r.besuRoleBinding(instance))
		if result != nil {
			return *result, err
		}

		result, err = r.ensureServiceAccount(request, instance, resources.NewServiceAccount(instance.ObjectMeta.Name+"-sa", instance.GetNamespace()))
		if result != nil {
			return *result, err
		}

		result, err = r.ensureJob(request, instance, r.besuInitJob(instance))
		if result != nil {
			return *result, err
		}
		instance.Status.HaveKeys = true
	}

	result, err = r.ensureConfigMap(request, instance, r.besuConfigMap(instance))
	if result != nil {
		return *result, err
	}

	for i := 0; i < instance.Spec.BootnodesCount; i++ {
		node := r.newBesuNode(instance, "bootnode"+strconv.Itoa(i+1), "Bootnode", instance.Spec.BootnodesCount)
		result, err = r.ensureBesuNode(request, instance, node)
		log.Error(err, "Failed to ensure bootnode BesuNode", "BesuNode.Namespace", instance.Namespace, "BesuNode.Name", "bootnode"+strconv.Itoa(i+1))
		if result != nil {
			return *result, err
		}
	}

	for i := 0; i < instance.Spec.ValidatorsCount; i++ {
		result, err = r.ensureBesuNode(request, instance, r.newBesuNode(instance, "validator"+strconv.Itoa(i+1), "Validator", instance.Spec.BootnodesCount))
		if result != nil {
			log.Error(err, "Failed to ensure validator BesuNode")
			return *result, err
		}
	}

	node := r.newBesuNode(instance, "member", "Member", instance.Spec.BootnodesCount)
	node.Spec.Replicas = instance.Spec.Members
	result, err = r.ensureBesuNode(request, instance, node)
	if result != nil {
		log.Error(err, "Failed to ensure member BesuNode")
		return *result, err
	}

	if instance.Spec.Monitoring {
		result, err = r.ensureGrafana(request, instance, r.newGrafana(instance))
		if result != nil {
			log.Error(err, "Failed to ensure Grafana")
			return *result, err
		}

		result, err = r.ensurePrometheus(request, instance, r.newPrometheus(instance))
		if result != nil {
			log.Error(err, "Failed to ensure Prometheus")
			return *result, err
		}
	}

	reqLogger.Info("Besu Reconciled ended : Everything went fine")
	return reconcile.Result{}, nil
}
func (r *ReconcileBesu) handleCleanupFinalizer(reqLogger logr.Logger, instance *hyperledgerv1alpha1.Besu) (*reconcile.Result, error) {
	isBesuMarkedToBeDeleted := instance.GetDeletionTimestamp() != nil
	if isBesuMarkedToBeDeleted {
		if contains(instance.GetFinalizers(), besuFinalizer) {
			// Run finalization logic for finalizer. If the
			// finalization logic fails, don't remove the finalizer so
			// that we can retry during the next reconciliation.
			sfs := r.besuCleanupJob(instance)
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
					log.Error(err, "Failed to create new Job", "Job.Namespace", sfs.Namespace, "Job.Name", sfs.Name)
					return &reconcile.Result{}, err
				} else {
					return &reconcile.Result{Requeue: true}, nil
				}
			} else if err != nil {
				// Error that isn't due to the Job not existing
				log.Error(err, "Failed to get Job")
				return &reconcile.Result{}, err
			} else {
				if found.Status.Succeeded == 0 {
					log.Info("Cleanup Job not completed yet")
					return &reconcile.Result{Requeue: true}, nil
				}
				log.Info("Cleanup Job completed")
			}

			// Remove finalizer. Once all finalizers have been
			// removed, the object will be deleted.
			controllerutil.RemoveFinalizer(instance, besuFinalizer)
			err = r.client.Update(context.TODO(), instance)
			if err != nil {
				return &reconcile.Result{}, err
			}
		}
		return &reconcile.Result{}, nil
	}

	// Add finalizer for this CR
	if !contains(instance.GetFinalizers(), besuFinalizer) {
		if err := r.addFinalizer(reqLogger, instance); err != nil {
			return &reconcile.Result{}, err
		}
	}

	return nil, nil
}

func (r *ReconcileBesu) addFinalizer(reqLogger logr.Logger, instance *hyperledgerv1alpha1.Besu) error {
	reqLogger.Info("Adding Finalizer for the Besu")
	controllerutil.AddFinalizer(instance, besuFinalizer)

	err := r.client.Update(context.TODO(), instance)
	if err != nil {
		reqLogger.Error(err, "Failed to update Besu with finalizer")
		return err
	}
	return nil
}

func contains(list []string, s string) bool {
	for _, v := range list {
		if v == s {
			return true
		}
	}
	return false
}
