package besu

import (
	"context"
	"strconv"

	hyperledgerv1alpha1 "github.com/Sumaid/besu-kubernetes/besu-operator/pkg/apis/hyperledger/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

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
	pubkeys := make(map[string]string)

	for i, bootspec := range instance.Spec.Bootnodes {
		pubkeys["bootnode"+strconv.Itoa(i+1)+"pubkey"] = bootspec.PubKey
	}

	result, err = r.ensureConfigMap(request, instance, r.besuConfigMap(instance, pubkeys))
	if result != nil {
		return *result, err
	}

	for i, bootspec := range instance.Spec.Bootnodes {
		bootspec.Type = "Bootnode"
		bootspec.Bootnodes = instance.Spec.BootnodesCount
		node := &hyperledgerv1alpha1.BesuNode{
			TypeMeta: metav1.TypeMeta{
				Kind:       "BesuNode",
				APIVersion: "hyperledger.org/v1alpha1",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      "bootnode" + strconv.Itoa(i+1),
				Namespace: instance.Namespace,
			},
			Spec: bootspec,
		}
		controllerutil.SetControllerReference(instance, node, r.scheme)
		result, err = r.ensureBesuNode(request, instance, node)
		log.Error(err, "Failed to ensure bootnode BesuNode")
		if result != nil {
			return *result, err
		}
	}

	for i, validatorSpec := range instance.Spec.Validators {
		validatorSpec.Type = "Validator"
		validatorSpec.Bootnodes = instance.Spec.BootnodesCount
		validatorNode := &hyperledgerv1alpha1.BesuNode{
			TypeMeta: metav1.TypeMeta{
				Kind:       "BesuNode",
				APIVersion: "hyperledger.org/v1alpha1",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      "validator" + strconv.Itoa(i+1),
				Namespace: instance.Namespace,
			},
			Spec: validatorSpec,
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

func (r *ReconcileBesu) besuConfigMap(instance *hyperledgerv1alpha1.Besu, data map[string]string) *corev1.ConfigMap {
	data["genesis.json"] = `
	{
		"config": {
		  "chainId": 2018,
		  "constantinoplefixblock": 0,
		  "ibft2": {
			"blockperiodseconds": 2,
			"epochlength": 30000,
			"requesttimeoutseconds": 10
		  }
		},
		"nonce": "0x0",
		"timestamp": "0x58ee40ba",
		"gasLimit": "0x47b760",
		"difficulty": "0x1",
		"mixHash": "0x63746963616c2062797a616e74696e65206661756c7420746f6c6572616e6365",
		"coinbase": "0x0000000000000000000000000000000000000000",
		"alloc": {
		  "fe3b557e8fb62b89f4916b721be55ceb828dbd73": {
			"privateKey": "8f2a55949038a9610f50fb23b5883af3b4ecb3c3bb792cbcefbd1542c692be63",
			"comment": "private key and this comment are ignored.  In a real chain, the private key should NOT be stored",
			"balance": "0xad78ebc5ac6200000"
		  },
		  "627306090abaB3A6e1400e9345bC60c78a8BEf57": {
			"privateKey": "c87509a1c067bbde78beb793e6fa76530b6382a4c0241e5e4a9ec0a0f44dc0d3",
			"comment": "private key and this comment are ignored.  In a real chain, the private key should NOT be stored",
			"balance": "90000000000000000000000"
		  },
		  "f17f52151EbEF6C7334FAD080c5704D77216b732": {
			"privateKey": "ae6ae8e5ccbfb04590405997ee2d52d2b330726137b875053c36d94e974d162f",
			"comment": "private key and this comment are ignored.  In a real chain, the private key should NOT be stored",
			"balance": "90000000000000000000000"
		  }
		},
		"extraData": "0xf87ea00000000000000000000000000000000000000000000000000000000000000000f85494ca6e9704586eb1fb38194308e2192e43b1e1979c94ce2276efc33fee3c321e634eac28a9476e53b71c94f466a7174230056004d11178d2647c12740fa58b94b83820d6cf4b7e5aa67a2b57969caa5cdf6dff49808400000000c0"
	  }`

	conf := &corev1.ConfigMap{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ConfigMap",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "besu-" + "configmap",
			Namespace: instance.Namespace,
			Labels: map[string]string{
				"app": "besu-" + "configmap",
			},
		},
		Data: data,
	}
	controllerutil.SetControllerReference(instance, conf, r.scheme)
	return conf
}
