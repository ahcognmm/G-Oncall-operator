/*
Copyright 2023.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

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
	ctrllog "sigs.k8s.io/controller-runtime/pkg/log"

	oncallv1 "github.com/ahcogn/grafana-oncall-operator/api/v1"
	oncallClient "github.com/ahcogn/grafana-oncall-operator/internal/client"
)

const (
	escalationURL         = "/api/v1/escalation_chains/"
	escalationPoliciesURL = "/api/v1/escalation_policies/"
)

// EscalationReconciler reconciles a Escalation object
type EscalationReconciler struct {
	client.Client
	Scheme *runtime.Scheme
	Config oncallClient.Config
}

//+kubebuilder:rbac:groups=oncall.ahcogn.com,resources=escalations,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=oncall.ahcogn.com,resources=escalations/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=oncall.ahcogn.com,resources=escalations/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Escalation object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.14.1/pkg/reconcile
func (r *EscalationReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := ctrllog.FromContext(ctx)
	log.Info("Start reconciling escalation")

	escalation := &oncallv1.Escalation{}

	err := r.Get(ctx, req.NamespacedName, escalation)
	if err != nil {
		return ctrl.Result{}, err
	}

	if err != nil {
		if errors.IsNotFound(err) {
			log.Info("Oncall resource not found. Ignoring since object must be deleted")
			return ctrl.Result{}, nil
		}
		log.Error(err, "Failed to get Oncall escalation")
		return ctrl.Result{}, err
	}

	err = CheckFinalizer(ctx, r, escalation, func() error {
		return r.finalize()
	})
	if err != nil {
		return ctrl.Result{}, nil
	}

	eId := escalation.Spec.ID
	client := &oncallClient.DefaultClient{
		Ctx:    ctx,
		Config: r.Config,
	}

	if eId == "" {
		err = r.createEscalation(ctx, r.Config.OncallUrl, client, escalation)
		if err != nil {
			return ctrl.Result{}, err
		}
	}
	err = r.updateEscalation(ctx, client, &r.Config.OncallUrl, escalation)
	if err != nil {
		return ctrl.Result{}, nil
	}

	err = r.Update(ctx, escalation)
	if err != nil {
		return ctrl.Result{}, nil
	}

	log.Info("Done reconciling escalation")
	return ctrl.Result{}, nil
}

func (r *EscalationReconciler) createEscalation(ctx context.Context, baseURL url.URL, oncallClient oncallClient.Client, escalation *oncallv1.Escalation) error {

	baseEsacalationURL := baseURL.JoinPath(escalationURL)
	escalationRq := make(map[string]interface{})
	escalationRq["name"] = escalation.Spec.Name

	resp, err := oncallClient.Post(ctx, *baseEsacalationURL, escalationRq)
	if err != nil {
		return err
	}

	escalationResp := &oncallv1.EscalationSpec{}
	err = json.NewDecoder(bytes.NewReader(resp)).Decode(&escalationResp)
	if err != nil {
		return err
	}

	escalation.Spec.ID = escalationResp.ID
	return nil
}
func (r *EscalationReconciler) updateEscalation(ctx context.Context, oncallClient oncallClient.Client, baseURL *url.URL, escalation *oncallv1.Escalation) error {
	var (
		policies      []oncallv1.EscalationPolicy
		policyRespRaw []byte
		err           error
	)
	pMap := &struct {
		oncallv1.EscalationPolicy
		EscalationChainId string `json:"escalation_chain_id"`
	}{}
	policyResp := &oncallv1.EscalationPolicy{}

	// TODO: create escalation policies
	baseEsacalationPolicesURl := baseURL.JoinPath(escalationPoliciesURL)
	for _, policy := range escalation.Spec.EscalationPolices {
		pMap.EscalationChainId = escalation.Spec.ID
		pMap.EscalationPolicy = policy
		pId := policy.ID
		if pId != "" {
			baseEsacalationPolicesURl = baseEsacalationPolicesURl.JoinPath(pId)
			policyRespRaw, err = oncallClient.Put(ctx, *baseEsacalationPolicesURl, pMap)
			if err != nil {
				return err
			}

		} else {
			ctrllog.FromContext(ctx).Info("POST policy")
			policyRespRaw, err = oncallClient.Post(ctx, *baseEsacalationPolicesURl, pMap)
			if err != nil {
				return err
			}
		}

		err := json.NewDecoder(bytes.NewReader(policyRespRaw)).Decode(&policyResp)
		if err != nil {
			return err
		}
		policies = append(policies, *policyResp)
	}
	escalation.Spec.EscalationPolices = policies
	return nil
}

func (r *EscalationReconciler) finalize() error {
	return nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *EscalationReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&oncallv1.Escalation{}).
		Complete(r)
}
