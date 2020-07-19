package besu

import (
	"context"
	"crypto/ecdsa"
	"strconv"

	hyperledgerv1alpha1 "github.com/Sumaid/besu-kubernetes/besu-operator/pkg/apis/hyperledger/v1alpha1"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	corev1 "k8s.io/api/core/v1"

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

	// If user has provided more keys than required, then only consider first count keys
	bootnodeKeys := instance.Spec.BootnodeKeys[:min(instance.Spec.BootnodesCount, len(instance.Spec.BootnodeKeys))]
	validatorKeys := instance.Spec.ValidatorKeys[:min(instance.Spec.ValidatorsCount, len(instance.Spec.ValidatorKeys))]

	// If user has provided less keys than, reqdKeys is extra keys which need to be generated
	reqdBootnodeKeys := instance.Spec.BootnodesCount - len(bootnodeKeys)
	reqdValidatorKeys := instance.Spec.ValidatorsCount - len(validatorKeys)

	for i := 0; i < reqdBootnodeKeys; i++ {
		privkey, pubkey := r.generateKeyPair()
		bootnodeKeys = append(bootnodeKeys, hyperledgerv1alpha1.Key{PubKey: pubkey, PrivKey: privkey})
	}

	for i := 0; i < reqdValidatorKeys; i++ {
		privkey, pubkey := r.generateKeyPair()
		validatorKeys = append(validatorKeys, hyperledgerv1alpha1.Key{PubKey: pubkey, PrivKey: privkey})
	}

	for i, key := range bootnodeKeys {
		result, err = r.ensureSecret(request, instance, r.besuSecret(instance, "bootnode"+strconv.Itoa(i+1), key.PrivKey, key.PubKey))
		if result != nil {
			return *result, err
		}
	}

	for i, key := range validatorKeys {
		result, err = r.ensureSecret(request, instance, r.besuSecret(instance, "validator"+strconv.Itoa(i+1), key.PrivKey, key.PubKey))
		if result != nil {
			return *result, err
		}
	}

	result, err = r.ensureConfigMap(request, instance, r.besuGenesisConfigMap(instance))
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
		node := r.newBesuNode(instance, "validator"+strconv.Itoa(i+1), "Validator", instance.Spec.BootnodesCount)
		result, err = r.ensureBesuNode(request, instance, node)
		log.Error(err, "Failed to ensure bootnode BesuNode", "BesuNode.Namespace", instance.Namespace, "BesuNode.Name", "bootnode"+strconv.Itoa(i+1))
		if result != nil {
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

func (r *ReconcileBesu) generateKeyPair() (string, string) {
	privateKey, err := crypto.GenerateKey()
	if err != nil {
		log.Error(err, "Failed to generate private key")
	}

	privateKeyBytes := crypto.FromECDSA(privateKey)
	privkey := hexutil.Encode(privateKeyBytes)[2:]

	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		log.Error(err, "Failed to retrieve public key")
	}

	publicKeyBytes := crypto.FromECDSAPub(publicKeyECDSA)
	pubkey := hexutil.Encode(publicKeyBytes)[4:]
	return privkey, pubkey
}

func min(a int, b int) int {
	if a < b {
		return a
	}
	return b
}
