package besu

import (
	"context"
	"strconv"

	hyperledgerv1alpha1 "github.com/Sumaid/besu-kubernetes/besu-operator/pkg/apis/hyperledger/v1alpha1"
	"github.com/Sumaid/besu-kubernetes/besu-operator/pkg/resources"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

var log = logf.Log.WithName("controller_besu")

/**
* USER ACTION REQUIRED: This is a scaffold file intended for the user to modify with their own Controller
* business logic.  Delete these comments after modifying this file.*
 */

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
// TODO(user): Modify this Reconcile function to implement your Controller logic.  This example creates
// a Pod as an example
// Note:
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
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}
	var result *reconcile.Result

	if len(instance.Spec.Bootnodes) > 0 && instance.Spec.Bootnodes[0].PubKey != "" {
		instance.Status.HaveKeys = true
	} else {
		instance.Status.HaveKeys = false
	}

	if instance.Status.HaveKeys == false {
		// generate keys
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

	// for i, bootspec := range instance.Spec.Bootnodes {
	// 	pubkeys["bootnode"+strconv.Itoa(i+1)+"pubkey"] = bootspec.PubKey
	// }

	result, err = r.ensureConfigMap(request, instance, r.besuConfigMap(instance))
	if result != nil {
		return *result, err
	}

	for i := 0; i < instance.Spec.BootnodesCount; i++ {
		node := &hyperledgerv1alpha1.BesuNode{
			TypeMeta: metav1.TypeMeta{
				Kind:       "BesuNode",
				APIVersion: "hyperledger.org/v1alpha1",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      "bootnode" + strconv.Itoa(i+1),
				Namespace: instance.Namespace,
			},
			Spec: hyperledgerv1alpha1.BesuNodeSpec{
				Type:      "Bootnode",
				Bootnodes: instance.Spec.BootnodesCount,
			},
		}
		controllerutil.SetControllerReference(instance, node, r.scheme)
		result, err = r.ensureBesuNode(request, instance, node)
		log.Error(err, "Failed to ensure bootnode BesuNode")
		if result != nil {
			return *result, err
		}
	}

	for i := 0; i < instance.Spec.ValidatorsCount; i++ {
		validatorNode := &hyperledgerv1alpha1.BesuNode{
			TypeMeta: metav1.TypeMeta{
				Kind:       "BesuNode",
				APIVersion: "hyperledger.org/v1alpha1",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      "validator" + strconv.Itoa(i+1),
				Namespace: instance.Namespace,
			},
			Spec: hyperledgerv1alpha1.BesuNodeSpec{
				Type:      "Validator",
				Bootnodes: instance.Spec.BootnodesCount,
			},
		}
		controllerutil.SetControllerReference(instance, validatorNode, r.scheme)
		result, err = r.ensureBesuNode(request, instance, validatorNode)
		if result != nil {
			log.Error(err, "Failed to ensure validator BesuNode")
			return *result, err
		}
	}

	MemberNode := &hyperledgerv1alpha1.BesuNode{
		TypeMeta: metav1.TypeMeta{
			Kind:       "BesuNode",
			APIVersion: "hyperledger.org/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "member",
			Namespace: instance.Namespace,
		},
		Spec: hyperledgerv1alpha1.BesuNodeSpec{
			Replicas:  instance.Spec.Members,
			Type:      "Member",
			Bootnodes: instance.Spec.BootnodesCount,
		},
	}
	controllerutil.SetControllerReference(instance, MemberNode, r.scheme)
	result, err = r.ensureBesuNode(request, instance, MemberNode)
	if result != nil {
		log.Error(err, "Failed to ensure member BesuNode")
		return *result, err
	}

	reqLogger.Info("Besu Reconciled ended : Everything went fine")
	return reconcile.Result{}, nil
}
