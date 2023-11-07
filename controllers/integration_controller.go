package controllers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/url"

	"k8s.io/apimachinery/pkg/api/errors"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	ctrllog "sigs.k8s.io/controller-runtime/pkg/log"

	oncallv1 "github.com/ahcogn/grafana-oncall-operator/api/v1"
	oncallClient "github.com/ahcogn/grafana-oncall-operator/internal/client"
	oncall_schema "github.com/ahcogn/grafana-oncall-operator/internal/schemas"
)

const (
	integrationUrl  = "/api/v1/integrations/"
	oncallFinalizer = "oncall.grafana.com/finalizer"
)

// IntegrationReconciler reconciles a Integration object
type IntegrationReconciler struct {
	client.Client
	Scheme  *runtime.Scheme
	Config  oncallClient.Config
	baseUrl string
}

//+kubebuilder:rbac:groups=oncall.ahcogn.com,resources=integrations,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=oncall.ahcogn.com,resources=integrations/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=oncall.ahcogn.com,resources=integrations/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Integration object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.14.1/pkg/reconcile
func (r *IntegrationReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := ctrllog.FromContext(ctx)
	log.Info("Start reconcile oncall resource...")
	baseUrl := r.Config.OncallUrl.JoinPath(integrationUrl)
	// TODO(user): your logic here
	integration := &oncallv1.Integration{}
	err := r.Get(ctx, req.NamespacedName, integration)

	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			log.Info("Oncall resource not found. Ignoring since object must be deleted")
			return ctrl.Result{}, nil
		}
		// Error reading the object - requeue the request.
		log.Error(err, "Failed to get Oncall")
		return ctrl.Result{}, err
	}

	iId := integration.Spec.ID

	client := &oncallClient.DefaultClient{
		Ctx:    ctx,
		Config: r.Config,
	}

	// handle event delete resource
	isIntegrationMarkedToBeDeleted := integration.GetDeletionTimestamp() != nil
	if isIntegrationMarkedToBeDeleted {
		log.Info("Cleaning resource " + integration.Name + ": " + integration.Spec.Name)
		resourceURLbyID := baseUrl.JoinPath(iId)
		if controllerutil.ContainsFinalizer(integration, oncallFinalizer) {
			// Run finalization logic for memcachedFinalizer. If the
			// finalization logic fails, don't remove the finalizer so
			// that we can retry during the next reconciliation.
			if err := r.finalizeIntegration(ctx, *resourceURLbyID, client, integration); err != nil {
				return ctrl.Result{}, err
			}

			// Remove memcachedFinalizer. Once all finalizers have been
			// removed, the object will be deleted.
			controllerutil.RemoveFinalizer(integration, oncallFinalizer)
			err := r.Update(ctx, integration)
			if err != nil {
				return ctrl.Result{}, err
			}
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, nil
	}

	if !controllerutil.ContainsFinalizer(integration, oncallFinalizer) {
		controllerutil.AddFinalizer(integration, oncallFinalizer)
		err = r.Update(ctx, integration)
		if err != nil {
			return ctrl.Result{}, err
		}
	}

	if iId == "" {
		liveIntegration, err := r.createIntegration(ctx, *baseUrl, client, integration)
		if err != nil {
			return ctrl.Result{}, err
		}

		integration.Spec.ID = liveIntegration.ID
		rawjs, err := json.Marshal(liveIntegration)

		log.Info(string(rawjs))

		err = r.Update(ctx, integration)
		if err != nil {
			return ctrl.Result{}, err
		}
		return ctrl.Result{Requeue: true}, nil
	}

	resourceURLbyID := baseUrl.JoinPath(iId)

	liveIntegration, err := r.getIntegration(ctx, *resourceURLbyID, client, iId)
	if err != nil {
	}

	integration.Status.HttpEndpoint = liveIntegration.Link
	r.Status().Update(ctx, integration)

	log.Info("Done reconcile oncall")
	return ctrl.Result{}, nil
}

func (r *IntegrationReconciler) getIntegration(ctx context.Context, url url.URL, client oncallClient.Client, integrationId string) (*oncall_schema.Integration, error) {

	resp, err := client.Get(ctx, url)
	if err != nil {
		return nil, err
	}

	integrationResp := &oncall_schema.Integration{}
	err = json.NewDecoder(bytes.NewReader(resp)).Decode(&integrationResp)
	if err != nil {
		return nil, err
	}

	return integrationResp, err
}

func (r *IntegrationReconciler) createIntegration(ctx context.Context, url url.URL, client oncallClient.Client, integration *oncallv1.Integration) (*oncall_schema.Integration, error) {

	if integration.Spec.Name == "" {
		integration.Spec.Name = integration.Name
	}
	integrationReq := &oncallv1.IntegrationSpec{
		Name: integration.Spec.Name,
		Type: integration.Spec.Type,
	}

	resp, err := client.Post(ctx, url, integrationReq)
	if err != nil {
		return nil, err
	}

	integrationResp := &oncall_schema.Integration{}
	err = json.NewDecoder(bytes.NewReader(resp)).Decode(&integrationResp)
	if err != nil {
		return nil, err
	}

	return integrationResp, err
}

func (r *IntegrationReconciler) updateIntegration(ctx context.Context, url url.URL, client oncallClient.Client, integration *oncallv1.Integration) (*oncall_schema.Integration, error) {

	return nil, nil
}

func (r *IntegrationReconciler) finalizeIntegration(ctx context.Context, url url.URL, client oncallClient.Client, i *oncallv1.Integration) error {
	err := client.Delete(ctx, url)
	return err
}

// SetupWithManager sets up the controller with the Manager.
func (r *IntegrationReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&oncallv1.Integration{}).
		Complete(r)
}
