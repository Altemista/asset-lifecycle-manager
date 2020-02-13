package operatedasset

import (
	"context"

	aalmv1alpha1 "github.com/Altemista/asset-lifecycle-manager/pkg/apis/aalm/v1alpha1"
	olmv1 "github.com/operator-framework/operator-lifecycle-manager/pkg/api/apis/operators/v1"
	olmv1alpha1 "github.com/operator-framework/operator-lifecycle-manager/pkg/api/apis/operators/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
	"sort"
)

var log = logf.Log.WithName("controller_operatedasset")

// Add creates a new OperatedAsset Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new ReconcileOperatedAsset that implements reconcile.Reconciler
func newReconciler(mgr manager.Manager) *ReconcileOperatedAsset {
	return &ReconcileOperatedAsset{client: mgr.GetClient(), scheme: mgr.GetScheme()}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler (ReconcileOperatedAsset)
func add(mgr manager.Manager, r *ReconcileOperatedAsset) error {
	// Create a new controller
	c, err := controller.New("operatedasset-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}
	//This reconciler needs the controller to watch resources dynamically
	r.SetController(c)

	// Watch for changes to primary resource OperatedAsset
	err = c.Watch(&source.Kind{Type: &aalmv1alpha1.OperatedAsset{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	// Transform requests for subresources created by OperatedAsset to OperatedAsset requests
	// Thanks for that, the same reconcile function can be used to attend to necessary resource events 
	enqueueRequestsFromMapFunc := &handler.EnqueueRequestsFromMapFunc{
		ToRequests: handler.ToRequestsFunc(func(a handler.MapObject) []reconcile.Request {
			annotations := a.Meta.GetAnnotations()
			return []reconcile.Request{
				{NamespacedName: types.NamespacedName{
					Name:      annotations["aalm.altemista.cloud/name"],
					Namespace: annotations["aalm.altemista.cloud/namespace"],
				}},
			}
		}),
	}

	// Only watch events from resources which are annotated with "aalm.altemista.cloud/name"
	predicateFuncs := predicate.Funcs{
		CreateFunc: func(e event.CreateEvent) bool {
			_, exists := e.Meta.GetAnnotations()["aalm.altemista.cloud/name"]
			return exists
		},
		UpdateFunc: func(e event.UpdateEvent) bool {
			_, existsOld := e.MetaOld.GetAnnotations()["aalm.altemista.cloud/name"]
			_, existsNew := e.MetaNew.GetAnnotations()["aalm.altemista.cloud/name"]
			return existsOld && existsNew
		},
		GenericFunc: func(e event.GenericEvent) bool {
			_, exists := e.Meta.GetAnnotations()["aalm.altemista.cloud/name"]
			return exists
		},
		DeleteFunc: func(e event.DeleteEvent) bool {
			_, exists := e.Meta.GetAnnotations()["aalm.altemista.cloud/name"]
			return exists
		},
	}

	// Watch Namespaces created by an OperatedAsset
	err = c.Watch(
		&source.Kind{Type: &corev1.Namespace{}},
		enqueueRequestsFromMapFunc,
		predicateFuncs,
	)

	// Watch OperatorGroups created by an OperatedAsset
	err = c.Watch(
		&source.Kind{Type: &olmv1.OperatorGroup{}},
		enqueueRequestsFromMapFunc,
		predicateFuncs,
	)

	// Watch Subscription created by an OperatedAsset
	err = c.Watch(
		&source.Kind{Type: &olmv1alpha1.Subscription{}},
		enqueueRequestsFromMapFunc,
		predicateFuncs,
	)

	return nil
}

// blank assignment to verify that ReconcileOperatedAsset implements reconcile.Reconciler
var _ reconcile.Reconciler = &ReconcileOperatedAsset{}

// ReconcileOperatedAsset reconciles a OperatedAsset object
type ReconcileOperatedAsset struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client     client.Client
	scheme     *runtime.Scheme
	controller controller.Controller
	// Slice of assets operated by this controller
	assets     []string
}

// Method to inject controller to reconciler once it's created
func (r *ReconcileOperatedAsset) SetController(c controller.Controller) {
	r.controller = c
}

// Reconcile reads that state of the cluster for a OperatedAsset object and makes changes based on the state read
// and what is in the OperatedAsset.Spec
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileOperatedAsset) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Reconciling OperatedAsset")

	// Fetch the OperatedAsset instance
	instance := &aalmv1alpha1.OperatedAsset{}
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

	// Define a new Namespace object
	namespace := newNamespaceForCR(instance)

	// Check if this Namespace already exists
	foundNamespace := &corev1.Namespace{}
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: namespace.Name, Namespace: ""}, foundNamespace)
	if err != nil && errors.IsNotFound(err) {
		reqLogger.Info("Creating a new Namespace", "Namespace.Name", namespace.Name)
		err = r.client.Create(context.TODO(), namespace)
		if err != nil {
			return reconcile.Result{}, err
		}

		// Namespace created successfully - don't requeue
		return reconcile.Result{}, nil
	} else if err != nil {
		return reconcile.Result{}, err
	}
	// Namespace already exists - don't requeue
	reqLogger.Info("Skip reconcile: Namespace already exists", "Namespace.name", foundNamespace.Name)

	// Define a new OperatorGroup object
	operatorgroup := newOperatorGroupForCR(instance)

	// Check if this OperatorGroup already exists
	foundOperatorGroup := &olmv1.OperatorGroup{}
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: operatorgroup.Name, Namespace: operatorgroup.Namespace}, foundOperatorGroup)
	if err != nil && errors.IsNotFound(err) {
		reqLogger.Info("Creating a new OperatorGroup", "OperatorGroup.Name", operatorgroup.Name)
		err = r.client.Create(context.TODO(), operatorgroup)
		if err != nil {
			return reconcile.Result{}, err
		}

		// OperatorGroup created successfully - don't requeue
		return reconcile.Result{}, nil
	} else if err != nil {
		return reconcile.Result{}, err
	}
	// OperatorGroup already exists - don't requeue
	reqLogger.Info("Skip reconcile: OperatorGroup already exists", "OperatorGroup.name", foundOperatorGroup.Name)

	// Define a new Subscription object
	subscription := newSubscriptionForCR(instance)

	// Check if this Subscription already exists
	foundSubscription := &olmv1alpha1.Subscription{}
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: subscription.Name, Namespace: subscription.Namespace}, foundSubscription)
	if err != nil && errors.IsNotFound(err) {
		reqLogger.Info("Creating a new Subscription", "Subscription.Name", subscription.Name)
		err = r.client.Create(context.TODO(), subscription)
		if err != nil {
			return reconcile.Result{}, err
		}

		// Subscription created successfully - don't requeue
		return reconcile.Result{}, nil
	} else if err != nil {
		return reconcile.Result{}, err
	}
	// Subscription already exists - don't requeue
	reqLogger.Info("Skip reconcile: Subscription already exists", "Subscription.name", foundSubscription.Name)

	asset := &unstructured.Unstructured{}
	err = asset.UnmarshalJSON(instance.Spec.Asset.Raw)
	if err != nil {
		reqLogger.Error(err, "ERRORRRR")
		return reconcile.Result{}, err
	}
	asset.SetNamespace(instance.Namespace)

	if err := controllerutil.SetControllerReference(instance, asset, r.scheme); err != nil {
		reqLogger.Error(err, "ERRORRRR")
		return reconcile.Result{}, err
	}

	// Watch dynamically for changes in the asset resource and translate to event from primary resource OperatedAsset
	// Only watch if the resource is new for this controller 
	if sort.SearchStrings(r.assets, asset.GetKind()) == len(r.assets) {
		err = r.controller.Watch(&source.Kind{Type: asset},
			&handler.EnqueueRequestForOwner{
				IsController: false,
				OwnerType:    &aalmv1alpha1.OperatedAsset{},
			},
			predicate.Funcs{
				CreateFunc: func(e event.CreateEvent) bool {
					reqLogger.Info("Create Asset Event")
					return false
				},
				UpdateFunc: func(e event.UpdateEvent) bool {
					reqLogger.Info("Update Asset Event")
					return false
				},
				GenericFunc: func(e event.GenericEvent) bool {
					reqLogger.Info("Generic Asset Event")
					return false
				},
				DeleteFunc: func(e event.DeleteEvent) bool {
					reqLogger.Info("Delete Asset Event")
					return true
				},
			},
		)
		if err != nil {
			reqLogger.Error(err, "WATCH   ERRORRRR")
			return reconcile.Result{}, err
		}
		r.assets = append(r.assets, asset.GetKind())
	}

	// Check if this Asset resource already exists
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: asset.GetName(), Namespace: asset.GetNamespace()}, asset)
	if err != nil && errors.IsNotFound(err) {
		reqLogger.Info("Creating a new Asset", "Asset.Name", asset.GetName())
		err = r.client.Create(context.TODO(), asset)
		if err != nil {
			reqLogger.Error(err, "ERRORRRR")
			return reconcile.Result{}, err
		}

		// Asset created successfully - don't requeue
		return reconcile.Result{}, nil
	} else if err != nil {
		return reconcile.Result{}, err
	}

	// Asset already exists - don't requeue
	reqLogger.Info("Skip reconcile: Asset already exists", "Asset.name", asset.GetName())	

	return reconcile.Result{}, nil
}

func newNamespaceForCR(cr *aalmv1alpha1.OperatedAsset) *corev1.Namespace {
	labels := map[string]string{
		"operatedasset": cr.Name,
	}
	annotations := map[string]string{
		"aalm.altemista.cloud/name":      cr.Name,
		"aalm.altemista.cloud/namespace": cr.Namespace,
	}
	return &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name:        cr.Namespace + "-operators",
			Labels:      labels,
			Annotations: annotations,
		},
	}
}

func newOperatorGroupForCR(cr *aalmv1alpha1.OperatedAsset) *olmv1.OperatorGroup {
	labels := map[string]string{
		"operatedasset": cr.Name,
	}
	annotations := map[string]string{
		"aalm.altemista.cloud/name":      cr.Name,
		"aalm.altemista.cloud/namespace": cr.Namespace,
	}
	return &olmv1.OperatorGroup{
		ObjectMeta: metav1.ObjectMeta{
			Name:        cr.Namespace + "-operators",
			Namespace:   cr.Namespace + "-operators",
			Labels:      labels,
			Annotations: annotations,
		},
		Spec: olmv1.OperatorGroupSpec{
			TargetNamespaces: []string{cr.Namespace},
		},
	}
}

func newSubscriptionForCR(cr *aalmv1alpha1.OperatedAsset) *olmv1alpha1.Subscription {
	labels := map[string]string{
		"operatedasset": cr.Name,
	}
	annotations := map[string]string{
		"aalm.altemista.cloud/name":      cr.Name,
		"aalm.altemista.cloud/namespace": cr.Namespace,
	}

	return &olmv1alpha1.Subscription{
		ObjectMeta: metav1.ObjectMeta{
			Name:        cr.Spec.Operator.Package + "-operators",
			Namespace:   cr.Namespace + "-operators",
			Labels:      labels,
			Annotations: annotations,
		},
		Spec: &cr.Spec.Operator,
	}
}
