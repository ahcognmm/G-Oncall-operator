package controllers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/url"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	ctrllog "sigs.k8s.io/controller-runtime/pkg/log"

	oncallv1 "github.com/ahcogn/grafana-oncall-operator/api/v1"
	oncallClient "github.com/ahcogn/grafana-oncall-operator/internal/client"
	customerror "github.com/ahcogn/grafana-oncall-operator/internal/error"
	oncall_schema "github.com/ahcogn/grafana-oncall-operator/internal/schemas"
)

const (
	integrationUrl = "/api/v1/integrations/"
	routeUrl       = "/api/v1/routes/"
)

// IntegrationReconciler reconciles a Integration object
type IntegrationReconciler struct {
	client.Client
	Scheme *runtime.Scheme
	Config oncallClient.Config
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
	baseIntegrationUrl := r.Config.OncallUrl.JoinPath(integrationUrl)
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
		resourceURLbyID := baseIntegrationUrl.JoinPath(iId)
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
		err := r.createIntegration(ctx, *baseIntegrationUrl, client, integration)
		if err != nil {
			return ctrl.Result{}, err
		}

		err = r.Update(ctx, integration)
		if err != nil {
			return ctrl.Result{}, err
		}

		err = r.Status().Update(ctx, integration)
		if err != nil {
			return ctrl.Result{}, err
		}
		// let operator reconcile again and update missing field
		return ctrl.Result{Requeue: true}, nil
	}

	err = r.updateIntegration(ctx, *&r.Config.OncallUrl, client, integration, req)
	if err != nil {
		return ctrl.Result{}, err
	}

	err = r.Status().Update(ctx, integration)
	if err != nil {
		return ctrl.Result{}, err
	}

	err = r.Update(ctx, integration)
	if err != nil {
		return ctrl.Result{}, err
	}

	log.Info("Done reconcile oncall")
	return ctrl.Result{}, nil
}

func (r *IntegrationReconciler) createIntegration(ctx context.Context, url url.URL, client oncallClient.Client, integration *oncallv1.Integration) error {

	if integration.Spec.Name == "" {
		integration.Spec.Name = integration.Name
	}
	integrationReq := &oncallv1.IntegrationSpec{
		Name: integration.Spec.Name,
		Type: integration.Spec.Type,
	}

	resp, err := client.Post(ctx, url, integrationReq)
	if err != nil {
		return err
	}

	integrationResp := &oncall_schema.Integration{}
	err = json.NewDecoder(bytes.NewReader(resp)).Decode(&integrationResp)
	if err != nil {
		return err
	}

	integration.Status.HttpEndpoint = integrationResp.Link
	integration.Spec.ID = integrationResp.ID
	integration.Spec.DefaultRoute = integrationResp.DefaultRoute
	return err
}

func (r *IntegrationReconciler) updateIntegration(ctx context.Context, url url.URL, client oncallClient.Client, integration *oncallv1.Integration, ctrlReq ctrl.Request) error {

	var routes []oncallv1.Route
	baseRouteUrl := url.JoinPath(routeUrl)
	for _, route := range integration.Spec.Routes {
		var (
			resp []byte
			err  error
		)
		routeRq := make(map[string]interface{})
		routeRq["routing_regex"] = route.RoutingRegex
		routeRq["integration_id"] = integration.Spec.ID
		routeRq["position"] = route.Position

		// if there escalationChain is set, replace it with its id
		if route.EscalationChain != "" {
			escalationChain, err := r.getEscalationChainByName(ctx, route.EscalationChain, ctrlReq)
			if err != nil {
				return err
			}
			routeRq["escalation_chain_id"] = &escalationChain.Spec.ID
		} else {
			routeRq["escalation_chain_id"] = nil
		}

		if route.ID != "" {
			baseURL := baseRouteUrl.JoinPath(route.ID)
			if routeRq["escalation_chain_id"] == nil {
				delete(routeRq, "escalation_chain_id")
			}
			resp, err = client.Put(ctx, *baseURL, routeRq)
			if err != nil {
				if err.(customerror.HTTPError).StatusCode == http.StatusNotFound {
					resp, err = client.Post(ctx, *baseRouteUrl, routeRq)
					if err != nil {
						return err
					}
					return nil
				}
				return err
			}
		} else {
			resp, err = client.Post(ctx, *baseRouteUrl, routeRq)
			ctrllog.FromContext(ctx).Info("POST route")
			if err != nil {
				return err
			}
		}

		routeResponse := &oncallv1.Route{}
		err = json.NewDecoder(bytes.NewBuffer(resp)).Decode(&routeResponse)
		ctrllog.FromContext(ctx).Info(string(resp))
		if err != nil {
			return err
		}
		routes = append(routes, *routeResponse)
	}
	integration.Spec.Routes = routes

	integrationReq := &oncallv1.IntegrationSpec{
		Templates: integration.Spec.Templates,
	}

	target := url.JoinPath(integrationUrl).JoinPath(integration.Spec.ID)

	resp, err := client.Put(ctx, *target, integrationReq)
	if err != nil {
		return err
	}

	integrationResp := &oncall_schema.Integration{}
	err = json.NewDecoder(bytes.NewReader(resp)).Decode(&integrationResp)
	if err != nil {
		return err
	}

	integration.Status.HttpEndpoint = integrationResp.Link
	integration.Spec.DefaultRoute = integrationResp.DefaultRoute

	return nil
}

func (r *IntegrationReconciler) getEscalationChainByName(ctx context.Context, escalationChainName string, req ctrl.Request) (*oncallv1.Escalation, error) {

	escalation := &oncallv1.Escalation{}
	err := r.Get(ctx, types.NamespacedName{Name: escalationChainName, Namespace: req.Namespace}, escalation)

	if err != nil && errors.IsNotFound(err) {
		return nil, err
	}

	return escalation, nil

}

func (r *IntegrationReconciler) finalizeIntegration(ctx context.Context, url url.URL, client oncallClient.Client, i *oncallv1.Integration) error {
	err := client.Delete(ctx, url)

	// TODO: delete routes
	return err
}

// SetupWithManager sets up the controller with the Manager.
func (r *IntegrationReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&oncallv1.Integration{}).
		Complete(r)
}
