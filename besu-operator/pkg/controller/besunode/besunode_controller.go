package besunode

import (
	"context"

	hyperledgerv1alpha1 "github.com/Sumaid/besu-kubernetes/besu-operator/pkg/apis/hyperledger/v1alpha1"
	"github.com/Sumaid/besu-kubernetes/besu-operator/pkg/resources"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

var log = logf.Log.WithName("controller_besunode")

// Add creates a new BesuNode Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileBesuNode{client: mgr.GetClient(), scheme: mgr.GetScheme()}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("besunode-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource BesuNode
	err = c.Watch(&source.Kind{Type: &hyperledgerv1alpha1.BesuNode{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	// Watch for changes to secondary resource Pods and requeue the owner BesuNode
	err = c.Watch(&source.Kind{Type: &appsv1.StatefulSet{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &hyperledgerv1alpha1.BesuNode{},
	})
	if err != nil {
		return err
	}

	// Watch for changes to secrets that we create
	err = c.Watch(&source.Kind{Type: &corev1.Secret{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &hyperledgerv1alpha1.BesuNode{},
	})
	if err != nil {
		return err
	}

	// Watch for changes to services that we create
	err = c.Watch(&source.Kind{Type: &corev1.Service{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &hyperledgerv1alpha1.BesuNode{},
	})
	if err != nil {
		return err
	}

	err = c.Watch(&source.Kind{Type: &corev1.ServiceAccount{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &hyperledgerv1alpha1.BesuNode{},
	})
	if err != nil {
		return err
	}

	err = c.Watch(&source.Kind{Type: &rbacv1.Role{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &hyperledgerv1alpha1.BesuNode{},
	})
	if err != nil {
		return err
	}

	err = c.Watch(&source.Kind{Type: &rbacv1.RoleBinding{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &hyperledgerv1alpha1.BesuNode{},
	})
	if err != nil {
		return err
	}

	return nil
}

// blank assignment to verify that ReconcileBesuNode implements reconcile.Reconciler
var _ reconcile.Reconciler = &ReconcileBesuNode{}

// ReconcileBesuNode reconciles a BesuNode object
type ReconcileBesuNode struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client client.Client
	scheme *runtime.Scheme
}

// Reconcile reads that state of the cluster for a BesuNode object and makes changes based on the state read
// and what is in the BesuNode.Spec
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileBesuNode) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Reconciling BesuNode")

	// Fetch the BesuNode instance
	instance := &hyperledgerv1alpha1.BesuNode{}
	err := r.client.Get(context.TODO(), request.NamespacedName, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}
	namespace := instance.GetNamespace()
	var result *reconcile.Result

	//ensure existence of secretaccount
	if instance.Spec.Type != "Member" {
		//ensure existence of secret child resources
		result, err = r.ensureRole(request, instance, r.besunodeRole(instance))
		if result != nil {
			return *result, err
		}

		result, err = r.ensureRoleBinding(request, instance, r.besunodeRoleBinding(instance))
		if result != nil {
			return *result, err
		}
	}

	result, err = r.ensureServiceAccount(request, instance, resources.NewServiceAccount(instance.ObjectMeta.Name+"-sa", namespace))
	if result != nil {
		return *result, err
	}

	//ensure existence of service child resources
	result, err = r.ensureService(request, instance, r.besunodeService(instance))
	if result != nil {
		return *result, err
	}

	//ensure existence of statefulset child resources
	result, err = r.ensureStatefulSet(request, instance, r.besunodeStatefulSet(instance))
	if result != nil {
		return *result, err
	}

	// == Finish ==========
	// Everything went fine, don't requeue
	reqLogger.Info("BesuNode Reconciled ended : Everything went fine")
	return reconcile.Result{}, nil
}
